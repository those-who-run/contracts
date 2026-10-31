package contracts_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	pitwallv1 "github.com/those-who-run/contracts/gen/go/twr/pitwall/v1"
)

// twr.pitwall.v1 against v1's pit wall API (paddock#121, AC2):
// testdata/v1/pitwall_openapi.json is v1's web/openapi.json at the v1.0.0
// tag, generated from api_models.py. Every field of every model there has a
// field of the same meaning here, with presence wherever v1's could be null,
// and every route and parameter has its method and request field; what
// differs on purpose is listed below, with why. Every field here is v1's or
// listed as new, so nothing is added by accident either.

// The v1 models and the message each maps to.
var pitwallModels = map[string]protoreflect.FullName{
	"Session":       "twr.pitwall.v1.Session",
	"Track":         "twr.pitwall.v1.Track",
	"LapRow":        "twr.pitwall.v1.LapRow",
	"LapList":       "twr.pitwall.v1.ListLapsResponse",
	"Trace":         "twr.pitwall.v1.Trace",
	"FilterOption":  "twr.pitwall.v1.FilterOption",
	"KindOption":    "twr.pitwall.v1.KindOption",
	"FilterOptions": "twr.pitwall.v1.FilterOptions",
	"SessionList":   "twr.pitwall.v1.ListSessionsResponse",
	"SameTrack":     "twr.pitwall.v1.ListSameTrackResponse",
	"EngineStatus":  "twr.pitwall.v1.EngineStatus",
	"TelemetryLink": "twr.pitwall.v1.TelemetryLink",
	"LiveSession":   "twr.pitwall.v1.LiveSession",
	"TrackSeen":     "twr.pitwall.v1.LiveTrack",
}

// The v1 models with no message, each with why.
var pitwallModelsLeftOut = map[string]string{
	"Health":   "the pit wall's /healthz, plain HTTP beside the API, not a method (paddock#124, R5)",
	"ApiError": "Connect's own error body, with its code: not_found, invalid_argument, unavailable (paddock#124, R3)",
}

// The v1 fields whose name differs, each with its name here and why.
var pitwallRenamed = map[string]struct{ to, why string }{
	"Track.id":                        {"track_id", "as TrackVerdict's and LiveTrack's in twr.events.v1 and twr.state.v1"},
	"TrackSeen.id":                    {"track_id", "as Track's"},
	"Session.best_lap_no":             {"best_lap", "as SessionSummary's best_lap; best_lap_ms is its time"},
	"Session.fuel_used":               {"fuel_burned", "v1's value is the fuel burnt, refuels left out (SessionSummary's fuel_burned); SessionSummary's fuel_used is net of refuels, and the same name would mislead"},
	"TelemetryLink.last_packet_age_s": {"last_packet_age_secs", "as EcuStatus's, and idle_secs beside it"},
	"LiveSession.id":                  {"session_id", "as Session's"},
	"LiveSession.laps_total":          {"total_laps", "as Session's"},
	"LiveSession.cars":                {"total_cars", "as Session's"},
}

// The v1 fields with no field here, each with why.
var pitwallLeftOut = map[string]string{
	"LapRow.file": "the lap's Parquet file: logbook's own layout, which is not a contract (ADR 0003, point 2)",
}

// The v1 fields whose type differs, each with why.
var pitwallRetyped = map[string]string{
	"Session.started_at":      "a google.protobuf.Timestamp, contracts' time everywhere (an RFC 3339 string in the JSON form), not Unix seconds",
	"Session.ended_at":        "a google.protobuf.Timestamp, as started_at",
	"EngineStatus.started_at": "a google.protobuf.Timestamp, as Session's started_at",
	"EngineStatus.activity":   "the Activity enum, ACTIVITY_ and v1's value in upper case, and ACTIVITY_OFFLINE for an engineer v1's proxy answered 502 for",
	"LiveSession.kind":        "the kind's key as a string, as Session's kind: a later kind adds a key, not an enum value",
}

// The v1 fields that are never null, but are absent here when they mean
// nothing, each with why.
var pitwallNowOptional = map[string]string{
	"Session.total_laps":           "absent for a timed session, or until the lap total is known, where v1 said 0, as SessionSummary's total_laps; an open lap race has it, as in v1",
	"EngineStatus.voice_connected": "radio's, in its own live key: absent when radio says nothing of it, so the chip says \"Voice: unknown\" rather than off (paddock#125, R6)",
}

// The fields here with no v1 field, each with why.
var pitwallNew = map[string]string{
	"Session.card_posted_at":          "when the results card was posted, from card.posted (paddock#122's triage)",
	"Session.from_v1_rows":            "whether an imported session came from v1's rows or was re-derived (ADR 0009)",
	"GetSessionResponse.session":      "a method's response is its own message (buf's STANDARD lint); v1's body was the Session",
	"GetTraceResponse.trace":          "a method's response is its own message; v1's body was the Trace",
	"GetEngineStatusResponse.status":  "a method's response is its own message; v1's body was the EngineStatus",
	"ListSessionsRequest.track":       "v1's track query parameter",
	"ListSessionsRequest.car":         "v1's car query parameter",
	"ListSessionsRequest.kind":        "v1's kind query parameter",
	"ListSessionsRequest.from_day":    "v1's from query parameter, renamed: from is a keyword in Python",
	"ListSessionsRequest.to_day":      "v1's to query parameter, renamed as from_day",
	"ListSessionsRequest.time_zone":   "v1's tz query parameter",
	"GetSessionRequest.session_id":    "v1's id path parameter",
	"ListLapsRequest.session_id":      "v1's id path parameter",
	"GetTraceRequest.session_id":      "v1's id path parameter",
	"GetTraceRequest.lap":             "v1's lap path parameter",
	"GetTraceRequest.points":          "v1's points query parameter",
	"ListSameTrackRequest.session_id": "v1's id path parameter",
	"Activity.ACTIVITY_OFFLINE":       "the engineer cannot be reached: v1's proxy answered 502, and the web app called it offline",
}

// v1's routes, each with its method and its parameters' request fields.
var pitwallRoutes = map[string]struct {
	method string
	params map[string]string
}{
	"/api/v1/sessions": {"ListSessions", map[string]string{
		"track": "track", "car": "car", "kind": "kind", "from": "from_day", "to": "to_day", "tz": "time_zone",
	}},
	"/api/v1/sessions/{id}":                  {"GetSession", map[string]string{"id": "session_id"}},
	"/api/v1/sessions/{id}/laps":             {"ListLaps", map[string]string{"id": "session_id"}},
	"/api/v1/sessions/{id}/laps/{lap}/trace": {"GetTrace", map[string]string{"id": "session_id", "lap": "lap", "points": "points"}},
	"/api/v1/sessions/{id}/same-track":       {"ListSameTrack", map[string]string{"id": "session_id"}},
	"/api/engine/status":                     {"GetEngineStatus", nil},
}

// v1's routes with no method, each with why.
var pitwallRoutesLeftOut = map[string]string{
	"/api/v1/openapi.json": "the schema: the protos in contracts are it now",
	"/healthz":             "the pit wall's own health, plain HTTP beside the API (paddock#124, R5)",
}

type openAPI struct {
	Paths map[string]map[string]struct {
		Parameters []struct {
			Name string `json:"name"`
		} `json:"parameters"`
	} `json:"paths"`
	Components struct {
		Schemas map[string]struct {
			Properties map[string]map[string]any `json:"properties"`
		} `json:"schemas"`
	} `json:"components"`
}

func loadPitwallOpenAPI(t *testing.T) openAPI {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "v1", "pitwall_openapi.json"))
	if err != nil {
		t.Fatal(err)
	}
	var api openAPI
	if err := json.Unmarshal(data, &api); err != nil {
		t.Fatal(err)
	}
	return api
}

var pitwallFile = pitwallv1.File_twr_pitwall_v1_pitwall_proto

// v1Type is a v1 property's JSON type, and whether it may be null: a type
// name, "array" with its items' in items, "ref" with its model in ref, or
// "enum".
type v1Type struct {
	kind, items, ref string
	enum             []string
	nullable         bool
}

func typeOf(prop map[string]any) v1Type {
	var out v1Type
	switch ty := prop["type"].(type) {
	case string:
		out.kind = ty
	case []any:
		for _, x := range ty {
			if x == "null" {
				out.nullable = true
			} else {
				out.kind, _ = x.(string)
			}
		}
	}
	if ref, ok := prop["$ref"].(string); ok {
		out.kind, out.ref = "ref", strings.TrimPrefix(ref, "#/components/schemas/")
	}
	for _, key := range []string{"anyOf", "allOf"} {
		alts, _ := prop[key].([]any)
		for _, a := range alts {
			alt, _ := a.(map[string]any)
			if ref, ok := alt["$ref"].(string); ok {
				out.kind, out.ref = "ref", strings.TrimPrefix(ref, "#/components/schemas/")
			} else if alt["type"] == "null" {
				out.nullable = true
			}
		}
	}
	if values, ok := prop["enum"].([]any); ok {
		out.kind = "enum"
		for _, v := range values {
			s, _ := v.(string)
			out.enum = append(out.enum, s)
		}
	}
	if out.kind == "array" {
		items, _ := prop["items"].(map[string]any)
		inner := typeOf(items)
		out.items = inner.kind
		out.ref = inner.ref
	}
	return out
}

// sameType says whether a v1 property's type is the field's, apart from
// presence.
func sameType(t *testing.T, v1 v1Type, f protoreflect.FieldDescriptor) bool {
	t.Helper()
	kind := v1.kind
	if kind == "array" {
		if !f.IsList() {
			return false
		}
		kind = v1.items
	} else if f.IsList() {
		return false
	}
	switch kind {
	case "integer":
		return f.Kind() == protoreflect.Int32Kind
	case "number":
		return f.Kind() == protoreflect.DoubleKind
	case "string":
		return f.Kind() == protoreflect.StringKind
	case "boolean":
		return f.Kind() == protoreflect.BoolKind
	case "ref":
		return f.Kind() == protoreflect.MessageKind && f.Message().FullName() == pitwallModels[v1.ref]
	}
	return false
}

// Every field of every v1 model is here, with the same meaning and type and
// presence where v1's could be null, or is listed as a difference; and every
// field here is v1's or listed as new.
func TestPitwallCarriesV1sModels(t *testing.T) {
	api := loadPitwallOpenAPI(t)
	accounted := map[string]bool{}
	used := map[string]bool{}
	for name, schema := range api.Components.Schemas {
		if _, ok := pitwallModelsLeftOut[name]; ok {
			used[name] = true
			continue
		}
		full, ok := pitwallModels[name]
		if !ok {
			t.Errorf("v1's %s maps to no message and is not in pitwallModelsLeftOut", name)
			continue
		}
		md := pitwallFile.Messages().ByName(full.Name())
		if md == nil {
			t.Errorf("%s (v1's %s) is not in pitwall.proto", full, name)
			continue
		}
		for field, prop := range schema.Properties {
			key := name + "." + field
			if _, ok := pitwallLeftOut[key]; ok {
				used[key] = true
				continue
			}
			to := field
			if r, ok := pitwallRenamed[key]; ok {
				to, used[key] = r.to, true
			}
			f := md.Fields().ByName(protoreflect.Name(to))
			if f == nil {
				t.Errorf("v1's %s has no field %s in %s, and is not in pitwallRenamed or pitwallLeftOut", key, to, full)
				continue
			}
			accounted[string(md.Name())+"."+to] = true
			v1 := typeOf(prop)
			if _, ok := pitwallRetyped[key]; ok {
				used[key] = true
			} else if !sameType(t, v1, f) {
				t.Errorf("v1's %s is %+v, but %s is %v: list it in pitwallRetyped if it is on purpose", key, v1, f.FullName(), f.Kind())
			}
			if v1.kind == "enum" && f.Kind() == protoreflect.EnumKind {
				for _, v := range v1.enum {
					value := "ACTIVITY_" + strings.ToUpper(v)
					if f.Enum().Values().ByName(protoreflect.Name(value)) == nil {
						t.Errorf("v1's %s value %q has no %s in %s", key, v, value, f.Enum().FullName())
					}
					accounted[string(f.Enum().Name())+"."+value] = true
				}
			}
			// R3: optional wherever v1's model has `| None`. A message field
			// always has presence.
			if v1.nullable && !f.HasPresence() {
				t.Errorf("v1's %s may be null, but %s is not optional", key, f.FullName())
			}
			if !v1.nullable && f.HasPresence() && f.Kind() != protoreflect.MessageKind {
				if _, ok := pitwallNowOptional[key]; !ok {
					t.Errorf("v1's %s is never null, but %s is optional: list it in pitwallNowOptional if it is on purpose", key, f.FullName())
				}
				used[key] = true
			}
		}
	}
	// Every field and enum value here is v1's or new.
	for i := range pitwallFile.Messages().Len() {
		md := pitwallFile.Messages().Get(i)
		for j := range md.Fields().Len() {
			key := string(md.Name()) + "." + string(md.Fields().Get(j).Name())
			if _, ok := pitwallNew[key]; ok {
				used[key] = true
			} else if !accounted[key] {
				t.Errorf("%s is no v1 field, and is not in pitwallNew", key)
			}
		}
	}
	for i := range pitwallFile.Enums().Len() {
		ed := pitwallFile.Enums().Get(i)
		for j := 1; j < ed.Values().Len(); j++ { // 0 is _UNSPECIFIED
			key := string(ed.Name()) + "." + string(ed.Values().Get(j).Name())
			if _, ok := pitwallNew[key]; ok {
				used[key] = true
			} else if !accounted[key] {
				t.Errorf("%s is no v1 value, and is not in pitwallNew", key)
			}
		}
	}
	// A listed difference that no longer differs is stale.
	var lists []string
	for _, m := range []map[string]string{pitwallModelsLeftOut, pitwallLeftOut, pitwallRetyped, pitwallNowOptional, pitwallNew} {
		lists = append(lists, keys(m)...)
	}
	for k := range pitwallRenamed {
		lists = append(lists, k)
	}
	for _, k := range lists {
		if !used[k] {
			t.Errorf("%s is listed as a difference, but nothing needed it", k)
		}
	}
}

// Every v1 route is a method, and each of its parameters a field of the
// method's request; every method is read-only, so Connect serves it as GET
// and the pit wall can answer a 304 (R1).
func TestPitwallMethodsAreV1sRoutes(t *testing.T) {
	api := loadPitwallOpenAPI(t)
	service := pitwallFile.Services().ByName("PitwallService")
	if service == nil {
		t.Fatal("no PitwallService")
	}
	methods := map[string]bool{}
	for path, ops := range api.Paths {
		if _, ok := pitwallRoutesLeftOut[path]; ok {
			continue
		}
		route, ok := pitwallRoutes[path]
		if !ok {
			t.Errorf("v1's %s maps to no method, and is not in pitwallRoutesLeftOut", path)
			continue
		}
		m := service.Methods().ByName(protoreflect.Name(route.method))
		if m == nil {
			t.Errorf("v1's %s: no method %s", path, route.method)
			continue
		}
		methods[route.method] = true
		for _, p := range ops["get"].Parameters {
			field, ok := route.params[p.Name]
			if !ok {
				t.Errorf("v1's %s parameter %s maps to no field of %s", path, p.Name, m.Input().Name())
			} else if m.Input().Fields().ByName(protoreflect.Name(field)) == nil {
				t.Errorf("v1's %s parameter %s: %s has no %s", path, p.Name, m.Input().Name(), field)
			}
		}
	}
	for i := range service.Methods().Len() {
		m := service.Methods().Get(i)
		if !methods[string(m.Name())] {
			t.Errorf("%s is no v1 route", m.Name())
		}
		opts, _ := m.Options().(*descriptorpb.MethodOptions)
		if opts.GetIdempotencyLevel() != descriptorpb.MethodOptions_NO_SIDE_EFFECTS {
			t.Errorf("%s is not NO_SIDE_EFFECTS, so Connect would not serve it as GET", m.Name())
		}
	}
	want := []string{"GetEngineStatus", "GetSession", "GetTrace", "ListLaps", "ListSameTrack", "ListSessions"}
	got := keys(methods)
	if !slices.Equal(got, want) {
		t.Errorf("methods %v, want %v", got, want)
	}
}

func keys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
