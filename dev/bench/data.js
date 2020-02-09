window.BENCHMARK_DATA = {
  "lastUpdate": 1581208232816,
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
          "id": "bad3584f59b883a143d0999691b9496872557ff6",
          "message": "Merge pull request #141 from Nerzal/keycloak-8.0.1\n\nUsing keycloak 8.0.1 in tests",
          "timestamp": "2020-02-09T00:07:48+01:00",
          "tree_id": "08fcdbe4abd9fc9823bc830b1fdc8c38bfd3acb1",
          "url": "https://github.com/Nerzal/gocloak/commit/bad3584f59b883a143d0999691b9496872557ff6"
        },
        "date": 1581203439344,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 97703833,
            "unit": "ns/op\t   68760 B/op\t     185 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 102664844,
            "unit": "ns/op\t   43998 B/op\t     182 allocs/op",
            "extra": "12 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 97232808,
            "unit": "ns/op\t   74672 B/op\t     182 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 62370732,
            "unit": "ns/op\t   52516 B/op\t     184 allocs/op",
            "extra": "18 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "sergey@vilgelm.info",
            "name": "Sergey Vilgelm",
            "username": "SVilgelm"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "a587a05061a2a14bd67ec1e5232160ac707c1355",
          "message": "Merge pull request #140 from agorman/get-composite-client-roles\n\nAdding GetCompositeClientRolesByRoleID method",
          "timestamp": "2020-02-08T18:28:13-06:00",
          "tree_id": "51d2f771d9dbca19e06535674712e82a991d57da",
          "url": "https://github.com/Nerzal/gocloak/commit/a587a05061a2a14bd67ec1e5232160ac707c1355"
        },
        "date": 1581208232433,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 88050831,
            "unit": "ns/op\t   56115 B/op\t     182 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 91259825,
            "unit": "ns/op\t   64976 B/op\t     185 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 88056838,
            "unit": "ns/op\t   50678 B/op\t     183 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 56963482,
            "unit": "ns/op\t   58410 B/op\t     184 allocs/op",
            "extra": "20 times\n2 procs"
          }
        ]
      }
    ]
  }
}