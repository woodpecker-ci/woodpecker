# External Configuration API

To provide additional management and preprocessing capabilities for pipeline configurations Woodpecker supports an HTTP api which can be enabled to call an external config service.
Before the run or restart of any pipeline Woodpecker will make a POST request to an external HTTP api sending the current repository, build information and all current config files retrieved from the repository. The external api can then send back new pipeline configurations that will be used immediately or respond with `HTTP 204` to tell the system to use the existing configuration.

Every request sent by Woodpecker is signed using a http-signature using the provided secret from `WOODPECKER_CONFIG_SERVICE_SECRET`. This way the external api can verify the authenticity request from the Woodpecker instance.

## Config

```shell
# Server
# ...
WOODPECKER_YAML_ENDPOINT=https://example.com/ciconfig
WOODPECKER_YAML_SECRET=mysecretsigningkey

```

### Example request made by Woodpecker

```json
{
  "repo": {
    "id": 100,
    "uid": "",
    "user_id": 0,
    "namespace": "",
    "name": "woodpecker-testpipe",
    "slug": "",
    "scm": "git",
    "git_http_url": "",
    "git_ssh_url": "",
    "link": "",
    "default_branhc": "",
    "private": true,
    "visibility": "private",
    "active": true,
    "config": "",
    "trusted": false,
    "protected": false,
    "ignore_forks": false,
    "ignore_pulls": false,
    "cancel_pulls": false,
    "timeout": 60,
    "counter": 0,
    "synced": 0,
    "created": 0,
    "updated": 0,
    "version": 0
  },
  "build": {
    "author": "myUser",
    "author_avatar": "https://myscm.com/avatars/d6b3f7787a685fcdf2a44e2c685c7e03",
    "author_email": "my@email.com",
    "branch": "master",
    "changed_files": [
      "somefilename.txt"
    ],
    "commit": "2fff90f8d288a4640e90f05049fe30e61a14fd50",
    "created_at": 0,
    "deploy_to": "",
    "enqueued_at": 0,
    "error": "",
    "event": "push",
    "finished_at": 0,
    "id": 0,
    "link_url": "https://myscm.com/myUser/woodpecker-testpipe/commit/2fff90f8d288a4640e90f05049fe30e61a14fd50",
    "message": "test old config\n",
    "number": 0,
    "parent": 0,
    "ref": "refs/heads/master",
    "refspec": "",
    "remote": "",
    "reviewed_at": 0,
    "reviewed_by": "",
    "sender": "myUser",
    "signed": false,
    "started_at": 0,
    "status": "",
    "timestamp": 1645962783,
    "title": "",
    "updated_at": 0,
    "verified": false
  },
  "config": [
    {
      "name": ".woodpecekr.yml",
      "data": "pipeline:\n  backend:\n    image: alpine\n    commands:\n      - echo \"Hello there from Repo (.woodpecekr.yml)\"\n"
    }
  ]
}
```

### Example response structure

```json
{
  "pipelines": [
    {
      "name": "central-override",
      "data": "pipeline:\n  backend:\n    image: alpine\n    commands:\n      - echo \"Hello there from ConfigAPI\"\n"
    }
  ]
}
```
