# Contracts

The contracts every Those Who Run v2 component builds against: the
protobuf messages and services, and `subjects.yaml`, the catalogue of
NATS subjects.
Components use them only by released tag, through the code generated
from them ([ADR 0002](https://github.com/those-who-run/paddock/blob/main/docs/adr/0002-contracts.md) in paddock).

Issues are off in this repository. Every issue, the roadmap and the decisions
(ADRs) live in [those-who-run/paddock](https://github.com/those-who-run/paddock). A pull request here closes
its paddock issue with `Closes those-who-run/paddock#N.`

## Checks

`make check` runs what CI runs; [`AGENTS.md`](AGENTS.md) says what that is and
what a new clone runs first.
