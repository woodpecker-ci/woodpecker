// Copyright 2018 Drone.IO Inc.
// Copyright 2022 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fixtures

const HookPush = `
{
  "actor": {
    "display_name": "Martin Herren",
    "links": {
      "self": {
        "href": "https://api.bitbucket.org/2.0/users/%7Bc5a0d676-fd27-4bd4-ac69-a7540d7b495b%7D"
      },
      "avatar": {
        "href": "https://secure.gravatar.com/avatar/37de364488b2ec474b5458ca86442bbb?d=https%3A%2F%2Favatar-management--avatars.us-west-2.prod.public.atl-paas.net%2Finitials%2FMH-2.png"
      },
      "html": {
        "href": "https://bitbucket.org/%7Bc5a0d676-fd27-4bd4-ac69-a7540d7b495b%7D/"
      }
    },
    "type": "user",
    "uuid": "{c5a0d676-fd27-4bd4-ac69-a7540d7b495b}",
    "account_id": "5cf8e3a9678ca90f8e7cc8a8",
    "nickname": "Martin Herren"
  },
  "repository": {
    "type": "repository",
    "full_name": "martinherren1984/publictestrepo",
    "links": {
      "self": {
        "href": "https://api.bitbucket.org/2.0/repositories/martinherren1984/publictestrepo"
      },
      "html": {
        "href": "https://bitbucket.org/martinherren1984/publictestrepo"
      },
      "avatar": {
        "href": "https://bytebucket.org/ravatar/%7B898477b2-a080-4089-b385-597a783db392%7D?ts=default"
      }
    },
    "name": "PublicTestRepo",
    "scm": "git",
    "website": null,
    "owner": {
      "display_name": "Martin Herren",
      "links": {
        "self": {
          "href": "https://api.bitbucket.org/2.0/users/%7Bc5a0d676-fd27-4bd4-ac69-a7540d7b495b%7D"
        },
        "avatar": {
          "href": "https://secure.gravatar.com/avatar/37de364488b2ec474b5458ca86442bbb?d=https%3A%2F%2Favatar-management--avatars.us-west-2.prod.public.atl-paas.net%2Finitials%2FMH-2.png"
        },
        "html": {
          "href": "https://bitbucket.org/%7Bc5a0d676-fd27-4bd4-ac69-a7540d7b495b%7D/"
        }
      },
      "type": "user",
      "uuid": "{c5a0d676-fd27-4bd4-ac69-a7540d7b495b}",
      "account_id": "5cf8e3a9678ca90f8e7cc8a8",
      "nickname": "Martin Herren"
    },
    "workspace": {
      "type": "workspace",
      "uuid": "{c5a0d676-fd27-4bd4-ac69-a7540d7b495b}",
      "name": "Martin Herren",
      "slug": "martinherren1984",
      "links": {
        "avatar": {
          "href": "https://bitbucket.org/workspaces/martinherren1984/avatar/?ts=1658761964"
        },
        "html": {
          "href": "https://bitbucket.org/martinherren1984/"
        },
        "self": {
          "href": "https://api.bitbucket.org/2.0/workspaces/martinherren1984"
        }
      }
    },
    "is_private": false,
    "project": {
      "type": "project",
      "key": "PUB",
      "uuid": "{2cede481-f59e-49ec-88d0-a85629b7925d}",
      "name": "PublicTestProject",
      "links": {
        "self": {
          "href": "https://api.bitbucket.org/2.0/workspaces/martinherren1984/projects/PUB"
        },
        "html": {
          "href": "https://bitbucket.org/martinherren1984/workspace/projects/PUB"
        },
        "avatar": {
          "href": "https://bitbucket.org/account/user/martinherren1984/projects/PUB/avatar/32?ts=1658768453"
        }
      }
    },
    "uuid": "{898477b2-a080-4089-b385-597a783db392}"
  },
  "push": {
    "changes": [
      {
        "old": {
          "name": "main",
          "target": {
            "type": "commit",
            "hash": "a51241ae1f00cbe728930db48e890b18fd527f99",
            "date": "2022-08-17T15:24:29+00:00",
            "author": {
              "type": "author",
              "raw": "Martin Herren <martin.herren@xxx.com>",
              "user": {
                "display_name": "Martin Herren",
                "links": {
                  "self": {
                    "href": "https://api.bitbucket.org/2.0/users/%7B69cc59f2-706b-4a9c-b99c-eac2ace320da%7D"
                  },
                  "avatar": {
                    "href": "https://secure.gravatar.com/avatar/7b2e50690b4ab7bb9e1db18ea3b8ae95?d=https%3A%2F%2Favatar-management--avatars.us-west-2.prod.public.atl-paas.net%2Finitials%2FMH-5.png"
                  },
                  "html": {
                    "href": "https://bitbucket.org/%7B69cc59f2-706b-4a9c-b99c-eac2ace320da%7D/"
                  }
                },
                "type": "user",
                "uuid": "{69cc59f2-706b-4a9c-b99c-eac2ace320da}",
                "account_id": "5d286e857133f10c17e026cb",
                "nickname": "Martin Herren"
              }
            },
            "message": "Add test .woodpecker.yml\n",
            "summary": {
              "type": "rendered",
              "raw": "Add test .woodpecker.yml\n",
              "markup": "markdown",
              "html": "<p>Add test .woodpecker.yml</p>"
            },
            "links": {
              "self": {
                "href": "https://api.bitbucket.org/2.0/repositories/martinherren1984/publictestrepo/commit/a51241ae1f00cbe728930db48e890b18fd527f99"
              },
              "html": {
                "href": "https://bitbucket.org/martinherren1984/publictestrepo/commits/a51241ae1f00cbe728930db48e890b18fd527f99"
              }
            },
            "parents": [],
            "rendered": {},
            "properties": {}
          },
          "links": {
            "self": {
              "href": "https://api.bitbucket.org/2.0/repositories/martinherren1984/publictestrepo/refs/branches/main"
            },
            "commits": {
              "href": "https://api.bitbucket.org/2.0/repositories/martinherren1984/publictestrepo/commits/main"
            },
            "html": {
              "href": "https://bitbucket.org/martinherren1984/publictestrepo/branch/main"
            }
          },
          "type": "branch",
          "merge_strategies": [
            "merge_commit",
            "squash",
            "fast_forward"
          ],
          "default_merge_strategy": "merge_commit"
        },
        "new": {
          "name": "main",
          "target": {
            "type": "commit",
            "hash": "c14c1bb05dfb1fdcdf06b31485fff61b0ea44277",
            "date": "2022-09-07T20:19:25+00:00",
            "author": {
              "type": "author",
              "raw": "Martin Herren <martin.herren@yyy.com>",
              "user": {
                "display_name": "Martin Herren",
                "links": {
                  "self": {
                    "href": "https://api.bitbucket.org/2.0/users/%7Bc5a0d676-fd27-4bd4-ac69-a7540d7b495b%7D"
                  },
                  "avatar": {
                    "href": "https://secure.gravatar.com/avatar/37de364488b2ec474b5458ca86442bbb?d=https%3A%2F%2Favatar-management--avatars.us-west-2.prod.public.atl-paas.net%2Finitials%2FMH-2.png"
                  },
                  "html": {
                    "href": "https://bitbucket.org/%7Bc5a0d676-fd27-4bd4-ac69-a7540d7b495b%7D/"
                  }
                },
                "type": "user",
                "uuid": "{c5a0d676-fd27-4bd4-ac69-a7540d7b495b}",
                "account_id": "5cf8e3a9678ca90f8e7cc8a8",
                "nickname": "Martin Herren"
              }
            },
            "message": "a\n",
            "summary": {
              "type": "rendered",
              "raw": "a\n",
              "markup": "markdown",
              "html": "<p>a</p>"
            },
            "links": {
              "self": {
                "href": "https://api.bitbucket.org/2.0/repositories/martinherren1984/publictestrepo/commit/c14c1bb05dfb1fdcdf06b31485fff61b0ea44277"
              },
              "html": {
                "href": "https://bitbucket.org/martinherren1984/publictestrepo/commits/c14c1bb05dfb1fdcdf06b31485fff61b0ea44277"
              }
            },
            "parents": [
              {
                "type": "commit",
                "hash": "a51241ae1f00cbe728930db48e890b18fd527f99",
                "links": {
                  "self": {
                    "href": "https://api.bitbucket.org/2.0/repositories/martinherren1984/publictestrepo/commit/a51241ae1f00cbe728930db48e890b18fd527f99"
                  },
                  "html": {
                    "href": "https://bitbucket.org/martinherren1984/publictestrepo/commits/a51241ae1f00cbe728930db48e890b18fd527f99"
                  }
                }
              }
            ],
            "rendered": {},
            "properties": {}
          },
          "links": {
            "self": {
              "href": "https://api.bitbucket.org/2.0/repositories/martinherren1984/publictestrepo/refs/branches/main"
            },
            "commits": {
              "href": "https://api.bitbucket.org/2.0/repositories/martinherren1984/publictestrepo/commits/main"
            },
            "html": {
              "href": "https://bitbucket.org/martinherren1984/publictestrepo/branch/main"
            }
          },
          "type": "branch",
          "merge_strategies": [
            "merge_commit",
            "squash",
            "fast_forward"
          ],
          "default_merge_strategy": "merge_commit"
        },
        "truncated": false,
        "created": false,
        "forced": false,
        "closed": false,
        "links": {
          "commits": {
            "href": "https://api.bitbucket.org/2.0/repositories/martinherren1984/publictestrepo/commits?include=c14c1bb05dfb1fdcdf06b31485fff61b0ea44277&exclude=a51241ae1f00cbe728930db48e890b18fd527f99"
          },
          "diff": {
            "href": "https://api.bitbucket.org/2.0/repositories/martinherren1984/publictestrepo/diff/c14c1bb05dfb1fdcdf06b31485fff61b0ea44277..a51241ae1f00cbe728930db48e890b18fd527f99"
          },
          "html": {
            "href": "https://bitbucket.org/martinherren1984/publictestrepo/branches/compare/c14c1bb05dfb1fdcdf06b31485fff61b0ea44277..a51241ae1f00cbe728930db48e890b18fd527f99"
          }
        },
        "commits": [
          {
            "type": "commit",
            "hash": "c14c1bb05dfb1fdcdf06b31485fff61b0ea44277",
            "date": "2022-09-07T20:19:25+00:00",
            "author": {
              "type": "author",
              "raw": "Martin Herren <martin.herren@yyy.com>",
              "user": {
                "display_name": "Martin Herren",
                "links": {
                  "self": {
                    "href": "https://api.bitbucket.org/2.0/users/%7Bc5a0d676-fd27-4bd4-ac69-a7540d7b495b%7D"
                  },
                  "avatar": {
                    "href": "https://secure.gravatar.com/avatar/37de364488b2ec474b5458ca86442bbb?d=https%3A%2F%2Favatar-management--avatars.us-west-2.prod.public.atl-paas.net%2Finitials%2FMH-2.png"
                  },
                  "html": {
                    "href": "https://bitbucket.org/%7Bc5a0d676-fd27-4bd4-ac69-a7540d7b495b%7D/"
                  }
                },
                "type": "user",
                "uuid": "{c5a0d676-fd27-4bd4-ac69-a7540d7b495b}",
                "account_id": "5cf8e3a9678ca90f8e7cc8a8",
                "nickname": "Martin Herren"
              }
            },
            "message": "a\n",
            "summary": {
              "type": "rendered",
              "raw": "a\n",
              "markup": "markdown",
              "html": "<p>a</p>"
            },
            "links": {
              "self": {
                "href": "https://api.bitbucket.org/2.0/repositories/martinherren1984/publictestrepo/commit/c14c1bb05dfb1fdcdf06b31485fff61b0ea44277"
              },
              "html": {
                "href": "https://bitbucket.org/martinherren1984/publictestrepo/commits/c14c1bb05dfb1fdcdf06b31485fff61b0ea44277"
              },
              "diff": {
                "href": "https://api.bitbucket.org/2.0/repositories/martinherren1984/publictestrepo/diff/c14c1bb05dfb1fdcdf06b31485fff61b0ea44277"
              },
              "approve": {
                "href": "https://api.bitbucket.org/2.0/repositories/martinherren1984/publictestrepo/commit/c14c1bb05dfb1fdcdf06b31485fff61b0ea44277/approve"
              },
              "comments": {
                "href": "https://api.bitbucket.org/2.0/repositories/martinherren1984/publictestrepo/commit/c14c1bb05dfb1fdcdf06b31485fff61b0ea44277/comments"
              },
              "statuses": {
                "href": "https://api.bitbucket.org/2.0/repositories/martinherren1984/publictestrepo/commit/c14c1bb05dfb1fdcdf06b31485fff61b0ea44277/statuses"
              },
              "patch": {
                "href": "https://api.bitbucket.org/2.0/repositories/martinherren1984/publictestrepo/patch/c14c1bb05dfb1fdcdf06b31485fff61b0ea44277"
              }
            },
            "parents": [
              {
                "type": "commit",
                "hash": "a51241ae1f00cbe728930db48e890b18fd527f99",
                "links": {
                  "self": {
                    "href": "https://api.bitbucket.org/2.0/repositories/martinherren1984/publictestrepo/commit/a51241ae1f00cbe728930db48e890b18fd527f99"
                  },
                  "html": {
                    "href": "https://bitbucket.org/martinherren1984/publictestrepo/commits/a51241ae1f00cbe728930db48e890b18fd527f99"
                  }
                }
              }
            ],
            "rendered": {},
            "properties": {}
          }
        ]
      }
    ]
  }
}
`

const HookPushEmptyHash = `
{
  "push": {
    "changes": [
      {
        "new": {
          "type": "branch",
          "target": { "hash": "" }
        }
      }
    ]
  }
}
`

const HookPull = `
{
  "actor": {
    "username": "emmap1",
    "links": {
      "avatar": {
        "href": "https:\/\/bitbucket-api-assetroot.s3.amazonaws.com\/c\/photos\/2015\/Feb\/26\/3613917261-0-emmap1-avatar_avatar.png"
      }
    }
  },
  "pullrequest": {
    "id": 1,
    "title": "Title of pull request",
    "description": "Description of pull request",
    "state": "OPEN",
    "author": {
      "username": "emmap1",
      "links": {
        "avatar": {
          "href": "https:\/\/bitbucket-api-assetroot.s3.amazonaws.com\/c\/photos\/2015\/Feb\/26\/3613917261-0-emmap1-avatar_avatar.png"
        }
      }
    },
    "source": {
      "branch": {
        "name": "branch2"
      },
      "commit": {
        "hash": "d3022fc0ca3d"
      },
      "repository": {
        "links": {
          "html": {
            "href": "https:\/\/api.bitbucket.org\/team_name\/repo_name"
          },
          "avatar": {
            "href": "https:\/\/api-staging-assetroot.s3.amazonaws.com\/c\/photos\/2014\/Aug\/01\/bitbucket-logo-2629490769-3_avatar.png"
          }
        },
        "full_name": "user_name\/repo_name",
        "scm": "git",
        "is_private": true
      }
    },
    "destination": {
      "branch": {
        "name": "main"
      },
      "commit": {
        "hash": "ce5965ddd289"
      },
      "repository": {
        "links": {
          "html": {
            "href": "https:\/\/api.bitbucket.org\/team_name\/repo_name"
          },
          "avatar": {
            "href": "https:\/\/api-staging-assetroot.s3.amazonaws.com\/c\/photos\/2014\/Aug\/01\/bitbucket-logo-2629490769-3_avatar.png"
          }
        },
        "full_name": "user_name\/repo_name",
        "scm": "git",
        "is_private": true
      }
    },
    "links": {
      "self": {
        "href": "https:\/\/api.bitbucket.org\/api\/2.0\/pullrequests\/pullrequest_id"
      },
      "html": {
        "href": "https:\/\/api.bitbucket.org\/pullrequest_id"
      }
    }
  },
  "repository": {
    "links": {
      "html": {
        "href": "https:\/\/api.bitbucket.org\/team_name\/repo_name"
      },
      "avatar": {
        "href": "https:\/\/api-staging-assetroot.s3.amazonaws.com\/c\/photos\/2014\/Aug\/01\/bitbucket-logo-2629490769-3_avatar.png"
      }
    },
    "full_name": "user_name\/repo_name",
    "scm": "git",
    "is_private": true
  }
}
`

const HookPullRequestMerged = `
{
  "repository": {
    "type": "repository",
    "full_name": "anbraten/test-2",
    "links": {
      "self": {
        "href": "https://api.bitbucket.org/2.0/repositories/anbraten/test-2"
      },
      "html": {
        "href": "https://bitbucket.org/anbraten/test-2"
      },
      "avatar": {
        "href": "https://bytebucket.org/ravatar/%7B26554729-595f-47d1-aedd-302625cb4a97%7D?ts=default"
      }
    },
    "name": "test-2",
    "scm": "git",
    "website": null,
    "owner": {
      "display_name": "Anbraten",
      "links": {
        "self": {
          "href": "https://api.bitbucket.org/2.0/users/%7Bb1b7beef-77ca-452d-b059-fa092504ebd7%7D"
        },
        "avatar": {
          "href": "https://avatar-management--avatars.us-west-2.prod.public.atl-paas.net/70121:3046ad5f-946f-48fa-bcb4-a399eef48f0e/784add1f-95cc-42a5-a562-38a0e12de4fa/128"
        },
        "html": {
          "href": "https://bitbucket.org/%7Bb1b7beef-77ca-452d-b059-fa092504ebd7%7D/"
        }
      },
      "type": "user",
      "uuid": "{b1b7beef-77ca-452d-b059-fa092504ebd7}",
      "account_id": "70121:3046ad5f-946f-48fa-bcb4-a399eef48f0e",
      "nickname": "Anbraten"
    },
    "workspace": {
      "type": "workspace",
      "uuid": "{b1b7beef-77ca-452d-b059-fa092504ebd7}",
      "name": "Anbraten",
      "slug": "anbraten",
      "links": {
        "avatar": {
          "href": "https://bitbucket.org/workspaces/anbraten/avatar/?ts=1651865281"
        },
        "html": {
          "href": "https://bitbucket.org/anbraten/"
        },
        "self": {
          "href": "https://api.bitbucket.org/2.0/workspaces/anbraten"
        }
      }
    },
    "is_private": true,
    "project": {
      "type": "project",
      "key": "TEST",
      "uuid": "{3fa6429f-95e1-4c5a-875c-1753abcd8ace}",
      "name": "test",
      "links": {
        "self": {
          "href": "https://api.bitbucket.org/2.0/workspaces/anbraten/projects/TEST"
        },
        "html": {
          "href": "https://bitbucket.org/anbraten/workspace/projects/TEST"
        },
        "avatar": {
          "href": "https://bitbucket.org/account/user/anbraten/projects/TEST/avatar/32?ts=1690725373"
        }
      }
    },
    "uuid": "{26554729-595f-47d1-aedd-302625cb4a97}",
    "parent": null
  },
  "actor": {
    "display_name": "Anbraten",
    "links": {
      "self": {
        "href": "https://api.bitbucket.org/2.0/users/%7Bb1b7beef-77ca-452d-b059-fa092504ebd7%7D"
      },
      "avatar": {
        "href": "https://avatar-management--avatars.us-west-2.prod.public.atl-paas.net/70121:3046ad5f-946f-48fa-bcb4-a399eef48f0e/784add1f-95cc-42a5-a562-38a0e12de4fa/128"
      },
      "html": {
        "href": "https://bitbucket.org/%7Bb1b7beef-77ca-452d-b059-fa092504ebd7%7D/"
      }
    },
    "type": "user",
    "uuid": "{b1b7beef-77ca-452d-b059-fa092504ebd7}",
    "account_id": "70121:3046ad5f-946f-48fa-bcb4-a399eef48f0e",
    "nickname": "Anbraten"
  },
  "pullrequest": {
    "comment_count": 0,
    "task_count": 0,
    "type": "pullrequest",
    "id": 1,
    "title": "README.md created online with Bitbucket",
    "description": "README.md created online with Bitbucket",
    "rendered": {
      "title": {
        "type": "rendered",
        "raw": "README.md created online with Bitbucket",
        "markup": "markdown",
        "html": "<p>README.md created online with Bitbucket</p>"
      },
      "description": {
        "type": "rendered",
        "raw": "README.md created online with Bitbucket",
        "markup": "markdown",
        "html": "<p>README.md created online with Bitbucket</p>"
      }
    },
    "state": "MERGED",
    "merge_commit": {
      "type": "commit",
      "hash": "006704dbeab2",
      "links": {
        "self": {
          "href": "https://api.bitbucket.org/2.0/repositories/anbraten/test-2/commit/006704dbeab2"
        },
        "html": {
          "href": "https://bitbucket.org/anbraten/test-2/commits/006704dbeab2"
        }
      }
    },
    "close_source_branch": true,
    "closed_by": {
      "display_name": "Anbraten",
      "links": {
        "self": {
          "href": "https://api.bitbucket.org/2.0/users/%7Bb1b7beef-77ca-452d-b059-fa092504ebd7%7D"
        },
        "avatar": {
          "href": "https://avatar-management--avatars.us-west-2.prod.public.atl-paas.net/70121:3046ad5f-946f-48fa-bcb4-a399eef48f0e/784add1f-95cc-42a5-a562-38a0e12de4fa/128"
        },
        "html": {
          "href": "https://bitbucket.org/%7Bb1b7beef-77ca-452d-b059-fa092504ebd7%7D/"
        }
      },
      "type": "user",
      "uuid": "{b1b7beef-77ca-452d-b059-fa092504ebd7}",
      "account_id": "70121:3046ad5f-946f-48fa-bcb4-a399eef48f0e",
      "nickname": "Anbraten"
    },
    "author": {
      "display_name": "Anbraten",
      "links": {
        "self": {
          "href": "https://api.bitbucket.org/2.0/users/%7Bb1b7beef-77ca-452d-b059-fa092504ebd7%7D"
        },
        "avatar": {
          "href": "https://avatar-management--avatars.us-west-2.prod.public.atl-paas.net/70121:3046ad5f-946f-48fa-bcb4-a399eef48f0e/784add1f-95cc-42a5-a562-38a0e12de4fa/128"
        },
        "html": {
          "href": "https://bitbucket.org/%7Bb1b7beef-77ca-452d-b059-fa092504ebd7%7D/"
        }
      },
      "type": "user",
      "uuid": "{b1b7beef-77ca-452d-b059-fa092504ebd7}",
      "account_id": "70121:3046ad5f-946f-48fa-bcb4-a399eef48f0e",
      "nickname": "Anbraten"
    },
    "reason": "",
    "created_on": "2023-12-05T18:28:16.861881+00:00",
    "updated_on": "2023-12-05T18:29:44.785393+00:00",
    "destination": {
      "branch": {
        "name": "main"
      },
      "commit": {
        "type": "commit",
        "hash": "6c5f0bc9b2aa",
        "links": {
          "self": {
            "href": "https://api.bitbucket.org/2.0/repositories/anbraten/test-2/commit/6c5f0bc9b2aa"
          },
          "html": {
            "href": "https://bitbucket.org/anbraten/test-2/commits/6c5f0bc9b2aa"
          }
        }
      },
      "repository": {
        "type": "repository",
        "full_name": "anbraten/test-2",
        "links": {
          "self": {
            "href": "https://api.bitbucket.org/2.0/repositories/anbraten/test-2"
          },
          "html": {
            "href": "https://bitbucket.org/anbraten/test-2"
          },
          "avatar": {
            "href": "https://bytebucket.org/ravatar/%7B26554729-595f-47d1-aedd-302625cb4a97%7D?ts=default"
          }
        },
        "name": "test-2",
        "uuid": "{26554729-595f-47d1-aedd-302625cb4a97}"
      }
    },
    "source": {
      "branch": {
        "name": "patch-2"
      },
      "commit": {
        "type": "commit",
        "hash": "668218c13e04",
        "links": {
          "self": {
            "href": "https://api.bitbucket.org/2.0/repositories/anbraten/test-2/commit/668218c13e04"
          },
          "html": {
            "href": "https://bitbucket.org/anbraten/test-2/commits/668218c13e04"
          }
        }
      },
      "repository": {
        "type": "repository",
        "full_name": "anbraten/test-2",
        "links": {
          "self": {
            "href": "https://api.bitbucket.org/2.0/repositories/anbraten/test-2"
          },
          "html": {
            "href": "https://bitbucket.org/anbraten/test-2"
          },
          "avatar": {
            "href": "https://bytebucket.org/ravatar/%7B26554729-595f-47d1-aedd-302625cb4a97%7D?ts=default"
          }
        },
        "name": "test-2",
        "uuid": "{26554729-595f-47d1-aedd-302625cb4a97}"
      }
    },
    "reviewers": [],
    "participants": [
      {
        "type": "participant",
        "user": {
          "display_name": "Anbraten",
          "links": {
            "self": {
              "href": "https://api.bitbucket.org/2.0/users/%7Bb1b7beef-77ca-452d-b059-fa092504ebd7%7D"
            },
            "avatar": {
              "href": "https://avatar-management--avatars.us-west-2.prod.public.atl-paas.net/70121:3046ad5f-946f-48fa-bcb4-a399eef48f0e/784add1f-95cc-42a5-a562-38a0e12de4fa/128"
            },
            "html": {
              "href": "https://bitbucket.org/%7Bb1b7beef-77ca-452d-b059-fa092504ebd7%7D/"
            }
          },
          "type": "user",
          "uuid": "{b1b7beef-77ca-452d-b059-fa092504ebd7}",
          "account_id": "70121:3046ad5f-946f-48fa-bcb4-a399eef48f0e",
          "nickname": "Anbraten"
        },
        "role": "PARTICIPANT",
        "approved": true,
        "state": "approved",
        "participated_on": "2023-12-05T18:29:25.611876+00:00"
      }
    ],
    "links": {
      "self": {
        "href": "https://api.bitbucket.org/2.0/repositories/anbraten/test-2/pullrequests/1"
      },
      "html": {
        "href": "https://bitbucket.org/anbraten/test-2/pull-requests/1"
      },
      "commits": {
        "href": "https://api.bitbucket.org/2.0/repositories/anbraten/test-2/pullrequests/1/commits"
      },
      "approve": {
        "href": "https://api.bitbucket.org/2.0/repositories/anbraten/test-2/pullrequests/1/approve"
      },
      "request-changes": {
        "href": "https://api.bitbucket.org/2.0/repositories/anbraten/test-2/pullrequests/1/request-changes"
      },
      "diff": {
        "href": "https://api.bitbucket.org/2.0/repositories/anbraten/test-2/diff/anbraten/test-2:668218c13e04%0D6c5f0bc9b2aa?from_pullrequest_id=1&topic=true"
      },
      "diffstat": {
        "href": "https://api.bitbucket.org/2.0/repositories/anbraten/test-2/diffstat/anbraten/test-2:668218c13e04%0D6c5f0bc9b2aa?from_pullrequest_id=1&topic=true"
      },
      "comments": {
        "href": "https://api.bitbucket.org/2.0/repositories/anbraten/test-2/pullrequests/1/comments"
      },
      "activity": {
        "href": "https://api.bitbucket.org/2.0/repositories/anbraten/test-2/pullrequests/1/activity"
      },
      "merge": {
        "href": "https://api.bitbucket.org/2.0/repositories/anbraten/test-2/pullrequests/1/merge"
      },
      "decline": {
        "href": "https://api.bitbucket.org/2.0/repositories/anbraten/test-2/pullrequests/1/decline"
      },
      "statuses": {
        "href": "https://api.bitbucket.org/2.0/repositories/anbraten/test-2/pullrequests/1/statuses"
      }
    },
    "summary": {
      "type": "rendered",
      "raw": "README.md created online with Bitbucket",
      "markup": "markdown",
      "html": "<p>README.md created online with Bitbucket</p>"
    }
  }
}
`

const HookPullRequestDeclined = `
{
  "repository": {
    "type": "repository",
    "full_name": "anbraten/test-2",
    "links": {
      "self": {
        "href": "https://api.bitbucket.org/2.0/repositories/anbraten/test-2"
      },
      "html": {
        "href": "https://bitbucket.org/anbraten/test-2"
      },
      "avatar": {
        "href": "https://bytebucket.org/ravatar/%7B26554729-595f-47d1-aedd-302625cb4a97%7D?ts=default"
      }
    },
    "name": "test-2",
    "scm": "git",
    "website": null,
    "owner": {
      "display_name": "Anbraten",
      "links": {
        "self": {
          "href": "https://api.bitbucket.org/2.0/users/%7Bb1b7beef-77ca-452d-b059-fa092504ebd7%7D"
        },
        "avatar": {
          "href": "https://avatar-management--avatars.us-west-2.prod.public.atl-paas.net/70121:3046ad5f-946f-48fa-bcb4-a399eef48f0e/784add1f-95cc-42a5-a562-38a0e12de4fa/128"
        },
        "html": {
          "href": "https://bitbucket.org/%7Bb1b7beef-77ca-452d-b059-fa092504ebd7%7D/"
        }
      },
      "type": "user",
      "uuid": "{b1b7beef-77ca-452d-b059-fa092504ebd7}",
      "account_id": "70121:3046ad5f-946f-48fa-bcb4-a399eef48f0e",
      "nickname": "Anbraten"
    },
    "workspace": {
      "type": "workspace",
      "uuid": "{b1b7beef-77ca-452d-b059-fa092504ebd7}",
      "name": "Anbraten",
      "slug": "anbraten",
      "links": {
        "avatar": {
          "href": "https://bitbucket.org/workspaces/anbraten/avatar/?ts=1651865281"
        },
        "html": {
          "href": "https://bitbucket.org/anbraten/"
        },
        "self": {
          "href": "https://api.bitbucket.org/2.0/workspaces/anbraten"
        }
      }
    },
    "is_private": true,
    "project": {
      "type": "project",
      "key": "TEST",
      "uuid": "{3fa6429f-95e1-4c5a-875c-1753abcd8ace}",
      "name": "test",
      "links": {
        "self": {
          "href": "https://api.bitbucket.org/2.0/workspaces/anbraten/projects/TEST"
        },
        "html": {
          "href": "https://bitbucket.org/anbraten/workspace/projects/TEST"
        },
        "avatar": {
          "href": "https://bitbucket.org/account/user/anbraten/projects/TEST/avatar/32?ts=1690725373"
        }
      }
    },
    "uuid": "{26554729-595f-47d1-aedd-302625cb4a97}",
    "parent": null
  },
  "actor": {
    "display_name": "Anbraten",
    "links": {
      "self": {
        "href": "https://api.bitbucket.org/2.0/users/%7Bb1b7beef-77ca-452d-b059-fa092504ebd7%7D"
      },
      "avatar": {
        "href": "https://avatar-management--avatars.us-west-2.prod.public.atl-paas.net/70121:3046ad5f-946f-48fa-bcb4-a399eef48f0e/784add1f-95cc-42a5-a562-38a0e12de4fa/128"
      },
      "html": {
        "href": "https://bitbucket.org/%7Bb1b7beef-77ca-452d-b059-fa092504ebd7%7D/"
      }
    },
    "type": "user",
    "uuid": "{b1b7beef-77ca-452d-b059-fa092504ebd7}",
    "account_id": "70121:3046ad5f-946f-48fa-bcb4-a399eef48f0e",
    "nickname": "Anbraten"
  },
  "pullrequest": {
    "comment_count": 0,
    "task_count": 0,
    "type": "pullrequest",
    "id": 2,
    "title": "CHANGELOG.md created online with Bitbucket",
    "description": "CHANGELOG.md created online with Bitbucket",
    "rendered": {
      "title": {
        "type": "rendered",
        "raw": "CHANGELOG.md created online with Bitbucket",
        "markup": "markdown",
        "html": "<p>CHANGELOG.md created online with Bitbucket</p>"
      },
      "description": {
        "type": "rendered",
        "raw": "CHANGELOG.md created online with Bitbucket",
        "markup": "markdown",
        "html": "<p>CHANGELOG.md created online with Bitbucket</p>"
      }
    },
    "state": "DECLINED",
    "merge_commit": null,
    "close_source_branch": false,
    "closed_by": {
      "display_name": "Anbraten",
      "links": {
        "self": {
          "href": "https://api.bitbucket.org/2.0/users/%7Bb1b7beef-77ca-452d-b059-fa092504ebd7%7D"
        },
        "avatar": {
          "href": "https://avatar-management--avatars.us-west-2.prod.public.atl-paas.net/70121:3046ad5f-946f-48fa-bcb4-a399eef48f0e/784add1f-95cc-42a5-a562-38a0e12de4fa/128"
        },
        "html": {
          "href": "https://bitbucket.org/%7Bb1b7beef-77ca-452d-b059-fa092504ebd7%7D/"
        }
      },
      "type": "user",
      "uuid": "{b1b7beef-77ca-452d-b059-fa092504ebd7}",
      "account_id": "70121:3046ad5f-946f-48fa-bcb4-a399eef48f0e",
      "nickname": "Anbraten"
    },
    "author": {
      "display_name": "Anbraten",
      "links": {
        "self": {
          "href": "https://api.bitbucket.org/2.0/users/%7Bb1b7beef-77ca-452d-b059-fa092504ebd7%7D"
        },
        "avatar": {
          "href": "https://avatar-management--avatars.us-west-2.prod.public.atl-paas.net/70121:3046ad5f-946f-48fa-bcb4-a399eef48f0e/784add1f-95cc-42a5-a562-38a0e12de4fa/128"
        },
        "html": {
          "href": "https://bitbucket.org/%7Bb1b7beef-77ca-452d-b059-fa092504ebd7%7D/"
        }
      },
      "type": "user",
      "uuid": "{b1b7beef-77ca-452d-b059-fa092504ebd7}",
      "account_id": "70121:3046ad5f-946f-48fa-bcb4-a399eef48f0e",
      "nickname": "Anbraten"
    },
    "reason": "",
    "created_on": "2023-12-05T18:36:27.667680+00:00",
    "updated_on": "2023-12-05T18:36:57.260672+00:00",
    "destination": {
      "branch": {
        "name": "main"
      },
      "commit": {
        "type": "commit",
        "hash": "006704dbeab2",
        "links": {
          "self": {
            "href": "https://api.bitbucket.org/2.0/repositories/anbraten/test-2/commit/006704dbeab2"
          },
          "html": {
            "href": "https://bitbucket.org/anbraten/test-2/commits/006704dbeab2"
          }
        }
      },
      "repository": {
        "type": "repository",
        "full_name": "anbraten/test-2",
        "links": {
          "self": {
            "href": "https://api.bitbucket.org/2.0/repositories/anbraten/test-2"
          },
          "html": {
            "href": "https://bitbucket.org/anbraten/test-2"
          },
          "avatar": {
            "href": "https://bytebucket.org/ravatar/%7B26554729-595f-47d1-aedd-302625cb4a97%7D?ts=default"
          }
        },
        "name": "test-2",
        "uuid": "{26554729-595f-47d1-aedd-302625cb4a97}"
      }
    },
    "source": {
      "branch": {
        "name": "patch-2"
      },
      "commit": {
        "type": "commit",
        "hash": "f90e18fc9d45",
        "links": {
          "self": {
            "href": "https://api.bitbucket.org/2.0/repositories/anbraten/test-2/commit/f90e18fc9d45"
          },
          "html": {
            "href": "https://bitbucket.org/anbraten/test-2/commits/f90e18fc9d45"
          }
        }
      },
      "repository": {
        "type": "repository",
        "full_name": "anbraten/test-2",
        "links": {
          "self": {
            "href": "https://api.bitbucket.org/2.0/repositories/anbraten/test-2"
          },
          "html": {
            "href": "https://bitbucket.org/anbraten/test-2"
          },
          "avatar": {
            "href": "https://bytebucket.org/ravatar/%7B26554729-595f-47d1-aedd-302625cb4a97%7D?ts=default"
          }
        },
        "name": "test-2",
        "uuid": "{26554729-595f-47d1-aedd-302625cb4a97}"
      }
    },
    "reviewers": [],
    "participants": [],
    "links": {
      "self": {
        "href": "https://api.bitbucket.org/2.0/repositories/anbraten/test-2/pullrequests/2"
      },
      "html": {
        "href": "https://bitbucket.org/anbraten/test-2/pull-requests/2"
      },
      "commits": {
        "href": "https://api.bitbucket.org/2.0/repositories/anbraten/test-2/pullrequests/2/commits"
      },
      "approve": {
        "href": "https://api.bitbucket.org/2.0/repositories/anbraten/test-2/pullrequests/2/approve"
      },
      "request-changes": {
        "href": "https://api.bitbucket.org/2.0/repositories/anbraten/test-2/pullrequests/2/request-changes"
      },
      "diff": {
        "href": "https://api.bitbucket.org/2.0/repositories/anbraten/test-2/diff/anbraten/test-2:f90e18fc9d45%0D006704dbeab2?from_pullrequest_id=2&topic=true"
      },
      "diffstat": {
        "href": "https://api.bitbucket.org/2.0/repositories/anbraten/test-2/diffstat/anbraten/test-2:f90e18fc9d45%0D006704dbeab2?from_pullrequest_id=2&topic=true"
      },
      "comments": {
        "href": "https://api.bitbucket.org/2.0/repositories/anbraten/test-2/pullrequests/2/comments"
      },
      "activity": {
        "href": "https://api.bitbucket.org/2.0/repositories/anbraten/test-2/pullrequests/2/activity"
      },
      "merge": {
        "href": "https://api.bitbucket.org/2.0/repositories/anbraten/test-2/pullrequests/2/merge"
      },
      "decline": {
        "href": "https://api.bitbucket.org/2.0/repositories/anbraten/test-2/pullrequests/2/decline"
      },
      "statuses": {
        "href": "https://api.bitbucket.org/2.0/repositories/anbraten/test-2/pullrequests/2/statuses"
      }
    },
    "summary": {
      "type": "rendered",
      "raw": "CHANGELOG.md created online with Bitbucket",
      "markup": "markdown",
      "html": "<p>CHANGELOG.md created online with Bitbucket</p>"
    }
  }
}
`
