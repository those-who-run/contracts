// Package callout builds the subject a callout travels on, from its priority
// (paddock#71, ADR 0001 point 2 in paddock): twr.v1.<src>.callout.p0, .p1 or
// .p2, so radio can take p0 before p1 and p2. The levels keep v1's order: p0
// is a priority under 10, p1 10 to 19 and p2 20 and over, and the start brief
// (8) is p0.
//
// A producer publishes with it and ce:
//
//	subject, err := callout.Subject("live", c)
//	h, data, err := ce.Encode(c, ce.Event{ID: c.GetId(), Source: "/twr/ecu/gt7", Time: c.GetCreatedAt().AsTime()})
//	msg := &nats.Msg{Subject: subject, Header: nats.Header(h), Data: data}
package callout

import (
	"errors"
	"fmt"
	"strings"

	eventsv1 "github.com/those-who-run/contracts/gen/go/twr/events/v1"
)

// Level is a callout's level, the subject's last token.
type Level string

// The levels, most urgent first.
const (
	P0 Level = "p0"
	P1 Level = "p1"
	P2 Level = "p2"
)

// Levels are every level, most urgent first: a consumer subscribes to each.
var Levels = []Level{P0, P1, P2}

// LevelOf is the level of a priority, v1's number where lower is more urgent:
// p0 under 10, p1 10 to 19, p2 20 and over.
func LevelOf(priority int32) Level {
	switch {
	case priority < 10:
		return P0
	case priority < 20:
		return P1
	default:
		return P2
	}
}

// LevelSubject is the subject of a level's callouts from src, such as
// twr.v1.live.callout.p0.
func LevelSubject(src string, l Level) (string, error) {
	if src == "" || strings.ContainsAny(src, ".*> \t\r\n") {
		return "", fmt.Errorf("source %q is not one subject token", src)
	}
	return "twr.v1." + src + ".callout." + string(l), nil
}

// Subject is the subject the callout travels on from src (`live`, `parity`
// or `replay-<name>`): its priority's level. A priority of 0 is unset, since
// a publisher sets it from 1, so it is refused rather than sent as p0.
func Subject(src string, c *eventsv1.Callout) (string, error) {
	if c.GetPriority() == 0 {
		return "", errors.New("the callout has no priority: a publisher sets it from 1")
	}
	return LevelSubject(src, LevelOf(c.GetPriority()))
}
