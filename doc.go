// Package contracts is the contracts every Those Who Run v2 component builds
// against; README.md says what they are and how a component pins a release.
//
// The messages are generated under gen/go/, one Go package per protobuf
// package, such as gen/go/twr/events/v1. Package ce encodes and decodes them
// with their CloudEvents headers, the same way in every component, and package
// format writes numbers as text, the same way on the results card and the pit
// wall.
package contracts
