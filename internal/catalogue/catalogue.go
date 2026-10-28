// Package catalogue reads subjects.yaml, the catalogue of NATS subjects and KV
// keys, checks it against ADR 0001's subject scheme, ADR 0002's rules and the
// generated messages, and renders it as docs/subjects.md
// (those-who-run/paddock#25).
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
// it, or a KV bucket and key and the value kept there. subjects.yaml's header
// says what each key holds.
type Subject struct {
	Pattern     string   `yaml:"pattern"`
	CEType      string   `yaml:"ce_type"`
	Message     string   `yaml:"message"`
	ContentType string   `yaml:"content_type"`
	Stream      string   `yaml:"stream"`
	KV          string   `yaml:"kv"`
	Key         string   `yaml:"key"`
	Producers   []string `yaml:"producers"`
	Consumers   []string `yaml:"consumers"`
	MsgID       string   `yaml:"msg_id"`
	Description string   `yaml:"description"`
}

// IsKV is whether the entry is a KV key rather than a subject.
func (s Subject) IsKV() bool { return s.KV != "" }

// name is how an error names the entry.
func (s Subject) name() string {
	if s.IsKV() {
		return "kv " + s.KV + " " + s.Key
	}
	return s.Pattern
}

// Catalogue is subjects.yaml.
type Catalogue struct {
	Subjects []Subject `yaml:"subjects"`
	// Helpers are the messages in an events package that are not events, only
	// parts of one, such as SessionSummary: they need no entry.
	Helpers []string `yaml:"helpers"`
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

// EventsPackages is the prefix of every events package, whatever its version:
// each message there is an event and needs an entry, unless it is a helper.
const EventsPackages = "twr.events."

// NoSession is the token an event subject has in place of <sid> when it is
// published while no session is open (paddock#70).
const NoSession = "_"

// Placeholder is a token for a variable part of a pattern, a KV key or a msg_id
// scheme, and where it may appear. The check and the rendered doc both use
// Placeholders, so a new one is explained where it is allowed.
type Placeholder struct {
	Token, Meaning   string
	Pattern, Key, ID bool
}

// Placeholders are the known tokens.
var Placeholders = []Placeholder{
	{"<src>", "where the telemetry comes from: `live`, `parity`, or `replay-` followed by the replay's name", true, true, true},
	{"<sid>", "the session id, a ULID ecu mints", true, true, true},
	{"<run>", "the ecu run's id, a ULID ecu mints when it starts (not a replay's name)", false, false, true},
	{"<seq>", "a frame's sequence number within the ecu run: a frame's own, or the `at` of the frame an event happened on", false, false, true},
	{"<n>", "a lap's number: lap.completed's own, or for lap.best the lap whose end published it (LapBestChanged's ended_lap)", false, false, true},
	{"<pause>", "the pause's number in the session, from 1, which ecu carries across a restart: a pause open across one keeps its number", false, false, true},
}

// word is a token of a subject's <area> or <what>, or of a ce-type: lower-case
// letters and digits, with single underscores between them.
const word = `[a-z0-9]+(?:_[a-z0-9]+)*`

var (
	// ADR 0002, point 4: reverse DNS, at least an area and a what, and a version.
	// A token is lower-case words joined by single underscores, such as
	// final_lap (paddock#69).
	ceType = regexp.MustCompile(`^run\.thosewho(\.` + word + `){2,}\.v[1-9][0-9]*$`)
	// ADR 0001, point 2, with the stream point 1 captures each kind of subject
	// in. Callouts are listed one entry per level, p0, p1 and p2, each a subject
	// a consumer subscribes to, sharing one message and ce-type.
	schemes = []struct {
		kind    string
		pattern *regexp.Regexp
		stream  string
	}{
		{"frame", regexp.MustCompile(`^twr\.v1\.<src>\.frame\.[a-z0-9]+$`), "TELEMETRY"},
		{"event", regexp.MustCompile(`^twr\.v1\.<src>\.evt\.(<sid>|_)\.` + word + `\.` + word + `$`), "EVENTS"},
		{"callout", regexp.MustCompile(`^twr\.v1\.<src>\.callout\.p[0-2]$`), "CALLOUTS"},
	}
	placeholder = regexp.MustCompile(`<[^<>]*>`)
	// An event's msg_id (paddock#25, 2026-10-26, with the source in front from
	// paddock#279): its source, its session, its ce-type's <area>.<what>, then
	// one or more parts, each a word or a placeholder.
	eventID = regexp.MustCompile(`^<src>:(<sid>|<run>):(` + word + `\.` + word + `):([a-z0-9]+|<[a-z]+>)(:([a-z0-9]+|<[a-z]+>))*$`)
	// A ce-type's <area>.<what>: what comes between run.thosewho. and the version.
	typeName = regexp.MustCompile(`^run\.thosewho\.(.+)\.v[1-9][0-9]*$`)
	// A KV bucket's name, and a key: dot-separated tokens or placeholders.
	bucket = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)
	key    = regexp.MustCompile(`^([a-z0-9_-]+|<[a-z]+>)(\.([a-z0-9_-]+|<[a-z]+>))*$`)
)

// Messages is what the generated code registers, such as
// protoregistry.GlobalFiles.
type Messages interface {
	FindDescriptorByName(protoreflect.FullName) (protoreflect.Descriptor, error)
	RangeFiles(func(protoreflect.FileDescriptor) bool)
}

// problems collects the check's errors, each naming its entry.
type problems []error

func (p *problems) add(s Subject, i int, format string, args ...any) {
	*p = append(*p, fmt.Errorf("subjects[%d] (%s): %s", i, s.name(), fmt.Sprintf(format, args...)))
}

// Check returns every problem with the catalogue, or nil. messages is what the
// generated code registers.
func (c *Catalogue) Check(messages Messages) error {
	var p problems
	seen := map[string]int{}
	byType := map[string]string{}
	typeOf := map[string]string{}
	listed := map[string]bool{}
	for i, s := range c.Subjects {
		fail := func(format string, args ...any) { p.add(s, i, format, args...) }
		if first, ok := seen[s.name()]; ok {
			fail("subjects[%d] is the same subject or key", first)
		}
		seen[s.name()] = i
		if s.IsKV() {
			checkKV(s, fail)
		} else {
			checkSubject(s, fail)
			listed[s.Message] = true
			// ADR 0002, point 4: a ce-type is one to one with a message. A message
			// may travel on several subjects, with its one ce-type on each.
			if m, ok := byType[s.CEType]; ok && s.CEType != "" && m != s.Message {
				fail("ce_type %s is %s's too: a ce_type names one message", s.CEType, m)
			}
			if t, ok := typeOf[s.Message]; ok && t != s.CEType {
				fail("%s has the ce_type %s elsewhere: a message has one ce_type", s.Message, t)
			}
			byType[s.CEType], typeOf[s.Message] = s.Message, s.CEType
		}
		checkMessage(s, messages, fail)
		if len(s.Producers) == 0 {
			fail("no producers")
		}
		if len(s.Consumers) == 0 {
			fail("no consumers")
		}
		if s.Description == "" {
			fail("no description")
		}
	}
	for _, h := range c.Helpers {
		if !strings.HasPrefix(h, EventsPackages) {
			p = append(p, fmt.Errorf("helper %s is not in an events package", h))
		} else if d, err := messages.FindDescriptorByName(protoreflect.FullName(h)); err != nil {
			p = append(p, fmt.Errorf("helper %s is not in the generated code", h))
		} else if _, ok := d.(protoreflect.MessageDescriptor); !ok {
			p = append(p, fmt.Errorf("helper %s is not a message", h))
		}
		if listed[h] {
			p = append(p, fmt.Errorf("helper %s has an entry: a helper is not an event", h))
		}
	}
	for _, name := range c.EventMessages(messages) {
		if !listed[string(name)] {
			p = append(p, fmt.Errorf("event message %s has no entry in the catalogue, and is not a helper", name))
		}
	}
	return errors.Join(p...)
}

func checkMessage(s Subject, messages Messages, fail func(string, ...any)) {
	if s.Message == "" {
		fail("no message")
	} else if d, err := messages.FindDescriptorByName(protoreflect.FullName(s.Message)); err != nil {
		fail("message %s is not in the generated code", s.Message)
	} else if _, ok := d.(protoreflect.MessageDescriptor); !ok {
		fail("%s is not a message", s.Message)
	}
	if s.ContentType != ContentType {
		fail("content_type %q is not %s", s.ContentType, ContentType)
	}
}

func checkSubject(s Subject, fail func(string, ...any)) {
	if s.Key != "" {
		fail("a key without a kv bucket")
	}
	switch {
	case s.Pattern == "":
		fail("no pattern")
	case !strings.HasPrefix(s.Pattern, "twr.v1."):
		fail("the pattern does not start with twr.v1.")
	default:
		checkTokens(s.Pattern, "the pattern", func(p Placeholder) bool { return p.Pattern }, fail)
		matched := false
		for _, scheme := range schemes {
			if scheme.pattern.MatchString(s.Pattern) {
				matched = true
				if s.Stream != scheme.stream {
					fail("a %s subject is captured by the %s stream, not %q", scheme.kind, scheme.stream, s.Stream)
				}
				if scheme.kind == "event" {
					checkEventID(s, fail)
				}
			}
		}
		if !matched {
			fail("the pattern does not follow ADR 0001's subject scheme (frame, evt.<sid> or evt._, or callout.p0 to p2)")
		}
	}
	if !ceType.MatchString(s.CEType) {
		fail("ce_type %q is not run.thosewho.<area>.<what>.v<N>", s.CEType)
	}
	if s.MsgID == "" {
		fail("no msg_id: every message carries a deterministic Nats-Msg-Id")
	}
	checkTokens(s.MsgID, "msg_id", func(p Placeholder) bool { return p.ID }, fail)
	for _, token := range placeholder.FindAllString(s.MsgID, -1) {
		if (token == "<sid>" || token == "<src>") && !strings.Contains(s.Pattern, token) {
			fail("msg_id uses %s, which the pattern does not carry", token)
		}
	}
}

// checkEventID checks an event's msg_id against the one style
// (paddock#25, 2026-10-26; paddock#279): `<src>:<sid>:<area>.<what>:<n>`,
// where <area>.<what> is the event's ce-type's, such as
// `<src>:<sid>:lap.completed:<n>`. The source comes first because EVENTS
// deduplicates by Nats-Msg-Id across the whole stream, and two sources can
// carry the same session id (two replays of one recording, or a replay and the
// parity harness). An event published with no session open (evt._) has
// `<run>`, the ecu run, in place of `<sid>`.
func checkEventID(s Subject, fail func(string, ...any)) {
	m := eventID.FindStringSubmatch(s.MsgID)
	if m == nil {
		fail("msg_id %q is not <src>:<sid>:<area>.<what>:<part>[:<part>…]", s.MsgID)
		return
	}
	want := "<sid>"
	if strings.Contains(s.Pattern, ".evt."+NoSession+".") {
		want = "<run>"
	}
	if m[1] != want {
		fail("msg_id %q has %s after the source, want %s for this subject", s.MsgID, m[1], want)
	}
	if t := typeName.FindStringSubmatch(s.CEType); t != nil && t[1] != m[2] {
		fail("msg_id %q names %s, but the ce_type is %s's", s.MsgID, m[2], t[1])
	}
}

// checkKV checks a KV entry: a bucket and a key, and no subject, stream, ce_type
// or msg_id, since a value written with Put carries no headers.
func checkKV(s Subject, fail func(string, ...any)) {
	if !bucket.MatchString(s.KV) {
		fail("kv bucket %q is not lower-case letters, digits, - and _", s.KV)
	}
	if !key.MatchString(s.Key) {
		fail("key %q is not dot-separated tokens and placeholders", s.Key)
	}
	checkTokens(s.Key, "the key", func(p Placeholder) bool { return p.Key }, fail)
	for _, f := range [...]struct{ name, value string }{
		{"pattern", s.Pattern}, {"stream", s.Stream}, {"ce_type", s.CEType}, {"msg_id", s.MsgID},
	} {
		if f.value != "" {
			fail("a KV entry has no %s", f.name)
		}
	}
}

// checkTokens fails on a placeholder that isn't known, or isn't allowed where it
// is.
func checkTokens(text, where string, allowed func(Placeholder) bool, fail func(string, ...any)) {
	var known []string
	for _, p := range Placeholders {
		if allowed(p) {
			known = append(known, p.Token)
		}
	}
	for _, token := range placeholder.FindAllString(text, -1) {
		if !slices.Contains(known, token) {
			fail("unknown placeholder %s in %s: known are %s", token, where, strings.Join(known, ", "))
		}
	}
}

// EventMessages are the messages in every events package, whatever its version,
// that are not helpers: each needs an entry.
func (c *Catalogue) EventMessages(messages Messages) []protoreflect.FullName {
	var events []protoreflect.FullName
	messages.RangeFiles(func(f protoreflect.FileDescriptor) bool {
		if !strings.HasPrefix(string(f.Package()), EventsPackages) {
			return true
		}
		for i := range f.Messages().Len() {
			if name := f.Messages().Get(i).FullName(); !slices.Contains(c.Helpers, string(name)) {
				events = append(events, name)
			}
		}
		return true
	})
	slices.Sort(events)
	return events
}

// Render writes the catalogue as Markdown, the docs/subjects.md that
// `make generate` writes.
func (c *Catalogue) Render() []byte {
	var b bytes.Buffer
	b.WriteString("# Subjects\n\n")
	b.WriteString("Rendered from [`subjects.yaml`](../subjects.yaml) by `make generate`; edit that file,\n")
	b.WriteString("not this one.\n\n")
	b.WriteString("## Subjects\n\n")
	b.WriteString("Every message is sent as protobuf's JSON form (`ce-datacontenttype: " + ContentType + "`)\n")
	b.WriteString("with the full CloudEvents header set, and its `Nats-Msg-Id` is the same as its\n")
	b.WriteString("`ce-id`. An event subject has `" + NoSession + "` in place of `<sid>` when it is published while\n")
	b.WriteString("no session is open. An event's id is `<src>:<sid>:<area>.<what>:<part>`, with its\n")
	b.WriteString("`ce-type`'s `<area>.<what>`, such as `<src>:<sid>:lap.completed:<n>`; one that happens\n")
	b.WriteString("once per session ends in `:1`, and with no session open `<run>` takes `<sid>`'s place.\n")
	b.WriteString("The source comes first because `EVENTS` deduplicates by `Nats-Msg-Id` across the\n")
	b.WriteString("whole stream, and two sources can carry the same session id. A frame's id is\n")
	b.WriteString("`<run>:frame:<seq>`, without the source.\n\n")
	b.WriteString("| Subject | Message | `ce-type` | Stream | Producers | Consumers | `ce-id` and `Nats-Msg-Id` |\n")
	b.WriteString("| --- | --- | --- | --- | --- | --- | --- |\n")
	var kv []Subject
	for _, s := range c.Subjects {
		if s.IsKV() {
			kv = append(kv, s)
			continue
		}
		fmt.Fprintf(&b, "| `%s` | `%s` | `%s` | `%s` | %s | %s | `%s` |\n",
			s.Pattern, s.Message, s.CEType, s.Stream, strings.Join(s.Producers, ", "), strings.Join(s.Consumers, ", "), s.MsgID)
	}
	if len(kv) > 0 {
		b.WriteString("\n## KV keys\n\n")
		b.WriteString("Each value is the message in protobuf's JSON form, written with `Put`, so it has no\n")
		b.WriteString("headers.\n\n")
		b.WriteString("| Bucket | Key | Message | Writers | Readers |\n")
		b.WriteString("| --- | --- | --- | --- | --- |\n")
		for _, s := range kv {
			fmt.Fprintf(&b, "| `%s` | `%s` | `%s` | %s | %s |\n",
				s.KV, s.Key, s.Message, strings.Join(s.Producers, ", "), strings.Join(s.Consumers, ", "))
		}
	}
	b.WriteString("\n## What each one is\n\n")
	for _, s := range c.Subjects {
		name := "`" + s.Pattern + "`"
		if s.IsKV() {
			name = "`" + s.KV + "` `" + s.Key + "`"
		}
		fmt.Fprintf(&b, "- %s: %s\n", name, s.Description)
	}
	if len(c.Helpers) > 0 {
		b.WriteString("\nParts of an event, not events themselves, so they have no subject: ")
		b.WriteString("`" + strings.Join(c.Helpers, "`, `") + "`.\n")
	}
	b.WriteString("\n## Placeholders\n\n")
	b.WriteString("| Placeholder | Meaning | In |\n| --- | --- | --- |\n")
	for _, p := range Placeholders {
		var in []string
		for _, w := range []struct {
			ok   bool
			name string
		}{{p.Pattern, "subjects"}, {p.Key, "KV keys"}, {p.ID, "`msg_id`"}} {
			if w.ok {
				in = append(in, w.name)
			}
		}
		fmt.Fprintf(&b, "| `%s` | %s | %s |\n", p.Token, p.Meaning, strings.Join(in, ", "))
	}
	return b.Bytes()
}

// Agrees returns every difference between the catalogue's ce-types and types,
// a ce-type by message full name such as the ce package's, or nil: a message
// with another ce-type, or one only one side has. KV entries have no ce-type.
func (c *Catalogue) Agrees(types map[protoreflect.FullName]string) error {
	var errs []error
	listed := map[protoreflect.FullName]bool{}
	for _, s := range c.Subjects {
		name := protoreflect.FullName(s.Message)
		if s.IsKV() || listed[name] {
			continue
		}
		listed[name] = true
		switch t, ok := types[name]; {
		case !ok:
			errs = append(errs, fmt.Errorf("%s (%s) has no ce-type in the helper", name, s.CEType))
		case t != s.CEType:
			errs = append(errs, fmt.Errorf("%s: the catalogue's ce-type is %s, the helper's %s", name, s.CEType, t))
		}
	}
	for name, t := range types {
		if !listed[name] {
			errs = append(errs, fmt.Errorf("%s (%s) is in the helper but not the catalogue", name, t))
		}
	}
	return errors.Join(errs...)
}
