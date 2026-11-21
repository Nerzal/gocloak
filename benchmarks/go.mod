module github.com/Nerzal/gocloak/v15/benchmarks

go 1.27.0

require (
	github.com/Nerzal/gocloak/v15 v15.0.0
	github.com/go-resty/resty/v2 v2.17.2
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/valyala/fasthttp v1.79.0
)

require (
	github.com/aclements/go-moremath v0.0.0-20210112150236-f10218a38794 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/klauspost/compress v1.20.4 // indirect
	github.com/molecule-man/go-brrr v1.3.1 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel v1.47.0 // indirect
	go.opentelemetry.io/otel/log v1.47.0 // indirect
	go.opentelemetry.io/otel/metric v1.47.0 // indirect
	go.opentelemetry.io/otel/trace v1.47.0 // indirect
	golang.org/x/net v0.63.0 // indirect
	golang.org/x/perf v0.0.0-20261117185222-4febdb082b2c // indirect
)

replace github.com/Nerzal/gocloak/v15 => ../

tool golang.org/x/perf/cmd/benchstat
