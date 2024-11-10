# About

Woodpecker has been originally forked from Drone 0.8 as the Drone CI license was changed after the 0.8 release from Apache 2.0 to a proprietary license. Woodpecker is based on this latest freely available version.

## History

Woodpecker was originally forked by [@laszlocph](https://github.com/laszlocph) in 2019.

A few important time points:

- [`2fbaa56`](https://github.com/woodpecker-ci/woodpecker/commit/2fbaa56eee0f4be7a3ca4be03dbd00c1bf5d1274) is the first commit of the fork, made on Apr 3, 2019.
- The first release [v0.8.91](https://github.com/woodpecker-ci/woodpecker/releases/tag/v0.8.91) was published on Apr 6, 2019.
- On Aug 27, 2019, the project was renamed to "Woodpecker" ([`630c383`](https://github.com/woodpecker-ci/woodpecker/commit/630c383181b10c4ec375e500c812c4b76b3c52b8)).
- The first release under the name "Woodpecker" was published on Sep 9, 2019 ([v0.8.104](https://github.com/woodpecker-ci/woodpecker/releases/tag/v0.8.104)).

## Differences to Drone

Woodpecker is a community-focused software that still stay free and open source forever, while Drone is managed by [Harness](https://harness.io/) and published under [Polyform Small Business](https://polyformproject.org/licenses/small-business/1.0.0/) license, as well as an [OSS edition](https://docs.drone.io/enterprise/#what-is-the-difference-between-open-source-and-enterprise) published under Apache 2.0 .

### Features

| Feature | Drone CI (Enterprise) | Woodpecker CI |
|---------|-----------------------|---------------|
| Trigger | CI, Manual, Cronjob   |  |
| Runner  | Docker, Exec, SSH, DigitalOcean |  |
| Extensions | Admission, Configuration, Conversion, Environment, Registry, Secrets, Validation |  |
| Templating | (/) |  |
| Manual steps | (!) (via promoting) |  |
| Report UI | [(/)](https://docs.drone.io/plugins/adaptive_cards/) |  |
| Secret management | (/) (via extensions) |  |
