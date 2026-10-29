# twr-contracts

The Python package of [`contracts`](https://github.com/those-who-run/contracts):
the messages generated from its `.proto` files, and the helpers every Python
component encodes and decodes them with. It is installed from a release tag,
with no package registry:

```toml
dependencies = [
  "twr-contracts @ git+https://github.com/those-who-run/contracts@vX.Y.Z#subdirectory=gen/python",
]
```

- `twr/`: the generated code, one module per `.proto` file, such as
  `twr.events.v1.callout_pb2`, with its type stubs. `make generate` writes it;
  it is never edited by hand.
- `twr_contracts.ce`: encodes and decodes a message and its CloudEvents
  headers, as the Go package `ce` does: protobuf's JSON form, unknown fields
  ignored on read.
- `twr_contracts.callout`: a callout's subject from its priority, as the Go
  package `callout` builds it.

```python
from twr.events.v1 import callout_pb2
from twr_contracts import callout, ce

c = callout_pb2.Callout()
event = ce.decode(msg.headers or {}, msg.data, c)  # ce.WrongTypeError on another ce-type
```

The repository's README has the rest.
