#!/usr/bin/env bash
# Repeatable gocloak benchmark run.
#   offline: JSON and HTTP stacks against an in-process fake Keycloak (no docker needed)
#   live:    the same HTTP stacks against the docker Keycloak of the main repo
# Every suite runs twice: on the encoding/json v1 engine (GOEXPERIMENT=nojsonv2) and on the v2 engine (Go 1.27
# default). Client comparisons use the v2 run.
# Results and benchstat comparisons land in results/<timestamp>/.
#
# Run on an otherwise idle machine (no builds, IDE indexing, other agents or containers): other load shows up as
# variance (benchstat "±") and can hide or fake differences. Compare runs only if env.txt matches.
#
# usage: ./run.sh [offline|live|all]
# env:   COUNT (default 10), CPU (GOMAXPROCS, default 4), BENCHTIME (default 1s), KEEP_KEYCLOAK=1
set -euo pipefail
cd "$(dirname "$0")"

MODE=${1:-all}
COUNT=${COUNT:-10}
CPU=${CPU:-4}
BENCHTIME=${BENCHTIME:-1s}
OUT="results/$(date +%Y%m%d-%H%M%S)"
mkdir -p "$OUT"

{
    go version
    go env GOOS GOARCH
    echo "COUNT=$COUNT CPU=$CPU BENCHTIME=$BENCHTIME MODE=$MODE"
    git rev-parse HEAD 2>/dev/null || true
} >"$OUT/env.txt"

if [ -r /proc/loadavg ] && awk '{exit !($1 > 1.0)}' /proc/loadavg; then
    echo "warning: load average $(cut -d' ' -f1 /proc/loadavg) > 1, results will be noisy" | tee -a "$OUT/env.txt"
fi

# bench <name> <pattern> [extra go test args]: runs one suite and stores the raw output as <name>.txt
bench() {
    local name=$1 pattern=$2
    shift 2
    go test -run '^$' -bench "$pattern" -benchmem -count "$COUNT" -cpu "$CPU" -benchtime "$BENCHTIME" \
        -timeout 0 "$@" . | tee "$OUT/$name.txt"
}

# compare <name> <benchstat args...>: stores a benchstat comparison as cmp-<name>.txt
compare() {
    local name=$1
    shift
    # MSYS_NO_PATHCONV: Git Bash on Windows would rewrite "/client" into a Windows path
    MSYS_NO_PATHCONV=1 go tool benchstat "$@" | tee "$OUT/cmp-$name.txt"
}

run_offline() {
    GOEXPERIMENT=nojsonv2 bench json-v1 '^BenchmarkJSON'
    bench json-v2 '^BenchmarkJSON'
    GOEXPERIMENT=nojsonv2 bench offline-v1 '^BenchmarkOffline'
    bench offline-v2 '^BenchmarkOffline'

    # v1 API on the v1 engine vs v1 API on the v2 engine (no code change, only the experiment flag)
    compare json-engine -filter '/api:v1' v1engine="$OUT/json-v1.txt" v2engine="$OUT/json-v2.txt"
    # v1 API vs direct v2 API, both on the v2 engine
    compare json-api -col /api "$OUT/json-v2.txt"
    # HTTP stacks against each other
    compare offline-clients -col /client "$OUT/offline-v2.txt"
    # whole request path with the v1 vs the v2 engine
    compare offline-jsonv2 v1engine="$OUT/offline-v1.txt" v2engine="$OUT/offline-v2.txt"
}

run_live() {
    local started=0
    if ! curl -fsS -o /dev/null http://localhost:8080/realms/master; then
        docker compose -f ../docker-compose.yml up -d --build --wait --wait-timeout 400
        started=1
    fi
    if [ "$started" = 1 ] && [ "${KEEP_KEYCLOAK:-0}" != 1 ]; then
        trap 'docker compose -f ../docker-compose.yml down' EXIT
    fi
    GOEXPERIMENT=nojsonv2 bench live-v1 '^BenchmarkKeycloak' -tags keycloak
    bench live-v2 '^BenchmarkKeycloak' -tags keycloak

    compare live-clients -col /client "$OUT/live-v2.txt"
    compare live-jsonv2 v1engine="$OUT/live-v1.txt" v2engine="$OUT/live-v2.txt"
}

case "$MODE" in
offline) run_offline ;;
live) run_live ;;
all)
    run_offline
    run_live
    ;;
*)
    echo "usage: $0 [offline|live|all]" >&2
    exit 2
    ;;
esac

echo "results: $OUT"
