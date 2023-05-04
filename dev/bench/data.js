window.BENCHMARK_DATA = {
  "lastUpdate": 1683201041653,
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
          "id": "5bb5db7c033742f59702d7854c55ee18c3552adb",
          "message": "Merge pull request #150 from Nerzal/release-v5\n\nRelease v5",
          "timestamp": "2020-02-22T16:35:50-06:00",
          "tree_id": "4911fdaf47611c0a596a609fea40c272b030a0a2",
          "url": "https://github.com/Nerzal/gocloak/commit/5bb5db7c033742f59702d7854c55ee18c3552adb"
        },
        "date": 1582411080618,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 84164259,
            "unit": "ns/op\t   68827 B/op\t     207 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 86712339,
            "unit": "ns/op\t   65960 B/op\t     206 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 85199598,
            "unit": "ns/op\t   68652 B/op\t     205 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 54893892,
            "unit": "ns/op\t   53961 B/op\t     204 allocs/op",
            "extra": "21 times\n2 procs"
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
          "id": "0f1d3fac7fae2f24665f649ed83c1883c0238c9b",
          "message": "Merge pull request #157 from Nerzal/fix-get-policies\n\nFix GetPolicies function",
          "timestamp": "2020-03-16T10:25:18-05:00",
          "tree_id": "dfb2a8188686cffbd02711e4884ccf932a1fc378",
          "url": "https://github.com/Nerzal/gocloak/commit/0f1d3fac7fae2f24665f649ed83c1883c0238c9b"
        },
        "date": 1584372462544,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 92022613,
            "unit": "ns/op\t   60000 B/op\t     203 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 89719410,
            "unit": "ns/op\t   54428 B/op\t     204 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 92259120,
            "unit": "ns/op\t   57155 B/op\t     202 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 56946941,
            "unit": "ns/op\t   58739 B/op\t     202 allocs/op",
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
          "id": "07381de0fa4fb1728a8b3cd87be4c4886192e253",
          "message": "Merge pull request #162 from SVilgelm/error-description\n\nError description",
          "timestamp": "2020-04-01T09:19:52-05:00",
          "tree_id": "87bf864d48e13863e5e3545fd9b48d44a4c0b480",
          "url": "https://github.com/Nerzal/gocloak/commit/07381de0fa4fb1728a8b3cd87be4c4886192e253"
        },
        "date": 1585750960135,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 94430550,
            "unit": "ns/op\t   60008 B/op\t     203 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 92160402,
            "unit": "ns/op\t   51062 B/op\t     205 allocs/op",
            "extra": "12 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 92471493,
            "unit": "ns/op\t   41734 B/op\t     203 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 60957962,
            "unit": "ns/op\t   55280 B/op\t     204 allocs/op",
            "extra": "19 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "andreas.walter@easy.de",
            "name": "a.walter"
          },
          "committer": {
            "email": "sergey@vilgelm.info",
            "name": "Sergey Vilgelm",
            "username": "SVilgelm"
          },
          "distinct": true,
          "id": "ce76a8ca4d851b10c93eeb7567aada41f5bb4b8c",
          "message": "#160 added support for user federated identities",
          "timestamp": "2020-04-01T10:38:31-05:00",
          "tree_id": "af2dd47a2d7ac1badb0e35c1d6b2928d19c3efda",
          "url": "https://github.com/Nerzal/gocloak/commit/ce76a8ca4d851b10c93eeb7567aada41f5bb4b8c"
        },
        "date": 1585755871068,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 83219581,
            "unit": "ns/op\t   57091 B/op\t     205 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 84734943,
            "unit": "ns/op\t   65569 B/op\t     204 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 79864876,
            "unit": "ns/op\t   70256 B/op\t     205 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 50101016,
            "unit": "ns/op\t   55599 B/op\t     206 allocs/op",
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
          "id": "2cf3e2a21d4c8925802bf36d92936917382fa58a",
          "message": "Add files via upload",
          "timestamp": "2020-04-18T11:27:39+02:00",
          "tree_id": "68e9ef2ff9c27aaf4647e61f4ff9ace5776be0ef",
          "url": "https://github.com/Nerzal/gocloak/commit/2cf3e2a21d4c8925802bf36d92936917382fa58a"
        },
        "date": 1588194157570,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 79818653,
            "unit": "ns/op\t   54424 B/op\t     203 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 77921692,
            "unit": "ns/op\t   66749 B/op\t     204 allocs/op",
            "extra": "15 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 82656984,
            "unit": "ns/op\t   54183 B/op\t     202 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 56324564,
            "unit": "ns/op\t   58493 B/op\t     203 allocs/op",
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
          "id": "33311f5e5b1d62da4005dcc15026b17974ac7eef",
          "message": "Merge pull request #164 from FridaFino/master\n\nadd groups count",
          "timestamp": "2020-05-04T10:36:25-05:00",
          "tree_id": "660f956fbf479f951c38ff8980277da4e1d157ad",
          "url": "https://github.com/Nerzal/gocloak/commit/33311f5e5b1d62da4005dcc15026b17974ac7eef"
        },
        "date": 1588606743699,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 83575417,
            "unit": "ns/op\t   60522 B/op\t     213 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 83239007,
            "unit": "ns/op\t   47010 B/op\t     211 allocs/op",
            "extra": "14 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 83699756,
            "unit": "ns/op\t   47036 B/op\t     211 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 60295833,
            "unit": "ns/op\t   54420 B/op\t     212 allocs/op",
            "extra": "22 times\n2 procs"
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
          "id": "ee24c8acd55a4f30f8e6c8b14a0594563c24a016",
          "message": "Merge pull request #165 from SVilgelm/lint\n\nGolangCi enable comments validation",
          "timestamp": "2020-05-05T09:49:55+02:00",
          "tree_id": "d25200d8501441a7cd1fcbda1fa9bd3b361046ee",
          "url": "https://github.com/Nerzal/gocloak/commit/ee24c8acd55a4f30f8e6c8b14a0594563c24a016"
        },
        "date": 1588665152191,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 77652436,
            "unit": "ns/op\t   49740 B/op\t     212 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 78198613,
            "unit": "ns/op\t   52622 B/op\t     211 allocs/op",
            "extra": "15 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 79292694,
            "unit": "ns/op\t   47509 B/op\t     209 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 46831683,
            "unit": "ns/op\t   52561 B/op\t     210 allocs/op",
            "extra": "22 times\n2 procs"
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
          "id": "f451073fd2f5dc9504f8a68759b202866ec0e098",
          "message": "Merge pull request #167 from KosToZyB/master\n\nadd json tag string for structs *Params",
          "timestamp": "2020-05-18T06:26:12-05:00",
          "tree_id": "7630a240f65a3801c6a0ab0c6a1ec3edee099ad9",
          "url": "https://github.com/Nerzal/gocloak/commit/f451073fd2f5dc9504f8a68759b202866ec0e098"
        },
        "date": 1589801365963,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 85016839,
            "unit": "ns/op\t   57550 B/op\t     211 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 86546937,
            "unit": "ns/op\t   57701 B/op\t     213 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 84902390,
            "unit": "ns/op\t   69155 B/op\t     213 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 48666484,
            "unit": "ns/op\t   48902 B/op\t     211 allocs/op",
            "extra": "21 times\n2 procs"
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
          "id": "f2a58619f46d037ec4ebe8f932ce2c68fa0744d1",
          "message": "Merge pull request #169 from KosToZyB/master\n\nadd Required field to RoleDefinition",
          "timestamp": "2020-05-21T09:48:06-05:00",
          "tree_id": "9b7be083e7756d1dad9673079dfff26631253f8f",
          "url": "https://github.com/Nerzal/gocloak/commit/f2a58619f46d037ec4ebe8f932ce2c68fa0744d1"
        },
        "date": 1590072872621,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 92235406,
            "unit": "ns/op\t   48573 B/op\t     213 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 89967744,
            "unit": "ns/op\t   60680 B/op\t     214 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 89044007,
            "unit": "ns/op\t   68926 B/op\t     211 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 57404391,
            "unit": "ns/op\t   54117 B/op\t     210 allocs/op",
            "extra": "21 times\n2 procs"
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
          "id": "e3515becd74f46d5340642ec91a994e4bc9743e3",
          "message": "Merge pull request #171 from KosToZyB/GetAvailableClientRolesByUserID\n\nadd method GetAvailableClientRolesByUserID",
          "timestamp": "2020-05-21T12:03:36-05:00",
          "tree_id": "1aa195e19441fb3d73d2c15cc0776bd597128d78",
          "url": "https://github.com/Nerzal/gocloak/commit/e3515becd74f46d5340642ec91a994e4bc9743e3"
        },
        "date": 1590081107669,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 85616337,
            "unit": "ns/op\t   57574 B/op\t     211 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 90520445,
            "unit": "ns/op\t   49241 B/op\t     212 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 86719058,
            "unit": "ns/op\t   71917 B/op\t     212 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 51423309,
            "unit": "ns/op\t   56040 B/op\t     210 allocs/op",
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
          "id": "4dc0866dc8536dc9508afdb02d2d4c4d52740f74",
          "message": "Merge pull request #172 from SVilgelm/user-credentials\n\nImplement Credential APIs",
          "timestamp": "2020-06-02T07:55:35+02:00",
          "tree_id": "d24b4bd4dd87cb3c02e0db6ac3565b5ff3010ae9",
          "url": "https://github.com/Nerzal/gocloak/commit/4dc0866dc8536dc9508afdb02d2d4c4d52740f74"
        },
        "date": 1591077534252,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 90899638,
            "unit": "ns/op\t   57668 B/op\t     212 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 91300726,
            "unit": "ns/op\t   48497 B/op\t     211 allocs/op",
            "extra": "12 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 89511283,
            "unit": "ns/op\t   54612 B/op\t     211 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 50964992,
            "unit": "ns/op\t   61012 B/op\t     211 allocs/op",
            "extra": "22 times\n2 procs"
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
          "id": "ce6672c246759fd352631a5d43e6638c901e022d",
          "message": "Merge pull request #176 from agorman/user-group-count-search\n\nAdding search parameters when counting users and groups",
          "timestamp": "2020-06-20T12:42:07-05:00",
          "tree_id": "ca39a03dea86dc49591bcc4a1a4ddfbefc9e651a",
          "url": "https://github.com/Nerzal/gocloak/commit/ce6672c246759fd352631a5d43e6638c901e022d"
        },
        "date": 1592677209450,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 86985840,
            "unit": "ns/op\t   51977 B/op\t     212 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 85099214,
            "unit": "ns/op\t   54850 B/op\t     214 allocs/op",
            "extra": "12 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 85235714,
            "unit": "ns/op\t   64106 B/op\t     215 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 52724810,
            "unit": "ns/op\t   52481 B/op\t     211 allocs/op",
            "extra": "21 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "sergey@vilgelm.com",
            "name": "Sergey Vilgelm",
            "username": "SVilgelm"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "9070cf0ccf654c0866d1d924fd4a94a88546a323",
          "message": "Merge pull request #178 from viniciussousazup/feature/loginOtp\n\nFeature/login otp",
          "timestamp": "2020-06-22T07:38:48-05:00",
          "tree_id": "9363750fc76248f29a0602d402bc659c38bc375b",
          "url": "https://github.com/Nerzal/gocloak/commit/9070cf0ccf654c0866d1d924fd4a94a88546a323"
        },
        "date": 1592829848351,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 70502725,
            "unit": "ns/op\t   82782 B/op\t     211 allocs/op",
            "extra": "16 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 71421572,
            "unit": "ns/op\t   69904 B/op\t     213 allocs/op",
            "extra": "15 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 70661424,
            "unit": "ns/op\t   64838 B/op\t     211 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 45018322,
            "unit": "ns/op\t   57500 B/op\t     210 allocs/op",
            "extra": "28 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "sergey@vilgelm.com",
            "name": "Sergey Vilgelm",
            "username": "SVilgelm"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "5bdb05fcb34bf8b3e3e6340dc64e9e027e22fdcc",
          "message": "Merge pull request #180 from alexsn/context\n\nAdding context.Context support",
          "timestamp": "2020-06-22T10:36:33-05:00",
          "tree_id": "6b39e8ecd548694f5626e64bc80fb4fdb7806eda",
          "url": "https://github.com/Nerzal/gocloak/commit/5bdb05fcb34bf8b3e3e6340dc64e9e027e22fdcc"
        },
        "date": 1592840337900,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 81632787,
            "unit": "ns/op\t   52528 B/op\t     212 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 82227524,
            "unit": "ns/op\t   68613 B/op\t     215 allocs/op",
            "extra": "14 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 81885336,
            "unit": "ns/op\t   46571 B/op\t     212 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 51653694,
            "unit": "ns/op\t   58115 B/op\t     214 allocs/op",
            "extra": "22 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "sergey@vilgelm.com",
            "name": "Sergey Vilgelm",
            "username": "SVilgelm"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "fb884346883ef5f07a48c198e3a213287dbf54a9",
          "message": "Merge pull request #182 from Nerzal/fix-174\n\nUsing pointers for array types",
          "timestamp": "2020-06-25T07:45:35-05:00",
          "tree_id": "9a9bebaa638728ed4a1edf5c99ae033c709c968c",
          "url": "https://github.com/Nerzal/gocloak/commit/fb884346883ef5f07a48c198e3a213287dbf54a9"
        },
        "date": 1593089261679,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 88615229,
            "unit": "ns/op\t   66641 B/op\t     216 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 85411129,
            "unit": "ns/op\t   58199 B/op\t     216 allocs/op",
            "extra": "12 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 87689990,
            "unit": "ns/op\t   55119 B/op\t     213 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 59373849,
            "unit": "ns/op\t   56514 B/op\t     214 allocs/op",
            "extra": "24 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "sergey@vilgelm.com",
            "name": "Sergey Vilgelm",
            "username": "SVilgelm"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "f29422bed50c086f9a59baacb2cc6b1916fa9cc1",
          "message": "Merge pull request #183 from Nerzal/fix-173\n\nExtend UserInfo according to the specification",
          "timestamp": "2020-06-27T14:18:45-05:00",
          "tree_id": "56998bc0f1a30044191e86102c9388aad0717fa9",
          "url": "https://github.com/Nerzal/gocloak/commit/f29422bed50c086f9a59baacb2cc6b1916fa9cc1"
        },
        "date": 1593285645053,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 84744363,
            "unit": "ns/op\t   57861 B/op\t     213 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 84482570,
            "unit": "ns/op\t   49560 B/op\t     214 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 84285301,
            "unit": "ns/op\t   52212 B/op\t     212 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 56216535,
            "unit": "ns/op\t   64153 B/op\t     214 allocs/op",
            "extra": "24 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "sergey@vilgelm.com",
            "name": "Sergey Vilgelm",
            "username": "SVilgelm"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "40a37bb5044ea7aea74a1c939e64abe5a1738b7a",
          "message": "Merge pull request #184 from Nerzal/fix-163\n\nSignedJWT",
          "timestamp": "2020-06-28T11:22:12-05:00",
          "tree_id": "6a3f4c087dcd38d77a3fce66ab1af881185c4433",
          "url": "https://github.com/Nerzal/gocloak/commit/40a37bb5044ea7aea74a1c939e64abe5a1738b7a"
        },
        "date": 1593362326569,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 70600260,
            "unit": "ns/op\t   71521 B/op\t     212 allocs/op",
            "extra": "16 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 72209856,
            "unit": "ns/op\t   63353 B/op\t     215 allocs/op",
            "extra": "14 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 72285470,
            "unit": "ns/op\t   68566 B/op\t     215 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 47833318,
            "unit": "ns/op\t   55196 B/op\t     213 allocs/op",
            "extra": "27 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "sergey@vilgelm.com",
            "name": "Sergey Vilgelm",
            "username": "SVilgelm"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "323bb5bbbcbb7722d3c7992c9006dce9372c1227",
          "message": "Merge pull request #186 from ydstaana/Logout\n\nLogout all sessions for a user",
          "timestamp": "2020-06-30T08:44:16-05:00",
          "tree_id": "9c7b5bb18aa60f0378502ff53a38a87ac4f6ffb1",
          "url": "https://github.com/Nerzal/gocloak/commit/323bb5bbbcbb7722d3c7992c9006dce9372c1227"
        },
        "date": 1593524786056,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 88559370,
            "unit": "ns/op\t   67618 B/op\t     218 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 91078909,
            "unit": "ns/op\t   66844 B/op\t     218 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 86339632,
            "unit": "ns/op\t   49494 B/op\t     213 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 52674999,
            "unit": "ns/op\t   56364 B/op\t     214 allocs/op",
            "extra": "21 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "sergey@vilgelm.com",
            "name": "Sergey Vilgelm",
            "username": "SVilgelm"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "735c1c5be863c89dc416debc6e09377c88c2dd04",
          "message": "Merge pull request #188 from Nerzal/fix-187\n\nReplace *interface{} to interface{} in models",
          "timestamp": "2020-07-04T11:09:06-05:00",
          "tree_id": "98b28b4c0eb6361828e6a871b9741af95a3b1978",
          "url": "https://github.com/Nerzal/gocloak/commit/735c1c5be863c89dc416debc6e09377c88c2dd04"
        },
        "date": 1593879088728,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 100325339,
            "unit": "ns/op\t   64520 B/op\t     217 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 99477637,
            "unit": "ns/op\t   54976 B/op\t     214 allocs/op",
            "extra": "12 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 101969968,
            "unit": "ns/op\t   47135 B/op\t     214 allocs/op",
            "extra": "10 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 67621549,
            "unit": "ns/op\t   66289 B/op\t     220 allocs/op",
            "extra": "15 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "sergey@vilgelm.com",
            "name": "Sergey Vilgelm",
            "username": "SVilgelm"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "5853888e047530f39cb1923c66cee8c5fe765bee",
          "message": "Merge pull request #189 from agorman/composite-role-represenation\n\nAdding CompositesRepresentation struct",
          "timestamp": "2020-07-05T05:47:14-05:00",
          "tree_id": "9117a00ab9e2f050a934301ed215056c424e2c0c",
          "url": "https://github.com/Nerzal/gocloak/commit/5853888e047530f39cb1923c66cee8c5fe765bee"
        },
        "date": 1593946169422,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 66005772,
            "unit": "ns/op\t   62410 B/op\t     213 allocs/op",
            "extra": "16 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 69103346,
            "unit": "ns/op\t   78764 B/op\t     216 allocs/op",
            "extra": "16 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 66524417,
            "unit": "ns/op\t   71787 B/op\t     215 allocs/op",
            "extra": "16 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 38590462,
            "unit": "ns/op\t   56707 B/op\t     214 allocs/op",
            "extra": "28 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "sergey@vilgelm.com",
            "name": "Sergey Vilgelm",
            "username": "SVilgelm"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "c0de0212f8f48f8e1f08fa047d876ee2760100b2",
          "message": "Merge pull request #190 from Nerzal/omitempty\n\nAdd omitempty flag to all fields",
          "timestamp": "2020-07-06T11:37:04-05:00",
          "tree_id": "468f0aba6476d8dbef0c433801fec1f3460502ee",
          "url": "https://github.com/Nerzal/gocloak/commit/c0de0212f8f48f8e1f08fa047d876ee2760100b2"
        },
        "date": 1594053566185,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 72690532,
            "unit": "ns/op\t   80074 B/op\t     215 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 72921485,
            "unit": "ns/op\t   57978 B/op\t     214 allocs/op",
            "extra": "15 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 69963790,
            "unit": "ns/op\t   55565 B/op\t     215 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 47641510,
            "unit": "ns/op\t   66402 B/op\t     213 allocs/op",
            "extra": "30 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "sergey@vilgelm.com",
            "name": "Sergey Vilgelm",
            "username": "SVilgelm"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "1896b075de784346ffa6d7b56d737d489ba72efe",
          "message": "Merge pull request #194 from eduardev/feature/logout-single-session\n\nimplement DELETE /{realm}/sessions/{session}",
          "timestamp": "2020-07-22T21:00:24-05:00",
          "tree_id": "245b038d0099e175a5b980a8ef22ea0d774e51d1",
          "url": "https://github.com/Nerzal/gocloak/commit/1896b075de784346ffa6d7b56d737d489ba72efe"
        },
        "date": 1595469809593,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 99162464,
            "unit": "ns/op\t   58191 B/op\t     216 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 92804483,
            "unit": "ns/op\t   52508 B/op\t     215 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 98346210,
            "unit": "ns/op\t   73558 B/op\t     216 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 54681238,
            "unit": "ns/op\t   61620 B/op\t     216 allocs/op",
            "extra": "22 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "sergey@vilgelm.com",
            "name": "Sergey Vilgelm",
            "username": "SVilgelm"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "1561f183aae1210975ba917e4f6afd1f698d8edb",
          "message": "Merge pull request #195 from VoIdemar/feature/fields-for-x509-certs\n\nAdd fields for X5C/X509 public key certificates",
          "timestamp": "2020-07-24T12:35:04-05:00",
          "tree_id": "08509173a3dcb8cac6b6034d8311c41d64e4d397",
          "url": "https://github.com/Nerzal/gocloak/commit/1561f183aae1210975ba917e4f6afd1f698d8edb"
        },
        "date": 1595612253842,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 85192774,
            "unit": "ns/op\t   64422 B/op\t     217 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 83116112,
            "unit": "ns/op\t   63920 B/op\t     217 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 80676597,
            "unit": "ns/op\t   58066 B/op\t     215 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 51441863,
            "unit": "ns/op\t   46463 B/op\t     211 allocs/op",
            "extra": "26 times\n2 procs"
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
          "id": "9248a1d541dad9e1af77ebb5984559b4ff544290",
          "message": "Merge pull request #196 from Nerzal/upgrade_jwt\n\nsupport audience array & strings",
          "timestamp": "2020-07-28T17:44:09+02:00",
          "tree_id": "584f945e45ef423db6529fe11841c99a5a2f2d71",
          "url": "https://github.com/Nerzal/gocloak/commit/9248a1d541dad9e1af77ebb5984559b4ff544290"
        },
        "date": 1595951172463,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 86699915,
            "unit": "ns/op\t   51931 B/op\t     214 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 85545788,
            "unit": "ns/op\t   74991 B/op\t     214 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 84884655,
            "unit": "ns/op\t   49587 B/op\t     214 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 53424326,
            "unit": "ns/op\t   73369 B/op\t     216 allocs/op",
            "extra": "22 times\n2 procs"
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
          "id": "6839a9297e492368255d4c9dac7a32ac284b013a",
          "message": "Update README.md",
          "timestamp": "2020-07-28T17:45:36+02:00",
          "tree_id": "ee7da9ff32b44a0829f4a71cb91dff793f33023a",
          "url": "https://github.com/Nerzal/gocloak/commit/6839a9297e492368255d4c9dac7a32ac284b013a"
        },
        "date": 1595951271219,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 67715696,
            "unit": "ns/op\t   44067 B/op\t     214 allocs/op",
            "extra": "16 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 68533328,
            "unit": "ns/op\t   60265 B/op\t     215 allocs/op",
            "extra": "16 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 69862522,
            "unit": "ns/op\t   55419 B/op\t     213 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 39466509,
            "unit": "ns/op\t   57837 B/op\t     212 allocs/op",
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
          "id": "51ced33ef0a2b77c8f53ea833cdd5c70ad907f30",
          "message": "Update README.md",
          "timestamp": "2020-07-28T17:46:10+02:00",
          "tree_id": "9391d3c52bb5415a6b949180aac5a7a8bf7eabf0",
          "url": "https://github.com/Nerzal/gocloak/commit/51ced33ef0a2b77c8f53ea833cdd5c70ad907f30"
        },
        "date": 1595951356185,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 88420458,
            "unit": "ns/op\t   54987 B/op\t     214 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 90153160,
            "unit": "ns/op\t   61535 B/op\t     219 allocs/op",
            "extra": "12 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 90011447,
            "unit": "ns/op\t   60891 B/op\t     214 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 53058810,
            "unit": "ns/op\t   58208 B/op\t     215 allocs/op",
            "extra": "20 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "tt@fino.digital",
            "name": "Tobias Theel"
          },
          "committer": {
            "email": "tt@fino.digital",
            "name": "Tobias Theel"
          },
          "distinct": true,
          "id": "3654444a6d6c4bd339eb3f66341545de9c049431",
          "message": "update import paths to v7",
          "timestamp": "2020-07-28T17:49:49+02:00",
          "tree_id": "2c1313a106bf29d18ae12128b48e86ebf162857a",
          "url": "https://github.com/Nerzal/gocloak/commit/3654444a6d6c4bd339eb3f66341545de9c049431"
        },
        "date": 1595996581605,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 83889631,
            "unit": "ns/op\t   60869 B/op\t     215 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 83816796,
            "unit": "ns/op\t   75324 B/op\t     218 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 84790371,
            "unit": "ns/op\t   61103 B/op\t     217 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 52728481,
            "unit": "ns/op\t   56456 B/op\t     213 allocs/op",
            "extra": "24 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "sergey@vilgelm.com",
            "name": "Sergey Vilgelm",
            "username": "SVilgelm"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "70f32cc8289581a8157d6c6d0e4c7ba81b453eef",
          "message": "Merge pull request #197 from agorman/available-client-roles-by-group\n\nAdding GetAvailableClientRolesByGroupID method",
          "timestamp": "2020-08-09T12:45:49-05:00",
          "tree_id": "e1139f1c3120f9f57727cefe09e531f77965d6e9",
          "url": "https://github.com/Nerzal/gocloak/commit/70f32cc8289581a8157d6c6d0e4c7ba81b453eef"
        },
        "date": 1596995293368,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 87265736,
            "unit": "ns/op\t   58255 B/op\t     217 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 86276760,
            "unit": "ns/op\t   64532 B/op\t     218 allocs/op",
            "extra": "12 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 84804949,
            "unit": "ns/op\t   66753 B/op\t     217 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 50567425,
            "unit": "ns/op\t   56317 B/op\t     213 allocs/op",
            "extra": "21 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "sergey@vilgelm.com",
            "name": "Sergey Vilgelm",
            "username": "SVilgelm"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "a0d0d0692d339e2128cd6ad4decda698987a24b1",
          "message": "Merge pull request #199 from agorman/add-group-params-get-user-groups\n\nAdding GetGroupParams to GetUserGroups method",
          "timestamp": "2020-08-09T12:46:20-05:00",
          "tree_id": "3819abfcca029a830bbb501eb376a0d4db6447d7",
          "url": "https://github.com/Nerzal/gocloak/commit/a0d0d0692d339e2128cd6ad4decda698987a24b1"
        },
        "date": 1596995321224,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 81113124,
            "unit": "ns/op\t   55182 B/op\t     215 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 81198570,
            "unit": "ns/op\t   72879 B/op\t     217 allocs/op",
            "extra": "15 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 80454970,
            "unit": "ns/op\t   65848 B/op\t     214 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 52830801,
            "unit": "ns/op\t   48787 B/op\t     213 allocs/op",
            "extra": "24 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "sergey@vilgelm.com",
            "name": "Sergey Vilgelm",
            "username": "SVilgelm"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "34d9e0244f6bcd14cc3b4d26b684ac44342f2d8e",
          "message": "Merge pull request #198 from life1347/return-original-error-message\n\nLet checkForError return original error message",
          "timestamp": "2020-08-09T12:47:04-05:00",
          "tree_id": "3d1c1384e7cb0df11d2193a09746c90a894a3510",
          "url": "https://github.com/Nerzal/gocloak/commit/34d9e0244f6bcd14cc3b4d26b684ac44342f2d8e"
        },
        "date": 1596995362668,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 73564681,
            "unit": "ns/op\t   62878 B/op\t     215 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 72445126,
            "unit": "ns/op\t   55541 B/op\t     214 allocs/op",
            "extra": "15 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 72669394,
            "unit": "ns/op\t   50442 B/op\t     212 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 45430297,
            "unit": "ns/op\t   64437 B/op\t     213 allocs/op",
            "extra": "28 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "sergey@vilgelm.com",
            "name": "Sergey Vilgelm",
            "username": "SVilgelm"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "136f02c1ede1caa8e69d2772bd2ac220b5a0e2a7",
          "message": "Merge pull request #192 from Nerzal/alternative_urls\n\nAdd functional options",
          "timestamp": "2020-08-09T12:48:19-05:00",
          "tree_id": "2c8841013e2b0bbdc7e3903caa351ff10a42a25d",
          "url": "https://github.com/Nerzal/gocloak/commit/136f02c1ede1caa8e69d2772bd2ac220b5a0e2a7"
        },
        "date": 1596995405709,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 62481675,
            "unit": "ns/op\t   53745 B/op\t     214 allocs/op",
            "extra": "18 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 62151241,
            "unit": "ns/op\t   59666 B/op\t     213 allocs/op",
            "extra": "19 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 61843728,
            "unit": "ns/op\t   59551 B/op\t     210 allocs/op",
            "extra": "18 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 35627065,
            "unit": "ns/op\t   50967 B/op\t     211 allocs/op",
            "extra": "33 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "sergey@vilgelm.com",
            "name": "Sergey Vilgelm",
            "username": "SVilgelm"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "7c13d06c0f3acd8b99b817ca8cc06dc5568bcd90",
          "message": "Merge pull request #202 from Nerzal/update-readme\n\nAdd installation section",
          "timestamp": "2020-08-12T09:46:05-05:00",
          "tree_id": "f2f4b7a80c16dbc8dd2e32e944d296002e3f543c",
          "url": "https://github.com/Nerzal/gocloak/commit/7c13d06c0f3acd8b99b817ca8cc06dc5568bcd90"
        },
        "date": 1597249828677,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 85675803,
            "unit": "ns/op\t   67510 B/op\t     217 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 81555593,
            "unit": "ns/op\t   50226 B/op\t     215 allocs/op",
            "extra": "14 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 81110882,
            "unit": "ns/op\t   63258 B/op\t     215 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 50449638,
            "unit": "ns/op\t   62501 B/op\t     214 allocs/op",
            "extra": "25 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "sergey@vilgelm.com",
            "name": "Sergey Vilgelm",
            "username": "SVilgelm"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "e78566240578d8ed250f90c96a6ee45dd6d3211a",
          "message": "Merge pull request #206 from agorman/fix-get-gealm-roles-by-group-id\n\nAdding missing SetResult to GetRealmRolesByGroupID method",
          "timestamp": "2020-08-18T16:13:18-05:00",
          "tree_id": "0b74421e60f78227609bd6b999ca2dac05a3f53d",
          "url": "https://github.com/Nerzal/gocloak/commit/e78566240578d8ed250f90c96a6ee45dd6d3211a"
        },
        "date": 1597785342054,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 84729397,
            "unit": "ns/op\t   44597 B/op\t     211 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 90752195,
            "unit": "ns/op\t   76672 B/op\t     217 allocs/op",
            "extra": "12 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 91440588,
            "unit": "ns/op\t   60686 B/op\t     215 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 52148684,
            "unit": "ns/op\t   57887 B/op\t     212 allocs/op",
            "extra": "24 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "sergey@vilgelm.com",
            "name": "Sergey Vilgelm",
            "username": "SVilgelm"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "844a1d56d64fd05782b1d1fc6cdc35e071b33541",
          "message": "Merge pull request #207 from agorman/realm-role-methods\n\nAdding multiple methods for dealing with realm roles",
          "timestamp": "2020-08-18T16:29:31-05:00",
          "tree_id": "1104b95c26ff27ca2d2df5fff418f645d6937cb8",
          "url": "https://github.com/Nerzal/gocloak/commit/844a1d56d64fd05782b1d1fc6cdc35e071b33541"
        },
        "date": 1597786303887,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 91327226,
            "unit": "ns/op\t   58234 B/op\t     216 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 91630311,
            "unit": "ns/op\t   58375 B/op\t     217 allocs/op",
            "extra": "12 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 97236915,
            "unit": "ns/op\t   70515 B/op\t     216 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 56620881,
            "unit": "ns/op\t   58155 B/op\t     214 allocs/op",
            "extra": "21 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "sergey@vilgelm.com",
            "name": "Sergey Vilgelm",
            "username": "SVilgelm"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "1a6fb6749cc6a6d2d52ffb8780c955d0689896b4",
          "message": "Merge pull request #209 from madest92/scope-mappings\n\nScope mappings",
          "timestamp": "2020-08-21T11:55:02-05:00",
          "tree_id": "61331aafc74862058e2e0e711555d7a10f16f6c6",
          "url": "https://github.com/Nerzal/gocloak/commit/1a6fb6749cc6a6d2d52ffb8780c955d0689896b4"
        },
        "date": 1598029039481,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 92883134,
            "unit": "ns/op\t   63830 B/op\t     218 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 85513650,
            "unit": "ns/op\t   79083 B/op\t     223 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 83223051,
            "unit": "ns/op\t   63856 B/op\t     219 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 58747091,
            "unit": "ns/op\t   66763 B/op\t     218 allocs/op",
            "extra": "25 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "sergey@vilgelm.com",
            "name": "Sergey Vilgelm",
            "username": "SVilgelm"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "862bcfb7e12940b1db32707658d5d58c6e34d757",
          "message": "Merge pull request #205 from goolanger/master\n\nAdd groups to user model",
          "timestamp": "2020-08-21T12:04:41-05:00",
          "tree_id": "b4e8a0ddb4538ff20192d3d03245a0599eca6093",
          "url": "https://github.com/Nerzal/gocloak/commit/862bcfb7e12940b1db32707658d5d58c6e34d757"
        },
        "date": 1598035353123,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 67528573,
            "unit": "ns/op\t   54873 B/op\t     218 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 70919855,
            "unit": "ns/op\t   64406 B/op\t     220 allocs/op",
            "extra": "16 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 68157927,
            "unit": "ns/op\t   54523 B/op\t     218 allocs/op",
            "extra": "16 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 41785676,
            "unit": "ns/op\t   58546 B/op\t     218 allocs/op",
            "extra": "31 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "sergey@vilgelm.com",
            "name": "Sergey Vilgelm",
            "username": "SVilgelm"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "4cbc389eb3c8037c7a7b993a0fbaea8eebb6bf6d",
          "message": "Merge pull request #211 from toddkazakov/master\n\nAdd support for enabled and exact query params in list users query",
          "timestamp": "2020-08-25T12:21:17-05:00",
          "tree_id": "6608ee07af7c6a99a70f07bcacd6b1d3caad3099",
          "url": "https://github.com/Nerzal/gocloak/commit/4cbc389eb3c8037c7a7b993a0fbaea8eebb6bf6d"
        },
        "date": 1598376219471,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 88774686,
            "unit": "ns/op\t   68991 B/op\t     220 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 87413523,
            "unit": "ns/op\t   84793 B/op\t     224 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 87811764,
            "unit": "ns/op\t   48139 B/op\t     219 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 60874134,
            "unit": "ns/op\t   64386 B/op\t     219 allocs/op",
            "extra": "24 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "sergey@vilgelm.com",
            "name": "Sergey Vilgelm",
            "username": "SVilgelm"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "1f024b90dfeb18b1a22d2cbb3a48106c343cd2a5",
          "message": "Merge pull request #210 from Nerzal/update/modules\n\nupgrade to go 1.15 and update some dependencies",
          "timestamp": "2020-08-25T12:21:53-05:00",
          "tree_id": "3fb0b4557e5de09ccafea0b6f50704025e0170a7",
          "url": "https://github.com/Nerzal/gocloak/commit/1f024b90dfeb18b1a22d2cbb3a48106c343cd2a5"
        },
        "date": 1598376239155,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 64279526,
            "unit": "ns/op\t   63390 B/op\t     217 allocs/op",
            "extra": "18 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 63109872,
            "unit": "ns/op\t   65477 B/op\t     217 allocs/op",
            "extra": "16 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 63863832,
            "unit": "ns/op\t   57286 B/op\t     217 allocs/op",
            "extra": "18 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 44424948,
            "unit": "ns/op\t   47670 B/op\t     217 allocs/op",
            "extra": "32 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "sergey@vilgelm.com",
            "name": "Sergey Vilgelm",
            "username": "SVilgelm"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "ccaf08026370c16a107e01df201486534c4c5720",
          "message": "Write go.list file for nancy (#218)",
          "timestamp": "2020-09-08T08:05:06-05:00",
          "tree_id": "2f940b7f3bc127aec367383d1671f62e50636972",
          "url": "https://github.com/Nerzal/gocloak/commit/ccaf08026370c16a107e01df201486534c4c5720"
        },
        "date": 1599570436618,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 80078527,
            "unit": "ns/op\t   63195 B/op\t     218 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 78423425,
            "unit": "ns/op\t   68610 B/op\t     217 allocs/op",
            "extra": "14 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 80162711,
            "unit": "ns/op\t   83449 B/op\t     220 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 46477398,
            "unit": "ns/op\t   64129 B/op\t     220 allocs/op",
            "extra": "24 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "todd.kazakov@gmail.com",
            "name": "Todd Kazakov",
            "username": "toddkazakov"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "daf5d72801c591472fa818d9e99b4ba86e311928",
          "message": "Serialize enabled and exact query params as strings (#220)",
          "timestamp": "2020-09-09T10:30:04-05:00",
          "tree_id": "5861274a85b87f78858b135dd05808ad9fb92285",
          "url": "https://github.com/Nerzal/gocloak/commit/daf5d72801c591472fa818d9e99b4ba86e311928"
        },
        "date": 1599665555879,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 81834088,
            "unit": "ns/op\t   60366 B/op\t     217 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 84653804,
            "unit": "ns/op\t   62094 B/op\t     219 allocs/op",
            "extra": "14 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 80838637,
            "unit": "ns/op\t   58374 B/op\t     217 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 55811400,
            "unit": "ns/op\t   56817 B/op\t     220 allocs/op",
            "extra": "21 times\n2 procs"
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
          "id": "38400d49944a23a7a4dc11e42266fc52096f76fa",
          "message": "Merge pull request #225 from sachinjohar2/sachin/PermissionEntities\n\nAPIs to fetch resource and scopes associated with permission",
          "timestamp": "2020-11-07T18:29:43+01:00",
          "tree_id": "100ec4ac9acb77ff55b5b8932f5d2cf5f782bcb3",
          "url": "https://github.com/Nerzal/gocloak/commit/38400d49944a23a7a4dc11e42266fc52096f76fa"
        },
        "date": 1604770319140,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 85754626,
            "unit": "ns/op\t   47404 B/op\t     217 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 95508583,
            "unit": "ns/op\t   59157 B/op\t     218 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 83871794,
            "unit": "ns/op\t   63267 B/op\t     219 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 60539164,
            "unit": "ns/op\t   53983 B/op\t     217 allocs/op",
            "extra": "22 times\n2 procs"
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
          "id": "eb2ff7d1627794b3860d2f01944ee3a7b649042d",
          "message": "Merge pull request #231 from sachinjohar2/sachin/RoleById\n\nAdded api to fetch role using role ID",
          "timestamp": "2020-11-24T14:38:35+01:00",
          "tree_id": "6a76c024239633a6b67c1b6cb2f81f8adcf0b353",
          "url": "https://github.com/Nerzal/gocloak/commit/eb2ff7d1627794b3860d2f01944ee3a7b649042d"
        },
        "date": 1606225258747,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 78726327,
            "unit": "ns/op\t   51653 B/op\t     217 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 79803296,
            "unit": "ns/op\t   57257 B/op\t     218 allocs/op",
            "extra": "14 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 77741298,
            "unit": "ns/op\t   57685 B/op\t     220 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 47150547,
            "unit": "ns/op\t   48924 B/op\t     216 allocs/op",
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
          "id": "ce3567f439e6eb128ad61693b92f998b02ab0633",
          "message": "Merge pull request #230 from timdrysdale/clientpat\n\nAdd client methods for Get, Update resource",
          "timestamp": "2020-11-26T11:41:34+01:00",
          "tree_id": "ac0edfc0ab367054ffc3aa0bfb6fbbc2306cd609",
          "url": "https://github.com/Nerzal/gocloak/commit/ce3567f439e6eb128ad61693b92f998b02ab0633"
        },
        "date": 1606387423323,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 84753534,
            "unit": "ns/op\t   61406 B/op\t     219 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 84499617,
            "unit": "ns/op\t   55154 B/op\t     220 allocs/op",
            "extra": "14 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 82312248,
            "unit": "ns/op\t   65954 B/op\t     217 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 51249074,
            "unit": "ns/op\t   51157 B/op\t     218 allocs/op",
            "extra": "21 times\n2 procs"
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
          "id": "4ce82f58b38c36ca0f8704c8f6384a975b1fd6f6",
          "message": "Merge pull request #240 from timdrysdale/master\n\nUpdate interface methods in README.md",
          "timestamp": "2020-12-09T12:42:25+01:00",
          "tree_id": "2abbc1b25e3199a511857a392a596ed43da5b23b",
          "url": "https://github.com/Nerzal/gocloak/commit/4ce82f58b38c36ca0f8704c8f6384a975b1fd6f6"
        },
        "date": 1607514278758,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 85245864,
            "unit": "ns/op\t   46860 B/op\t     217 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 85268903,
            "unit": "ns/op\t   61333 B/op\t     221 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 80008234,
            "unit": "ns/op\t   49861 B/op\t     217 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 55711014,
            "unit": "ns/op\t   63735 B/op\t     216 allocs/op",
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
          "id": "c466757115d4eb939647e5259688074850f45f76",
          "message": "Merge pull request #239 from jpughcs/master\n\nAdd MatchingURI to GetResourceParams model",
          "timestamp": "2020-12-09T20:48:58+01:00",
          "tree_id": "5647fc218136b8008badd62e070f11f8dc574db8",
          "url": "https://github.com/Nerzal/gocloak/commit/c466757115d4eb939647e5259688074850f45f76"
        },
        "date": 1607543497561,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 73728175,
            "unit": "ns/op\t   60340 B/op\t     217 allocs/op",
            "extra": "16 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 70529372,
            "unit": "ns/op\t   63980 B/op\t     217 allocs/op",
            "extra": "16 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 69828839,
            "unit": "ns/op\t   54392 B/op\t     217 allocs/op",
            "extra": "16 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 38933662,
            "unit": "ns/op\t   71879 B/op\t     218 allocs/op",
            "extra": "28 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "timothy.d.drysdale@gmail.com",
            "name": "Tim Drysdale",
            "username": "timdrysdale"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "6874eb3a7accf3bb8553ce32816fffbe2cb8f26f",
          "message": "Add missing interface methods to README.md (#241)\n\nThis is a continuation of issue #223 to add missing methods to\r\ninterface in README.md and remove one stale method, now that\r\ngithub.com/timdrysdale/ifcmp updated to report these.\r\n\r\n[Issue: 223]",
          "timestamp": "2020-12-09T15:03:07-06:00",
          "tree_id": "b6d386e76b76668767bd3e58cdf0892bf5c1aacf",
          "url": "https://github.com/Nerzal/gocloak/commit/6874eb3a7accf3bb8553ce32816fffbe2cb8f26f"
        },
        "date": 1607547916086,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 91530936,
            "unit": "ns/op\t   72488 B/op\t     222 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 92491522,
            "unit": "ns/op\t   59802 B/op\t     218 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 88291366,
            "unit": "ns/op\t   55084 B/op\t     218 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 54861618,
            "unit": "ns/op\t   59672 B/op\t     218 allocs/op",
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
          "id": "0bd22a70abd689ac2c36e3815dca45c880fc9ec9",
          "message": "Merge pull request #259 from 2785/fix-tests\n\nWIP: Fix client token related test failure",
          "timestamp": "2021-03-22T13:18:51+01:00",
          "tree_id": "084e02d4718c4ba90506b4ec0998e06014a6e46d",
          "url": "https://github.com/Nerzal/gocloak/commit/0bd22a70abd689ac2c36e3815dca45c880fc9ec9"
        },
        "date": 1616415682191,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 87510306,
            "unit": "ns/op\t   74230 B/op\t     218 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 88642658,
            "unit": "ns/op\t   75974 B/op\t     219 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 87175221,
            "unit": "ns/op\t   60174 B/op\t     217 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 58754202,
            "unit": "ns/op\t   51932 B/op\t     218 allocs/op",
            "extra": "19 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "tt@fino.digital",
            "name": "Tobias Theel"
          },
          "committer": {
            "email": "tt@fino.digital",
            "name": "Tobias Theel"
          },
          "distinct": true,
          "id": "0bdb1134b12213545f53c414192afd2b85f703d0",
          "message": "fix github action syntax error",
          "timestamp": "2021-09-02T14:07:05+02:00",
          "tree_id": "cda50f96a6642234894f7021bd567a4431be1427",
          "url": "https://github.com/Nerzal/gocloak/commit/0bdb1134b12213545f53c414192afd2b85f703d0"
        },
        "date": 1630584570168,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 71286667,
            "unit": "ns/op\t   72449 B/op\t     218 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 71802680,
            "unit": "ns/op\t   62883 B/op\t     217 allocs/op",
            "extra": "15 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 71674145,
            "unit": "ns/op\t   59808 B/op\t     217 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 44761084,
            "unit": "ns/op\t   65663 B/op\t     217 allocs/op",
            "extra": "28 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "tt@fino.digital",
            "name": "Tobias Theel"
          },
          "committer": {
            "email": "tt@fino.digital",
            "name": "Tobias Theel"
          },
          "distinct": true,
          "id": "92d46f98a6162c85a43c4ef79e6282ec08f9e85d",
          "message": "v9 update fixes",
          "timestamp": "2021-09-02T14:29:04+02:00",
          "tree_id": "68b9108fe2df7f7d94e90f0f5d1d42e76238f3b1",
          "url": "https://github.com/Nerzal/gocloak/commit/92d46f98a6162c85a43c4ef79e6282ec08f9e85d"
        },
        "date": 1630585901837,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 88125512,
            "unit": "ns/op\t   64777 B/op\t     218 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 86922784,
            "unit": "ns/op\t   61228 B/op\t     217 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 85975586,
            "unit": "ns/op\t   67707 B/op\t     220 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 47254681,
            "unit": "ns/op\t   56970 B/op\t     216 allocs/op",
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
          "id": "55a9df0b418207ed0a166820ff41fa0fb8282d5e",
          "message": "Merge pull request #303 from mpawlowski/main\n\nadd Users to ResourcePolicyRepresentation",
          "timestamp": "2021-09-03T10:24:16+02:00",
          "tree_id": "d2ce12ad510de5168b838ed1fea678587224f860",
          "url": "https://github.com/Nerzal/gocloak/commit/55a9df0b418207ed0a166820ff41fa0fb8282d5e"
        },
        "date": 1630657600826,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 71716813,
            "unit": "ns/op\t   61337 B/op\t     219 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 73035706,
            "unit": "ns/op\t   60789 B/op\t     218 allocs/op",
            "extra": "14 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 72880815,
            "unit": "ns/op\t   63017 B/op\t     219 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 41935828,
            "unit": "ns/op\t   56186 B/op\t     216 allocs/op",
            "extra": "28 times\n2 procs"
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
          "id": "028d38203f4a3b0e69848da4f148fd8afc236525",
          "message": "Merge pull request #305 from moritz-muecke/wrong-function-signature-token-exchange\n\nFix wrong function signature",
          "timestamp": "2021-09-14T10:27:30+02:00",
          "tree_id": "53aba0130b4d2f47f7c0df45440b3fb76929b231",
          "url": "https://github.com/Nerzal/gocloak/commit/028d38203f4a3b0e69848da4f148fd8afc236525"
        },
        "date": 1631608210480,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 73618318,
            "unit": "ns/op\t   65050 B/op\t     218 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 82876739,
            "unit": "ns/op\t   76887 B/op\t     219 allocs/op",
            "extra": "15 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 71096240,
            "unit": "ns/op\t   48632 B/op\t     217 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 47089474,
            "unit": "ns/op\t   54029 B/op\t     215 allocs/op",
            "extra": "30 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "tt@fino.digital",
            "name": "Tobias Theel"
          },
          "committer": {
            "email": "tt@fino.digital",
            "name": "Tobias Theel"
          },
          "distinct": true,
          "id": "5bf3c296a0bc016a70800182bed927a20287949c",
          "message": "prepare readme for new tag",
          "timestamp": "2021-09-14T11:27:41+02:00",
          "tree_id": "274a029d9fa4cdbe2d3bb7c333fd95d84a489377",
          "url": "https://github.com/Nerzal/gocloak/commit/5bf3c296a0bc016a70800182bed927a20287949c"
        },
        "date": 1631611816717,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 91933137,
            "unit": "ns/op\t   57619 B/op\t     217 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 91294113,
            "unit": "ns/op\t   78503 B/op\t     219 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 89897963,
            "unit": "ns/op\t   64172 B/op\t     219 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 58626461,
            "unit": "ns/op\t   64616 B/op\t     219 allocs/op",
            "extra": "21 times\n2 procs"
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
          "id": "be5eb89ecb833e341414209403f7e7dffc5b61f7",
          "message": "Merge pull request #310 from Omnia-Wahid/main\n\ntype are added to the path of update policy and update permission fix #308",
          "timestamp": "2021-10-07T13:37:32+02:00",
          "tree_id": "1fbcdc32a967be514052cb6948228d48a487edbd",
          "url": "https://github.com/Nerzal/gocloak/commit/be5eb89ecb833e341414209403f7e7dffc5b61f7"
        },
        "date": 1633606833631,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 88279456,
            "unit": "ns/op\t   52704 B/op\t     216 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 90349972,
            "unit": "ns/op\t   66934 B/op\t     221 allocs/op",
            "extra": "12 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 89268310,
            "unit": "ns/op\t   42735 B/op\t     219 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 52468756,
            "unit": "ns/op\t   63360 B/op\t     217 allocs/op",
            "extra": "25 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "tt@fino.digital",
            "name": "Tobias Theel"
          },
          "committer": {
            "email": "tt@fino.digital",
            "name": "Tobias Theel"
          },
          "distinct": true,
          "id": "07319b20140613c7b53da30bcf4d06a5e6971474",
          "message": "fix codebeat badge",
          "timestamp": "2021-10-13T13:34:42+02:00",
          "tree_id": "d59e47c4580b808c28e635122b9ff3467285f80d",
          "url": "https://github.com/Nerzal/gocloak/commit/07319b20140613c7b53da30bcf4d06a5e6971474"
        },
        "date": 1634125038767,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 72078887,
            "unit": "ns/op\t   63806 B/op\t     216 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 73391095,
            "unit": "ns/op\t   51326 B/op\t     218 allocs/op",
            "extra": "14 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 71811007,
            "unit": "ns/op\t   63488 B/op\t     217 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 42242212,
            "unit": "ns/op\t   60547 B/op\t     217 allocs/op",
            "extra": "28 times\n2 procs"
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
          "id": "42526f4576b0abe7fefe44420bfc70d4dd0e55ff",
          "message": "Create codeql-analysis.yml",
          "timestamp": "2021-10-13T13:36:06+02:00",
          "tree_id": "73b627fcf44615548aba4c21fd5534a192025e8e",
          "url": "https://github.com/Nerzal/gocloak/commit/42526f4576b0abe7fefe44420bfc70d4dd0e55ff"
        },
        "date": 1634125123830,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 80901275,
            "unit": "ns/op\t   63634 B/op\t     217 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 71548663,
            "unit": "ns/op\t   58593 B/op\t     218 allocs/op",
            "extra": "15 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 77056284,
            "unit": "ns/op\t   60416 B/op\t     217 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 51975812,
            "unit": "ns/op\t   43484 B/op\t     219 allocs/op",
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
          "id": "1735136de139ba898279ce74771381c161e73983",
          "message": "Merge pull request #314 from yue-wen/main\n\nAdd query parameter search to model GetClientsParams",
          "timestamp": "2021-10-28T14:49:52+02:00",
          "tree_id": "7b959b9a2f2ee269c7af306051307c87843145be",
          "url": "https://github.com/Nerzal/gocloak/commit/1735136de139ba898279ce74771381c161e73983"
        },
        "date": 1635425541090,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 75409311,
            "unit": "ns/op\t   63494 B/op\t     219 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 72036706,
            "unit": "ns/op\t   60201 B/op\t     218 allocs/op",
            "extra": "14 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 75531300,
            "unit": "ns/op\t   52242 B/op\t     218 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 48347439,
            "unit": "ns/op\t   66037 B/op\t     218 allocs/op",
            "extra": "25 times\n2 procs"
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
          "id": "b135fbbe50569720821bcb1e1cc48df27297daa2",
          "message": "Merge pull request #313 from vennekilde/main\n\nAdded ExactName param to GetResourceParams",
          "timestamp": "2021-10-28T14:50:13+02:00",
          "tree_id": "339ee0783e44f1ebadd8a4fdbcd3b27091657b3c",
          "url": "https://github.com/Nerzal/gocloak/commit/b135fbbe50569720821bcb1e1cc48df27297daa2"
        },
        "date": 1635425589642,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 87357349,
            "unit": "ns/op\t   73580 B/op\t     222 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 86884421,
            "unit": "ns/op\t   57550 B/op\t     218 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 88032798,
            "unit": "ns/op\t   55009 B/op\t     218 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 58060379,
            "unit": "ns/op\t   57929 B/op\t     219 allocs/op",
            "extra": "19 times\n2 procs"
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
          "id": "9226d49bdefa96af08f400558525a2c6d4518346",
          "message": "Merge pull request #311 from bastianccm/add-default-role\n\nadd defaultRole to RealmRepresentation",
          "timestamp": "2021-10-28T14:51:20+02:00",
          "tree_id": "7cb8d395b86043c3388aa044c80d4df3601b75ea",
          "url": "https://github.com/Nerzal/gocloak/commit/9226d49bdefa96af08f400558525a2c6d4518346"
        },
        "date": 1635425628817,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 77061388,
            "unit": "ns/op\t   81720 B/op\t     219 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 74217088,
            "unit": "ns/op\t   92733 B/op\t     219 allocs/op",
            "extra": "15 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 75323486,
            "unit": "ns/op\t   66492 B/op\t     219 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 46006097,
            "unit": "ns/op\t   56578 B/op\t     217 allocs/op",
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
          "id": "05ae2abca43317b3da8f7d58e7bfba496222cb95",
          "message": "Merge pull request #312 from teemuteemu/client-registration-api\n\nAdd support for client registration API",
          "timestamp": "2021-10-28T14:51:03+02:00",
          "tree_id": "73b8fb387947436bac35b221eabfeec0615b5437",
          "url": "https://github.com/Nerzal/gocloak/commit/05ae2abca43317b3da8f7d58e7bfba496222cb95"
        },
        "date": 1635425658848,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 94955886,
            "unit": "ns/op\t   64344 B/op\t     216 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 93544310,
            "unit": "ns/op\t   55740 B/op\t     221 allocs/op",
            "extra": "12 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 93543219,
            "unit": "ns/op\t   57774 B/op\t     219 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 57707749,
            "unit": "ns/op\t   65862 B/op\t     219 allocs/op",
            "extra": "22 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "tt@fino.digital",
            "name": "Tobias Theel"
          },
          "committer": {
            "email": "tt@fino.digital",
            "name": "Tobias Theel"
          },
          "distinct": true,
          "id": "7401dd274b458fcec552189ca701cc60a2e889db",
          "message": "update changelog",
          "timestamp": "2021-10-28T14:54:04+02:00",
          "tree_id": "7750c285c34b1e920e78c48da0b096b774c45f36",
          "url": "https://github.com/Nerzal/gocloak/commit/7401dd274b458fcec552189ca701cc60a2e889db"
        },
        "date": 1635425808214,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 82767796,
            "unit": "ns/op\t   79121 B/op\t     220 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 82169091,
            "unit": "ns/op\t   81826 B/op\t     219 allocs/op",
            "extra": "15 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 83124071,
            "unit": "ns/op\t   59457 B/op\t     220 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 43802851,
            "unit": "ns/op\t   59509 B/op\t     214 allocs/op",
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
          "id": "4f624d6513eff0d193aba7bbeb75fa157c42ab2a",
          "message": "Merge pull request #318 from bastianccm/auth-flows-and-executions\n\nAuth flows and executions",
          "timestamp": "2021-11-17T12:29:50+01:00",
          "tree_id": "f3c0ac615885c13f452911691d026b6ee9d3b772",
          "url": "https://github.com/Nerzal/gocloak/commit/4f624d6513eff0d193aba7bbeb75fa157c42ab2a"
        },
        "date": 1637148771684,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 73671797,
            "unit": "ns/op\t   64601 B/op\t     219 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 73813406,
            "unit": "ns/op\t   60770 B/op\t     216 allocs/op",
            "extra": "15 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 73200118,
            "unit": "ns/op\t   74588 B/op\t     218 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 45611964,
            "unit": "ns/op\t   49189 B/op\t     216 allocs/op",
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
          "id": "a4e6ca852cdfa34e61fa8435dbddc43cdfb2e3f2",
          "message": "Merge pull request #320 from bastianccm/passwordpolicy\n\nadd password policies to server info",
          "timestamp": "2021-11-17T12:30:35+01:00",
          "tree_id": "7a35d1bd8fdb65804bfc40b38438d679219a71f5",
          "url": "https://github.com/Nerzal/gocloak/commit/a4e6ca852cdfa34e61fa8435dbddc43cdfb2e3f2"
        },
        "date": 1637148800853,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 86743561,
            "unit": "ns/op\t   55781 B/op\t     220 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 74207733,
            "unit": "ns/op\t   70343 B/op\t     218 allocs/op",
            "extra": "15 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 83900901,
            "unit": "ns/op\t   90496 B/op\t     219 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 52290228,
            "unit": "ns/op\t   69204 B/op\t     218 allocs/op",
            "extra": "26 times\n2 procs"
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
          "id": "e1f44489760d364959abaa896104e19efe4ffab4",
          "message": "Merge pull request #319 from bastianccm/utils\n\ndefaults for P* utils",
          "timestamp": "2021-11-17T12:31:08+01:00",
          "tree_id": "c78e84383b7cf35d404a3564b7dc65809d6b7b51",
          "url": "https://github.com/Nerzal/gocloak/commit/e1f44489760d364959abaa896104e19efe4ffab4"
        },
        "date": 1637148856617,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 84280726,
            "unit": "ns/op\t   62817 B/op\t     218 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 82151043,
            "unit": "ns/op\t   72948 B/op\t     221 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 79733205,
            "unit": "ns/op\t   59070 B/op\t     219 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 54251566,
            "unit": "ns/op\t   54230 B/op\t     218 allocs/op",
            "extra": "21 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "tt@fino.digital",
            "name": "Tobias Theel"
          },
          "committer": {
            "email": "tt@fino.digital",
            "name": "Tobias Theel"
          },
          "distinct": true,
          "id": "2e4a11e5ea4e6231e7cd01eb48df32e880082d7d",
          "message": "fix test",
          "timestamp": "2021-11-19T15:49:11+01:00",
          "tree_id": "86a171c0c9288d411db0aeb48ef6cbb7b4f4f351",
          "url": "https://github.com/Nerzal/gocloak/commit/2e4a11e5ea4e6231e7cd01eb48df32e880082d7d"
        },
        "date": 1637333497412,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 70121234,
            "unit": "ns/op\t   52595 B/op\t     218 allocs/op",
            "extra": "16 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 70521935,
            "unit": "ns/op\t   51389 B/op\t     218 allocs/op",
            "extra": "16 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 71402283,
            "unit": "ns/op\t   61850 B/op\t     219 allocs/op",
            "extra": "16 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 42775098,
            "unit": "ns/op\t   62385 B/op\t     217 allocs/op",
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
          "id": "44f36fb244a4e6d8cf5996d6f7b72dbc1e5c62e9",
          "message": "Merge pull request #322 from latipovsharif/patch-1\n\nUpdate README.md",
          "timestamp": "2021-11-19T16:10:40+01:00",
          "tree_id": "8a81a9365850b7c733e94a46465cf46381fa220c",
          "url": "https://github.com/Nerzal/gocloak/commit/44f36fb244a4e6d8cf5996d6f7b72dbc1e5c62e9"
        },
        "date": 1637334807982,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 103189547,
            "unit": "ns/op\t   59216 B/op\t     220 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 102121590,
            "unit": "ns/op\t   63318 B/op\t     222 allocs/op",
            "extra": "12 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 99421664,
            "unit": "ns/op\t   56188 B/op\t     220 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 66241736,
            "unit": "ns/op\t   69928 B/op\t     221 allocs/op",
            "extra": "19 times\n2 procs"
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
          "id": "dd6b55b25888e7b9327ad957902b638aa4604d93",
          "message": "Merge pull request #329 from igorkim/main\n\nAdd RedirectURI field in TokenOptions structure",
          "timestamp": "2021-12-10T11:41:50+01:00",
          "tree_id": "5a0170415aec4b13fe8a95b73eb67982ed3fa116",
          "url": "https://github.com/Nerzal/gocloak/commit/dd6b55b25888e7b9327ad957902b638aa4604d93"
        },
        "date": 1639133072810,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 82706074,
            "unit": "ns/op\t   50024 B/op\t     219 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 83573742,
            "unit": "ns/op\t   73092 B/op\t     220 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 84790736,
            "unit": "ns/op\t   55755 B/op\t     218 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 52308590,
            "unit": "ns/op\t   56123 B/op\t     218 allocs/op",
            "extra": "25 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "bastian.ike@aoe.com",
            "name": "Bastian",
            "username": "bastianccm"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "392e397a4023a2be4728f6e355994a67af380039",
          "message": "fix naming for authentication executions (#330)",
          "timestamp": "2021-12-10T12:36:34+01:00",
          "tree_id": "332255a05a755794ae73077d7e78b58b8b128362",
          "url": "https://github.com/Nerzal/gocloak/commit/392e397a4023a2be4728f6e355994a67af380039"
        },
        "date": 1639136333092,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 72666968,
            "unit": "ns/op\t   61557 B/op\t     218 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 72946305,
            "unit": "ns/op\t   75236 B/op\t     222 allocs/op",
            "extra": "15 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 71538036,
            "unit": "ns/op\t   60323 B/op\t     218 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 40685151,
            "unit": "ns/op\t   63465 B/op\t     218 allocs/op",
            "extra": "26 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "sirockin@gmail.com",
            "name": "Dave Sirockin",
            "username": "sirockin"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "76b1bd1e3b71ab7f376d5ec41650620c207404af",
          "message": "add `Q` and `EmailVerified` to GetUsersParams (#337)\n\nCo-authored-by: Dave Sirockin <dave@flow-systems.net>",
          "timestamp": "2022-01-24T11:39:20+01:00",
          "tree_id": "0b71fcbb303d817b2feebf7467ab6613b0d20a39",
          "url": "https://github.com/Nerzal/gocloak/commit/76b1bd1e3b71ab7f376d5ec41650620c207404af"
        },
        "date": 1643020922854,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 85958892,
            "unit": "ns/op\t   52392 B/op\t     220 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 83533600,
            "unit": "ns/op\t   71292 B/op\t     221 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 83754964,
            "unit": "ns/op\t   60714 B/op\t     221 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 49329060,
            "unit": "ns/op\t   55621 B/op\t     218 allocs/op",
            "extra": "24 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "chris@shibumi.dev",
            "name": "Christian Rebischke",
            "username": "shibumi"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "7fa5389b9f83a2be4ee2d372b9d7cbaf54402154",
          "message": "feat: add requiredActionProviderRepresentation (#339)\n\nThis commit adds the requiredActionProviderRepresentation and a function for updating required actions",
          "timestamp": "2022-02-17T14:24:22+01:00",
          "tree_id": "076bec138675a63029ab111614f6522b772a155c",
          "url": "https://github.com/Nerzal/gocloak/commit/7fa5389b9f83a2be4ee2d372b9d7cbaf54402154"
        },
        "date": 1645104439844,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 79408505,
            "unit": "ns/op\t   75061 B/op\t     221 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 80309614,
            "unit": "ns/op\t   69213 B/op\t     220 allocs/op",
            "extra": "14 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 85229091,
            "unit": "ns/op\t   57287 B/op\t     219 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 46200397,
            "unit": "ns/op\t   59328 B/op\t     219 allocs/op",
            "extra": "25 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "github@mdrone.de",
            "name": "Merlin Dienst",
            "username": "doktormerlin"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "4716d38a7db252adb5f3b3b5ebaeb89166538927",
          "message": "Changed ComponentConfig to be a map[string][]string instead of its own struct (#325)\n\n* Feature: #321 ComponentConfig now is a map[string]string to generalize it's usage\r\n\r\n* changed module name\r\n\r\n* changed all nerzal to doktormerlin\r\n\r\n* added replacement directive\r\n\r\n* Added newline to end of file\r\n\r\n* removed replacement directive\r\n\r\n* map[string]string now map[string][]string in ComponentConfig\r\n\r\n* Changed ComponentConfig to be a map[string][]string instead of its own struct\r\n\r\n* Fixed issues caused by renaming. Go and forking just doesnt go well together\r\n\r\n* Renamed some stuff from Nerzal to nerzal\r\n\r\n* tests sleeping 10 seconds again\r\n\r\n* reverted tests\r\n\r\nCo-authored-by: merl_umlaut <merlin.dienst@umlaut.com>",
          "timestamp": "2022-02-17T14:26:26+01:00",
          "tree_id": "b6d7b6ff2bf49073e2e3d44a4cd34fbe8e0fbeca",
          "url": "https://github.com/Nerzal/gocloak/commit/4716d38a7db252adb5f3b3b5ebaeb89166538927"
        },
        "date": 1645104536679,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 73474112,
            "unit": "ns/op\t   52667 B/op\t     218 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 79426320,
            "unit": "ns/op\t   60300 B/op\t     220 allocs/op",
            "extra": "15 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 72432054,
            "unit": "ns/op\t   65299 B/op\t     218 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 52293455,
            "unit": "ns/op\t   64149 B/op\t     219 allocs/op",
            "extra": "26 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "tt@fino.digital",
            "name": "Tobias Theel"
          },
          "committer": {
            "email": "tt@fino.digital",
            "name": "Tobias Theel"
          },
          "distinct": true,
          "id": "b55ab25c3e22f9da84028faeb254a9375502c093",
          "message": "ignore gocognit in test files",
          "timestamp": "2022-02-22T17:07:18+01:00",
          "tree_id": "df087a57c4c9e06761fb09a6b2187db1f5c83d4e",
          "url": "https://github.com/Nerzal/gocloak/commit/b55ab25c3e22f9da84028faeb254a9375502c093"
        },
        "date": 1645546205381,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 92228484,
            "unit": "ns/op\t   59866 B/op\t     220 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 90883401,
            "unit": "ns/op\t   54122 B/op\t     219 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 90788877,
            "unit": "ns/op\t   63380 B/op\t     218 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 52479600,
            "unit": "ns/op\t   59701 B/op\t     219 allocs/op",
            "extra": "22 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "tt@fino.digital",
            "name": "Tobias Theel"
          },
          "committer": {
            "email": "tt@fino.digital",
            "name": "Tobias Theel"
          },
          "distinct": true,
          "id": "c0c3fdca4d0b0c5cecac477552066b50cfbd14c8",
          "message": "upgrade to V11",
          "timestamp": "2022-03-01T10:34:58+01:00",
          "tree_id": "174d10d984a02dd5ccd28333220025ecb7b62729",
          "url": "https://github.com/Nerzal/gocloak/commit/c0c3fdca4d0b0c5cecac477552066b50cfbd14c8"
        },
        "date": 1646127471767,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 95459026,
            "unit": "ns/op\t   61645 B/op\t     221 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 93690640,
            "unit": "ns/op\t   63234 B/op\t     223 allocs/op",
            "extra": "12 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 90154550,
            "unit": "ns/op\t   46215 B/op\t     220 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 56046811,
            "unit": "ns/op\t   60192 B/op\t     219 allocs/op",
            "extra": "20 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "77924577+tjarkmeyer@users.noreply.github.com",
            "name": "Tjark Meyer",
            "username": "tjarkmeyer"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "44d3a1f80082495a5bc11e23f806cb0541741feb",
          "message": "fix: json tags in GetEventsParams (#335) (#340)\n\nCo-authored-by: Tjark Meyer <t.meyer@e-mission.de>",
          "timestamp": "2022-03-01T10:31:56+01:00",
          "tree_id": "1488f732d394854e411f0feee4435447e093737c",
          "url": "https://github.com/Nerzal/gocloak/commit/44d3a1f80082495a5bc11e23f806cb0541741feb"
        },
        "date": 1646127776331,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 77091331,
            "unit": "ns/op\t   70689 B/op\t     222 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 77233511,
            "unit": "ns/op\t   62671 B/op\t     219 allocs/op",
            "extra": "15 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 76410301,
            "unit": "ns/op\t   49217 B/op\t     219 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 43413906,
            "unit": "ns/op\t   64435 B/op\t     220 allocs/op",
            "extra": "25 times\n2 procs"
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
          "id": "30870aa195a7975adfcb929c5b90a4100c675a18",
          "message": "Update README.md",
          "timestamp": "2022-03-01T13:02:15+01:00",
          "tree_id": "0d991379ccda946c0dba94fa3e61ff8fb6e92a07",
          "url": "https://github.com/Nerzal/gocloak/commit/30870aa195a7975adfcb929c5b90a4100c675a18"
        },
        "date": 1646136258935,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 64864154,
            "unit": "ns/op\t   60524 B/op\t     215 allocs/op",
            "extra": "16 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 63950068,
            "unit": "ns/op\t   63517 B/op\t     217 allocs/op",
            "extra": "18 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 63464242,
            "unit": "ns/op\t   73776 B/op\t     219 allocs/op",
            "extra": "18 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 35395355,
            "unit": "ns/op\t   66496 B/op\t     217 allocs/op",
            "extra": "34 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "quantonganh@gmail.com",
            "name": "quantonganh",
            "username": "quantonganh"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "4f92e9c55399685d71f96f3645f802de3d2ca350",
          "message": "token exchange: make requested_subject optional (#343)",
          "timestamp": "2022-03-08T08:54:33+01:00",
          "tree_id": "d6d6d91df11fef2660d91260acf8013052905603",
          "url": "https://github.com/Nerzal/gocloak/commit/4f92e9c55399685d71f96f3645f802de3d2ca350"
        },
        "date": 1646726223639,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 80520614,
            "unit": "ns/op\t   59627 B/op\t     218 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 76158342,
            "unit": "ns/op\t   61976 B/op\t     222 allocs/op",
            "extra": "14 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 79773612,
            "unit": "ns/op\t   60030 B/op\t     217 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 46309110,
            "unit": "ns/op\t   70669 B/op\t     220 allocs/op",
            "extra": "26 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "bastian.ike@aoe.com",
            "name": "Bastian",
            "username": "bastianccm"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "fec776484b690e0823d4318415cbeff02387badf",
          "message": "fix(client): remove unnecessary fmt.Print (#344)",
          "timestamp": "2022-03-11T10:22:28+01:00",
          "tree_id": "0b4192ba8c53fcd0abcd40689eb8e095e809f87f",
          "url": "https://github.com/Nerzal/gocloak/commit/fec776484b690e0823d4318415cbeff02387badf"
        },
        "date": 1646990689373,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 71365700,
            "unit": "ns/op\t   62275 B/op\t     221 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 80572081,
            "unit": "ns/op\t   70020 B/op\t     219 allocs/op",
            "extra": "15 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 68807842,
            "unit": "ns/op\t   64540 B/op\t     217 allocs/op",
            "extra": "16 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 42423852,
            "unit": "ns/op\t   63488 B/op\t     218 allocs/op",
            "extra": "28 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "tt@fino.digital",
            "name": "Tobias Theel"
          },
          "committer": {
            "email": "tt@fino.digital",
            "name": "Tobias Theel"
          },
          "distinct": true,
          "id": "94b56cad73b9be260a44601a6bb560d6f7ab4a64",
          "message": "update github actions",
          "timestamp": "2022-05-04T18:35:56+02:00",
          "tree_id": "910d356bdac3ba1d0cf96cb67849cd70f4e03962",
          "url": "https://github.com/Nerzal/gocloak/commit/94b56cad73b9be260a44601a6bb560d6f7ab4a64"
        },
        "date": 1651682335180,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 86834484,
            "unit": "ns/op\t   48898 B/op\t     216 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 86836001,
            "unit": "ns/op\t   61004 B/op\t     221 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 87110710,
            "unit": "ns/op\t   71526 B/op\t     220 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 47913477,
            "unit": "ns/op\t   65818 B/op\t     218 allocs/op",
            "extra": "22 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "tt@fino.digital",
            "name": "Tobias Theel"
          },
          "committer": {
            "email": "tt@fino.digital",
            "name": "Tobias Theel"
          },
          "distinct": true,
          "id": "e32d443550835cf2be738f9886884d5cd6d348de",
          "message": "update direct dependencies",
          "timestamp": "2022-05-04T18:38:01+02:00",
          "tree_id": "03b706e211726692eaa7c45dc521c6f53d576908",
          "url": "https://github.com/Nerzal/gocloak/commit/e32d443550835cf2be738f9886884d5cd6d348de"
        },
        "date": 1651682442544,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 72236492,
            "unit": "ns/op\t   51729 B/op\t     220 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 72367704,
            "unit": "ns/op\t   58117 B/op\t     222 allocs/op",
            "extra": "15 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 71710819,
            "unit": "ns/op\t   56758 B/op\t     220 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 41149889,
            "unit": "ns/op\t   66854 B/op\t     220 allocs/op",
            "extra": "25 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "tt@fino.digital",
            "name": "Tobias Theel"
          },
          "committer": {
            "email": "tt@fino.digital",
            "name": "Tobias Theel"
          },
          "distinct": true,
          "id": "b2d9f950b45dec1393bbcc44850bb8e624d86a59",
          "message": "update golang-ci lint action to newest version",
          "timestamp": "2022-05-04T18:39:36+02:00",
          "tree_id": "64537c94abc292243444ed500b3a0433bdef6fc4",
          "url": "https://github.com/Nerzal/gocloak/commit/b2d9f950b45dec1393bbcc44850bb8e624d86a59"
        },
        "date": 1651682521910,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 70718342,
            "unit": "ns/op\t   57311 B/op\t     219 allocs/op",
            "extra": "16 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 70550560,
            "unit": "ns/op\t   66410 B/op\t     221 allocs/op",
            "extra": "16 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 70659348,
            "unit": "ns/op\t   62901 B/op\t     219 allocs/op",
            "extra": "16 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 41978063,
            "unit": "ns/op\t   57036 B/op\t     220 allocs/op",
            "extra": "27 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "tt@fino.digital",
            "name": "Tobias Theel"
          },
          "committer": {
            "email": "tt@fino.digital",
            "name": "Tobias Theel"
          },
          "distinct": true,
          "id": "3348765652e493e9b2110d27bd54113d75908b89",
          "message": "update install go action to newest version",
          "timestamp": "2022-05-04T18:40:12+02:00",
          "tree_id": "9cb9893d2d7122e6c4cb7c4fcdede1a538317a9c",
          "url": "https://github.com/Nerzal/gocloak/commit/3348765652e493e9b2110d27bd54113d75908b89"
        },
        "date": 1651682579068,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 80605121,
            "unit": "ns/op\t   62339 B/op\t     222 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 73608046,
            "unit": "ns/op\t   70253 B/op\t     221 allocs/op",
            "extra": "15 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 78537459,
            "unit": "ns/op\t   55375 B/op\t     220 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 50987091,
            "unit": "ns/op\t   61495 B/op\t     220 allocs/op",
            "extra": "27 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "mpawlowski@users.noreply.github.com",
            "name": "Marcin Pawlowski",
            "username": "mpawlowski"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "289eed6740763c589dc44bee0a0ba16178371fec",
          "message": " update GetResourcePoliciesParams to use correct filter key (#350)",
          "timestamp": "2022-05-04T21:53:08+02:00",
          "tree_id": "19626bf7dd811fc0cf942c9da131393398557e7b",
          "url": "https://github.com/Nerzal/gocloak/commit/289eed6740763c589dc44bee0a0ba16178371fec"
        },
        "date": 1651694137921,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 71480217,
            "unit": "ns/op\t   74522 B/op\t     220 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 71955918,
            "unit": "ns/op\t   68379 B/op\t     222 allocs/op",
            "extra": "15 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 70994658,
            "unit": "ns/op\t   56851 B/op\t     221 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 41705048,
            "unit": "ns/op\t   72375 B/op\t     221 allocs/op",
            "extra": "30 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "himanshu.malpande@gmail.com",
            "name": "Himanshu Malpande",
            "username": "HimanshuM"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "47ca5d67f90cab8d2efc262f5fa1f6e53054f1b3",
          "message": "Added ClientScope ProtocolMappers CRUD (#351)",
          "timestamp": "2022-05-05T10:00:17+02:00",
          "tree_id": "042130b275f99f9cf457172265f7698c21012d9c",
          "url": "https://github.com/Nerzal/gocloak/commit/47ca5d67f90cab8d2efc262f5fa1f6e53054f1b3"
        },
        "date": 1651737779594,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 83072505,
            "unit": "ns/op\t   62422 B/op\t     221 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 91332764,
            "unit": "ns/op\t   58006 B/op\t     221 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 86669467,
            "unit": "ns/op\t   72600 B/op\t     221 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 55396409,
            "unit": "ns/op\t   66468 B/op\t     221 allocs/op",
            "extra": "22 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "msdmelo@inf.ufpel.edu.br",
            "name": "Mateus Santos de Melo",
            "username": "mateussm"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "c740220171ec1414a232fafb2543d213d23b8232",
          "message": "add: full path field (#349)",
          "timestamp": "2022-05-05T10:46:21+02:00",
          "tree_id": "1d2be99346470ea14b153f9a6b4e928bfb4c735b",
          "url": "https://github.com/Nerzal/gocloak/commit/c740220171ec1414a232fafb2543d213d23b8232"
        },
        "date": 1651740525247,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 75481509,
            "unit": "ns/op\t   55396 B/op\t     220 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 73372828,
            "unit": "ns/op\t   74990 B/op\t     223 allocs/op",
            "extra": "15 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 74513140,
            "unit": "ns/op\t   78169 B/op\t     222 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 45152997,
            "unit": "ns/op\t   65058 B/op\t     222 allocs/op",
            "extra": "25 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "git@lit.plus",
            "name": "Philipp Nowak",
            "username": "literalplus"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "04f3ff43ffe1b5683374a05f8b83fb63ad601589",
          "message": "Add missing endpoints for client scopes -> scope mappings -> client roles (#348)\n\n#348 Add missing endpoints for client scopes -> scope mappings -> client roles",
          "timestamp": "2022-05-10T10:51:56+02:00",
          "tree_id": "9f50b2fa3633c084941fb984d735833b41b983ef",
          "url": "https://github.com/Nerzal/gocloak/commit/04f3ff43ffe1b5683374a05f8b83fb63ad601589"
        },
        "date": 1652172882760,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 89152601,
            "unit": "ns/op\t   54610 B/op\t     225 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 88266469,
            "unit": "ns/op\t   77385 B/op\t     224 allocs/op",
            "extra": "12 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 88491658,
            "unit": "ns/op\t   60559 B/op\t     223 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 63098721,
            "unit": "ns/op\t   68609 B/op\t     225 allocs/op",
            "extra": "18 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "49190107+dmnkdmnt@users.noreply.github.com",
            "name": "Dominik Dumont",
            "username": "dmnkdmnt"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "9556250acd6a20d04682540e7a40230daaac9ca7",
          "message": "resolve protocol mapper types and builtin protocol mappers on server info (#354)",
          "timestamp": "2022-05-19T13:59:43+02:00",
          "tree_id": "0ccf59556839b12783539482ff14f28a586656cb",
          "url": "https://github.com/Nerzal/gocloak/commit/9556250acd6a20d04682540e7a40230daaac9ca7"
        },
        "date": 1652961769593,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 92195903,
            "unit": "ns/op\t   63526 B/op\t     222 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 91831893,
            "unit": "ns/op\t   70396 B/op\t     222 allocs/op",
            "extra": "12 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 92842482,
            "unit": "ns/op\t   56756 B/op\t     222 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 55249346,
            "unit": "ns/op\t   61928 B/op\t     223 allocs/op",
            "extra": "21 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "vovakirishi@gmail.com",
            "name": "Vladimir",
            "username": "VladimirStepanov"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "1c89abf61f250a344b37bd27fb79c64363892a3e",
          "message": "feat: add supporing ecdsa algorithm (#356)\n\n* feat: add supporing es256 algorithm\r\n\r\n* feat: add supporing es256 algorithm in method DecodeAccessTokenCustomClaims\r\n\r\n* fix: rename method\r\n\r\n* fix: fix an algorithm detecting bug\r\n\r\n* refactor: remove code duplicates\r\n\r\n* fix: remove useless println\r\n\r\n* refactor: remove code duplicates\r\n\r\n* test: add tests\r\n\r\n* refactor: add comments\r\n\r\n* refactor: refactor of decodeECDSAPublicKey func\r\n\r\nCo-authored-by: Vladimir Stepanov <v.stepanov@redmadrobot.com>",
          "timestamp": "2022-06-27T15:50:24+02:00",
          "tree_id": "ce8334963734a3b433042a5884415876cfa4095f",
          "url": "https://github.com/Nerzal/gocloak/commit/1c89abf61f250a344b37bd27fb79c64363892a3e"
        },
        "date": 1656337979619,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 78672246,
            "unit": "ns/op\t   64972 B/op\t     222 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 78230641,
            "unit": "ns/op\t   61238 B/op\t     221 allocs/op",
            "extra": "14 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 78359433,
            "unit": "ns/op\t   65915 B/op\t     222 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 47167474,
            "unit": "ns/op\t   52671 B/op\t     220 allocs/op",
            "extra": "27 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "ziggo1879@gmail.com",
            "name": "Sven Ziegler",
            "username": "svzieg"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "fe4f627eaf1bff988ee5df2fd0d0b87daac6c074",
          "message": "feat: add functionality to register, get, list and delete required ac… (#360)\n\n* feat: add functionality to register, get, list and delete required actions\r\n\r\n* fix: golangci-lint errors by using go 1.19 binaries\r\n\r\n* chore: add some more tests\r\n\r\n* chore: rename misspelled key in test",
          "timestamp": "2022-08-11T09:51:44+02:00",
          "tree_id": "e312138c6a8e9906b8c7c575ff8326e588a82c3e",
          "url": "https://github.com/Nerzal/gocloak/commit/fe4f627eaf1bff988ee5df2fd0d0b87daac6c074"
        },
        "date": 1660204462683,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 82452269,
            "unit": "ns/op\t   60774 B/op\t     221 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 79047554,
            "unit": "ns/op\t   79628 B/op\t     222 allocs/op",
            "extra": "14 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 82729958,
            "unit": "ns/op\t   42814 B/op\t     222 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 56870020,
            "unit": "ns/op\t   60842 B/op\t     223 allocs/op",
            "extra": "21 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "bastian.ike@aoe.com",
            "name": "Bastian",
            "username": "bastianccm"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "1137160ec90381be67ca78acafec3f7bbaa8e56d",
          "message": "fix(realm): add webauthn config (#371)",
          "timestamp": "2022-10-12T18:15:09+02:00",
          "tree_id": "a93765bf1eaf3fba95e1513d34a71a4c264c05e2",
          "url": "https://github.com/Nerzal/gocloak/commit/1137160ec90381be67ca78acafec3f7bbaa8e56d"
        },
        "date": 1665591463837,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 75285985,
            "unit": "ns/op\t   55793 B/op\t     220 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 75556873,
            "unit": "ns/op\t   55536 B/op\t     221 allocs/op",
            "extra": "15 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 75756430,
            "unit": "ns/op\t   54078 B/op\t     220 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 47263445,
            "unit": "ns/op\t   60388 B/op\t     223 allocs/op",
            "extra": "22 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "bastian.ike@aoe.com",
            "name": "Bastian",
            "username": "bastianccm"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "2cb779b591064a777c9333313006f6ee680d457b",
          "message": "feat(Realm): get/alter localizations (#372)\n\nCo-authored-by: Dirk Gerretz <dirk@gerretz.de>",
          "timestamp": "2022-10-12T18:15:41+02:00",
          "tree_id": "a0198da1ec7c2dd84229cf7989ab49d50c436fe6",
          "url": "https://github.com/Nerzal/gocloak/commit/2cb779b591064a777c9333313006f6ee680d457b"
        },
        "date": 1665591528069,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 88640377,
            "unit": "ns/op\t   60594 B/op\t     224 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 93325599,
            "unit": "ns/op\t   63607 B/op\t     224 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 90568117,
            "unit": "ns/op\t   55835 B/op\t     222 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 51069803,
            "unit": "ns/op\t   61352 B/op\t     221 allocs/op",
            "extra": "22 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "53079496+ingoaf@users.noreply.github.com",
            "name": "ingoaf",
            "username": "ingoaf"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "287fa81adca19561577f38fa03d586b3ffa5a34d",
          "message": "remove deprecated linter (#369)",
          "timestamp": "2022-10-12T18:16:47+02:00",
          "tree_id": "cb95003d113ba1354822ed5722f0676ddfbf497d",
          "url": "https://github.com/Nerzal/gocloak/commit/287fa81adca19561577f38fa03d586b3ffa5a34d"
        },
        "date": 1665591545446,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 71561619,
            "unit": "ns/op\t   65889 B/op\t     222 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 71388715,
            "unit": "ns/op\t   62806 B/op\t     220 allocs/op",
            "extra": "15 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 74997544,
            "unit": "ns/op\t   53777 B/op\t     221 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 39520749,
            "unit": "ns/op\t   65080 B/op\t     220 allocs/op",
            "extra": "26 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "moritz.muecke@aoe.com",
            "name": "Moritz Mücke",
            "username": "moritz-muecke"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "f15f3bd541e6a6ba54e5ce1bce8667a68336ee3f",
          "message": "fix(realm): error handling for update required action (#374)",
          "timestamp": "2022-10-12T18:18:17+02:00",
          "tree_id": "b6654ebaaf7251ec582e6b55994e152930e63510",
          "url": "https://github.com/Nerzal/gocloak/commit/f15f3bd541e6a6ba54e5ce1bce8667a68336ee3f"
        },
        "date": 1665591633505,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 73507811,
            "unit": "ns/op\t   62710 B/op\t     220 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 73790814,
            "unit": "ns/op\t   59026 B/op\t     221 allocs/op",
            "extra": "15 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 75206305,
            "unit": "ns/op\t   70213 B/op\t     221 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 39929706,
            "unit": "ns/op\t   58572 B/op\t     219 allocs/op",
            "extra": "27 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "bastian.ike@aoe.com",
            "name": "Bastian",
            "username": "bastianccm"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "4f04556a7a037f207eb3ae9b4193f426d54f6551",
          "message": "feat(gocloak): add GetCompositeRolesByRoleID (#376)",
          "timestamp": "2022-10-12T18:19:08+02:00",
          "tree_id": "f180c1422591172630d7bd6ab539b34f26904953",
          "url": "https://github.com/Nerzal/gocloak/commit/4f04556a7a037f207eb3ae9b4193f426d54f6551"
        },
        "date": 1665591688874,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 74863393,
            "unit": "ns/op\t   67288 B/op\t     221 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 72627169,
            "unit": "ns/op\t   54220 B/op\t     222 allocs/op",
            "extra": "15 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 72012527,
            "unit": "ns/op\t   73189 B/op\t     221 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 40837251,
            "unit": "ns/op\t   47395 B/op\t     220 allocs/op",
            "extra": "27 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "moritz.muecke@aoe.com",
            "name": "Moritz Mücke",
            "username": "moritz-muecke"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "e42788eda8ea7e19c85829ab5d17261f002a4075",
          "message": "feat(roles): get groups by client role (#373)\n\n* feat(roles): get groups by client role\r\n\r\n* fix(lint): rename id parameters\r\n\r\n* test: add unit tests for GroupsByRole and GroupsByClientRole",
          "timestamp": "2022-10-12T18:18:43+02:00",
          "tree_id": "7b47c03145c29de73f9312558e37408dc2c542ed",
          "url": "https://github.com/Nerzal/gocloak/commit/e42788eda8ea7e19c85829ab5d17261f002a4075"
        },
        "date": 1665591694908,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 84955652,
            "unit": "ns/op\t   63095 B/op\t     222 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 87664533,
            "unit": "ns/op\t   69095 B/op\t     225 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 86909963,
            "unit": "ns/op\t   64421 B/op\t     223 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 47511400,
            "unit": "ns/op\t   59382 B/op\t     222 allocs/op",
            "extra": "22 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "timothee@indie.host",
            "name": "Timothee Gosselin",
            "username": "unteem"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "1cd6e0f75ba0f7b6be619b0f275894c5d2cc8353",
          "message": "add missing attributes to ProtocolMappersConfig (#363)",
          "timestamp": "2022-10-12T18:19:40+02:00",
          "tree_id": "7ba7347dd1331c30e92da03ffbea3d242fcaf494",
          "url": "https://github.com/Nerzal/gocloak/commit/1cd6e0f75ba0f7b6be619b0f275894c5d2cc8353"
        },
        "date": 1665591763105,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 93442727,
            "unit": "ns/op\t   49521 B/op\t     222 allocs/op",
            "extra": "12 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 92762582,
            "unit": "ns/op\t   48765 B/op\t     220 allocs/op",
            "extra": "12 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 94762775,
            "unit": "ns/op\t   51568 B/op\t     220 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 49406447,
            "unit": "ns/op\t   53768 B/op\t     222 allocs/op",
            "extra": "21 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "favoursmoe@gmail.com",
            "name": "Undefined K",
            "username": "Kyya"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "1a6a98e02c843c1866c824aa644113df2a06e865",
          "message": "feat: support getting group by path (#375)",
          "timestamp": "2022-10-12T18:20:04+02:00",
          "tree_id": "a9d5027b36794d67a0215ccddad41c47be96f4d5",
          "url": "https://github.com/Nerzal/gocloak/commit/1a6a98e02c843c1866c824aa644113df2a06e865"
        },
        "date": 1665591775264,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 85003254,
            "unit": "ns/op\t   60890 B/op\t     222 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 81974256,
            "unit": "ns/op\t   54711 B/op\t     223 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 85140999,
            "unit": "ns/op\t   50088 B/op\t     221 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 57658433,
            "unit": "ns/op\t   67877 B/op\t     220 allocs/op",
            "extra": "27 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "Tolyar@users.noreply.github.com",
            "name": "Tolyar",
            "username": "Tolyar"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "b4b5d20412546ecd710d3d211f543b201fb0fcd5",
          "message": "GetComponentsWithParams and UpdateComponent added (#359)\n\n* GetComponentsWithParams and test for it\r\n\r\n* Possible fix for GetComponent tests\r\n\r\n* UpdateComponent and tests for it\r\n\r\n* Possible fix for UpdateComponent tests\r\n\r\n* Check component name after UpdateComponent. It should be changed.\r\n\r\n* Allow to get component by ID\r\n\r\n* Add GetComponent for fetch component by ID.\r\n\r\n* GetComponent always return only one component.\r\n\r\n* Fix GetComponents query params\r\n\r\n* Fix whitespace\r\n\r\nCo-authored-by: Vladimir Fidunin <v.fidunin@vk.team>",
          "timestamp": "2022-10-13T10:57:57+02:00",
          "tree_id": "a3d05fc1ce59f40c8c0fad0de00a368d5e58e21e",
          "url": "https://github.com/Nerzal/gocloak/commit/b4b5d20412546ecd710d3d211f543b201fb0fcd5"
        },
        "date": 1665651623286,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 78857345,
            "unit": "ns/op\t   54810 B/op\t     222 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 79474063,
            "unit": "ns/op\t   70240 B/op\t     221 allocs/op",
            "extra": "14 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 79939025,
            "unit": "ns/op\t   65022 B/op\t     219 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 44646315,
            "unit": "ns/op\t   59058 B/op\t     221 allocs/op",
            "extra": "26 times\n2 procs"
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
          "id": "18f0f3d14cb8dd31a49d357ca37fab0aa8ca337f",
          "message": "WIP: Support for Keycloak 17+ / Version 19 as target (#361)\n\n* accept interfaces, return structs\r\n\r\n* update dependencies\r\n\r\n* update go version in CI\r\n\r\n* Fix linter errors\r\n\r\n* BreakingChange: fix typo in ServerInfoRepresentation\r\n\r\n* BreakingChange: rename RetrospectTokenResult to IntrospectTokenResult\r\n\r\n* fix more linter errors\r\n\r\n* BreakingChange: Bugfix - Issue #358 CreateClientRepresentation does now require a token\r\n\r\n* upgrade for keycloak 19\r\n\r\n* update CI\r\n\r\n* jub\r\n\r\n* cleanup\r\n\r\n* add SendVerifyEmail function\r\n\r\n* test\r\n\r\n* Fix #248\r\n\r\n* Fix realm import (#367)\r\n\r\n* Fix realm import\r\n\r\n* Fix gocloak-realm\r\n\r\n* Add test permissions, policy\r\n\r\n* Change JS policies to client, remove test for uploading JS policy\r\n\r\n* Fix register client tests\r\n\r\nCo-authored-by: Pavol Ipoth <pavol.ipoth@external.t-systems.com>\r\n\r\n* fix tests (#379)\r\n\r\n* fix tests\r\n\r\n* Fix Comment go-linter\r\n\r\n* update dependencies and add nancy ignore case\r\n\r\n* Add -d to make file\r\n\r\nCo-authored-by: Jonas Heinemann <jonas.heinemann@clarilab.de>\r\n\r\n* upgrade gocloak version (#380)\r\n\r\nCo-authored-by: Tobias Theel <tt@fino.digital>\r\nCo-authored-by: p53 <pavol.ipoth@protonmail.com>\r\nCo-authored-by: Pavol Ipoth <pavol.ipoth@external.t-systems.com>\r\nCo-authored-by: Jonas <57955592+JonasHeinemann@users.noreply.github.com>\r\nCo-authored-by: Jonas Heinemann <jonas.heinemann@clarilab.de>\r\nCo-authored-by: WilliPkv <105049959+WilliPkv@users.noreply.github.com>",
          "timestamp": "2022-10-19T14:34:15+02:00",
          "tree_id": "57a9851ba20ccf27d578a783eea661ef45c22d45",
          "url": "https://github.com/Nerzal/gocloak/commit/18f0f3d14cb8dd31a49d357ca37fab0aa8ca337f"
        },
        "date": 1666183014012,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 74286812,
            "unit": "ns/op\t   65727 B/op\t     217 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 74941268,
            "unit": "ns/op\t   74548 B/op\t     221 allocs/op",
            "extra": "15 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 73772205,
            "unit": "ns/op\t   68953 B/op\t     217 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 41933081,
            "unit": "ns/op\t   65543 B/op\t     219 allocs/op",
            "extra": "26 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "bastian.ike@aoe.com",
            "name": "Bastian",
            "username": "bastianccm"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "c59ba6ef1f9f187bbb0f80cbda56222a36e77ac5",
          "message": "fix(login): specify openid scope (#396)",
          "timestamp": "2023-01-05T14:34:40+01:00",
          "tree_id": "4c9d18253aa9e9ef2e0d02365a25f825d8375e78",
          "url": "https://github.com/Nerzal/gocloak/commit/c59ba6ef1f9f187bbb0f80cbda56222a36e77ac5"
        },
        "date": 1672925832182,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 76115197,
            "unit": "ns/op\t   60420 B/op\t     226 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 82573459,
            "unit": "ns/op\t   63584 B/op\t     226 allocs/op",
            "extra": "15 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 74376821,
            "unit": "ns/op\t   55658 B/op\t     227 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 46202054,
            "unit": "ns/op\t   69344 B/op\t     228 allocs/op",
            "extra": "25 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "amir4895@gmail.com",
            "name": "amir4895",
            "username": "amir4895"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "94bb9404b0989b12ec89f0ea72f5de735f6d7fa1",
          "message": "Adding new function GetUserBruteForceDetectionStatus to retrieve the … (#388)\n\n* Adding new function GetUserBruteForceDetectionStatus to retrieve the brute force protection status per user\r\n\r\n* code convention\r\n\r\n* code convention\r\n\r\n* code convention\r\n\r\nCo-authored-by: amirlevinzon <amir.levinzon@broadcom.com>",
          "timestamp": "2023-01-09T14:28:25+01:00",
          "tree_id": "a87c088c8e5612cfe09f624608c3042e931c60a2",
          "url": "https://github.com/Nerzal/gocloak/commit/94bb9404b0989b12ec89f0ea72f5de735f6d7fa1"
        },
        "date": 1673271054405,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 76274048,
            "unit": "ns/op\t   73196 B/op\t     229 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 76172509,
            "unit": "ns/op\t   55790 B/op\t     228 allocs/op",
            "extra": "15 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 75253156,
            "unit": "ns/op\t   77205 B/op\t     227 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 41053034,
            "unit": "ns/op\t   70032 B/op\t     227 allocs/op",
            "extra": "26 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "5464641+upils@users.noreply.github.com",
            "name": "Upils",
            "username": "upils"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "b2ed4a59a4cbf93f274f014375b1eebdf1216334",
          "message": "Add params to GetUsersByRoleName (#392)\n\n* Add params to GetUsersByRoleName\r\n\r\nSigned-off-by: Paul MARS <paul.mars@intrinsec.com>\r\n\r\n* ✅ Add a test case on GetUsersByRoleName\r\n\r\nSigned-off-by: Paul MARS <paul.mars@intrinsec.com>\r\n\r\n* 📌 Pin keycloak version in test to avoid breaking test in newly release keycloak version\r\n\r\nSigned-off-by: Paul MARS <paul.mars@intrinsec.com>\r\n\r\n---------\r\n\r\nSigned-off-by: Paul MARS <paul.mars@intrinsec.com>",
          "timestamp": "2023-02-15T14:01:56+01:00",
          "tree_id": "457cd37974180e12a10df203736eb0cd27aeb9d1",
          "url": "https://github.com/Nerzal/gocloak/commit/b2ed4a59a4cbf93f274f014375b1eebdf1216334"
        },
        "date": 1676466274761,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 77661491,
            "unit": "ns/op\t   60274 B/op\t     227 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 78082637,
            "unit": "ns/op\t   73639 B/op\t     229 allocs/op",
            "extra": "14 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 77081854,
            "unit": "ns/op\t   70384 B/op\t     229 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 42996120,
            "unit": "ns/op\t   67836 B/op\t     227 allocs/op",
            "extra": "25 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "tt@fino.digital",
            "name": "Tobias Theel"
          },
          "committer": {
            "email": "tt@fino.digital",
            "name": "Tobias Theel"
          },
          "distinct": true,
          "id": "9cd3f14839feea38219224ab49072e6dc71ca1e3",
          "message": "update to glocloak v13",
          "timestamp": "2023-02-22T12:09:23+01:00",
          "tree_id": "47888c650177f3f73463c60e5e25110fd79f6520",
          "url": "https://github.com/Nerzal/gocloak/commit/9cd3f14839feea38219224ab49072e6dc71ca1e3"
        },
        "date": 1677064329488,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 77644867,
            "unit": "ns/op\t   54959 B/op\t     227 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 77790144,
            "unit": "ns/op\t   66729 B/op\t     228 allocs/op",
            "extra": "14 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 76842708,
            "unit": "ns/op\t   68270 B/op\t     229 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 42967921,
            "unit": "ns/op\t   61788 B/op\t     226 allocs/op",
            "extra": "27 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "36693488+r3st@users.noreply.github.com",
            "name": "r3st",
            "username": "r3st"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "b703322166ecf2bdbdd4c128847f4cb6160f4622",
          "message": "fix(model): wrong value type (#402)\n\nCo-authored-by: Rene Stolle <r.stolle@prosoz.de>",
          "timestamp": "2023-02-28T12:18:24+01:00",
          "tree_id": "390c7b1cbaf18793da25b954cef67ddeb22e2474",
          "url": "https://github.com/Nerzal/gocloak/commit/b703322166ecf2bdbdd4c128847f4cb6160f4622"
        },
        "date": 1677583255449,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 76573257,
            "unit": "ns/op\t   60176 B/op\t     226 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 77994998,
            "unit": "ns/op\t   80860 B/op\t     228 allocs/op",
            "extra": "14 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 76819466,
            "unit": "ns/op\t   78792 B/op\t     229 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 42486784,
            "unit": "ns/op\t   71974 B/op\t     227 allocs/op",
            "extra": "27 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "49699333+dependabot[bot]@users.noreply.github.com",
            "name": "dependabot[bot]",
            "username": "dependabot[bot]"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "27921c7833f07a6752ba1d555ab35950bc929167",
          "message": "Bump golang.org/x/net from 0.0.0-20221019024206-cb67ada4b0ad to 0.7.0 (#401)\n\nBumps [golang.org/x/net](https://github.com/golang/net) from 0.0.0-20221019024206-cb67ada4b0ad to 0.7.0.\r\n- [Release notes](https://github.com/golang/net/releases)\r\n- [Commits](https://github.com/golang/net/commits/v0.7.0)\r\n\r\n---\r\nupdated-dependencies:\r\n- dependency-name: golang.org/x/net\r\n  dependency-type: indirect\r\n...\r\n\r\nSigned-off-by: dependabot[bot] <support@github.com>\r\nCo-authored-by: dependabot[bot] <49699333+dependabot[bot]@users.noreply.github.com>",
          "timestamp": "2023-02-28T12:19:05+01:00",
          "tree_id": "095faa3e3b5d9afd6345fef73a9632bb79b6fa21",
          "url": "https://github.com/Nerzal/gocloak/commit/27921c7833f07a6752ba1d555ab35950bc929167"
        },
        "date": 1677583313906,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 74114402,
            "unit": "ns/op\t   67688 B/op\t     228 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 75596020,
            "unit": "ns/op\t   62467 B/op\t     229 allocs/op",
            "extra": "15 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 74918498,
            "unit": "ns/op\t   63554 B/op\t     228 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 41564602,
            "unit": "ns/op\t   72621 B/op\t     229 allocs/op",
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
          "id": "846cfda50dba49abe432cd092f70d4124c34a6ac",
          "message": "update deps (#404)\n\nCo-authored-by: Tobias Theel <tt@fino.digital>",
          "timestamp": "2023-02-28T12:41:43+01:00",
          "tree_id": "1fa28819bab2c198a7b4d15da778f6b39d0376f5",
          "url": "https://github.com/Nerzal/gocloak/commit/846cfda50dba49abe432cd092f70d4124c34a6ac"
        },
        "date": 1677584664524,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 78963475,
            "unit": "ns/op\t   62100 B/op\t     225 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 80138766,
            "unit": "ns/op\t   75806 B/op\t     231 allocs/op",
            "extra": "14 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 79611924,
            "unit": "ns/op\t   61514 B/op\t     225 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 44791599,
            "unit": "ns/op\t   58525 B/op\t     225 allocs/op",
            "extra": "27 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "19890779+mattkeeler@users.noreply.github.com",
            "name": "mattkeeler",
            "username": "mattkeeler"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "64210723be229939e149f15e42616b7b77b74331",
          "message": "Fix token introspection example (#408)",
          "timestamp": "2023-03-21T11:17:02+01:00",
          "tree_id": "3379656686e9de61925b03da962fff183ae54ae8",
          "url": "https://github.com/Nerzal/gocloak/commit/64210723be229939e149f15e42616b7b77b74331"
        },
        "date": 1679393989978,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 80804662,
            "unit": "ns/op\t   67532 B/op\t     227 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 80300871,
            "unit": "ns/op\t   76770 B/op\t     228 allocs/op",
            "extra": "14 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 79615380,
            "unit": "ns/op\t   79108 B/op\t     226 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 44122598,
            "unit": "ns/op\t   67671 B/op\t     226 allocs/op",
            "extra": "26 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "ozman4@gmail.com",
            "name": "Mike",
            "username": "osmian"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "a521bf02a43a487b0a3212aa3273eb5d15012862",
          "message": "initial (#412)\n\nCo-authored-by: Mike Osmian <mikeosmian@Mikes-MacBook-Pro-2.local>",
          "timestamp": "2023-04-05T12:39:55+02:00",
          "tree_id": "8a5814d4b11c3bb3bb34dd4e17653b8ee04ae9b7",
          "url": "https://github.com/Nerzal/gocloak/commit/a521bf02a43a487b0a3212aa3273eb5d15012862"
        },
        "date": 1680691376598,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 71605610,
            "unit": "ns/op\t   52905 B/op\t     227 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 71496649,
            "unit": "ns/op\t   83448 B/op\t     229 allocs/op",
            "extra": "15 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 70225929,
            "unit": "ns/op\t   63473 B/op\t     225 allocs/op",
            "extra": "16 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 41413837,
            "unit": "ns/op\t   56668 B/op\t     227 allocs/op",
            "extra": "25 times\n2 procs"
          },
          {
            "name": "BenchmarkGetGroups",
            "value": 2135576,
            "unit": "ns/op\t   36107 B/op\t      97 allocs/op",
            "extra": "549 times"
          },
          {
            "name": "BenchmarkGetGroups",
            "value": 2301093,
            "unit": "ns/op\t   35344 B/op\t      97 allocs/op",
            "extra": "523 times\n2 procs"
          },
          {
            "name": "BenchmarkGetGroupsFull",
            "value": 2184702,
            "unit": "ns/op\t   36880 B/op\t     119 allocs/op",
            "extra": "600 times"
          },
          {
            "name": "BenchmarkGetGroupsFull",
            "value": 2187924,
            "unit": "ns/op\t   37094 B/op\t     119 allocs/op",
            "extra": "572 times\n2 procs"
          },
          {
            "name": "BenchmarkGetGroupsBrief",
            "value": 2206436,
            "unit": "ns/op\t   37129 B/op\t     119 allocs/op",
            "extra": "603 times"
          },
          {
            "name": "BenchmarkGetGroupsBrief",
            "value": 2288313,
            "unit": "ns/op\t   37120 B/op\t     119 allocs/op",
            "extra": "579 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "19890779+mattkeeler@users.noreply.github.com",
            "name": "mattkeeler",
            "username": "mattkeeler"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "21898478be85b381d9c3aa9ebecf66ae21da02bb",
          "message": "add missing fields to GetGroupsParams struct (#413)",
          "timestamp": "2023-04-06T16:50:40+02:00",
          "tree_id": "587f3813c32640e768f189327e15e34e7cd59c2d",
          "url": "https://github.com/Nerzal/gocloak/commit/21898478be85b381d9c3aa9ebecf66ae21da02bb"
        },
        "date": 1680792863214,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 64658432,
            "unit": "ns/op\t   66165 B/op\t     227 allocs/op",
            "extra": "18 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 65802960,
            "unit": "ns/op\t   71792 B/op\t     229 allocs/op",
            "extra": "18 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 64655336,
            "unit": "ns/op\t   56603 B/op\t     224 allocs/op",
            "extra": "18 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 35182746,
            "unit": "ns/op\t   67524 B/op\t     224 allocs/op",
            "extra": "32 times\n2 procs"
          },
          {
            "name": "BenchmarkGetGroups",
            "value": 1852328,
            "unit": "ns/op\t   36268 B/op\t      97 allocs/op",
            "extra": "564 times"
          },
          {
            "name": "BenchmarkGetGroups",
            "value": 1808766,
            "unit": "ns/op\t   35584 B/op\t      97 allocs/op",
            "extra": "658 times\n2 procs"
          },
          {
            "name": "BenchmarkGetGroupsFull",
            "value": 1770875,
            "unit": "ns/op\t   39523 B/op\t     119 allocs/op",
            "extra": "663 times"
          },
          {
            "name": "BenchmarkGetGroupsFull",
            "value": 1853824,
            "unit": "ns/op\t   38840 B/op\t     119 allocs/op",
            "extra": "667 times\n2 procs"
          },
          {
            "name": "BenchmarkGetGroupsBrief",
            "value": 1735783,
            "unit": "ns/op\t   38698 B/op\t     119 allocs/op",
            "extra": "739 times"
          },
          {
            "name": "BenchmarkGetGroupsBrief",
            "value": 1662970,
            "unit": "ns/op\t   36268 B/op\t     119 allocs/op",
            "extra": "724 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "ozman4@gmail.com",
            "name": "Mike",
            "username": "osmian"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "ff2d01beda367e68a08a55722a355988bf027821",
          "message": "get group and get group by path benchmarks (#414)\n\nCo-authored-by: Mike Osmian <mikeosmian@Mikes-MacBook-Pro-2.local>",
          "timestamp": "2023-04-13T12:35:36+02:00",
          "tree_id": "e46130ec31ab828d239c713d762a3f7e80939cec",
          "url": "https://github.com/Nerzal/gocloak/commit/ff2d01beda367e68a08a55722a355988bf027821"
        },
        "date": 1681382306331,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 76492679,
            "unit": "ns/op\t   70417 B/op\t     227 allocs/op",
            "extra": "15 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 77244845,
            "unit": "ns/op\t   74266 B/op\t     230 allocs/op",
            "extra": "14 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 76252288,
            "unit": "ns/op\t   78563 B/op\t     227 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 42450136,
            "unit": "ns/op\t   67171 B/op\t     226 allocs/op",
            "extra": "28 times\n2 procs"
          },
          {
            "name": "BenchmarkGetGroups",
            "value": 1654961,
            "unit": "ns/op\t   35814 B/op\t      96 allocs/op",
            "extra": "746 times"
          },
          {
            "name": "BenchmarkGetGroups",
            "value": 1613386,
            "unit": "ns/op\t   35066 B/op\t      97 allocs/op",
            "extra": "686 times\n2 procs"
          },
          {
            "name": "BenchmarkGetGroupsFull",
            "value": 1608134,
            "unit": "ns/op\t   37441 B/op\t     119 allocs/op",
            "extra": "718 times"
          },
          {
            "name": "BenchmarkGetGroupsFull",
            "value": 1584946,
            "unit": "ns/op\t   36629 B/op\t     119 allocs/op",
            "extra": "756 times\n2 procs"
          },
          {
            "name": "BenchmarkGetGroupsBrief",
            "value": 1507759,
            "unit": "ns/op\t   38182 B/op\t     119 allocs/op",
            "extra": "784 times"
          },
          {
            "name": "BenchmarkGetGroupsBrief",
            "value": 1560392,
            "unit": "ns/op\t   38369 B/op\t     119 allocs/op",
            "extra": "766 times\n2 procs"
          },
          {
            "name": "BenchmarkGetGroup",
            "value": 1604832,
            "unit": "ns/op\t   35743 B/op\t     135 allocs/op",
            "extra": "741 times"
          },
          {
            "name": "BenchmarkGetGroup",
            "value": 1589651,
            "unit": "ns/op\t   37355 B/op\t     135 allocs/op",
            "extra": "727 times\n2 procs"
          },
          {
            "name": "BenchmarkGetGroupByPath",
            "value": 1373990,
            "unit": "ns/op\t   37010 B/op\t     122 allocs/op",
            "extra": "801 times"
          },
          {
            "name": "BenchmarkGetGroupByPath",
            "value": 1296871,
            "unit": "ns/op\t   37495 B/op\t     122 allocs/op",
            "extra": "901 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "307427+mweibel@users.noreply.github.com",
            "name": "Michael Weibel",
            "username": "mweibel"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "781427a6df84d524cb60d9aa7d0f1c410eb130df",
          "message": "feat: expose getRequest* methods for custom endpoints (#405)\n\nkeycloak is able to have extensions and it would be nice to not have to\r\nconstruct everything againn and just use a custom endpoint.\r\n\r\nfixes #219",
          "timestamp": "2023-04-13T12:36:45+02:00",
          "tree_id": "4000e46df4ef3091643ba6ac61a42d4d61f71d93",
          "url": "https://github.com/Nerzal/gocloak/commit/781427a6df84d524cb60d9aa7d0f1c410eb130df"
        },
        "date": 1681382370253,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 76106122,
            "unit": "ns/op\t   78183 B/op\t     228 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 76299014,
            "unit": "ns/op\t   67734 B/op\t     229 allocs/op",
            "extra": "14 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 76508863,
            "unit": "ns/op\t   73548 B/op\t     229 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 44427758,
            "unit": "ns/op\t   62398 B/op\t     226 allocs/op",
            "extra": "28 times\n2 procs"
          },
          {
            "name": "BenchmarkGetGroups",
            "value": 1741646,
            "unit": "ns/op\t   37077 B/op\t      97 allocs/op",
            "extra": "578 times"
          },
          {
            "name": "BenchmarkGetGroups",
            "value": 1684831,
            "unit": "ns/op\t   37416 B/op\t      97 allocs/op",
            "extra": "644 times\n2 procs"
          },
          {
            "name": "BenchmarkGetGroupsFull",
            "value": 1697278,
            "unit": "ns/op\t   38516 B/op\t     119 allocs/op",
            "extra": "747 times"
          },
          {
            "name": "BenchmarkGetGroupsFull",
            "value": 1671338,
            "unit": "ns/op\t   36081 B/op\t     119 allocs/op",
            "extra": "704 times\n2 procs"
          },
          {
            "name": "BenchmarkGetGroupsBrief",
            "value": 1648739,
            "unit": "ns/op\t   36491 B/op\t     119 allocs/op",
            "extra": "760 times"
          },
          {
            "name": "BenchmarkGetGroupsBrief",
            "value": 1580584,
            "unit": "ns/op\t   41676 B/op\t     119 allocs/op",
            "extra": "726 times\n2 procs"
          },
          {
            "name": "BenchmarkGetGroup",
            "value": 1698369,
            "unit": "ns/op\t   40119 B/op\t     135 allocs/op",
            "extra": "666 times"
          },
          {
            "name": "BenchmarkGetGroup",
            "value": 1565256,
            "unit": "ns/op\t   37827 B/op\t     135 allocs/op",
            "extra": "776 times\n2 procs"
          },
          {
            "name": "BenchmarkGetGroupByPath",
            "value": 1382418,
            "unit": "ns/op\t   36915 B/op\t     122 allocs/op",
            "extra": "844 times"
          },
          {
            "name": "BenchmarkGetGroupByPath",
            "value": 1322585,
            "unit": "ns/op\t   37500 B/op\t     122 allocs/op",
            "extra": "856 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "tt@fino.digital",
            "name": "Tobias Theel"
          },
          "committer": {
            "email": "tt@fino.digital",
            "name": "Tobias Theel"
          },
          "distinct": true,
          "id": "de1cf0b4d7298bbc47da5e74edc5f815e8740e66",
          "message": "fix lint issue",
          "timestamp": "2023-04-13T13:23:56+02:00",
          "tree_id": "dd843330c9e25f8b9caa1e1d62b230d4773e9a1e",
          "url": "https://github.com/Nerzal/gocloak/commit/de1cf0b4d7298bbc47da5e74edc5f815e8740e66"
        },
        "date": 1681385202112,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 65856158,
            "unit": "ns/op\t   82211 B/op\t     227 allocs/op",
            "extra": "18 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 68576800,
            "unit": "ns/op\t   70032 B/op\t     227 allocs/op",
            "extra": "18 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 65068206,
            "unit": "ns/op\t   67699 B/op\t     227 allocs/op",
            "extra": "18 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 36172440,
            "unit": "ns/op\t   56201 B/op\t     225 allocs/op",
            "extra": "33 times\n2 procs"
          },
          {
            "name": "BenchmarkGetGroups",
            "value": 2044875,
            "unit": "ns/op\t   36781 B/op\t      97 allocs/op",
            "extra": "663 times"
          },
          {
            "name": "BenchmarkGetGroups",
            "value": 1929172,
            "unit": "ns/op\t   37070 B/op\t      97 allocs/op",
            "extra": "687 times\n2 procs"
          },
          {
            "name": "BenchmarkGetGroupsFull",
            "value": 1884471,
            "unit": "ns/op\t   37936 B/op\t     119 allocs/op",
            "extra": "703 times"
          },
          {
            "name": "BenchmarkGetGroupsFull",
            "value": 1758944,
            "unit": "ns/op\t   37553 B/op\t     119 allocs/op",
            "extra": "646 times\n2 procs"
          },
          {
            "name": "BenchmarkGetGroupsBrief",
            "value": 1700164,
            "unit": "ns/op\t   38514 B/op\t     119 allocs/op",
            "extra": "711 times"
          },
          {
            "name": "BenchmarkGetGroupsBrief",
            "value": 1709772,
            "unit": "ns/op\t   37062 B/op\t     119 allocs/op",
            "extra": "703 times\n2 procs"
          },
          {
            "name": "BenchmarkGetGroup",
            "value": 1836242,
            "unit": "ns/op\t   38772 B/op\t     135 allocs/op",
            "extra": "654 times"
          },
          {
            "name": "BenchmarkGetGroup",
            "value": 1796161,
            "unit": "ns/op\t   39097 B/op\t     135 allocs/op",
            "extra": "662 times\n2 procs"
          },
          {
            "name": "BenchmarkGetGroupByPath",
            "value": 1540371,
            "unit": "ns/op\t   36355 B/op\t     122 allocs/op",
            "extra": "726 times"
          },
          {
            "name": "BenchmarkGetGroupByPath",
            "value": 1712400,
            "unit": "ns/op\t   37711 B/op\t     122 allocs/op",
            "extra": "740 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "bastian.ike@aoe.com",
            "name": "Bastian",
            "username": "bastianccm"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "70f6ad9988a208b1ee57cd053af1c7ea6506bd40",
          "message": "fix(client): correct serverinfo endpoint (#382)",
          "timestamp": "2023-05-03T13:54:00+02:00",
          "tree_id": "6716a7b963fcd923c1c8e4b021825fe7404d5476",
          "url": "https://github.com/Nerzal/gocloak/commit/70f6ad9988a208b1ee57cd053af1c7ea6506bd40"
        },
        "date": 1683115058065,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 84420872,
            "unit": "ns/op\t   74837 B/op\t     227 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 83510846,
            "unit": "ns/op\t   72036 B/op\t     229 allocs/op",
            "extra": "13 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 82095965,
            "unit": "ns/op\t   51410 B/op\t     228 allocs/op",
            "extra": "13 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 48752230,
            "unit": "ns/op\t   72895 B/op\t     227 allocs/op",
            "extra": "25 times\n2 procs"
          },
          {
            "name": "BenchmarkGetGroups",
            "value": 2665099,
            "unit": "ns/op\t   37812 B/op\t      97 allocs/op",
            "extra": "448 times"
          },
          {
            "name": "BenchmarkGetGroups",
            "value": 3006411,
            "unit": "ns/op\t   37814 B/op\t      97 allocs/op",
            "extra": "478 times\n2 procs"
          },
          {
            "name": "BenchmarkGetGroupsFull",
            "value": 2445401,
            "unit": "ns/op\t   36674 B/op\t     119 allocs/op",
            "extra": "501 times"
          },
          {
            "name": "BenchmarkGetGroupsFull",
            "value": 2595469,
            "unit": "ns/op\t   37334 B/op\t     119 allocs/op",
            "extra": "603 times\n2 procs"
          },
          {
            "name": "BenchmarkGetGroupsBrief",
            "value": 2418232,
            "unit": "ns/op\t   37693 B/op\t     119 allocs/op",
            "extra": "516 times"
          },
          {
            "name": "BenchmarkGetGroupsBrief",
            "value": 2519735,
            "unit": "ns/op\t   36686 B/op\t     119 allocs/op",
            "extra": "519 times\n2 procs"
          },
          {
            "name": "BenchmarkGetGroup",
            "value": 2367654,
            "unit": "ns/op\t   36403 B/op\t     135 allocs/op",
            "extra": "489 times"
          },
          {
            "name": "BenchmarkGetGroup",
            "value": 2599435,
            "unit": "ns/op\t   38331 B/op\t     135 allocs/op",
            "extra": "482 times\n2 procs"
          },
          {
            "name": "BenchmarkGetGroupByPath",
            "value": 2277134,
            "unit": "ns/op\t   35390 B/op\t     122 allocs/op",
            "extra": "544 times"
          },
          {
            "name": "BenchmarkGetGroupByPath",
            "value": 2231825,
            "unit": "ns/op\t   35599 B/op\t     122 allocs/op",
            "extra": "567 times\n2 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "brendan@glaulabs.com",
            "name": "Brendan Le Glaunec",
            "username": "Ullaakut"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "640934dd78692930603e556a389a4a6f660d4ac9",
          "message": "feat: add support for resource server settings endpoint (#421)\n\n* feat: add support for resource server settings endpoint\r\n\r\n* fix: endpoint URL\r\n\r\n* test: add unit test",
          "timestamp": "2023-05-04T13:47:47+02:00",
          "tree_id": "76fe866e7dd8d2a1c476db8fa81a7cb30693610b",
          "url": "https://github.com/Nerzal/gocloak/commit/640934dd78692930603e556a389a4a6f660d4ac9"
        },
        "date": 1683201040532,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLogin",
            "value": 76248611,
            "unit": "ns/op\t   74022 B/op\t     228 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLogin",
            "value": 77325327,
            "unit": "ns/op\t   72413 B/op\t     229 allocs/op",
            "extra": "14 times\n2 procs"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 77116871,
            "unit": "ns/op\t   75586 B/op\t     229 allocs/op",
            "extra": "14 times"
          },
          {
            "name": "BenchmarkLoginParallel",
            "value": 42243073,
            "unit": "ns/op\t   77293 B/op\t     227 allocs/op",
            "extra": "26 times\n2 procs"
          },
          {
            "name": "BenchmarkGetGroups",
            "value": 1643343,
            "unit": "ns/op\t   36730 B/op\t      97 allocs/op",
            "extra": "807 times"
          },
          {
            "name": "BenchmarkGetGroups",
            "value": 1529090,
            "unit": "ns/op\t   37629 B/op\t      97 allocs/op",
            "extra": "716 times\n2 procs"
          },
          {
            "name": "BenchmarkGetGroupsFull",
            "value": 1527224,
            "unit": "ns/op\t   37972 B/op\t     119 allocs/op",
            "extra": "793 times"
          },
          {
            "name": "BenchmarkGetGroupsFull",
            "value": 1509692,
            "unit": "ns/op\t   37253 B/op\t     119 allocs/op",
            "extra": "788 times\n2 procs"
          },
          {
            "name": "BenchmarkGetGroupsBrief",
            "value": 1485096,
            "unit": "ns/op\t   37260 B/op\t     119 allocs/op",
            "extra": "703 times"
          },
          {
            "name": "BenchmarkGetGroupsBrief",
            "value": 1533235,
            "unit": "ns/op\t   36521 B/op\t     119 allocs/op",
            "extra": "830 times\n2 procs"
          },
          {
            "name": "BenchmarkGetGroup",
            "value": 1563716,
            "unit": "ns/op\t   39127 B/op\t     135 allocs/op",
            "extra": "703 times"
          },
          {
            "name": "BenchmarkGetGroup",
            "value": 1543812,
            "unit": "ns/op\t   36559 B/op\t     135 allocs/op",
            "extra": "657 times\n2 procs"
          },
          {
            "name": "BenchmarkGetGroupByPath",
            "value": 1362603,
            "unit": "ns/op\t   36992 B/op\t     122 allocs/op",
            "extra": "885 times"
          },
          {
            "name": "BenchmarkGetGroupByPath",
            "value": 1331493,
            "unit": "ns/op\t   38007 B/op\t     122 allocs/op",
            "extra": "878 times\n2 procs"
          }
        ]
      }
    ]
  }
}