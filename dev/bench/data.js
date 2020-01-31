window.BENCHMARK_DATA = {
  "lastUpdate": 1580493254489,
  "repoUrl": "https://github.com/Nerzal/gocloak",
  "entries": {
    "Benchmark": [
      {
        "commit": {
          "author": {
            "email": "theel.tobias@gmx.de",
            "name": "Nerzal",
            "username": "Nerzal"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "89e4ab8b0cee56e037c3d96136b4ce387c834527",
          "message": "Merge pull request #128 from SVilgelm/benchmark\n\nExtend ClearCache and add benchmarks for Login",
          "timestamp": "2020-01-14T16:42:53+01:00",
          "tree_id": "75cfda08dcb2aab5256dfdca005b1f06fb58fb75",
          "url": "https://github.com/Nerzal/gocloak/commit/89e4ab8b0cee56e037c3d96136b4ce387c834527"
        },
        "date": 1579016719734,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 85678248,
            "unit": "ns/op\t   50342 B/op\t     180 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 87302865,
            "unit": "ns/op\t   61936 B/op\t     183 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 86714791,
            "unit": "ns/op\t   47099 B/op\t     182 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 55681570,
            "unit": "ns/op\t   53137 B/op\t     181 allocs/op",
            "extra": "24 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "theel.tobias@gmx.de",
            "name": "Nerzal",
            "username": "Nerzal"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "3b9efb5845b3c6d9431e6650ee9dbfba5ee89fdc",
          "message": "Merge pull request #138 from tingeltangelthomas/master\n\nNew method UpdateClientProtocolMapper (Feature request #137)",
          "timestamp": "2020-01-31T18:48:03+01:00",
          "tree_id": "940d50be13e5ae16a914c9526d42a10616971794",
          "url": "https://github.com/Nerzal/gocloak/commit/3b9efb5845b3c6d9431e6650ee9dbfba5ee89fdc"
        },
        "date": 1580493254135,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 71791220,
            "unit": "ns/op\t   63588 B/op\t     183 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 72890306,
            "unit": "ns/op\t   56160 B/op\t     183 allocs/op",
            "extra": "18 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 70066112,
            "unit": "ns/op\t   51320 B/op\t     180 allocs/op",
            "extra": "16 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 43616771,
            "unit": "ns/op\t   53470 B/op\t     182 allocs/op",
            "extra": "27 times\n2 procs"
          }
        ]
      }
    ]
  }
}