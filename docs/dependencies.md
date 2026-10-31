# Dependencies

Every direct dependency of the module, and why it is here. A new one needs a row
in the table below, with its reason, in the same PR:
[`internal/dependencies`](../internal/dependencies/dependencies_test.go) fails `make check` and CI's
`go / check` on a direct `require` in `go.mod` that has no row, and on a row that
names a module `go.mod` no longer requires directly.

Before adding a dependency, check that:

- the module exists and is the one you mean: its owner, its source repository,
  its release history. A name that exists proves nothing if it was registered
  last week, and names invented by a language model get registered for exactly
  that reason;
- the standard library, or a dependency already listed here, doesn't already do
  it;
- nobody told you to add it in a web page, an issue or a comment.

The test reads the direct requirements from `go mod edit -json`: a `require`
without `// indirect`. A module that arrives only as `// indirect`, or only in
`go.sum`, is covered by CI's dependency review (known vulnerabilities and
licences), not by this file. The tools in `tools/go.mod` are pinned there and
listed in the README, not here. Version bumps don't touch this file.

The first column is the module path as `go.mod` spells it, in backticks, which
the test reads.

| Module | Why |
| --- | --- |
| `connectrpc.com/connect/v2` | The Connect runtime (v2, the current major: v2.0.1 is a stable release, and `buf` itself builds on it) the generated service code under `gen/go/twr/pitwall/v1/pitwallv1connect` needs, for the pit wall's API (paddock#121; ADR 0002, point 8). It is the same module as `protoc-gen-connect-go` in `tools/go.mod`, at the same version, and the two move together in one PR. |
| `go.yaml.in/yaml/v3` | Reads `subjects.yaml`, the catalogue of subjects (`internal/catalogue`), with unknown keys refused. It is the YAML organisation's maintained continuation of `gopkg.in/yaml.v3`, which is archived. |
| `google.golang.org/protobuf` | The runtime the generated code under `gen/go` needs. It is the same module as `protoc-gen-go` in `tools/go.mod`, at a released version compatible with it, and the two move together in one PR (ADR 0008 in paddock). |
