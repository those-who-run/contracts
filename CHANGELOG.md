# Changelog

## [0.2.2](https://github.com/those-who-run/contracts/compare/v0.2.1...v0.2.2) (2026-10-29)


### Features

* the callout message on callout.&lt;p0|p1|p2&gt; ([#26](https://github.com/those-who-run/contracts/issues/26)) ([cf1a938](https://github.com/those-who-run/contracts/commit/cf1a93836a5d48b2a6aaf75909a120dd6086171d))

## [0.2.1](https://github.com/those-who-run/contracts/compare/v0.2.0...v0.2.1) (2026-10-29)


### Features

* the pit, car and telemetry events ([#25](https://github.com/those-who-run/contracts/issues/25)) ([8e89653](https://github.com/those-who-run/contracts/commit/8e8965317357c5fded0a138ce4217e4fbef516fd))

## [0.2.0](https://github.com/those-who-run/contracts/compare/v0.1.5...v0.2.0) (2026-10-28)


### Features

* the session and race events, and the full session summary ([#22](https://github.com/those-who-run/contracts/issues/22)) ([ded14ae](https://github.com/those-who-run/contracts/commit/ded14ae96742d3aa04d93579964d909680bba58e))

## [0.1.5](https://github.com/those-who-run/contracts/compare/v0.1.4...v0.1.5) (2026-10-28)


### Features

* **subjects:** event ids carry the source ([#17](https://github.com/those-who-run/contracts/issues/17)) ([0bb1a72](https://github.com/those-who-run/contracts/commit/0bb1a7261a992462c8d146b9d0d0fd1ea95b007a))

## [0.1.4](https://github.com/those-who-run/contracts/compare/v0.1.3...v0.1.4) (2026-10-27)


### Features

* EndReason RESTARTED (6), NEXT_SESSION_STARTED (7) and INPUT_ENDED (8) ([#15](https://github.com/those-who-run/contracts/issues/15)) ([491cd3c](https://github.com/those-who-run/contracts/commit/491cd3cdaa0ffc71f4e39783f6a052736436d7a4))
* Frame.seq counts from 1; finishing_position only for a final LAP_RACE at the flag; the joined lap is complete; position absent outside a field, unlike v1 ([#15](https://github.com/those-who-run/contracts/issues/15)) ([491cd3c](https://github.com/those-who-run/contracts/commit/491cd3cdaa0ffc71f4e39783f6a052736436d7a4))
* SessionEnded.last_seq (4), the seq of the session's last frame ([#15](https://github.com/those-who-run/contracts/issues/15)) ([491cd3c](https://github.com/those-who-run/contracts/commit/491cd3cdaa0ffc71f4e39783f6a052736436d7a4))
* SessionStarted.fuel_at_start (10), the fuel on the session's first frame, in litres ([#15](https://github.com/those-who-run/contracts/issues/15)) ([491cd3c](https://github.com/those-who-run/contracts/commit/491cd3cdaa0ffc71f4e39783f6a052736436d7a4))
* SessionSummary.final_position (6), the position on the last frame, absent outside a field ([#15](https://github.com/those-who-run/contracts/issues/15)) ([491cd3c](https://github.com/those-who-run/contracts/commit/491cd3cdaa0ffc71f4e39783f6a052736436d7a4))
* SessionSummary.fuel_used (5), the session's fuel, in litres, as v1 computes it ([#15](https://github.com/those-who-run/contracts/issues/15)) ([491cd3c](https://github.com/those-who-run/contracts/commit/491cd3cdaa0ffc71f4e39783f6a052736436d7a4))
* SessionSummary.session_type (7), ecu's final verdict on the session's type ([#15](https://github.com/those-who-run/contracts/issues/15)) ([491cd3c](https://github.com/those-who-run/contracts/commit/491cd3cdaa0ffc71f4e39783f6a052736436d7a4))
* SessionType LAP_RACE (5), SOLO_LAPS (6), TIMED_FIELD (7) and SOLO_TIMED (8), one to one with v1's session kinds ([#15](https://github.com/those-who-run/contracts/issues/15)) ([491cd3c](https://github.com/those-who-run/contracts/commit/491cd3cdaa0ffc71f4e39783f6a052736436d7a4))

## [0.1.3](https://github.com/those-who-run/contracts/compare/v0.1.2...v0.1.3) (2026-10-26)


### Features

* package ce, the Go helpers for the JSON form and CloudEvents headers ([#8](https://github.com/those-who-run/contracts/issues/8)) ([00c74ec](https://github.com/those-who-run/contracts/commit/00c74ec66db97f3cd09f4512a6934e59299ab6e0))

## [0.1.2](https://github.com/those-who-run/contracts/compare/v0.1.1...v0.1.2) (2026-10-26)


### Features

* subjects.yaml, the catalogue of subjects, and its checks ([#7](https://github.com/those-who-run/contracts/issues/7)) ([97b9be0](https://github.com/those-who-run/contracts/commit/97b9be07c692884942c283c052982b3e3b71dea0))

## [0.1.1](https://github.com/those-who-run/contracts/compare/v0.1.0...v0.1.1) (2026-10-26)


### Features

* the v0.1 messages: Frame, SessionStarted, LapCompleted, SessionEnded ([#5](https://github.com/those-who-run/contracts/issues/5)) ([bd7518f](https://github.com/those-who-run/contracts/commit/bd7518feaf38b439d3174d60a7ccf9d4b2798d49))

## 0.1.0 (2026-10-26)


### Features

* buf, Go generation and the proto checks, from the chassis template ([#3](https://github.com/those-who-run/contracts/issues/3)) ([52bff3c](https://github.com/those-who-run/contracts/commit/52bff3c33f4bc1467467d0f10a0b5005115ff533))
