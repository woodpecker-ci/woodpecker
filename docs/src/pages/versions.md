# Versions

Woodpecker is having two different kinds of releases: **stable** and **next**.

If you want all (new) features of Woodpecker while supporting us with feedback and are willing to accept some possible bugs from time to time, you should use the next release, otherwise use the stable release.

We plan to release a new version every four weeks and will release the next version as a stable version.

## Stable version

The **stable** releases are official versions following [semver](https://semver.org/). By default, only the latest stable release will receive bug fixes. Once a new major or minor release is available, previous minor versions might receive security patches, but won't be updated with bug fixes anymore (so called backporting) by default.

### Breaking changes

As of semver guidelines, breaking changes will be released as a major version. We will hold back
breaking changes to not release many majors each containing just a few breaking changes.
Prior to the release of a major version, a release candidate (RC) will be published to allow easy testing,
the actual release will be about a week later.

### Deprecations & migrations

All deprecations and migrations for Woodpecker users and instance admins are documented in the [migration guide](/migrations).

## Next version (current state of the `main` branch)

The **next** version contains all bugfixes and features from `main` branch. Normally it should be pretty stable, but as its frequently updated, it might contain some bugs from time to time. There are no binaries for this version.

## Past versions (Not maintained anymore)

Here you can find documentation for previous versions of Woodpecker.

[Changelog](https://github.com/woodpecker-ci/woodpecker/blob/main/CHANGELOG.md)

|         |            |                                                                                       |
| ------- | ---------- | ------------------------------------------------------------------------------------- |
| 2.7.3   | 2024-11-28 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v2.7.3/docs/docs/)   |
| 2.7.2   | 2024-11-03 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v2.7.2/docs/docs/)   |
| 2.7.1   | 2024-09-07 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v2.7.1/docs/docs/)   |
| 2.7.0   | 2024-07-18 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v2.7.0/docs/docs/)   |
| 2.6.1   | 2024-07-19 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v2.6.1/docs/docs/)   |
| 2.6.0   | 2024-06-13 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v2.6.0/docs/docs/)   |
| 2.5.0   | 2024-06-01 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v2.5.0/docs/docs/)   |
| 2.4.1   | 2024-03-20 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v2.4.1/docs/docs/)   |
| 2.4.0   | 2024-03-19 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v2.4.0/docs/docs/)   |
| 2.3.0   | 2024-01-31 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v2.3.0/docs/docs/)   |
| 2.2.2   | 2024-01-21 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v2.2.2/docs/docs/)   |
| 2.2.1   | 2024-01-21 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v2.2.1/docs/docs/)   |
| 2.2.0   | 2024-01-21 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v2.2.0/docs/docs/)   |
| 2.1.1   | 2023-12-27 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v2.1.1/docs/docs/)   |
| 2.1.0   | 2023-12-26 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v2.1.0/docs/docs/)   |
| 2.0.0   | 2023-12-23 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v2.0.0/docs/docs/)   |
| 1.0.5   | 2023-11-09 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v1.0.5/docs/docs/)   |
| 1.0.4   | 2023-11-05 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v1.0.4/docs/docs/)   |
| 1.0.3   | 2023-10-14 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v1.0.3/docs/docs/)   |
| 1.0.2   | 2023-08-16 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v1.0.2/docs/docs/)   |
| 1.0.1   | 2023-08-08 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v1.0.1/docs/docs/)   |
| 1.0.0   | 2023-07-29 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v1.0.0/docs/docs/)   |
| 0.15.11 | 2023-07-12 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v0.15.11/docs/docs/) |
| 0.15.10 | 2023-07-09 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v0.15.10/docs/docs/) |
| 0.15.9  | 2023-05-11 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v0.15.9/docs/docs/)  |
| 0.15.8  | 2023-04-29 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v0.15.8/docs/docs/)  |
| 0.15.7  | 2023-03-14 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v0.15.7/docs/docs/)  |
| 0.15.6  | 2022-12-23 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v0.15.6/docs/docs/)  |
| 0.15.5  | 2022-10-13 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v0.15.5/docs/docs/)  |
| 0.15.4  | 2022-09-06 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v0.15.4/docs/docs/)  |
| 0.15.3  | 2022-06-16 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v0.15.3/docs/docs/)  |
| 0.15.2  | 2022-06-14 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v0.15.2/docs/docs/)  |
| 0.15.1  | 2022-04-13 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v0.15.1/docs/docs/)  |
| 0.15.0  | 2022-02-24 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v0.15.0/docs/docs/)  |
| 0.14.4  | 2022-01-31 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v0.14.4/docs/docs/)  |
| 0.14.3  | 2021-10-30 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v0.14.3/docs/docs/)  |
| 0.14.2  | 2021-10-19 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v0.14.2/docs/docs/)  |
| 0.14.1  | 2021-09-21 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v0.14.1/docs/docs/)  |
| 0.14.0  | 2021-08-01 | [Documentation](https://github.com/woodpecker-ci/woodpecker/tree/v0.14.0/docs/docs/)  |

If you are using an older version of Woodpecker and would like to view docs for this version, please use GitHub to browse the repository at your tag.
