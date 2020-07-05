window.BENCHMARK_DATA = {
  "lastUpdate": 1593946169927,
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
      }
    ]
  }
}