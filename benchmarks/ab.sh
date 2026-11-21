#!/usr/bin/env bash
# Interleaved A/B comparison: builds both variants once and alternates them round by round (also swapping which one
# goes first), so load that changes over time (other processes, thermal throttling, a warming JVM) hits both alike.
# Use this for "old vs new" claims; run.sh runs whole suites one after another and is only fair on an idle machine.
#
# usage: ./ab.sh json        encoding/json on the v1 engine (GOEXPERIMENT=nojsonv2) vs the v2 engine (Go 1.27 default)
#        ./ab.sh ref [REF]   gocloak at git REF (default: main) built with BASE_GO vs the working tree with the
#                            current toolchain; also compares the working tree with SetDirectBodyRead
# env:   ROUNDS (default 10), CPU (default 4), BENCHTIME (default 1s), BASE_GO (default go1.26.2),
#        LIVE=1 adds the Keycloak benchmarks to ref mode (needs the docker Keycloak, see run.sh)
set -euo pipefail
cd "$(dirname "$0")"
export MSYS_NO_PATHCONV=1 # Git Bash on Windows would rewrite "/client" style arguments into paths

MODE=${1:-json}
ROUNDS=${ROUNDS:-10}
CPU=${CPU:-4}
BENCHTIME=${BENCHTIME:-1s}
OUT="results/ab-$MODE-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$OUT"

# run <binary> <pattern> <output file>: one count of the selected benchmarks, appended to the output file
run() {
    "$1" -test.run '^$' -test.bench "$2" -test.benchmem -test.count 1 -test.cpu "$CPU" \
        -test.benchtime "$BENCHTIME" -test.timeout 0 | grep -E '^(Benchmark|goos|goarch|pkg|cpu)' >>"$3"
}

# interleave <round> <cmd a> <cmd b>: runs both commands, a first in odd rounds, b first in even rounds
interleave() {
    if [ $(($1 % 2)) = 1 ]; then
        $2
        $3
    else
        $3
        $2
    fi
}

json_mode() {
    GOEXPERIMENT=nojsonv2 go test -c -o "$OUT/a.test" .
    go test -c -o "$OUT/b.test" .
    # -bench splits patterns at "/" per sub-benchmark level, so each selection is its own pattern
    local patterns=('^BenchmarkJSON/./^api=v1$' '^BenchmarkOfflineHTTP$/./^client=gocloak$' '^BenchmarkOfflineDecodeAccessToken$')
    a() { for p in "${patterns[@]}"; do run "$OUT/a.test" "$p" "$OUT/v1engine.txt"; done; }
    b() { for p in "${patterns[@]}"; do run "$OUT/b.test" "$p" "$OUT/v2engine.txt"; done; }
    for i in $(seq 1 "$ROUNDS"); do
        echo "round $i/$ROUNDS"
        interleave "$i" a b
    done
    go tool benchstat v1engine="$OUT/v1engine.txt" v2engine="$OUT/v2engine.txt" | tee "$OUT/cmp-json.txt"
}

# build_baseline <ref> <tags>: builds the benchmarks against gocloak at ref with BASE_GO into $OUT/base.test
build_baseline() {
    # absolute paths in Windows form under Git Bash (pwd -W): MSYS_NO_PATHCONV stops MSYS from converting them for
    # native programs like git and go
    local ref=$1 tags=$2 root worktree
    root=$(pwd -W 2>/dev/null || pwd)
    worktree="$root/$OUT/gocloak-$1"
    git -C .. worktree add --detach "$worktree" "$ref" >/dev/null
    trap 'git -C .. worktree remove --force "'"$worktree"'" >/dev/null 2>&1 || true' EXIT
    # a copy of go.mod that points the replace at the worktree and allows the older toolchain
    cp go.mod "$OUT/base.mod"
    cp go.sum "$OUT/base.sum"
    go mod edit -modfile="$OUT/base.mod" -replace "github.com/Nerzal/gocloak/v15=$worktree" -go=1.26.0
    GOTOOLCHAIN=${BASE_GO:-go1.26.2} go test -c -modfile="$OUT/base.mod" -tags "$tags" -o "$OUT/base.test" .
}

ref_mode() {
    local ref=${1:-main} tags=baseline
    if [ "${LIVE:-0}" = 1 ]; then
        tags="baseline,keycloak"
    fi
    build_baseline "$ref" "$tags"
    go test -c -tags "${tags#baseline,}" -o "$OUT/new.test" .
    if [ "${LIVE:-0}" != 1 ]; then
        go test -c -o "$OUT/new.test" .
    fi
    {
        echo "baseline: gocloak $ref ($(git -C .. rev-parse --short "$ref")), ${BASE_GO:-go1.26.2}"
        echo "new: working tree ($(git -C .. rev-parse --short HEAD) + changes), $(go env GOVERSION)"
        echo "ROUNDS=$ROUNDS CPU=$CPU BENCHTIME=$BENCHTIME LIVE=${LIVE:-0}"
    } >"$OUT/env.txt"

    local patterns=('^BenchmarkOfflineHTTP$/./^client=gocloak$' '^BenchmarkOfflinePool$/^proto=tls-h1$/^client=gocloak$')
    local direct=('^BenchmarkOfflineHTTP$/./^client=gocloak-direct$')
    if [ "${LIVE:-0}" = 1 ]; then
        # every process start seeds Keycloak with 1000 users, so live mode runs only the live benchmarks
        patterns=('^BenchmarkKeycloakHTTP$/^endpoint=(token|user|users100|users1000)$/^client=gocloak$' '^BenchmarkKeycloakPool$/^client=gocloak$')
        direct=('^BenchmarkKeycloakHTTP$/^endpoint=(token|user|users100|users1000)$/^client=gocloak-direct$')
    fi
    base() { for p in "${patterns[@]}"; do run "$OUT/base.test" "$p" "$OUT/base.txt"; done; }
    new() {
        for p in "${patterns[@]}"; do run "$OUT/new.test" "$p" "$OUT/new.txt"; done
        for p in "${direct[@]}"; do run "$OUT/new.test" "$p" "$OUT/new-direct.txt"; done
    }
    for i in $(seq 1 "$ROUNDS"); do
        echo "round $i/$ROUNDS"
        interleave "$i" base new
    done
    go tool benchstat base="$OUT/base.txt" new="$OUT/new.txt" | tee "$OUT/cmp-base-new.txt"
    # resty body read vs SetDirectBodyRead, both on the working tree; names differ, so compare them as columns
    {
        grep -E '^(goos|goarch|pkg|cpu):' "$OUT/new.txt" | head -4
        grep -hE '^Benchmark(OfflineHTTP|KeycloakHTTP)/' "$OUT/new.txt" "$OUT/new-direct.txt"
    } >"$OUT/new-both.txt"
    go tool benchstat -col /client "$OUT/new-both.txt" | tee "$OUT/cmp-direct.txt"
}

case "$MODE" in
json) json_mode ;;
ref) ref_mode "${2:-main}" ;;
*)
    echo "usage: $0 json | ref [REF]" >&2
    exit 2
    ;;
esac
echo "results: $OUT"
