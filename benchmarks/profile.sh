#!/usr/bin/env bash
# CPU and heap profiles "old vs new" for selected benchmarks, with text tops, SVG call graphs and diff graphs.
#   old = gocloak before v15.1.0: encoding/json v1 engine (GOEXPERIMENT=nojsonv2), resty's connection pool
#   new = v15.1.0 on Go 1.27: v2 engine, gocloak's connection pool, SetDirectBodyRead where named
# Needs Graphviz (dot) for SVGs. Run on an idle machine, like run.sh.
#
# usage: ./profile.sh [offline|live|all]   env: BENCHTIME (default 5s), CPU (default 4)
# output: profiles/<timestamp>/<scenario>/{old,new}/{cpu,mem}.pprof, *.txt, *.svg and diff-*.svg
set -euo pipefail
cd "$(dirname "$0")"
export MSYS_NO_PATHCONV=1 # Git Bash on Windows would rewrite benchmark patterns starting with "/"

if ! command -v dot >/dev/null && [ -x "/c/Program Files/Graphviz/bin/dot.exe" ]; then
    export PATH="$PATH:/c/Program Files/Graphviz/bin"
fi
command -v dot >/dev/null || echo "warning: Graphviz dot not found, SVGs will be skipped" >&2

MODE=${1:-all}
BENCHTIME=${BENCHTIME:-5s}
CPU=${CPU:-4}
OUT="profiles/$(date +%Y%m%d-%H%M%S)"
mkdir -p "$OUT"
{
    go version
    echo "BENCHTIME=$BENCHTIME CPU=$CPU MODE=$MODE"
    git rev-parse HEAD 2>/dev/null || true
} >"$OUT/env.txt"

# profile <dir> <pattern> <experiment> [extra go test args]: runs one benchmark with CPU and heap profiling
profile() {
    local dir=$1 pattern=$2 experiment=$3
    shift 3
    mkdir -p "$dir"
    GOEXPERIMENT=$experiment go test -run '^$' -bench "$pattern" -benchmem -count 1 -cpu "$CPU" \
        -benchtime "$BENCHTIME" -timeout 0 -o "$dir/bench.test" \
        -cpuprofile "$dir/cpu.pprof" -memprofile "$dir/mem.pprof" "$@" . | tee "$dir/bench.txt"
    go tool pprof -top -nodecount=40 "$dir/bench.test" "$dir/cpu.pprof" >"$dir/cpu-top.txt" 2>/dev/null
    go tool pprof -top -nodecount=40 -sample_index=alloc_space "$dir/bench.test" "$dir/mem.pprof" >"$dir/mem-top.txt" 2>/dev/null
    if command -v dot >/dev/null; then
        go tool pprof -svg -nodecount=60 -output "$dir/cpu.svg" "$dir/bench.test" "$dir/cpu.pprof" 2>/dev/null
        go tool pprof -svg -nodecount=60 -sample_index=alloc_space -output "$dir/mem.svg" "$dir/bench.test" "$dir/mem.pprof" 2>/dev/null
    fi
}

# diff <scenario dir>: old as base, new on top; red = more in new, green = less in new
diff_profiles() {
    local s=$1
    go tool pprof -top -nodecount=40 -diff_base "$s/old/cpu.pprof" "$s/new/bench.test" "$s/new/cpu.pprof" >"$s/diff-cpu-top.txt" 2>/dev/null
    go tool pprof -top -nodecount=40 -sample_index=alloc_space -diff_base "$s/old/mem.pprof" "$s/new/bench.test" "$s/new/mem.pprof" >"$s/diff-mem-top.txt" 2>/dev/null
    if command -v dot >/dev/null; then
        go tool pprof -svg -nodecount=60 -diff_base "$s/old/cpu.pprof" -output "$s/diff-cpu.svg" "$s/new/bench.test" "$s/new/cpu.pprof" 2>/dev/null
        go tool pprof -svg -nodecount=60 -sample_index=alloc_space -diff_base "$s/old/mem.pprof" -output "$s/diff-mem.svg" "$s/new/bench.test" "$s/new/mem.pprof" 2>/dev/null
    fi
}

# scenario <name> <old pattern> <old experiment> <new pattern> <new experiment> [extra go test args]
scenario() {
    local name=$1 old=$2 old_exp=$3 new=$4 new_exp=$5
    shift 5
    profile "$OUT/$name/old" "$old" "$old_exp" "$@"
    profile "$OUT/$name/new" "$new" "$new_exp" "$@"
    diff_profiles "$OUT/$name"
}

run_offline() {
    # pure decoding of 1000 users: v1 engine vs v2 engine
    scenario json-users1000 '^BenchmarkJSONUnmarshal/payload=users1000/api=v1$' nojsonv2         '^BenchmarkJSONUnmarshal/payload=users1000/api=v1$' ''
    # whole gocloak.GetUsers(max=1000) path against the fake Keycloak: v1 engine vs v2 engine
    scenario getusers1000 '^BenchmarkOfflineHTTP/endpoint=users1000/client=gocloak$' nojsonv2         '^BenchmarkOfflineHTTP/endpoint=users1000/client=gocloak$' ''
    # the same path with resty reading the body vs SetDirectBodyRead
    scenario getusers1000-direct '^BenchmarkOfflineHTTP/endpoint=users1000/client=gocloak$' ''         '^BenchmarkOfflineHTTP/endpoint=users1000/client=gocloak-direct$' ''
    # 16 concurrent GetUserByID over TLS/HTTP1.1: resty's pool + v1 engine vs gocloak's pool + v2 engine
    scenario pool-tls-h1 '^BenchmarkOfflinePool/proto=tls-h1/client=gocloak-restypool$' nojsonv2         '^BenchmarkOfflinePool/proto=tls-h1/client=gocloak$' ''
}

run_live() {
    # client side of gocloak.GetUsers(max=1000) against the docker Keycloak
    scenario live-getusers1000 '^BenchmarkKeycloakHTTP/endpoint=users1000/client=gocloak$' nojsonv2         '^BenchmarkKeycloakHTTP/endpoint=users1000/client=gocloak-direct$' '' -tags keycloak
    scenario live-pool '^BenchmarkKeycloakPool/client=gocloak-restypool$' nojsonv2         '^BenchmarkKeycloakPool/client=gocloak$' '' -tags keycloak
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

echo "profiles: $OUT"
