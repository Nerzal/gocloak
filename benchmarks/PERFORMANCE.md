# gocloak performance

Measurements behind the performance work in gocloak v15.1.0 (Go 1.27, resty v2.17.2, Keycloak 26.8.2 in Docker,
AMD Ryzen AI 9 HX 370, Windows 11). Timings are benchstat medians at GOMAXPROCS=4; allocations are deterministic.
The machine ran other workloads at the same time, so every before/after number below comes from `./ab.sh`, which
runs both builds alternately round by round. Separate runs minutes apart produced fake effects of up to ±40 % here
(for example `DecodeAccessToken` looked 35 % slower offline and 39 % faster live; interleaved, it did not change).

## Summary

| Change | Effect (interleaved A/B, n=10) | Status |
|---|---|---|
| 100 idle connections per host instead of resty's `GOMAXPROCS+1` | 16 concurrent callers over TLS/HTTP1.1: −57 % time, −98 % new connections | default in v15.1.0 |
| Go 1.27 (`encoding/json` on the v2 engine) | decoding −28 to −70 %, allocations −40 %; `GetUsers(1000)` −27 % | minimum Go version in v15.1.0 |
| `SetDirectBodyRead()` | `GetUsers(1000)`: −39 % allocated bytes, −22 % time; small responses unchanged | opt-in option |
| v15.0 on Go 1.26 → v15.1 on Go 1.27, all defaults | `GetUsers(1000)` −24 % time, −44 % allocations; geomean −21 % time, −27 % allocations | |
| Calling `encoding/json/v2` directly instead of `encoding/json` | small single objects and encoding faster, lists unchanged; different encoding semantics | rejected |
| Switching to fasthttp or plain net/http | live: within noise (0 to −10 % on small calls) | rejected |

A single call is dominated by Keycloak: about 0.9 ms for one user and 40 ms for 1000 users against a warm local
Keycloak; gocloak spends microseconds to a few milliseconds of CPU. The changes matter for throughput and garbage
under concurrent load, not for the latency of one call. Against the real Keycloak (plain HTTP, interleaved A/B, n=6):

| v15.0 / Go 1.26 → v15.1 / Go 1.27 | time | bytes | allocations |
|---|---|---|---|
| token | n.s. (1.95 ms) | −4 % | −10 % |
| user | n.s. (0.91 ms) | −6 % | −19 % |
| 100 users | n.s. (4.7 ms) | −5 % | −27 % |
| 1000 users | n.s. (41 → 39 ms, p=0.13) | −3 % | −27 % |
| 16 concurrent callers | n.s. | −12 % | −24 %, new connections −70 % |
| 1000 users with `SetDirectBodyRead` | n.s. | −29 % | |

## Connection pool

resty sets `MaxIdleConnsPerHost = GOMAXPROCS+1`. Keycloak is a single host, and since Go 1.25 GOMAXPROCS follows the
container CPU limit, so a pod limited to 1 CPU kept 2 idle connections. Every further concurrent request opened a new
connection (with a TLS handshake) and closed it afterwards. `NewClient` now sets `MaxIdleConns` and
`MaxIdleConnsPerHost` to 100; `RestyClient().Transport()` can still change them.

| 16 concurrent `GetUserByID`, TLS/HTTP1.1, offline | v15.0 / Go 1.26 | v15.1 / Go 1.27 |
|---|---|---|
| time per call | 37.2 µs | 15.8 µs (−57 %) |
| new connections per 1000 calls | 33.3 | 0.64 (−98 %) |
| allocations per call | 195 | 135 (−31 %) |

pprof of the old default: 27 % of all CPU went into TLS handshakes, 17 % of it RSA signatures on the server side;
with the larger pool it is 0.7 %. In production that handshake CPU is spent by Keycloak. With the old pool, 32 to 192
concurrent callers created so many connections that the listener (and Docker Desktop's port forwarder) started
refusing them. HTTP/2 (Keycloak over TLS) avoids the problem entirely because all requests share one connection.

## encoding/json v2 engine

Go 1.27 enables `GOEXPERIMENT=jsonv2` by default: the `encoding/json` API runs on the v2 engine without code changes.
Comparison on Go 1.27.4 with `GOEXPERIMENT=nojsonv2` as the baseline:

| | v1 engine | v2 engine | Δ |
|---|---|---|---|
| unmarshal token response | 13.6 µs | 4.0 µs | −70 % |
| unmarshal JWKS | 5.2 µs | 2.0 µs | −61 % |
| unmarshal user | 4.1 µs | 2.6 µs | −37 % |
| unmarshal 100 users | 393 µs | 274 µs | −30 % |
| unmarshal 1000 users | 3.87 ms | 2.53 ms | −35 % |
| unmarshal clients | 88 µs | 53 µs | −39 % |
| unmarshal realm export | 464 µs | 334 µs | −28 % |
| marshal user (create) | 1.35 µs | 2.02 µs | **+50 %** |
| marshal client | 1.72 µs | 2.47 µs | **+44 %** |
| gocloak `GetUsers(100)`, full path | 680 µs | 549 µs | −19 % |
| gocloak `GetUsers(1000)`, full path | 6.07 ms | 4.46 ms | −27 % |
| token, user, groups, clients, `DecodeAccessToken` | | | not significant |
| allocations, geomean | | | −40 % |

Why: pprof shows the v1 engine spending about 21 % of its CPU in `encoding/json.checkValid`, because it validates
the whole input before decoding it in a second pass. The v2 engine (`encoding/json/jsontext`) reads in one pass.
Encoding small structs got slower; request bodies are small and Keycloak takes milliseconds, so it does not show up
end to end. The full integration test suite passes against Keycloak on Go 1.27.

gocloak keeps the `encoding/json` API and does not call `encoding/json/v2` directly. Both APIs on the v2 engine
(single run, not interleaved, n=8): the v2 API decodes a user or the JWKS 28 to 43 % faster, lists of users, groups
and clients and the realm export show no significant difference, and encoding gets 13 % faster with 45 to 57 % fewer
allocations. `TestJSONv2Equivalence` shows that the v2 API decodes every payload exactly like v1, but encoding
differs: nil slices become `[]` instead of `null`, `omitempty` omits different values, and field names match
case-sensitively. On PUT, Keycloak treats a missing field as "keep", so those differences can change what an update
does. With `json.DefaultOptionsV1()` v2 behaves like v1 again and keeps no advantage. Callers who want the v2 API for
request bodies anyway can set it via `RestyClient().SetJSONMarshaler`.

## Reading response bodies

gocloak limits response bodies to 100 MiB. With a limit set, resty reads the body in 512-byte chunks and appends
them to a slice without using Content-Length (`readAllWithLimit`). For `GetUsers(max=1000)` (about 470 KB of JSON)
that was 61–67 % of all allocated bytes. `SetDirectBodyRead()` makes gocloak read the body itself into a buffer sized
from Content-Length (trusted only up to the limit), with the same limit, gzip handling and error decoding as resty;
a unit test checks that both paths return identical results and errors.

| v15.1 offline | resty reads | `SetDirectBodyRead` |
|---|---|---|
| `GetUsers(1000)` bytes / time | 3.65 MiB / 6.56 ms | 2.22 MiB (−39 %) / 5.14 ms (−22 %) |
| `GetUsers(100)` bytes / time | 357 KiB / 729 µs | 270 KiB (−24 %) / 591 µs (−19 %) |
| clients bytes | 91 KiB | 60 KiB (−34 %) |
| token, user | | unchanged |

It is opt-in because resty skips its response middleware when it does not read the body: `OnAfterResponse` hooks
registered on `RestyClient()` do not run for these calls and `SetDebug` does not log response bodies. The hooks are
unexported in resty v2, so gocloak cannot call them itself. Wrapping the transport instead would break
`Transport()`, `SetTLSClientConfig` and `SetProxy`, which need the concrete `*http.Transport`.

## Struct shape of large lists

Decoding 1000 users (Go 1.27, n=6):

| target | time | bytes | allocations |
|---|---|---|---|
| `[]*gocloak.User` (pointer scalars) | 2.49 ms | 1.19 MiB | 25.1k |
| same fields as values | not significant | +49 % | −40 % |
| 8-field struct of your own | −54 % | −72 % | −84 % |
| `json.NewDecoder` streaming instead of `Unmarshal` | +10 % | +82 % | same |

The pointer fields cost about 10 allocations per user, but replacing them with values does not make decoding faster
(larger slice elements) and would break updates: Keycloak treats a missing field as "keep" on PUT, so `false`, `0`
and `""` must stay sendable. Streaming decode is slower and allocates more than unmarshalling a byte slice. The real
saving is decoding into your own small struct when you only need a few fields; the README shows how with
`GetRequestWithBearerAuth`.

## HTTP clients

Sequential calls against Keycloak (plain HTTP, median per call, run.sh, clients compared within one run):

| endpoint | gocloak | resty | net/http | fasthttp |
|---|---|---|---|---|
| token | 1.81 ms | 1.74 ms | 1.75 ms | 1.63 ms (−10 %) |
| user | 906 µs | 952 µs | 959 µs | 820 µs (−9 %) |
| 100 users | 4.66 ms | 4.68 ms | 4.67 ms | 4.59 ms (n.s.) |
| 1000 users | 39.8 ms | 39.4 ms | 39.0 ms | 38.6 ms (n.s.) |

fasthttp allocates 55–85 % fewer bytes, but it has no HTTP/2, is not an `http.RoundTripper` and would break
`RestyClient()`/`SetRestyClient` and users' proxy, TLS and tracing configuration. resty v3 is still a release
candidate. gocloak stays on resty v2.

## Already fine

- JWKS certificates are cached and their public keys parsed once per fetch; local token verification costs about
  30 µs of CPU, against 1–2 ms for an introspection round trip.
- The server version is cached.

## Reproduce

```sh
cd benchmarks
./ab.sh ref main         # working tree vs main, interleaved; LIVE=1 for the Keycloak benchmarks
./ab.sh json             # encoding/json v1 engine vs v2 engine
./run.sh all             # full offline and live suites (COUNT, CPU, BENCHTIME via env)
./profile.sh all         # pprof CPU/heap old vs new, SVG call graphs and diff graphs (needs Graphviz)
```

`run.sh live` starts the Keycloak of the main repository via `docker compose` if it is not running; `profile.sh live`
and `ab.sh ref` with `LIVE=1` expect it running (`make start-keycloak` in the repository root). In the client
benchmarks, `gocloak` is the current default, `gocloak-restypool` the same client with resty's old pool size
(`GOMAXPROCS+1`) and `gocloak-direct` uses `SetDirectBodyRead`. Run on an idle machine for absolute numbers; use `ab.sh` for before/after
claims in any case. Raw results go to `results/` and `profiles/` (git-ignored).
