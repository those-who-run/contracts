// Package catalogue reads subjects.yaml, the catalogue of NATS subjects, checks
// it against ADR 0001's subject scheme, ADR 0002's rules and the generated
// messages, and renders it as docs/subjects.md (those-who-run/paddock#25).
package catalogue

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Subject is one entry of subjects.yaml: a subject pattern and what travels on
// it. subjects.yaml's header says what each key holds.
type Subject struct {
	Pattern     string   `yaml:"pattern"`
	CEType      string   `yaml:"ce_type"`
	Message     string   `yaml:"message"`
	ContentType string   `yaml:"content_type"`
	Stream      string   `yaml:"stream"`
	KV          string   `yaml:"kv"`
	Producer    string   `yaml:"producer"`
	Consumers   []string `yaml:"consumers"`
	MsgID       string   `yaml:"msg_id"`
	Description string   `yaml:"description"`
}

// Catalogue is subjects.yaml.
type Catalogue struct {
	Subjects []Subject `yaml:"subjects"`
}

// Parse reads a catalogue. A key it doesn't know is an error, so a misspelt key
// is never silently ignored.
func Parse(r io.Reader) (*Catalogue, error) {
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)
	var c Catalogue
	if err := dec.Decode(&c); err != nil {
		return nil, fmt.Errorf("reading the catalogue: %w", err)
	}
	return &c, nil
}

// ContentType is the one content type in v0.1: protobuf's JSON form (ADR 0002,
// point 2).
const ContentType = "application/json"

// EventsPackage holds the event messages, each of which needs an entry.
const EventsPackage = "twr.events.v1"

var (
	// ADR 0002, point 4: reverse DNS, at least an area and a what, and a version.
	ceType = regexp.MustCompile(`^run\.thosewho(\.[a-z0-9]+){2,}\.v[1-9][0-9]*$`)
	// ADR 0001, point 2, with the stream point 1 captures each kind of subject
	// in. A token is a placeholder or lower-case letters and digits.
	schemes = []struct {
		kind    string
		pattern *regexp.Regexp
		stream  string
	}{
		{"frame", regexp.MustCompile(`^twr\.v1\.<src>\.frame\.[a-z0-9]+$`), "TELEMETRY"},
		{"event", regexp.MustCompile(`^twr\.v1\.<src>\.evt\.<sid>\.[a-z0-9]+\.[a-z0-9]+$`), "EVENTS"},
		{"callout", regexp.MustCompile(`^twr\.v1\.<src>\.callout\.p[0-2]$`), "CALLOUTS"},
	}
	placeholder = regexp.MustCompile(`<[^<>]*>`)
	// The placeholders a pattern may have, and those a msg_id scheme may.
	patternTokens = []string{"<src>", "<sid>"}
	msgIDTokens   = []string{"<run>", "<seq>", "<sid>", "<n>"}
)

// Messages finds a message in the generated code by its full name, such as
// protoregistry.GlobalFiles.FindDescriptorByName.
type Messages interface {
	FindDescriptorByName(protoreflect.FullName) (protoreflect.Descriptor, error)
	RangeFilesByPackage(protoreflect.FullName, func(protoreflect.FileDescriptor) bool)
}

// Check returns every problem with the catalogue, or nil. messages is what the
// generated code registers.
func (c *Catalogue) Check(messages Messages) error {
	var errs []error
	fail := func(i int, s Subject, format string, args ...any) {
		errs = append(errs, fmt.Errorf("subjects[%d] (%s): %s", i, s.Pattern, fmt.Sprintf(format, args...)))
	}
	ceTypes := map[string]int{}
	listed := map[string]bool{}
	for i, s := range c.Subjects {
		checkPattern(s, func(format string, args ...any) { fail(i, s, format, args...) })
		checkHeaders(s, func(format string, args ...any) { fail(i, s, format, args...) })
		if first, seen := ceTypes[s.CEType]; seen && s.CEType != "" {
			fail(i, s, "ce_type %q is subjects[%d]'s too: a ce_type names one message", s.CEType, first)
		} else {
			ceTypes[s.CEType] = i
		}
		if s.Message == "" {
			fail(i, s, "no message")
		} else if d, err := messages.FindDescriptorByName(protoreflect.FullName(s.Message)); err != nil {
			fail(i, s, "message %s is not in the generated code", s.Message)
		} else if _, ok := d.(protoreflect.MessageDescriptor); !ok {
			fail(i, s, "%s is not a message", s.Message)
		}
		listed[s.Message] = true
		if s.Producer == "" {
			fail(i, s, "no producer")
		}
		if len(s.Consumers) == 0 {
			fail(i, s, "no consumers")
		}
		if s.Description == "" {
			fail(i, s, "no description")
		}
	}
	for _, name := range EventMessages(messages) {
		if !listed[string(name)] {
			errs = append(errs, fmt.Errorf("event message %s has no entry in the catalogue", name))
		}
	}
	return errors.Join(errs...)
}

func checkPattern(s Subject, fail func(string, ...any)) {
	if !strings.HasPrefix(s.Pattern, "twr.v1.") {
		fail("the pattern does not start with twr.v1.")
		return
	}
	for _, token := range placeholder.FindAllString(s.Pattern, -1) {
		if !slices.Contains(patternTokens, token) {
			fail("unknown placeholder %s in the pattern: known are %s", token, strings.Join(patternTokens, ", "))
		}
	}
	matched := false
	for _, scheme := range schemes {
		if scheme.pattern.MatchString(s.Pattern) {
			matched = true
			if s.Stream != scheme.stream {
				fail("a %s subject is captured by the %s stream, not %q", scheme.kind, scheme.stream, s.Stream)
			}
		}
	}
	if !matched {
		fail("the pattern does not follow ADR 0001's subject scheme (frame, evt or callout)")
	}
}

func checkHeaders(s Subject, fail func(string, ...any)) {
	if !ceType.MatchString(s.CEType) {
		fail("ce_type %q is not run.thosewho.<area>.<what>.v<N>", s.CEType)
	}
	if s.ContentType != ContentType {
		fail("content_type %q is not %s", s.ContentType, ContentType)
	}
	if (s.Stream == "") == (s.KV == "") {
		fail("it needs a stream or a kv, and not both")
	}
	if s.MsgID == "" {
		fail("no msg_id: every message carries a deterministic Nats-Msg-Id")
	}
	for _, token := range placeholder.FindAllString(s.MsgID, -1) {
		if !slices.Contains(msgIDTokens, token) {
			fail("unknown placeholder %s in msg_id: known are %s", token, strings.Join(msgIDTokens, ", "))
		}
		if token == "<sid>" && !strings.Contains(s.Pattern, "<sid>") {
			fail("msg_id uses <sid>, which the pattern does not carry")
		}
	}
}

// EventMessages are the messages in the events package that travel on their
// own: every message there that no other message there holds as a field, such
// as SessionSummary, which is part of SessionEnded.
func EventMessages(messages Messages) []protoreflect.FullName {
	var all []protoreflect.MessageDescriptor
	held := map[protoreflect.FullName]bool{}
	messages.RangeFilesByPackage(EventsPackage, func(f protoreflect.FileDescriptor) bool {
		for i := range f.Messages().Len() {
			m := f.Messages().Get(i)
			all = append(all, m)
			for j := range m.Fields().Len() {
				if field := m.Fields().Get(j).Message(); field != nil {
					held[field.FullName()] = true
				}
			}
		}
		return true
	})
	var events []protoreflect.FullName
	for _, m := range all {
		if !held[m.FullName()] {
			events = append(events, m.FullName())
		}
	}
	slices.Sort(events)
	return events
}

// Render writes the catalogue as Markdown, the docs/subjects.md that
// `make generate` writes.
func (c *Catalogue) Render() []byte {
	var b bytes.Buffer
	b.WriteString("# Subjects\n\n")
	b.WriteString("Rendered from [`subjects.yaml`](../subjects.yaml) by `make generate`; edit that file,\n")
	b.WriteString("not this one. Every message is sent as protobuf's JSON form\n")
	b.WriteString("(`ce-datacontenttype: " + ContentType + "`) with the full CloudEvents header set,\n")
	b.WriteString("and its `Nats-Msg-Id` is the same as its `ce-id`.\n\n")
	b.WriteString("| Subject | Message | `ce-type` | Captured by | Producer | Consumers | `ce-id` and `Nats-Msg-Id` |\n")
	b.WriteString("| --- | --- | --- | --- | --- | --- | --- |\n")
	for _, s := range c.Subjects {
		capture := "stream `" + s.Stream + "`"
		if s.KV != "" {
			capture = "KV `" + s.KV + "`"
		}
		fmt.Fprintf(&b, "| `%s` | `%s` | `%s` | %s | %s | %s | `%s` |\n",
			s.Pattern, s.Message, s.CEType, capture, s.Producer, strings.Join(s.Consumers, ", "), s.MsgID)
	}
	b.WriteString("\n## What each one is\n\n")
	for _, s := range c.Subjects {
		fmt.Fprintf(&b, "- `%s`: %s\n", s.Pattern, s.Description)
	}
	b.WriteString("\n## Placeholders\n\n")
	b.WriteString("- `<src>`: where the telemetry comes from: `live`, `replay-<run>` or `parity`.\n")
	b.WriteString("- `<sid>`: the session id, a ULID ecu mints.\n")
	b.WriteString("- `<run>`: the ecu run's id, a ULID ecu mints when it starts.\n")
	b.WriteString("- `<seq>`: the frame's sequence number within the run.\n")
	b.WriteString("- `<n>`: the lap's number.\n")
	return b.Bytes()
}
