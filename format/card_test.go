package format_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/those-who-run/contracts/format"
	eventsv1 "github.com/those-who-run/contracts/gen/go/twr/events/v1"
)

type cardCase struct {
	Fn   string            `json:"fn"`
	Args []json.RawMessage `json:"args"`
	Out  string            `json:"out"`
	V1   string            `json:"v1"`
}

func readCards(t *testing.T) []cardCase {
	t.Helper()
	var file struct {
		Cards struct {
			About string     `json:"about"`
			Cases []cardCase `json:"cases"`
		} `json:"cards"`
	}
	if err := json.Unmarshal(format.Cases(), &file); err != nil {
		t.Fatalf("format-cases.json: %v", err)
	}
	if file.Cards.About == "" || len(file.Cards.Cases) == 0 {
		t.Fatal("format-cases.json has no cards")
	}
	return file.Cards.Cases
}

func summaryArg(t *testing.T, raw json.RawMessage) *eventsv1.SessionSummary {
	t.Helper()
	s := &eventsv1.SessionSummary{}
	if err := protojson.Unmarshal(raw, s); err != nil {
		t.Fatalf("a summary: %v", err)
	}
	return s
}

func enumArg[E ~int32](t *testing.T, raw json.RawMessage, values map[string]int32) E {
	t.Helper()
	var name string
	if err := json.Unmarshal(raw, &name); err != nil {
		t.Fatal(err)
	}
	v, ok := values[name]
	if !ok {
		t.Fatalf("no value %q", name)
	}
	return E(v)
}

func intArg(t *testing.T, raw json.RawMessage) int {
	t.Helper()
	var n *int
	if err := json.Unmarshal(raw, &n); err != nil {
		t.Fatal(err)
	}
	if n == nil {
		return 0
	}
	return *n
}

// card is what the Go builder named fn writes for a case's args.
func card(t *testing.T, c cardCase) string {
	t.Helper()
	a := c.Args
	switch c.Fn {
	case "title":
		return format.Title(summaryArg(t, a[0]))
	case "result":
		var closed bool
		if err := json.Unmarshal(a[2], &closed); err != nil {
			t.Fatal(err)
		}
		reason := enumArg[eventsv1.EndReason](t, a[1], eventsv1.EndReason_value)
		return format.Result(summaryArg(t, a[0]), reason, closed, intArg(t, a[3]))
	case "kindLabel":
		kind := enumArg[eventsv1.SessionType](t, a[0], eventsv1.SessionType_value)
		return format.KindLabel(kind, intArg(t, a[1]))
	}
	t.Fatalf("no Go builder for %q: add one, and its case here", c.Fn)
	return ""
}

// The card's words: every "cards" case, most of them v1 1.0.0's cards for
// podium's recordings.
func TestEveryCardCasePasses(t *testing.T) {
	for i, c := range readCards(t) {
		t.Run(fmt.Sprintf("%d-%s", i, c.Fn), func(t *testing.T) {
			if got := card(t, c); got != c.Out {
				t.Errorf("%s(%s) = %q, want %q (%s)", c.Fn, a2s(c.Args), got, c.Out, c.V1)
			}
		})
	}
}

func a2s(args []json.RawMessage) string {
	b, _ := json.Marshal(args)
	return string(b)
}

// Every kind ecu can give has its words, and each builder its cases.
func TestTheCardCasesCoverEveryKind(t *testing.T) {
	labelled := map[string]bool{}
	fns := map[string]int{}
	for _, c := range readCards(t) {
		fns[c.Fn]++
		if c.Fn == "kindLabel" {
			var kind string
			_ = json.Unmarshal(c.Args[0], &kind)
			labelled[kind] = true
		}
	}
	for name := range eventsv1.SessionType_value {
		if !labelled[name] {
			t.Errorf("no kindLabel case for %s", name)
		}
	}
	for _, fn := range []string{"title", "result", "kindLabel"} {
		if fns[fn] == 0 {
			t.Errorf("no %s case", fn)
		}
	}
}

// What the cases don't reach: no summary at all, and a lap race with no total
// yet.
func TestCardEdges(t *testing.T) {
	for name, c := range map[string]struct{ got, want string }{
		"Title(nil)":            {format.Title(nil), "Unknown track · session"},
		"Result(nil, closed)":   {format.Result(nil, eventsv1.EndReason_END_REASON_QUIT, true, 0), "no full lap"},
		"Result(nil, open)":     {format.Result(nil, eventsv1.EndReason_END_REASON_UNSPECIFIED, false, 0), "Running, no full lap yet"},
		"KindLabel(lap race 0)": {format.KindLabel(eventsv1.SessionType_SESSION_TYPE_LAP_RACE, 0), "0-lap race"},
		"TrackLabel(nil)":       {format.TrackLabel(nil), "Unknown track"},
	} {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", name, c.got, c.want)
		}
	}
	if format.PositionShown(nil) {
		t.Error("PositionShown(nil) is true")
	}
}
