package contracts_test

import (
	"testing"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	placeholderv1 "github.com/those-who-run/contracts/gen/go/twr/placeholder/v1"
)

// The generated code works with the protobuf runtime go.mod requires: a message
// survives the wire form and protobuf's JSON form, the one the components send.
func TestGeneratedCodeRoundTrips(t *testing.T) {
	want := &placeholderv1.Placeholder{Note: "lap 1"}

	wire, err := proto.Marshal(want)
	if err != nil {
		t.Fatalf("proto.Marshal: %v", err)
	}
	fromWire := &placeholderv1.Placeholder{}
	if err := proto.Unmarshal(wire, fromWire); err != nil {
		t.Fatalf("proto.Unmarshal: %v", err)
	}
	if !proto.Equal(fromWire, want) {
		t.Errorf("from the wire form: got %v, want %v", fromWire, want)
	}

	json, err := protojson.Marshal(want)
	if err != nil {
		t.Fatalf("protojson.Marshal: %v", err)
	}
	fromJSON := &placeholderv1.Placeholder{}
	if err := protojson.Unmarshal(json, fromJSON); err != nil {
		t.Fatalf("protojson.Unmarshal: %v", err)
	}
	if !proto.Equal(fromJSON, want) {
		t.Errorf("from JSON %s: got %v, want %v", json, fromJSON, want)
	}
}
