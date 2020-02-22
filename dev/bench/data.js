window.BENCHMARK_DATA = {
  "lastUpdate": 1582409782747,
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
          "id": "91b2ddef620e855e54b6d52905d71aac1566d1dc",
          "message": "Merge pull request #142 from dlisin/master\n\nAdding UpdateRealm method",
          "timestamp": "2020-02-09T08:58:35-06:00",
          "tree_id": "9804554c7498eb72390663423b694a2d9cae561c",
          "url": "https://github.com/Nerzal/gocloak/commit/91b2ddef620e855e54b6d52905d71aac1566d1dc"
        },
        "date": 1581260445948,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 79392401,
            "unit": "ns/op\t   48281 B/op\t     182 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 81063390,
            "unit": "ns/op\t   64319 B/op\t     185 allocs/op",
            "extra": "14 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 79245780,
            "unit": "ns/op\t   53268 B/op\t     180 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 49443855,
            "unit": "ns/op\t   70101 B/op\t     182 allocs/op",
            "extra": "24 times\n2 procs"
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
          "id": "a7c121e2c634ae016c703e5db3e446f91f68a664",
          "message": "Merge pull request #134 from SVilgelm/remove-travis\n\nRemove .travis.yml",
          "timestamp": "2020-02-09T09:23:35-06:00",
          "tree_id": "8cd4ff89e819763537d30ed0222f9c3678c6224e",
          "url": "https://github.com/Nerzal/gocloak/commit/a7c121e2c634ae016c703e5db3e446f91f68a664"
        },
        "date": 1581261950289,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 91699270,
            "unit": "ns/op\t   56294 B/op\t     183 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 90316344,
            "unit": "ns/op\t   53527 B/op\t     183 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 90095496,
            "unit": "ns/op\t   59538 B/op\t     184 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 58104722,
            "unit": "ns/op\t   54507 B/op\t     182 allocs/op",
            "extra": "20 times\n2 procs"
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
          "id": "f2db20af7716766d0f485fa2b1f4324c86134db2",
          "message": "Merge pull request #143 from dlisin/master\n\nAdding GetDefaultGroups, AddDefaultGroup, RemoveDefaultGroup methods",
          "timestamp": "2020-02-10T14:34:01-06:00",
          "tree_id": "3a87b3ca2dc496bece9393755e062af16d9592e7",
          "url": "https://github.com/Nerzal/gocloak/commit/f2db20af7716766d0f485fa2b1f4324c86134db2"
        },
        "date": 1581366969317,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 76862817,
            "unit": "ns/op\t   56277 B/op\t     183 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 71876538,
            "unit": "ns/op\t   56373 B/op\t     184 allocs/op",
            "extra": "14 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 74087709,
            "unit": "ns/op\t   46450 B/op\t     183 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 49572834,
            "unit": "ns/op\t   53205 B/op\t     181 allocs/op",
            "extra": "25 times\n2 procs"
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
          "id": "2f4c873339d2b4d45711b8d24822f1071991826a",
          "message": "Merge pull request #145 from KosToZyB/master\n\nFixed example in README.md",
          "timestamp": "2020-02-18T06:26:31-06:00",
          "tree_id": "0b259db418de056b8a0ebb9af469cd1f78100483",
          "url": "https://github.com/Nerzal/gocloak/commit/2f4c873339d2b4d45711b8d24822f1071991826a"
        },
        "date": 1582028929415,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 91040582,
            "unit": "ns/op\t   64864 B/op\t     184 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 88488359,
            "unit": "ns/op\t   50697 B/op\t     184 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 100446244,
            "unit": "ns/op\t   47648 B/op\t     181 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 59484164,
            "unit": "ns/op\t   52703 B/op\t     181 allocs/op",
            "extra": "20 times\n2 procs"
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
          "id": "79707087a1359afe837c406057fcbdaec0e8c9b5",
          "message": "Update README.md",
          "timestamp": "2020-02-21T19:34:21+01:00",
          "tree_id": "1ad2d31424f6a10fde28ccb3f27eea6198add6e8",
          "url": "https://github.com/Nerzal/gocloak/commit/79707087a1359afe837c406057fcbdaec0e8c9b5"
        },
        "date": 1582310198804,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 92484789,
            "unit": "ns/op\t   67597 B/op\t     183 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 89792129,
            "unit": "ns/op\t   62064 B/op\t     184 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 92633694,
            "unit": "ns/op\t   40870 B/op\t     181 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 53552388,
            "unit": "ns/op\t   56674 B/op\t     185 allocs/op",
            "extra": "19 times\n2 procs"
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
          "id": "3ba9bbe12e819ffecf4217ac2318c76ac3f1bddd",
          "message": "Merge pull request #148 from SVilgelm/keycloak-9.0\n\nTesting keycloak:latest",
          "timestamp": "2020-02-21T12:41:05-06:00",
          "tree_id": "7de00c2a07c2f013a80ad81774b1b0c8a0ad0ef6",
          "url": "https://github.com/Nerzal/gocloak/commit/3ba9bbe12e819ffecf4217ac2318c76ac3f1bddd"
        },
        "date": 1582311951153,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 88890704,
            "unit": "ns/op\t   48763 B/op\t     204 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 85609134,
            "unit": "ns/op\t   70668 B/op\t     207 allocs/op",
            "extra": "14 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 87216441,
            "unit": "ns/op\t   74219 B/op\t     203 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 51417394,
            "unit": "ns/op\t   66676 B/op\t     204 allocs/op",
            "extra": "20 times\n2 procs"
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
          "id": "681aed0a300b73c261fc132437197679099fd1f7",
          "message": "Merge pull request #146 from Nerzal/error_refactoring\n\nError refactoring",
          "timestamp": "2020-02-22T16:14:11-06:00",
          "tree_id": "b369b3ec38bc34b50674c1b18153cd99d88d95c4",
          "url": "https://github.com/Nerzal/gocloak/commit/681aed0a300b73c261fc132437197679099fd1f7"
        },
        "date": 1582409782436,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 80197483,
            "unit": "ns/op\t   68673 B/op\t     205 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 81307817,
            "unit": "ns/op\t   65966 B/op\t     206 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 79944727,
            "unit": "ns/op\t   57192 B/op\t     203 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 49473322,
            "unit": "ns/op\t   64151 B/op\t     205 allocs/op",
            "extra": "22 times\n2 procs"
          }
        ]
      }
    ]
  }
}