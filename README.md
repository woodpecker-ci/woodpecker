# Woodpecker

Woodpecker is a Community fork of the Drone CI system.

[![Go Report Card](https://goreportcard.com/badge/github.com/woodpecker-ci/woodpecker)](https://goreportcard.com/report/github.com/woodpecker-ci/woodpecker) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![https://discord.gg/fcMQqSMXJy](https://img.shields.io/discord/838698813463724034.svg)](https://discord.gg/fcMQqSMXJy)

![woodpecker](docs/docs/images/woodpecker.png)

# Contribution

## Maintainers

To make sure every PR is checked, we have **team maintainers**.
Every PR **MUST** be reviewed by at least one maintainer (or owner) before it can get merged.
A maintainer should be a contributor and contributed at least 4 accepted PRs.
A contributor should apply as a maintainer in the [Discord](https://discord.gg/fcMQqSMXJy) #develop channel.
The owners or the team maintainers may invite the contributor.
A maintainer should spend some time on code reviews.
If a maintainer has no time to do that, they should apply to leave the maintainers team and
we will give them the honor of being a member of the [advisors
team](https://github.com/orgs/woodpecker-ci/teams/advisors/members).
Of course, if an advisor has time to code review, we will gladly welcome them back to the maintainers team.
If a maintainer is inactive for more than 3 months and forgets to leave the maintainers team,
the owners may move him or her from the maintainers team to the advisors team.
For security reasons, Maintainers should use 2FA for their accounts and if possible provide gpg signed commits.
https://help.github.com/articles/securing-your-account-with-two-factor-authentication-2fa/
https://help.github.com/articles/signing-commits-with-gpg/

## Owners

Since Woodpecker is a pure community organization without any company support, to keep the development healthy we will elect three owners every year.
All maintainers may vote to elect up to two candidates. When the new owners have been elected, the old owners will give up ownership to the newly elected owners.
If an owner is unable to do so, the other owners will assist in ceding ownership to the newly elected owners.
For security reasons. Owners must use 2FA. https://help.github.com/articles/securing-your-account-with-two-factor-authentication-2fa/

# Usage

## .woodpecker.yml

- Place your pipeline in a file named `.woodpecker.yml` in your repository
- Pipeline steps can be named as you like
- Run any command in the commands section

```yaml
# .woodpecker.yml
pipeline:
  build:
    image: debian
    commands:
      - echo "This is the build step"
  a-test-step:
    image: debian
    commands:
      - echo "Testing.."
```

## Build steps are containers

- Define any Docker image as context
- Install the needed tools in custom Docker images, use them as context

```diff
 pipeline:
   build:
-    image: debian
+    image: mycompany/image-with-awscli
     commands:
       - aws help
```

## File changes are incremental

- Woodpecker clones the source code in the beginning pipeline
- Changes to files are persisted through steps as the same volume is mounted to all steps

```yaml
# .woodpecker.yml
pipeline:
  build:
    image: debian
    commands:
      - touch myfile
  a-test-step:
    image: debian
    commands:
      - cat myfile
```

## Plugins are straightforward

- If you copy the same shell script from project to project
- Pack it into a plugin instead
- And make the yaml declarative
- Plugins are Docker images with your script as an entrypoint

```Dockerfile
# Dockerfile
FROM laszlocloud/kubectl
COPY deploy /usr/local/deploy
ENTRYPOINT ["/usr/local/deploy"]
```

```bash
# deploy
kubectl apply -f $PLUGIN_TEMPLATE
```

```yaml
# .woodpecker.yml
pipeline:
  deploy-to-k8s:
    image: laszlocloud/my-k8s-plugin
    template: config/k8s/service.yml
```

# Documentation

https://woodpecker.laszlo.cloud

## Who uses Woodpecker

Currently, I know of one organization using Woodpecker. With 50+ users, 130+ repos and more than 1100 builds a week.

Leave a [comment](https://github.com/woodpecker-ci/woodpecker/issues/122) if you're using it. 

## License

Woodpecker is Apache 2.0 licensed with the source files in this repository having a header indicating which license they are under and what copyrights apply.

Files under the `docs/` folder is licensed under Creative Commons Attribution-ShareAlike 4.0 International Public License.
