# Contributing

## Maintainers

To make sure every Pull Request (PR) is checked, we have **team maintainers**.  
Every PR **MUST** be reviewed by at least **one** maintainer (or owner) before it can get merged.  
A maintainer should be a contributor and contributed at least 4 accepted PRs.
A contributor should apply as a maintainer in the [Discord #develop](https://discord.gg/fcMQqSMXJy) or [Matrix Develop](https://matrix.to/#/#WoodpeckerCI-Develop:obermui.de) channel.
The owners or the team maintainers may invite the contributor.
A maintainer should spend some time on code reviews.

If a maintainer has no time to do that, they should apply to leave the maintainers team and we will give them the honor of being a member of the [advisors
team](https://github.com/orgs/woodpecker-ci/teams/advisors/members).
Of course, if an advisor has time to code review, we will gladly welcome them back to the maintainers team.
If a maintainer is inactive for more than 3 months and forgets to leave the maintainers team, the owners may move him or her from the maintainers team to the advisors team.

For security reasons, Maintainers must use 2FA for their accounts and if possible provide GPG signed commits.  
https://help.github.com/articles/securing-your-account-with-two-factor-authentication-2fa/
https://help.github.com/articles/signing-commits-with-gpg/

## Owners

Since Woodpecker is a pure community organization without any company support,
to keep the development healthy we will elect two owners every year.  
This can also happen when a owner propose a vote or the majority of the maintainers do so.  
All maintainers may vote to elect up to two candidates. When the new owners have been elected, the old owners will give up ownership to the newly elected owners.  
If an owner is unable to do so, the other owners will assist in ceding ownership to the newly elected owners.

For security reasons, Owners must use 2FA.  
([Docs: Securing your account with two-factor authentication](https://docs.github.com/en/authentication/securing-your-account-with-two-factor-authentication-2fa))

To honor the past owners, here's the history of the owners and the time
they served:

* 2023-01-01 ~ 2023-12-31 - https://github.com/woodpecker-ci/woodpecker/issues/1467
  * [6543](https://github.com/6543)
  * [Anbraten](https://github.com/anbraten)

* 2021-09-28 ~ 2022-12-31 - https://github.com/woodpecker-ci/woodpecker/issues/633
  * [6543](https://github.com/6543)
  * [Anbraten](https://github.com/anbraten)

* 2019-07-25 ~ 2021-09-28
  * [Laszlo Fogas](https://github.com/laszlocph)

## Code Review

Once code review starts on your PR, do not rebase nor squash your branch as it makes it
difficult to review the new changes. Only if there is a need, sync your branch by merging
the base branch into yours. Don't worry about merge commits messing up your tree as
the final merge process squashes all commits into one, with the visible commit message (first
line) being the PR title + PR index and description being the PR's first comment.

Once your PR gets approved, don't worry about keeping it up-to-date or breaking
builds (unless there's a merge conflict or a request is made by a maintainer to make
modifications). It is the maintainer team's responsibility from this point to get it merged.

## Versioning

We use [Semantic Versioning](https://semver.org/) to be able,
to communicate when admins have to do manual migration steps and when they can just bump versions up.
