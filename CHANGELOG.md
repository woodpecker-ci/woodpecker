# Changelog

## [2.7.1](https://github.com/woodpecker-ci/woodpecker/releases/tag/v2.7.1) - 2024-09-07

### ❤️ Thanks to all contributors! ❤️

@6543, @anbraten, @j04n-f, @qwerty287

### 🔒 Security

- Lint privileged plugin match and allow to be set empty [[#4084](https://github.com/woodpecker-ci/woodpecker/pull/4084)]
- Allow admins to specify privileged plugins by name **and tag** [[#4076](https://github.com/woodpecker-ci/woodpecker/pull/4076)]
- Warn if using secrets/env with plugin [[#4039](https://github.com/woodpecker-ci/woodpecker/pull/4039)]

### 🐛 Bug Fixes

- Set refspec for gitlab MR [[#4021](https://github.com/woodpecker-ci/woodpecker/pull/4021)]
- Change Bitbucket PR hook to point the source branch, commit & ref [[#3965](https://github.com/woodpecker-ci/woodpecker/pull/3965)]
- Add updated, merged and declined events to bb webhook activation [[#3963](https://github.com/woodpecker-ci/woodpecker/pull/3963)]
- Fix login via navbar [[#3962](https://github.com/woodpecker-ci/woodpecker/pull/3962)]
- Fix panic if forge is unreachable [[#3944](https://github.com/woodpecker-ci/woodpecker/pull/3944)]
- Fix org settings page [[#4093](https://github.com/woodpecker-ci/woodpecker/pull/4093)]

### Misc

- Bump github.com/docker/docker from v24.0.9 to v24.0.9+30 [[#4077](https://github.com/woodpecker-ci/woodpecker/pull/4077)]

## [2.7.0](https://github.com/woodpecker-ci/woodpecker/releases/tag/v2.7.0) - 2024-07-18

### ❤️ Thanks to all contributors! ❤️

@6543, @anbraten, @dvjn, @hhamalai, @lafriks, @pat-s, @qwerty287, @smainz, @tongjicoder, @zc-devs

### 🔒 Security

- Add blocklist of environment variables who could alter execution of plugins [[#3934](https://github.com/woodpecker-ci/woodpecker/pull/3934)]
- Make sure plugins only mount the workspace base in a predefinde location [[#3933](https://github.com/woodpecker-ci/woodpecker/pull/3933)]
- Disallow to set arbitrary environments for plugins [[#3909](https://github.com/woodpecker-ci/woodpecker/pull/3909)]
- Use proper oauth state [[#3847](https://github.com/woodpecker-ci/woodpecker/pull/3847)]
- Enhance token checking [[#3842](https://github.com/woodpecker-ci/woodpecker/pull/3842)]
- Bump github.com/hashicorp/go-retryablehttp v0.7.5 -> v0.7.7 [[#3834](https://github.com/woodpecker-ci/woodpecker/pull/3834)]

### ✨ Features

- Gracefully shutdown server [[#3896](https://github.com/woodpecker-ci/woodpecker/pull/3896)]
- Gracefully shutdown agent [[#3895](https://github.com/woodpecker-ci/woodpecker/pull/3895)]
- Convert urls in logs to links  [[#3904](https://github.com/woodpecker-ci/woodpecker/pull/3904)]
- Allow login using multiple forges [[#3822](https://github.com/woodpecker-ci/woodpecker/pull/3822)]
- Global and organization registries [[#1672](https://github.com/woodpecker-ci/woodpecker/pull/1672)]
- Cli get repo from git remote [[#3830](https://github.com/woodpecker-ci/woodpecker/pull/3830)]
- Add api for forges [[#3733](https://github.com/woodpecker-ci/woodpecker/pull/3733)]

### 📈 Enhancement

- Cli fix pipeline logs [[#3913](https://github.com/woodpecker-ci/woodpecker/pull/3913)]
- Migrate to github.com/urfave/cli/v3 [[#2951](https://github.com/woodpecker-ci/woodpecker/pull/2951)]
- Allow to change the working directory also for plugins and services [[#3914](https://github.com/woodpecker-ci/woodpecker/pull/3914)]
- Remove `unplugin-icons` [[#3809](https://github.com/woodpecker-ci/woodpecker/pull/3809)]
- Release windows binaries as zip file [[#3906](https://github.com/woodpecker-ci/woodpecker/pull/3906)]
- Convert to openapi 3.0 [[#3897](https://github.com/woodpecker-ci/woodpecker/pull/3897)]
- Add user registries UI [[#3888](https://github.com/woodpecker-ci/woodpecker/pull/3888)]
- Sort users by login [[#3891](https://github.com/woodpecker-ci/woodpecker/pull/3891)]
- Exclude dummy backend in production [[#3877](https://github.com/woodpecker-ci/woodpecker/pull/3877)]
- Fix deploy task env [[#3878](https://github.com/woodpecker-ci/woodpecker/pull/3878)]
- Get default branch and show message in pipeline list [[#3867](https://github.com/woodpecker-ci/woodpecker/pull/3867)]
- Add timestamp for last work done by agent [[#3844](https://github.com/woodpecker-ci/woodpecker/pull/3844)]
- Adjust logger types [[#3859](https://github.com/woodpecker-ci/woodpecker/pull/3859)]
- Cleanup state reporting [[#3850](https://github.com/woodpecker-ci/woodpecker/pull/3850)]
- Unify DB tables/columns [[#3806](https://github.com/woodpecker-ci/woodpecker/pull/3806)]
- Let webhook pass on pipeline parsing error [[#3829](https://github.com/woodpecker-ci/woodpecker/pull/3829)]
- Exclude mocks from release build [[#3831](https://github.com/woodpecker-ci/woodpecker/pull/3831)]
- K8s secrets reference from step [[#3655](https://github.com/woodpecker-ci/woodpecker/pull/3655)]

### 🐛 Bug Fixes

- Handle empty repositories in gitea when listing PRs [[#3925](https://github.com/woodpecker-ci/woodpecker/pull/3925)]
- Update alpine package dep for docker images [[#3917](https://github.com/woodpecker-ci/woodpecker/pull/3917)]
- Don't report error if agent was terminated gracefully [[#3894](https://github.com/woodpecker-ci/woodpecker/pull/3894)]
- Let agents continuously report their health [[#3893](https://github.com/woodpecker-ci/woodpecker/pull/3893)]
- Ignore warnings for cli exec [[#3868](https://github.com/woodpecker-ci/woodpecker/pull/3868)]
- Correct favicon states [[#3832](https://github.com/woodpecker-ci/woodpecker/pull/3832)]
- Cleanup of the login flow and tests [[#3810](https://github.com/woodpecker-ci/woodpecker/pull/3810)]
- Fix newlines in logs [[#3808](https://github.com/woodpecker-ci/woodpecker/pull/3808)]
- Fix authentication error handling [[#3807](https://github.com/woodpecker-ci/woodpecker/pull/3807)]

### 📚 Documentation

- Streamline docs for new users [[#3803](https://github.com/woodpecker-ci/woodpecker/pull/3803)]
- Add mastodon verification [[#3843](https://github.com/woodpecker-ci/woodpecker/pull/3843)]
- chore(deps): update docs npm deps non-major [[#3837](https://github.com/woodpecker-ci/woodpecker/pull/3837)]
- fix(deps): update docs npm deps non-major [[#3824](https://github.com/woodpecker-ci/woodpecker/pull/3824)]
- Add openSUSE package [[#3800](https://github.com/woodpecker-ci/woodpecker/pull/3800)]
- chore(deps): update docs npm deps non-major [[#3798](https://github.com/woodpecker-ci/woodpecker/pull/3798)]
- Add "Docker Tags" Plugin [[#3796](https://github.com/woodpecker-ci/woodpecker/pull/3796)]
- chore(deps): update dependency marked to v13 [[#3792](https://github.com/woodpecker-ci/woodpecker/pull/3792)]
- chore: fix some comments [[#3788](https://github.com/woodpecker-ci/woodpecker/pull/3788)]

### Misc

- chore(deps): update web npm deps non-major [[#3930](https://github.com/woodpecker-ci/woodpecker/pull/3930)]
- chore(deps): update dependency vitest to v2 [[#3905](https://github.com/woodpecker-ci/woodpecker/pull/3905)]
- fix(deps): update module github.com/google/go-github/v62 to v63 [[#3910](https://github.com/woodpecker-ci/woodpecker/pull/3910)]
- chore(deps): update docker.io/woodpeckerci/plugin-docker-buildx docker tag to v4.1.0 [[#3908](https://github.com/woodpecker-ci/woodpecker/pull/3908)]
- Update plugin-git and add renovate trigger [[#3901](https://github.com/woodpecker-ci/woodpecker/pull/3901)]
- chore(deps): update docker.io/mstruebing/editorconfig-checker docker tag to v3.0.3 [[#3903](https://github.com/woodpecker-ci/woodpecker/pull/3903)]
- fix(deps): update golang-packages [[#3875](https://github.com/woodpecker-ci/woodpecker/pull/3875)]
- chore(deps): lock file maintenance [[#3876](https://github.com/woodpecker-ci/woodpecker/pull/3876)]
- [pre-commit.ci] pre-commit autoupdate [[#3862](https://github.com/woodpecker-ci/woodpecker/pull/3862)]
- Add dummy backend [[#3820](https://github.com/woodpecker-ci/woodpecker/pull/3820)]
- chore(deps): update dependency replace-in-file to v8 [[#3852](https://github.com/woodpecker-ci/woodpecker/pull/3852)]
- Update forgejo sdk [[#3840](https://github.com/woodpecker-ci/woodpecker/pull/3840)]
- chore(deps): lock file maintenance [[#3838](https://github.com/woodpecker-ci/woodpecker/pull/3838)]
- Allow to set dist dir using env var [[#3814](https://github.com/woodpecker-ci/woodpecker/pull/3814)]
- chore(deps): lock file maintenance [[#3805](https://github.com/woodpecker-ci/woodpecker/pull/3805)]
- chore(deps): update docker.io/lycheeverse/lychee docker tag to v0.15.1 [[#3797](https://github.com/woodpecker-ci/woodpecker/pull/3797)]

## [2.6.1](https://github.com/woodpecker-ci/woodpecker/releases/tag/v2.6.1) - 2024-07-19

### 🔒 Security

- Add blocklist of environment variables who could alter execution of plugins [[#3934](https://github.com/woodpecker-ci/woodpecker/pull/3934)]
- Make sure plugins only mount the workspace base in a predefinde location [[#3933](https://github.com/woodpecker-ci/woodpecker/pull/3933)]
- Disalow to set arbitrary environments for plugins [[#3909](https://github.com/woodpecker-ci/woodpecker/pull/3909)]
- Bump trivy plugin version and remove unused variable [[#3833](https://github.com/woodpecker-ci/woodpecker/pull/3833)]

### 🐛 Bug Fixes

- Let webhook pass on pipeline parsion error [[#3829](https://github.com/woodpecker-ci/woodpecker/pull/3829)]
- Fix newlines in logs [[#3808](https://github.com/woodpecker-ci/woodpecker/pull/3808)]

## [2.6.0](https://github.com/woodpecker-ci/woodpecker/releases/tag/v2.6.0) - 2024-06-13

### ❤️ Thanks to all contributors! ❤️

@6543, @anbraten, @jcgl17, @pat-s, @qwerty287, @s00500, @wez, @zc-devs

### 🔒 Security

- Bump trivy plugin version and remove unused variable [[#3759](https://github.com/woodpecker-ci/woodpecker/pull/3759)]

### ✨ Features

- Allow to store logs in files [[#3568](https://github.com/woodpecker-ci/woodpecker/pull/3568)]
- Native forgejo support [[#3684](https://github.com/woodpecker-ci/woodpecker/pull/3684)]

### 🐛 Bug Fixes

- Add release event to webhooks [[#3784](https://github.com/woodpecker-ci/woodpecker/pull/3784)]
- Respect cli argument when checking docker backend availability [[#3770](https://github.com/woodpecker-ci/woodpecker/pull/3770)]
- Fix repo creation [[#3756](https://github.com/woodpecker-ci/woodpecker/pull/3756)]
- Fix config loading of cli [[#3764](https://github.com/woodpecker-ci/woodpecker/pull/3764)]
- Fix missing WOODPECKER_BITBUCKET_DC_URL [[#3761](https://github.com/woodpecker-ci/woodpecker/pull/3761)]
- Correct repo repair success message in cli [[#3757](https://github.com/woodpecker-ci/woodpecker/pull/3757)]

### 📈 Enhancement

- Improve step logging [[#3722](https://github.com/woodpecker-ci/woodpecker/pull/3722)]
- chore(deps): update dependency eslint to v9 [[#3594](https://github.com/woodpecker-ci/woodpecker/pull/3594)]
- Show workflow names if there are multiple configs [[#3767](https://github.com/woodpecker-ci/woodpecker/pull/3767)]
- Use http constants [[#3766](https://github.com/woodpecker-ci/woodpecker/pull/3766)]
- Spellcheck "server/*" [[#3753](https://github.com/woodpecker-ci/woodpecker/pull/3753)]
- Agent-wide node selector [[#3608](https://github.com/woodpecker-ci/woodpecker/pull/3608)]

### 📚 Documentation

- Remove misleading crontab guru suggestion from docs [[#3781](https://github.com/woodpecker-ci/woodpecker/pull/3781)]
- Add documentation for KUBERNETES_SERVICE_HOST in Agent [[#3747](https://github.com/woodpecker-ci/woodpecker/pull/3747)]
- Remove web.archive.org workaround in docs [[#3771](https://github.com/woodpecker-ci/woodpecker/pull/3771)]
- Serve plugin icons locally [[#3768](https://github.com/woodpecker-ci/woodpecker/pull/3768)]
- Docs: update local backend page [[#3765](https://github.com/woodpecker-ci/woodpecker/pull/3765)]
- Remove old docs versions [[#3743](https://github.com/woodpecker-ci/woodpecker/pull/3743)]
- Merge release plugins [[#3752](https://github.com/woodpecker-ci/woodpecker/pull/3752)]
- Split FAQ [[#3746](https://github.com/woodpecker-ci/woodpecker/pull/3746)]

### Misc

- Update nix flake [[#3780](https://github.com/woodpecker-ci/woodpecker/pull/3780)]
- chore(deps): lock file maintenance [[#3783](https://github.com/woodpecker-ci/woodpecker/pull/3783)]
- chore(deps): update pre-commit hook golangci/golangci-lint to v1.59.1 [[#3782](https://github.com/woodpecker-ci/woodpecker/pull/3782)]
- fix(deps): update codeberg.org/mvdkleijn/forgejo-sdk/forgejo digest to 168c988 [[#3776](https://github.com/woodpecker-ci/woodpecker/pull/3776)]
- chore(deps): lock file maintenance [[#3750](https://github.com/woodpecker-ci/woodpecker/pull/3750)]
- chore(deps): update gitea/gitea docker tag to v1.22 [[#3749](https://github.com/woodpecker-ci/woodpecker/pull/3749)]
- Fix setting name [[#3744](https://github.com/woodpecker-ci/woodpecker/pull/3744)]

## [2.5.0](https://github.com/woodpecker-ci/woodpecker/releases/tag/v2.5.0) - 2024-06-01

### ❤️ Thanks to all contributors! ❤️

@6543, @Andre601, @Elara6331, @OCram85, @anbraten, @aumetra, @da-Kai, @dominic-p, @dvjn, @eliasscosta, @fernandrone, @linghuying, @manuelluis, @nemunaire, @pat-s, @qwerty287, @sinlov, @stevapple, @xoxys, @zc-devs

### 🔒 Security

- bump golang.org/x/net to v0.24.0 [[#3628](https://github.com/woodpecker-ci/woodpecker/pull/3628)]

### ✨ Features

- Add DeletePipeline API [[#3506](https://github.com/woodpecker-ci/woodpecker/pull/3506)]
- CLI: remove step logs [[#3458](https://github.com/woodpecker-ci/woodpecker/pull/3458)]
- Step logs removing API and Button [[#3451](https://github.com/woodpecker-ci/woodpecker/pull/3451)]

### 📚 Documentation

- Create 2.5 docs [[#3732](https://github.com/woodpecker-ci/woodpecker/pull/3732)]
- Fix spelling in README [[#3741](https://github.com/woodpecker-ci/woodpecker/pull/3741)]
- chore: fix some comments [[#3740](https://github.com/woodpecker-ci/woodpecker/pull/3740)]
- Add "Is It Up Yet?" Plugin [[#3731](https://github.com/woodpecker-ci/woodpecker/pull/3731)]
- Remove discord as official community channel [[#3717](https://github.com/woodpecker-ci/woodpecker/pull/3717)]
- Add Gitea Package plugin [[#3707](https://github.com/woodpecker-ci/woodpecker/pull/3707)]
- Add documentation for setting Kubernetes labels and annotations [[#3687](https://github.com/woodpecker-ci/woodpecker/pull/3687)]
- Remove broken link to gobook.io [[#3694](https://github.com/woodpecker-ci/woodpecker/pull/3694)]
- docs: add `Gitea publisher-golang` plugin [[#3691](https://github.com/woodpecker-ci/woodpecker/pull/3691)]
- Add Ansible+Woodpecker blog post [[#3685](https://github.com/woodpecker-ci/woodpecker/pull/3685)]
- Clarify info on failing workflows/Steps [[#3679](https://github.com/woodpecker-ci/woodpecker/pull/3679)]
- Add discord plugin [[#3662](https://github.com/woodpecker-ci/woodpecker/pull/3662)]
- chore(deps): update dependency trim to v1 [[#3658](https://github.com/woodpecker-ci/woodpecker/pull/3658)]
- chore(deps): update dependency got to v14 [[#3657](https://github.com/woodpecker-ci/woodpecker/pull/3657)]
- Fail on broken anchors [[#3644](https://github.com/woodpecker-ci/woodpecker/pull/3644)]
- Fix step syntax in docs [[#3635](https://github.com/woodpecker-ci/woodpecker/pull/3635)]
- chore(deps): update docs npm deps non-major [[#3632](https://github.com/woodpecker-ci/woodpecker/pull/3632)]
- Add Twine plugin [[#3619](https://github.com/woodpecker-ci/woodpecker/pull/3619)]
- Fix docs [[#3615](https://github.com/woodpecker-ci/woodpecker/pull/3615)]
- Document how to enable parallel step exec for all steps [[#3605](https://github.com/woodpecker-ci/woodpecker/pull/3605)]
- Update dependency @types/marked to v6 [[#3544](https://github.com/woodpecker-ci/woodpecker/pull/3544)]
- Update docs npm deps non-major [[#3485](https://github.com/woodpecker-ci/woodpecker/pull/3485)]
- Docs updates and fixes [[#3535](https://github.com/woodpecker-ci/woodpecker/pull/3535)]

### 🐛 Bug Fixes

- Fix privileged steps in kubernetes [[#3711](https://github.com/woodpecker-ci/woodpecker/pull/3711)]
- Check for error in repo middleware [[#3688](https://github.com/woodpecker-ci/woodpecker/pull/3688)]
- Fix parent pipeline number env on restarts [[#3683](https://github.com/woodpecker-ci/woodpecker/pull/3683)]
- Fix bitbucket dir fetching [[#3668](https://github.com/woodpecker-ci/woodpecker/pull/3668)]
- Sanitize tag ref for gitea/forgejo [[#3664](https://github.com/woodpecker-ci/woodpecker/pull/3664)]
- Fix secret loading [[#3620](https://github.com/woodpecker-ci/woodpecker/pull/3620)]
- fix cli config loading and correct comment [[#3618](https://github.com/woodpecker-ci/woodpecker/pull/3618)]
- Handle ImagePullBackOff pod status [[#3580](https://github.com/woodpecker-ci/woodpecker/pull/3580)]
- Apply skip ci filter only on push events [[#3612](https://github.com/woodpecker-ci/woodpecker/pull/3612)]
- agent: Continue to retry indefinitely [[#3599](https://github.com/woodpecker-ci/woodpecker/pull/3599)]
- Fix cli version comparison and improve setup [[#3518](https://github.com/woodpecker-ci/woodpecker/pull/3518)]
- Fix flag name [[#3534](https://github.com/woodpecker-ci/woodpecker/pull/3534)]

### 📈 Enhancement

- Use IDs for tokens [[#3695](https://github.com/woodpecker-ci/woodpecker/pull/3695)]
- Lint go code with cspell [[#3706](https://github.com/woodpecker-ci/woodpecker/pull/3706)]
- Replace duplicated strings [[#3710](https://github.com/woodpecker-ci/woodpecker/pull/3710)]
- Cleanup server env settings [[#3670](https://github.com/woodpecker-ci/woodpecker/pull/3670)]
- Setting for empty commits on path condition [[#3708](https://github.com/woodpecker-ci/woodpecker/pull/3708)]
- Lint file names and directories via cSpell too [[#3703](https://github.com/woodpecker-ci/woodpecker/pull/3703)]
- Make retry count of config fetching form forge configure [[#3699](https://github.com/woodpecker-ci/woodpecker/pull/3699)]
- Ability to set pod annotations and labels from step [[#3609](https://github.com/woodpecker-ci/woodpecker/pull/3609)]
- Support github deploy task [[#3512](https://github.com/woodpecker-ci/woodpecker/pull/3512)]
- Rework entrypoints [[#3269](https://github.com/woodpecker-ci/woodpecker/pull/3269)]
- Add cli output handlers [[#3660](https://github.com/woodpecker-ci/woodpecker/pull/3660)]
- Cleanup api docs and ts api-client options [[#3663](https://github.com/woodpecker-ci/woodpecker/pull/3663)]
- Split client into multiple files and add more tests [[#3647](https://github.com/woodpecker-ci/woodpecker/pull/3647)]
- Add filter options to GetPipelines API [[#3645](https://github.com/woodpecker-ci/woodpecker/pull/3645)]
- Deprecate environment filter and improve errors [[#3634](https://github.com/woodpecker-ci/woodpecker/pull/3634)]
- Add task details to queue info in woodpecker-go [[#3636](https://github.com/woodpecker-ci/woodpecker/pull/3636)]
- Use forge from db [[#1417](https://github.com/woodpecker-ci/woodpecker/pull/1417)]
- Remove review button from approval view [[#3617](https://github.com/woodpecker-ci/woodpecker/pull/3617)]
- Rework addons (use rpc) [[#3268](https://github.com/woodpecker-ci/woodpecker/pull/3268)]
- Allow to disable deployments [[#3570](https://github.com/woodpecker-ci/woodpecker/pull/3570)]
- Add flag to only access public repositories on GitHub [[#3566](https://github.com/woodpecker-ci/woodpecker/pull/3566)]
- Add `runtimeClassName` in Kubernetes backend options [[#3474](https://github.com/woodpecker-ci/woodpecker/pull/3474)]
- Remove unused cache properties [[#3567](https://github.com/woodpecker-ci/woodpecker/pull/3567)]
- Allow separate gitea oauth URL  [[#3513](https://github.com/woodpecker-ci/woodpecker/pull/3513)]
- Add option to set the local repository path to the cli command exec. [[#3524](https://github.com/woodpecker-ci/woodpecker/pull/3524)]

### Misc

- chore(deps): update pre-commit non-major [[#3736](https://github.com/woodpecker-ci/woodpecker/pull/3736)]
- chore(deps): update docker.io/alpine docker tag to v3.20 [[#3735](https://github.com/woodpecker-ci/woodpecker/pull/3735)]
- fix(deps): update module github.com/google/go-github/v61 to v62 [[#3730](https://github.com/woodpecker-ci/woodpecker/pull/3730)]
- chore(deps): update docker.io/woodpeckerci/plugin-docker-buildx docker tag to v4 [[#3729](https://github.com/woodpecker-ci/woodpecker/pull/3729)]
- chore(deps): update docker.io/mstruebing/editorconfig-checker docker tag to v3 [[#3728](https://github.com/woodpecker-ci/woodpecker/pull/3728)]
- chore(deps): update woodpeckerci/plugin-ready-release-go docker tag to v1.1.2 [[#3724](https://github.com/woodpecker-ci/woodpecker/pull/3724)]
- fix(deps): update golang-packages [[#3713](https://github.com/woodpecker-ci/woodpecker/pull/3713)]
- chore(deps): update postgres docker tag to v16.3 [[#3719](https://github.com/woodpecker-ci/woodpecker/pull/3719)]
- chore(deps): update docker.io/appleboy/drone-discord docker tag to v1.3.2 [[#3718](https://github.com/woodpecker-ci/woodpecker/pull/3718)]
- Added steps to reproduce and expected behavior in bug_report.yaml [[#3714](https://github.com/woodpecker-ci/woodpecker/pull/3714)]
- flake: add flake-utils import and use eachDefaultSystem [[#3704](https://github.com/woodpecker-ci/woodpecker/pull/3704)]
- Add nix flake for dev shell [[#3702](https://github.com/woodpecker-ci/woodpecker/pull/3702)]
- Skip golangci in pre-commit.ci [[#3692](https://github.com/woodpecker-ci/woodpecker/pull/3692)]
- chore(deps): update woodpeckerci/plugin-github-release docker tag to v1.2.0 [[#3690](https://github.com/woodpecker-ci/woodpecker/pull/3690)]
- Switch back to upstream xgo image [[#3682](https://github.com/woodpecker-ci/woodpecker/pull/3682)]
- Allow running tests on arm64 runners [[#2605](https://github.com/woodpecker-ci/woodpecker/pull/2605)]
- chore(deps): update node.js to v22 [[#3659](https://github.com/woodpecker-ci/woodpecker/pull/3659)]
- chore(deps): lock file maintenance [[#3656](https://github.com/woodpecker-ci/woodpecker/pull/3656)]
- Add make target for spellcheck [[#3648](https://github.com/woodpecker-ci/woodpecker/pull/3648)]
- chore(deps): update woodpeckerci/plugin-ready-release-go docker tag to v1.1.1 [[#3641](https://github.com/woodpecker-ci/woodpecker/pull/3641)]
- chore(deps): update web npm deps non-major [[#3640](https://github.com/woodpecker-ci/woodpecker/pull/3640)]
- chore(deps): update web npm deps non-major [[#3631](https://github.com/woodpecker-ci/woodpecker/pull/3631)]
- Use our github-release plugin [[#3624](https://github.com/woodpecker-ci/woodpecker/pull/3624)]
- chore(deps): lock file maintenance [[#3622](https://github.com/woodpecker-ci/woodpecker/pull/3622)]
- Fix spellcheck and enable more dirs [[#3603](https://github.com/woodpecker-ci/woodpecker/pull/3603)]
- Update docker.io/golang Docker tag to v1.22.2 [[#3596](https://github.com/woodpecker-ci/woodpecker/pull/3596)]
- Update pre-commit hook pre-commit/pre-commit-hooks to v4.6.0 [[#3597](https://github.com/woodpecker-ci/woodpecker/pull/3597)]
- Update module github.com/google/go-github/v60 to v61 [[#3595](https://github.com/woodpecker-ci/woodpecker/pull/3595)]
- Update pre-commit hook golangci/golangci-lint to v1.57.2 [[#3575](https://github.com/woodpecker-ci/woodpecker/pull/3575)]
- Update docker.io/woodpeckerci/plugin-docker-buildx Docker tag to v3.2.1 [[#3574](https://github.com/woodpecker-ci/woodpecker/pull/3574)]
- Update web npm deps non-major [[#3576](https://github.com/woodpecker-ci/woodpecker/pull/3576)]
- Update dependency @intlify/unplugin-vue-i18n to v4 [[#3572](https://github.com/woodpecker-ci/woodpecker/pull/3572)]
- Update golang (packages) [[#3564](https://github.com/woodpecker-ci/woodpecker/pull/3564)]
- Update dependency typescript to v5.4.3 [[#3563](https://github.com/woodpecker-ci/woodpecker/pull/3563)]
- Lock file maintenance [[#3562](https://github.com/woodpecker-ci/woodpecker/pull/3562)]
- Update pre-commit non-major [[#3556](https://github.com/woodpecker-ci/woodpecker/pull/3556)]
- Update web npm deps non-major [[#3549](https://github.com/woodpecker-ci/woodpecker/pull/3549)]
- Update dependency @types/node-emoji to v2 [[#3545](https://github.com/woodpecker-ci/woodpecker/pull/3545)]
- Update golang (packages) [[#3543](https://github.com/woodpecker-ci/woodpecker/pull/3543)]
- Lock file maintenance [[#3541](https://github.com/woodpecker-ci/woodpecker/pull/3541)]
- Update docker.io/woodpeckerci/plugin-docker-buildx Docker tag to v3.2.0 [[#3540](https://github.com/woodpecker-ci/woodpecker/pull/3540)]

## [2.4.1](https://github.com/woodpecker-ci/woodpecker/releases/tag/v2.4.1) - 2024-03-20

### ❤️ Thanks to all contributors! ❤️

@manuelluis, @qwerty287, @xoxys

### 🔒 Security

- Only allow to deploy from push, tag and release [[#3522](https://github.com/woodpecker-ci/woodpecker/pull/3522)]

### 🐛 Bug Fixes

- Exclude setup from cli command exec. [[#3523](https://github.com/woodpecker-ci/woodpecker/pull/3523)]
- Fix uppercased env [[#3516](https://github.com/woodpecker-ci/woodpecker/pull/3516)]
- Fix env schema [[#3514](https://github.com/woodpecker-ci/woodpecker/pull/3514)]

### Misc

- Temp pin golangci version in makefile [[#3520](https://github.com/woodpecker-ci/woodpecker/pull/3520)]

## [2.4.0](https://github.com/woodpecker-ci/woodpecker/releases/tag/v2.4.0) - 2024-03-19

### ❤️ Thanks to all contributors! ❤️

@6543, @Ray-D-Song, @anbraten, @eliasscosta, @fernandrone, @kjuulh, @kytta, @langecode, @lukashass, @qwerty287, @rockdrilla, @sinlov, @smainz, @xoxys, @zc-devs, @zowhoey

### 🔒 Security

- Improve security context handling [[#3482](https://github.com/woodpecker-ci/woodpecker/pull/3482)]
- fix(deps): update module github.com/moby/moby to v24.0.9+incompatible [[#3323](https://github.com/woodpecker-ci/woodpecker/pull/3323)]

### ✨ Features

- Cli setup command [[#3384](https://github.com/woodpecker-ci/woodpecker/pull/3384)]
- Add bitbucket datacenter (server) support  [[#2503](https://github.com/woodpecker-ci/woodpecker/pull/2503)]
- Cli updater [[#3382](https://github.com/woodpecker-ci/woodpecker/pull/3382)]

### 📚 Documentation

- Delete docs for v0.15.x [[#3508](https://github.com/woodpecker-ci/woodpecker/pull/3508)]
- Add deployment plugin [[#3495](https://github.com/woodpecker-ci/woodpecker/pull/3495)]
- Bump follow-redirects and fix broken anchors [[#3488](https://github.com/woodpecker-ci/woodpecker/pull/3488)]
- fix: plugin doc page not found [[#3480](https://github.com/woodpecker-ci/woodpecker/pull/3480)]
- Documentation improvements [[#3376](https://github.com/woodpecker-ci/woodpecker/pull/3376)]
- fix(deps): update docs npm deps non-major [[#3455](https://github.com/woodpecker-ci/woodpecker/pull/3455)]
- Add "Sonatype Nexus" plugin [[#3446](https://github.com/woodpecker-ci/woodpecker/pull/3446)]
- Add blog post [[#3439](https://github.com/woodpecker-ci/woodpecker/pull/3439)]
- Add "Gradle Wrapper Validation" plugin [[#3435](https://github.com/woodpecker-ci/woodpecker/pull/3435)]
- Add blog post [[#3410](https://github.com/woodpecker-ci/woodpecker/pull/3410)]
- Extend core ideas documentation [[#3405](https://github.com/woodpecker-ci/woodpecker/pull/3405)]
- docs: fix contributions link [[#3363](https://github.com/woodpecker-ci/woodpecker/pull/3363)]
- Update/fix some docs [[#3359](https://github.com/woodpecker-ci/woodpecker/pull/3359)]
- chore(deps): update dependency marked to v12 [[#3325](https://github.com/woodpecker-ci/woodpecker/pull/3325)]

### 🐛 Bug Fixes

- Fix skip setup for some general cli commands [[#3498](https://github.com/woodpecker-ci/woodpecker/pull/3498)]
- Move generic agent flags to cmd/agent/core [[#3484](https://github.com/woodpecker-ci/woodpecker/pull/3484)]
- Fix usage of WOODPECKER_DATABASE_DATASOURCE_FILE [[#3404](https://github.com/woodpecker-ci/woodpecker/pull/3404)]
- Set pull-request id and labels on pr-closed event [[#3442](https://github.com/woodpecker-ci/woodpecker/pull/3442)]
- Update org name on login [[#3409](https://github.com/woodpecker-ci/woodpecker/pull/3409)]
- Do not alter secret key upper-/lowercase [[#3375](https://github.com/woodpecker-ci/woodpecker/pull/3375)]
- fix: can't run multiple services on k8s [[#3395](https://github.com/woodpecker-ci/woodpecker/pull/3395)]
- Fix agent polling [[#3378](https://github.com/woodpecker-ci/woodpecker/pull/3378)]
- Remove empty strings from slice before parsing agent config [[#3387](https://github.com/woodpecker-ci/woodpecker/pull/3387)]
- Set correct link for commit [[#3368](https://github.com/woodpecker-ci/woodpecker/pull/3368)]
- Fix schema links [[#3369](https://github.com/woodpecker-ci/woodpecker/pull/3369)]
- Fix correctly handle gitlab pr closed events [[#3362](https://github.com/woodpecker-ci/woodpecker/pull/3362)]
- fix: update schema event_enum to remove error warning when.event [[#3357](https://github.com/woodpecker-ci/woodpecker/pull/3357)]
- Fix version check on next [[#3340](https://github.com/woodpecker-ci/woodpecker/pull/3340)]
- Ignore gitlab merge request events without code changes [[#3338](https://github.com/woodpecker-ci/woodpecker/pull/3338)]
- Ignore gitlab push events without commits [[#3339](https://github.com/woodpecker-ci/woodpecker/pull/3339)]
- Consider gitlab inherited permissions [[#3308](https://github.com/woodpecker-ci/woodpecker/pull/3308)]
- fix: agent panic when node is terminated during step execution [[#3331](https://github.com/woodpecker-ci/woodpecker/pull/3331)]

### 📈 Enhancement

- Enable golangci linter gomnd [[#3171](https://github.com/woodpecker-ci/woodpecker/pull/3171)]
- Apply "grpcnotrace" go build tag [[#3448](https://github.com/woodpecker-ci/woodpecker/pull/3448)]
- Simplify store interfaces [[#3437](https://github.com/woodpecker-ci/woodpecker/pull/3437)]
- Deprecate alternative names on secrets [[#3406](https://github.com/woodpecker-ci/woodpecker/pull/3406)]
- Store workflows/steps for blocked pipeline [[#2757](https://github.com/woodpecker-ci/woodpecker/pull/2757)]
- Parse email from Gitea webhook [[#3420](https://github.com/woodpecker-ci/woodpecker/pull/3420)]
- Replace http types on forge interface [[#3374](https://github.com/woodpecker-ci/woodpecker/pull/3374)]
- Prevent agent deletion when it's still running tasks [[#3377](https://github.com/woodpecker-ci/woodpecker/pull/3377)]
- Refactor internal services [[#915](https://github.com/woodpecker-ci/woodpecker/pull/915)]
- Lint for event filter and deprecate `exclude` [[#3222](https://github.com/woodpecker-ci/woodpecker/pull/3222)]
- Allow editing all environment variables in pipeline popups [[#3314](https://github.com/woodpecker-ci/woodpecker/pull/3314)]
- Parse backend options in backend [[#3227](https://github.com/woodpecker-ci/woodpecker/pull/3227)]
- Make agent usable for external backends [[#3270](https://github.com/woodpecker-ci/woodpecker/pull/3270)]
- Add no branches text [[#3312](https://github.com/woodpecker-ci/woodpecker/pull/3312)]
- Add loading spinner to repo list [[#3310](https://github.com/woodpecker-ci/woodpecker/pull/3310)]

### Misc

- Post on mastodon when releasing a new version [[#3509](https://github.com/woodpecker-ci/woodpecker/pull/3509)]
- chore(deps): update dependency alpine_3_18/ca-certificates to v20240226 [[#3501](https://github.com/woodpecker-ci/woodpecker/pull/3501)]
- fix(deps): update module github.com/google/go-github/v59 to v60 [[#3493](https://github.com/woodpecker-ci/woodpecker/pull/3493)]
- fix(deps): update dependency @intlify/unplugin-vue-i18n to v3 [[#3492](https://github.com/woodpecker-ci/woodpecker/pull/3492)]
- chore(deps): update dependency vue-tsc to v2 [[#3491](https://github.com/woodpecker-ci/woodpecker/pull/3491)]
- chore(deps): update dependency eslint-config-airbnb-typescript to v18 [[#3490](https://github.com/woodpecker-ci/woodpecker/pull/3490)]
- chore(deps): update web npm deps non-major [[#3489](https://github.com/woodpecker-ci/woodpecker/pull/3489)]
- fix(deps): update golang (packages) [[#3486](https://github.com/woodpecker-ci/woodpecker/pull/3486)]
- fix(deps): update module google.golang.org/protobuf to v1.33.0 [security] [[#3487](https://github.com/woodpecker-ci/woodpecker/pull/3487)]
- chore(deps): update docker.io/techknowlogick/xgo docker tag to go-1.22.1 [[#3476](https://github.com/woodpecker-ci/woodpecker/pull/3476)]
- chore(deps): update docker.io/golang docker tag to v1.22.1 [[#3475](https://github.com/woodpecker-ci/woodpecker/pull/3475)]
- Update prettier version [[#3471](https://github.com/woodpecker-ci/woodpecker/pull/3471)]
- chore(deps): update woodpeckerci/plugin-ready-release-go docker tag to v1.1.0 [[#3464](https://github.com/woodpecker-ci/woodpecker/pull/3464)]
- chore(deps): lock file maintenance [[#3465](https://github.com/woodpecker-ci/woodpecker/pull/3465)]
- chore(deps): update postgres docker tag to v16.2 [[#3461](https://github.com/woodpecker-ci/woodpecker/pull/3461)]
- chore(deps): update lycheeverse/lychee docker tag to v0.14.3 [[#3429](https://github.com/woodpecker-ci/woodpecker/pull/3429)]
- fix(deps): update golang (packages) [[#3430](https://github.com/woodpecker-ci/woodpecker/pull/3430)]
- More `when` filters [[#3407](https://github.com/woodpecker-ci/woodpecker/pull/3407)]
- Apply `documentation`/`ui` label to corresponding renovate updates [[#3400](https://github.com/woodpecker-ci/woodpecker/pull/3400)]
- chore(deps): update dependency eslint-plugin-simple-import-sort to v12 [[#3396](https://github.com/woodpecker-ci/woodpecker/pull/3396)]
- chore(deps): update typescript-eslint monorepo to v7 (major) [[#3397](https://github.com/woodpecker-ci/woodpecker/pull/3397)]
- fix(deps): update module github.com/google/go-github/v58 to v59 [[#3398](https://github.com/woodpecker-ci/woodpecker/pull/3398)]
- chore(deps): update docker.io/techknowlogick/xgo docker tag to go-1.22.0 [[#3392](https://github.com/woodpecker-ci/woodpecker/pull/3392)]
- chore(deps): update docker.io/golang docker tag [[#3391](https://github.com/woodpecker-ci/woodpecker/pull/3391)]
- fix(deps): update golang (packages) [[#3393](https://github.com/woodpecker-ci/woodpecker/pull/3393)]
- chore(deps): update docker.io/woodpeckerci/plugin-docker-buildx docker tag to v3.1.0 [[#3394](https://github.com/woodpecker-ci/woodpecker/pull/3394)]
- Add link checking [[#3371](https://github.com/woodpecker-ci/woodpecker/pull/3371)]
- Apply `dependencies` label to all PRs [[#3358](https://github.com/woodpecker-ci/woodpecker/pull/3358)]
- chore(deps): update docker.io/woodpeckerci/plugin-docker-buildx docker tag to v3.0.1 [[#3324](https://github.com/woodpecker-ci/woodpecker/pull/3324)]

## [2.3.0](https://github.com/woodpecker-ci/woodpecker/releases/tag/v2.3.0) - 2024-01-31

### ❤️ Thanks to all contributors! ❤️

@anbraten, @HerHde, @qwerty287, @pat-s, @renovate[bot], @lukashass, @zc-devs, @Alonsohhl, @healdropper, @eliasscosta, @runephilosof-karnovgroup

### ✨ Features

- Add release event [[#3226](https://github.com/woodpecker-ci/woodpecker/pull/3226)]

### 📚 Documentation

- Add release types [[#3303](https://github.com/woodpecker-ci/woodpecker/pull/3303)]
- Add opencollective footer [[#3281](https://github.com/woodpecker-ci/woodpecker/pull/3281)]
- Use array syntax in docs [[#3242](https://github.com/woodpecker-ci/woodpecker/pull/3242)]

### 🐛 Bug Fixes

- Fix Gitpod: Gitea auth token creation [[#3299](https://github.com/woodpecker-ci/woodpecker/pull/3299)]
- Fix agent updating [[#3287](https://github.com/woodpecker-ci/woodpecker/pull/3287)]
- Sanitize pod's step label [[#3275](https://github.com/woodpecker-ci/woodpecker/pull/3275)]
- Pipeline errors must be an array [[#3276](https://github.com/woodpecker-ci/woodpecker/pull/3276)]
- fix bitbucket SSO using UUID from bitbucket api response as ForgeRemoteID [[#3265](https://github.com/woodpecker-ci/woodpecker/pull/3265)]
- fix: bug pod service without label service [[#3256](https://github.com/woodpecker-ci/woodpecker/pull/3256)]
- Fix disabling PRs [[#3258](https://github.com/woodpecker-ci/woodpecker/pull/3258)]
- fix: bug annotations [[#3255](https://github.com/woodpecker-ci/woodpecker/pull/3255)]

### 📈 Enhancement

- Update theme on system color mode change [[#3296](https://github.com/woodpecker-ci/woodpecker/pull/3296)]
- Improve secrets availability checks [[#3271](https://github.com/woodpecker-ci/woodpecker/pull/3271)]
- Load more pipeline log lines (500 => 5000) [[#3212](https://github.com/woodpecker-ci/woodpecker/pull/3212)]
- Clean up models [[#3228](https://github.com/woodpecker-ci/woodpecker/pull/3228)]

### Misc

- chore(deps): update docker.io/techknowlogick/xgo docker tag to go-1.21.6 [[#3294](https://github.com/woodpecker-ci/woodpecker/pull/3294)]
- fix(deps): update docs npm deps non-major [[#3295](https://github.com/woodpecker-ci/woodpecker/pull/3295)]
- Remove deprecated `group` from config [[#3289](https://github.com/woodpecker-ci/woodpecker/pull/3289)]
- Add spellcheck config [[#3018](https://github.com/woodpecker-ci/woodpecker/pull/3018)]
- fix(deps): update golang (packages) [[#3284](https://github.com/woodpecker-ci/woodpecker/pull/3284)]
- chore(deps): lock file maintenance [[#3274](https://github.com/woodpecker-ci/woodpecker/pull/3274)]
- chore(deps): update web npm deps non-major [[#3273](https://github.com/woodpecker-ci/woodpecker/pull/3273)]
- Pin prettier version [[#3260](https://github.com/woodpecker-ci/woodpecker/pull/3260)]
- Fix prettier [[#3259](https://github.com/woodpecker-ci/woodpecker/pull/3259)]
- Update UI building in Makefile [[#3250](https://github.com/woodpecker-ci/woodpecker/pull/3250)]

## [2.2.2](https://github.com/woodpecker-ci/woodpecker/releases/tag/2.2.2) - 2024-01-21

### ❤️ Thanks to all contributors! ❤️

@6543

### Misc

- build: fix nfpm path for server binary [[#3246](https://github.com/woodpecker-ci/woodpecker/pull/3246)]

## [2.2.1](https://github.com/woodpecker-ci/woodpecker/releases/tag/v2.2.1) - 2024-01-21

### ❤️ Thanks to all contributors! ❤️

@6543

### 🐛 Bug Fixes

- Add gitea/forgejo driver check, to handle ErrUnknownVersion error [[#3243](https://github.com/woodpecker-ci/woodpecker/pull/3243)]

### Misc

- Build tarball for distribution packages [[#3244](https://github.com/woodpecker-ci/woodpecker/pull/3244)]

## [2.2.0](https://github.com/woodpecker-ci/woodpecker/releases/tag/v2.2.0) - 2024-01-21

### ❤️ Thanks to all contributors! ❤️

@qwerty287, @zc-devs, @renovate[bot], @mzampetakis, @healdropper, @6543, @micash545, @xoxys, @pat-s, @miry, @lukashass, @lafriks, @pre-commit-ci[bot], @anbraten, @andyhan, @KamilaBorowska

### 🔒 Security

- Update web dependencies [[#3234](https://github.com/woodpecker-ci/woodpecker/pull/3234)]

### ✨ Features

- Support custom steps entrypoint [[#2985](https://github.com/woodpecker-ci/woodpecker/pull/2985)]

### 📚 Documentation

- Add 2.2 docs [[#3237](https://github.com/woodpecker-ci/woodpecker/pull/3237)]
- Fix/improve issue templates [[#3232](https://github.com/woodpecker-ci/woodpecker/pull/3232)]
- Delete `FUNDING.yaml` [[#3193](https://github.com/woodpecker-ci/woodpecker/pull/3193)]
- Remove contributing/security to use globally defined [[#3192](https://github.com/woodpecker-ci/woodpecker/pull/3192)]
- Add "Kaniko" Plugin [[#3183](https://github.com/woodpecker-ci/woodpecker/pull/3183)]
- Document core development ideas [[#3184](https://github.com/woodpecker-ci/woodpecker/pull/3184)]
- Add continuous deployment cookbook [[#3098](https://github.com/woodpecker-ci/woodpecker/pull/3098)]
- Make k8s backend configuration docs in the same format as others [[#3081](https://github.com/woodpecker-ci/woodpecker/pull/3081)]
- Hide backend config options from TOC [[#3126](https://github.com/woodpecker-ci/woodpecker/pull/3126)]
- Add X/Twitter account [[#3127](https://github.com/woodpecker-ci/woodpecker/pull/3127)]
- Add ansible plugin [[#3115](https://github.com/woodpecker-ci/woodpecker/pull/3115)]
- Format depends_on example [[#3118](https://github.com/woodpecker-ci/woodpecker/pull/3118)]
- Use WOODPECKER_AGENT_SECRET instead of deprecated alternative [[#3103](https://github.com/woodpecker-ci/woodpecker/pull/3103)]
- Add Reviewdog ESLint plugin [[#3102](https://github.com/woodpecker-ci/woodpecker/pull/3102)]
- Mark local backend as stable [[#3088](https://github.com/woodpecker-ci/woodpecker/pull/3088)]
- Update Owners 2024 [[#3075](https://github.com/woodpecker-ci/woodpecker/pull/3075)]
- Add reviewdog golangci plugin [[#3080](https://github.com/woodpecker-ci/woodpecker/pull/3080)]
- Add Codeberg Pages Deploy plugin to plugins list [[#3054](https://github.com/woodpecker-ci/woodpecker/pull/3054)]

### 🐛 Bug Fixes

- Fixed Pods creation of WP services [[#3236](https://github.com/woodpecker-ci/woodpecker/pull/3236)]
- Fix Bitbucket get pull requests that ignores pagination [[#3235](https://github.com/woodpecker-ci/woodpecker/pull/3235)]
- Make PipelineConfig unique again [[#3215](https://github.com/woodpecker-ci/woodpecker/pull/3215)]
- Fix feed sorting [[#3155](https://github.com/woodpecker-ci/woodpecker/pull/3155)]
- Step status update dont set to running again once it got stoped [[#3151](https://github.com/woodpecker-ci/woodpecker/pull/3151)]
- Use step uuid instead of name in GRPC status calls [[#3143](https://github.com/woodpecker-ci/woodpecker/pull/3143)]
- Use UUID instead of step name where possible [[#3136](https://github.com/woodpecker-ci/woodpecker/pull/3136)]
- Use step type to detect services in Kubernetes backend [[#3141](https://github.com/woodpecker-ci/woodpecker/pull/3141)]
- Fix config base64 parsing to utf-8 [[#3110](https://github.com/woodpecker-ci/woodpecker/pull/3110)]
- Pin Gitea version [[#3104](https://github.com/woodpecker-ci/woodpecker/pull/3104)]
- Fix step `depends_on` as string in schema [[#3099](https://github.com/woodpecker-ci/woodpecker/pull/3099)]
- Fix slice unmarshaling [[#3097](https://github.com/woodpecker-ci/woodpecker/pull/3097)]
- Allow PR secrets to be used on close [[#3084](https://github.com/woodpecker-ci/woodpecker/pull/3084)]
- make event in pipeline schema also a constraint_list [[#3082](https://github.com/woodpecker-ci/woodpecker/pull/3082)]
- Fix badge's repoUrl with rootpath [[#3076](https://github.com/woodpecker-ci/woodpecker/pull/3076)]
- Load changed files for closed PR [[#3067](https://github.com/woodpecker-ci/woodpecker/pull/3067)]
- Fix build output paths [[#3065](https://github.com/woodpecker-ci/woodpecker/pull/3065)]
- Fix `when` and `depends_on` [[#3063](https://github.com/woodpecker-ci/woodpecker/pull/3063)]
- Fix DAG cycle detection [[#3049](https://github.com/woodpecker-ci/woodpecker/pull/3049)]
- Fix duplicated icons [[#3045](https://github.com/woodpecker-ci/woodpecker/pull/3045)]

### 📈 Enhancement

- Retrieve all user repo perms with a single API call [[#3211](https://github.com/woodpecker-ci/woodpecker/pull/3211)]
- Secured kubernetes backend configuration [[#3204](https://github.com/woodpecker-ci/woodpecker/pull/3204)]
- Use `assert` for tests [[#3201](https://github.com/woodpecker-ci/woodpecker/pull/3201)]
- Replace `goimports` with `gci` [[#3202](https://github.com/woodpecker-ci/woodpecker/pull/3202)]
- Remove multipart logger [[#3200](https://github.com/woodpecker-ci/woodpecker/pull/3200)]
- Added protocol in port configuration [[#2993](https://github.com/woodpecker-ci/woodpecker/pull/2993)]
- Kubernetes AppArmor and seccomp [[#3123](https://github.com/woodpecker-ci/woodpecker/pull/3123)]
- `cli exec`: let override existing environment values but print a warning [[#3140](https://github.com/woodpecker-ci/woodpecker/pull/3140)]
- Enable golangci linter forcetypeassert [[#3168](https://github.com/woodpecker-ci/woodpecker/pull/3168)]
- Enable golangci linter contextcheck [[#3170](https://github.com/woodpecker-ci/woodpecker/pull/3170)]
- Remove panic recovering [[#3162](https://github.com/woodpecker-ci/woodpecker/pull/3162)]
- More docker backend test remove more undocumented [[#3156](https://github.com/woodpecker-ci/woodpecker/pull/3156)]
- Lowercase all log strings [[#3173](https://github.com/woodpecker-ci/woodpecker/pull/3173)]
- Cleanups + prefer .yaml [[#3069](https://github.com/woodpecker-ci/woodpecker/pull/3069)]
- Use UUID as podName and cleanup arguments for Kubernetes backend [[#3135](https://github.com/woodpecker-ci/woodpecker/pull/3135)]
- Enable golangci linter stylecheck [[#3167](https://github.com/woodpecker-ci/woodpecker/pull/3167)]
- Clean up logging [[#3161](https://github.com/woodpecker-ci/woodpecker/pull/3161)]
- Enable `gocritic` and don't ignore globally [[#3159](https://github.com/woodpecker-ci/woodpecker/pull/3159)]
- Remove steps for publishing release branches [[#3125](https://github.com/woodpecker-ci/woodpecker/pull/3125)]
- Enable `nolintlint` [[#3158](https://github.com/woodpecker-ci/woodpecker/pull/3158)]
- Enable some linters [[#3129](https://github.com/woodpecker-ci/woodpecker/pull/3129)]
- Use name in backend types instead of alias [[#3142](https://github.com/woodpecker-ci/woodpecker/pull/3142)]
- Make service icon rotate [[#3149](https://github.com/woodpecker-ci/woodpecker/pull/3149)]
- Add step name as label to docker containers [[#3137](https://github.com/woodpecker-ci/woodpecker/pull/3137)]
- Use js-base64 on pipeline log page [[#3146](https://github.com/woodpecker-ci/woodpecker/pull/3146)]
- Flexible image pull secret reference [[#3016](https://github.com/woodpecker-ci/woodpecker/pull/3016)]
- Always show pipeline step list [[#3114](https://github.com/woodpecker-ci/woodpecker/pull/3114)]
- Add loading spinner and no pull request text [[#3113](https://github.com/woodpecker-ci/woodpecker/pull/3113)]
- Fix timeout settings contrast [[#3112](https://github.com/woodpecker-ci/woodpecker/pull/3112)]
- Unfold workflow when opening via URL [[#3106](https://github.com/woodpecker-ci/woodpecker/pull/3106)]
- Remove env argument of addons [[#3100](https://github.com/woodpecker-ci/woodpecker/pull/3100)]
- Move `cmd/common` to `shared` [[#3092](https://github.com/woodpecker-ci/woodpecker/pull/3092)]
- use semver for version comparsion [[#3042](https://github.com/woodpecker-ci/woodpecker/pull/3042)]
- Extend create plugin docs [[#3062](https://github.com/woodpecker-ci/woodpecker/pull/3062)]
- Remove old files [[#3077](https://github.com/woodpecker-ci/woodpecker/pull/3077)]
- Indicate if step is service [[#3078](https://github.com/woodpecker-ci/woodpecker/pull/3078)]
- Add imports checks to linter [[#3056](https://github.com/woodpecker-ci/woodpecker/pull/3056)]
- Remove workflow version again [[#3052](https://github.com/woodpecker-ci/woodpecker/pull/3052)]
- Add option to disable version check in admin web UI [[#3040](https://github.com/woodpecker-ci/woodpecker/pull/3040)]

### Misc

- chore(deps): update docker.io/woodpeckerci/plugin-docker-buildx docker tag to v3 [[#3229](https://github.com/woodpecker-ci/woodpecker/pull/3229)]
- Docs: Fix expression syntax docs url [[#3208](https://github.com/woodpecker-ci/woodpecker/pull/3208)]
- Add schema test for depends_on [[#3205](https://github.com/woodpecker-ci/woodpecker/pull/3205)]
- chore(deps): lock file maintenance [[#3190](https://github.com/woodpecker-ci/woodpecker/pull/3190)]
- Do not run prettier with pre-commit [[#3196](https://github.com/woodpecker-ci/woodpecker/pull/3196)]
- fix(deps): update module github.com/google/go-github/v57 to v58 [[#3187](https://github.com/woodpecker-ci/woodpecker/pull/3187)]
- chore(deps): update docker.io/golang docker tag to v1.21.6 [[#3189](https://github.com/woodpecker-ci/woodpecker/pull/3189)]
- chore(deps): update docker.io/woodpeckerci/plugin-docker-buildx [[#3186](https://github.com/woodpecker-ci/woodpecker/pull/3186)]
- fix(deps): update golang (packages) [[#3185](https://github.com/woodpecker-ci/woodpecker/pull/3185)]
- declare different when statements once and reuse them [[#3176](https://github.com/woodpecker-ci/woodpecker/pull/3176)]
- Add `make clean-all` [[#3152](https://github.com/woodpecker-ci/woodpecker/pull/3152)]
- Fix `version.json` updates [[#3057](https://github.com/woodpecker-ci/woodpecker/pull/3057)]
- [pre-commit.ci] pre-commit autoupdate [[#3101](https://github.com/woodpecker-ci/woodpecker/pull/3101)]
- Update dependency @vitejs/plugin-vue to v5 [[#3074](https://github.com/woodpecker-ci/woodpecker/pull/3074)]
- Use CI vars for plugin [[#3061](https://github.com/woodpecker-ci/woodpecker/pull/3061)]
- Use `yamllint` [[#3066](https://github.com/woodpecker-ci/woodpecker/pull/3066)]
- Use dag in ci config [[#3010](https://github.com/woodpecker-ci/woodpecker/pull/3010)]

## [2.1.1](https://github.com/woodpecker-ci/woodpecker/releases/tag/v2.1.1) - 2023-12-27

### ❤️ Thanks to all contributors! ❤️

@6543, @andyhan, @qwerty287

### 🐛 Bug Fixes

- trim v on version check [[#3039](https://github.com/woodpecker-ci/woodpecker/pull/3039)]
- make backend step dag generation deterministic [[#3037](https://github.com/woodpecker-ci/woodpecker/pull/3037)]
- Fix showing wrong badge url when root path is set [[#3033](https://github.com/woodpecker-ci/woodpecker/pull/3033)]
- Fix docs label [[#3028](https://github.com/woodpecker-ci/woodpecker/pull/3028)]

### 📚 Documentation

- Update go report card badge [[#3029](https://github.com/woodpecker-ci/woodpecker/pull/3029)]

### Misc

- Add some tests [[#3030](https://github.com/woodpecker-ci/woodpecker/pull/3030)]

## [2.1.0](https://github.com/woodpecker-ci/woodpecker/releases/tag/v2.1.0) - 2023-12-26

### ❤️ Thanks to all contributors! ❤️

@anbraten, @lukashass, @qwerty287, @6543, @Lerentis, @renovate[bot], @zc-devs, @johanvdw, @lafriks, @runephilosof-karnovgroup, @allanger, @xoxys, @gapanyc, @mikhail-putilov, @kaylynb, @voidcontext, @robbie-cahill, @micash545, @dominic-p, @mzampetakis

### ✨ Features

- Add pull request closed event [[#2684](https://github.com/woodpecker-ci/woodpecker/pull/2684)]
- Add depends_on support for steps [[#2771](https://github.com/woodpecker-ci/woodpecker/pull/2771)]
- gitlab: support nested repos [[#2981](https://github.com/woodpecker-ci/woodpecker/pull/2981)]
- Support go plugins for forges and agent backends [[#2751](https://github.com/woodpecker-ci/woodpecker/pull/2751)]

### 📈 Enhancement

- Show default branch on top [[#3019](https://github.com/woodpecker-ci/woodpecker/pull/3019)]
- Support more addon types [[#2984](https://github.com/woodpecker-ci/woodpecker/pull/2984)]
- Hide PR tab if PRs are disabled [[#3004](https://github.com/woodpecker-ci/woodpecker/pull/3004)]
- Switch to ULID [[#2986](https://github.com/woodpecker-ci/woodpecker/pull/2986)]
- Ignore pipelines without config [[#2949](https://github.com/woodpecker-ci/woodpecker/pull/2949)]
- Link labels to input and select [[#2974](https://github.com/woodpecker-ci/woodpecker/pull/2974)]
- Register Agent with hostname [[#2936](https://github.com/woodpecker-ci/woodpecker/pull/2936)]
- Update slogan & logo [[#2962](https://github.com/woodpecker-ci/woodpecker/pull/2962)]
- Improve error handling when activating a repository [[#2965](https://github.com/woodpecker-ci/woodpecker/pull/2965)]
- Add check for storage where repo/org name is empty [[#2968](https://github.com/woodpecker-ci/woodpecker/pull/2968)]
- Update pipeline icons [[#2783](https://github.com/woodpecker-ci/woodpecker/pull/2783)]
- Kubernetes refactor [[#2794](https://github.com/woodpecker-ci/woodpecker/pull/2794)]
- Export changed files via builtin environment variables [[#2935](https://github.com/woodpecker-ci/woodpecker/pull/2935)]
- Show secrets from org and global level [[#2873](https://github.com/woodpecker-ci/woodpecker/pull/2873)]
- Only update pipelineStatus in one place [[#2952](https://github.com/woodpecker-ci/woodpecker/pull/2952)]
- Rename `engine` to `backend` [[#2950](https://github.com/woodpecker-ci/woodpecker/pull/2950)]
- Add linting for `log.Fatal()` [[#2946](https://github.com/woodpecker-ci/woodpecker/pull/2946)]
- Remove separate root path config [[#2943](https://github.com/woodpecker-ci/woodpecker/pull/2943)]
- init CI_COMMIT_TAG if commit ref is a tag [[#2934](https://github.com/woodpecker-ci/woodpecker/pull/2934)]
- Update go module path for major version 2 [[#2905](https://github.com/woodpecker-ci/woodpecker/pull/2905)]
- Unify date/time dependencies [[#2891](https://github.com/woodpecker-ci/woodpecker/pull/2891)]
- Add linting for `any` [[#2893](https://github.com/woodpecker-ci/woodpecker/pull/2893)]
- Fix vite deprecations [[#2885](https://github.com/woodpecker-ci/woodpecker/pull/2885)]
- Migrate to Xormigrate [[#2711](https://github.com/woodpecker-ci/woodpecker/pull/2711)]
- Simple security context options (Kubernetes) [[#2550](https://github.com/woodpecker-ci/woodpecker/pull/2550)]
- Changes PullRequest Index to ForgeRemoteID type [[#2823](https://github.com/woodpecker-ci/woodpecker/pull/2823)]

### 🐛 Bug Fixes

- Hide queue visualization if nothing to show [[#3003](https://github.com/woodpecker-ci/woodpecker/pull/3003)]
- fix and lint swagger file [[#3007](https://github.com/woodpecker-ci/woodpecker/pull/3007)]
- Fix IPv6 host aliases for kubernetes [[#2992](https://github.com/woodpecker-ci/woodpecker/pull/2992)]
- Fix cli lint throwing error on warnings  [[#2995](https://github.com/woodpecker-ci/woodpecker/pull/2995)]
- Fix static file caching [[#2975](https://github.com/woodpecker-ci/woodpecker/pull/2975)]
- Gitea driver: ignore GetOrg error if we get a valid user. [[#2967](https://github.com/woodpecker-ci/woodpecker/pull/2967)]
- feat(k8s): Add a port name to service definition [[#2933](https://github.com/woodpecker-ci/woodpecker/pull/2933)]
- Fix error container overflow [[#2957](https://github.com/woodpecker-ci/woodpecker/pull/2957)]
- ignore some errors on repairAllRepos [[#2792](https://github.com/woodpecker-ci/woodpecker/pull/2792)]
- Allow to restart pipelines that has warnings [[#2939](https://github.com/woodpecker-ci/woodpecker/pull/2939)]
- Fix skipped pipelines model [[#2923](https://github.com/woodpecker-ci/woodpecker/pull/2923)]
- fix: Add `backend_options` to service linter entry [[#2930](https://github.com/woodpecker-ci/woodpecker/pull/2930)]
- Fix flags added multiple times [[#2914](https://github.com/woodpecker-ci/woodpecker/pull/2914)]
- Fix schema validation with array syntax for clone and services [[#2920](https://github.com/woodpecker-ci/woodpecker/pull/2920)]
- Fix prometheus docs [[#2919](https://github.com/woodpecker-ci/woodpecker/pull/2919)]
- Fix podman agent container in v2 [[#2897](https://github.com/woodpecker-ci/woodpecker/pull/2897)]
- Fix bitbucket org fetching [[#2874](https://github.com/woodpecker-ci/woodpecker/pull/2874)]
- Only deploy docs on `main` [[#2892](https://github.com/woodpecker-ci/woodpecker/pull/2892)]
- Fix pipeline-related environment [[#2876](https://github.com/woodpecker-ci/woodpecker/pull/2876)]
- Fix version check partially [[#2871](https://github.com/woodpecker-ci/woodpecker/pull/2871)]
- Fix unregistering agents when using agent tokens [[#2870](https://github.com/woodpecker-ci/woodpecker/pull/2870)]

### 📚 Documentation

- [Awesome Woodpecker] added yet another autoscaler [[#3011](https://github.com/woodpecker-ci/woodpecker/pull/3011)]
- Add cookbook blog and improve docs [[#3002](https://github.com/woodpecker-ci/woodpecker/pull/3002)]
- Replace multi-pipelines with workflows on docs frontpage [[#2990](https://github.com/woodpecker-ci/woodpecker/pull/2990)]
- Update README badges [[#2956](https://github.com/woodpecker-ci/woodpecker/pull/2956)]
- Update 20-kubernetes.md [[#2927](https://github.com/woodpecker-ci/woodpecker/pull/2927)]
- Add release documentation to CONTRIBUTING [[#2917](https://github.com/woodpecker-ci/woodpecker/pull/2917)]
- Add nix-attic plugin to the index [[#2889](https://github.com/woodpecker-ci/woodpecker/pull/2889)]
- Add usage with Tunnelmole to docs [[#2881](https://github.com/woodpecker-ci/woodpecker/pull/2881)]
- Improve code blocks in docs [[#2879](https://github.com/woodpecker-ci/woodpecker/pull/2879)]
- Add a blog post [[#2877](https://github.com/woodpecker-ci/woodpecker/pull/2877)]
- Add documentation on Kubernetes securityContext [[#2822](https://github.com/woodpecker-ci/woodpecker/pull/2822)]
- Add default page to categories [[#2869](https://github.com/woodpecker-ci/woodpecker/pull/2869)]
- Use same format for Github docs as used for the other forges [[#2866](https://github.com/woodpecker-ci/woodpecker/pull/2866)]

### Misc

- chore(deps): update dependency isomorphic-dompurify to v2 [[#3001](https://github.com/woodpecker-ci/woodpecker/pull/3001)]
- fix(deps): update dependency @intlify/unplugin-vue-i18n to v2 [[#2998](https://github.com/woodpecker-ci/woodpecker/pull/2998)]
- Fix go in gitpod [[#2973](https://github.com/woodpecker-ci/woodpecker/pull/2973)]
- fix(deps): update module google.golang.org/grpc to v1.60.1 [[#2969](https://github.com/woodpecker-ci/woodpecker/pull/2969)]
- chore(deps): update docker.io/alpine docker tag to v3.19 [[#2970](https://github.com/woodpecker-ci/woodpecker/pull/2970)]
- Fix broken gated repos [[#2959](https://github.com/woodpecker-ci/woodpecker/pull/2959)]
- fix(deps): update golang (packages) [[#2958](https://github.com/woodpecker-ci/woodpecker/pull/2958)]
- Update docker.io/techknowlogick/xgo Docker tag to go-1.21.5 [[#2926](https://github.com/woodpecker-ci/woodpecker/pull/2926)]
- Update docker.io/golang Docker tag to v1.21.5 [[#2925](https://github.com/woodpecker-ci/woodpecker/pull/2925)]
- Lock file maintenance [[#2910](https://github.com/woodpecker-ci/woodpecker/pull/2910)]
- Update web npm deps non-major [[#2909](https://github.com/woodpecker-ci/woodpecker/pull/2909)]
- Update docs npm deps non-major [[#2908](https://github.com/woodpecker-ci/woodpecker/pull/2908)]
- Update golang (packages) [[#2904](https://github.com/woodpecker-ci/woodpecker/pull/2904)]
- Update module github.com/google/go-github/v56 to v57 [[#2899](https://github.com/woodpecker-ci/woodpecker/pull/2899)]
- Update dependency marked to v11 [[#2898](https://github.com/woodpecker-ci/woodpecker/pull/2898)]
- Update dependency vite-svg-loader to v5 [[#2837](https://github.com/woodpecker-ci/woodpecker/pull/2837)]
- Update golang (packages) [[#2894](https://github.com/woodpecker-ci/woodpecker/pull/2894)]
- Update web npm deps non-major [[#2895](https://github.com/woodpecker-ci/woodpecker/pull/2895)]
- Update web npm deps non-major [[#2884](https://github.com/woodpecker-ci/woodpecker/pull/2884)]
- Update docker.io/woodpeckerci/plugin-docker-buildx Docker tag to v2.2.1 [[#2883](https://github.com/woodpecker-ci/woodpecker/pull/2883)]

## [2.0.0](https://github.com/woodpecker-ci/woodpecker/releases/tag/v2.0.0) - 2023-11-23

### ❤️ Thanks to all contributors! ❤️

@qwerty287, @anbraten, @6543, @renovate[bot], @pat-s, @zc-devs, @xoxys, @lafriks, @silverwind, @pre-commit-ci[bot], @riczescaran, @J-Ha, @Janik-Haag, @jbiblio, @runephilosof-karnovgroup, @bitethecode, @HamburgerJungeJr, @nitram509, @JohnWalkerx, @OskarsPakers, @Exar04, @dominic-p, @categulario, @mzampetakis, @Timshel, @Denperidge, @tomix1024, @lonix1, @s3lph, @math3vz, @LTek-online, @testwill, @klinux, @pinpox, @hpidcock, @ChewingBever, @azdle, @praneeth-ovckd

### 💥 Breaking changes

- Rename `link` to `url` [[#2812](https://github.com/woodpecker-ci/woodpecker/pull/2812)]
- Revert to singular CLI args [[#2820](https://github.com/woodpecker-ci/woodpecker/pull/2820)]
- Use int64 for IDs in woodpecker client lib [[#2703](https://github.com/woodpecker-ci/woodpecker/pull/2703)]
- Woodpecker-go: Use Feed instead of Activity [[#2690](https://github.com/woodpecker-ci/woodpecker/pull/2690)]
- Do not sanitzie secrets with 3 or less chars [[#2680](https://github.com/woodpecker-ci/woodpecker/pull/2680)]
- fix(deps): update docker to v24 [[#2675](https://github.com/woodpecker-ci/woodpecker/pull/2675)]
- Remove `WOODPECKER_DOCS` config [[#2647](https://github.com/woodpecker-ci/woodpecker/pull/2647)]
- Remove plugin-only option from secrets [[#2213](https://github.com/woodpecker-ci/woodpecker/pull/2213)]
- Remove deprecated API paths [[#2639](https://github.com/woodpecker-ci/woodpecker/pull/2639)]
- Remove SSH backend [[#2635](https://github.com/woodpecker-ci/woodpecker/pull/2635)]
- Remove deprecated `build` command [[#2602](https://github.com/woodpecker-ci/woodpecker/pull/2602)]
- Deprecate "platform" filter in favour of "labels" [[#2181](https://github.com/woodpecker-ci/woodpecker/pull/2181)]
- Remove useless "sync" option from RepoListOpts from the client lib [[#2090](https://github.com/woodpecker-ci/woodpecker/pull/2090)]
- Drop deprecated built-in environment variables [[#2048](https://github.com/woodpecker-ci/woodpecker/pull/2048)]

### 🔒 Security

- Never log tokens [[#2466](https://github.com/woodpecker-ci/woodpecker/pull/2466)]
- Check permissions on repo lookup [[#2357](https://github.com/woodpecker-ci/woodpecker/pull/2357)]
- Change token logging to trace level [[#2247](https://github.com/woodpecker-ci/woodpecker/pull/2247)]
- Validate webhook before changing any data [[#2221](https://github.com/woodpecker-ci/woodpecker/pull/2221)]

### ✨ Features

- Add version and update notes [[#2722](https://github.com/woodpecker-ci/woodpecker/pull/2722)]
- Add repos list for admins [[#2347](https://github.com/woodpecker-ci/woodpecker/pull/2347)]
- Add org list [[#2338](https://github.com/woodpecker-ci/woodpecker/pull/2338)]
- Add option to configure tolerations in kubernetes backend [[#2249](https://github.com/woodpecker-ci/woodpecker/pull/2249)]
- Support user secrets [[#2126](https://github.com/woodpecker-ci/woodpecker/pull/2126)]
- Add opt save global log output to file [[#2115](https://github.com/woodpecker-ci/woodpecker/pull/2115)]
- Support bitbucket Dir() and support multi-workflows [[#2045](https://github.com/woodpecker-ci/woodpecker/pull/2045)]
- Add ping command to server to allow container healthchecks [[#2030](https://github.com/woodpecker-ci/woodpecker/pull/2030)]

### 📚 Documentation

- Add 2.0.0 post [[#2864](https://github.com/woodpecker-ci/woodpecker/pull/2864)]
- Add extend env plugin [[#2847](https://github.com/woodpecker-ci/woodpecker/pull/2847)]
- mark v1.0.x as unmaintained [[#2818](https://github.com/woodpecker-ci/woodpecker/pull/2818)]
- Update docs npm deps non-major [[#2799](https://github.com/woodpecker-ci/woodpecker/pull/2799)]
- Add docs about Gitea on same host and update docker-compose example [[#2752](https://github.com/woodpecker-ci/woodpecker/pull/2752)]
- Update docusaurus plugin [[#2804](https://github.com/woodpecker-ci/woodpecker/pull/2804)]
- Mark kubernetes backend as fully supported [[#2756](https://github.com/woodpecker-ci/woodpecker/pull/2756)]
- Update docusaurus to v3 [[#2732](https://github.com/woodpecker-ci/woodpecker/pull/2732)]
- Fix the wrong link to the cron job document [[#2740](https://github.com/woodpecker-ci/woodpecker/pull/2740)]
- Improve secrets documentation [[#2707](https://github.com/woodpecker-ci/woodpecker/pull/2707)]
- Add woodpecker-lint tool [[#2648](https://github.com/woodpecker-ci/woodpecker/pull/2648)]
- Add autoscaler docs [[#2631](https://github.com/woodpecker-ci/woodpecker/pull/2631)]
- Rework setup docs [[#2630](https://github.com/woodpecker-ci/woodpecker/pull/2630)]
- doc: improve prometheus docs [[#2617](https://github.com/woodpecker-ci/woodpecker/pull/2617)]
- docs add nixos install instructions [[#2616](https://github.com/woodpecker-ci/woodpecker/pull/2616)]
- Add prettier plugin [[#2621](https://github.com/woodpecker-ci/woodpecker/pull/2621)]
- [doc] improve documentation WOODPECKER_SESSION_EXPIRES [[#2603](https://github.com/woodpecker-ci/woodpecker/pull/2603)]
- Update documentation WRT to recent `$platform` changes [[#2531](https://github.com/woodpecker-ci/woodpecker/pull/2531)]
- Add plugin "GitHub release" [[#2592](https://github.com/woodpecker-ci/woodpecker/pull/2592)]
- Cleanup docs [[#2478](https://github.com/woodpecker-ci/woodpecker/pull/2478)]
- Add plugin "Release helper" [[#2584](https://github.com/woodpecker-ci/woodpecker/pull/2584)]
- Add plugin "Gitea Create Pull Request" to plugin index [[#2581](https://github.com/woodpecker-ci/woodpecker/pull/2581)]
- Adjust github scopes and clarify documentation. [[#2578](https://github.com/woodpecker-ci/woodpecker/pull/2578)]
- Remove redundant definition of webhook form docs [[#2561](https://github.com/woodpecker-ci/woodpecker/pull/2561)]
- Add notes about CRI-O specific config [[#2546](https://github.com/woodpecker-ci/woodpecker/pull/2546)]
- Fix incorrect yaml syntax for `ref` in docs [[#2518](https://github.com/woodpecker-ci/woodpecker/pull/2518)]
- Local image documentation [[#2521](https://github.com/woodpecker-ci/woodpecker/pull/2521)]
- Adds bitbucket tag support in docs [[#2536](https://github.com/woodpecker-ci/woodpecker/pull/2536)]
- Fix docs duplicate WOODPECKER_HOST assignment [[#2501](https://github.com/woodpecker-ci/woodpecker/pull/2501)]
- Update github auth install [[#2499](https://github.com/woodpecker-ci/woodpecker/pull/2499)]
- Update GH app installation instructions [[#2472](https://github.com/woodpecker-ci/woodpecker/pull/2472)]
- Add videos [[#2465](https://github.com/woodpecker-ci/woodpecker/pull/2465)]
- docs: missing info for runs_on [[#2457](https://github.com/woodpecker-ci/woodpecker/pull/2457)]
- Add hint about alternative pipeline skip syntax [[#2443](https://github.com/woodpecker-ci/woodpecker/pull/2443)]
- Fix typo in GitLab docs [[#2376](https://github.com/woodpecker-ci/woodpecker/pull/2376)]
- clarify setup with gitlab with RFC1918 nets and non standard TLDs [[#2363](https://github.com/woodpecker-ci/woodpecker/pull/2363)]
- Clarify env var `CI` in docs [[#2349](https://github.com/woodpecker-ci/woodpecker/pull/2349)]
- docs: yaml cheatsheet for advanced syntax [[#2329](https://github.com/woodpecker-ci/woodpecker/pull/2329)]
- Improve explanation for globs in when:path [[#2252](https://github.com/woodpecker-ci/woodpecker/pull/2252)]
- Fix usage description for backend-http-proxy flag [[#2250](https://github.com/woodpecker-ci/woodpecker/pull/2250)]
- Restructure k8s documentation [[#2193](https://github.com/woodpecker-ci/woodpecker/pull/2193)]
- Update list of "projects using Woodpecker" [[#2196](https://github.com/woodpecker-ci/woodpecker/pull/2196)]
- Update 92-awesome.md [[#2195](https://github.com/woodpecker-ci/woodpecker/pull/2195)]
- Better blog title/desc [[#2182](https://github.com/woodpecker-ci/woodpecker/pull/2182)]
- Fix version in FAQ [[#2101](https://github.com/woodpecker-ci/woodpecker/pull/2101)]
- Add blog posts/tutorials [[#2095](https://github.com/woodpecker-ci/woodpecker/pull/2095)]
- update version docs about versioning [[#2086](https://github.com/woodpecker-ci/woodpecker/pull/2086)]
- Fix client example [[#2085](https://github.com/woodpecker-ci/woodpecker/pull/2085)]
- Update docs deps to address cves [[#2080](https://github.com/woodpecker-ci/woodpecker/pull/2080)]
- fix: global registry docs [[#2070](https://github.com/woodpecker-ci/woodpecker/pull/2070)]
- Improve bitbucket docs [[#2066](https://github.com/woodpecker-ci/woodpecker/pull/2066)]
- update docs about versioning [[#2043](https://github.com/woodpecker-ci/woodpecker/pull/2043)]
- Set v1.0 documents as default and mark v0.15 as unmaintained [[#2034](https://github.com/woodpecker-ci/woodpecker/pull/2034)]

### 📈 Enhancement

- Cleanup plugins index [[#2856](https://github.com/woodpecker-ci/woodpecker/pull/2856)]
- Bump default clone image version to 2.4.0 [[#2852](https://github.com/woodpecker-ci/woodpecker/pull/2852)]
- Signal to clients the hook and event routes where removed [[#2826](https://github.com/woodpecker-ci/woodpecker/pull/2826)]
- Replace `interface{}` with `any` [[#2807](https://github.com/woodpecker-ci/woodpecker/pull/2807)]
- Fix repo owner filter [[#2808](https://github.com/woodpecker-ci/woodpecker/pull/2808)]
- Sort agents list by ID [[#2795](https://github.com/woodpecker-ci/woodpecker/pull/2795)]
- Fix css loading order in head [[#2785](https://github.com/woodpecker-ci/woodpecker/pull/2785)]
- Fix error color contrast in dark theme [[#2778](https://github.com/woodpecker-ci/woodpecker/pull/2778)]
- Replace linter icons to match theme [[#2765](https://github.com/woodpecker-ci/woodpecker/pull/2765)]
- Switch to go vanity urls [[#2706](https://github.com/woodpecker-ci/woodpecker/pull/2706)]
- Add workflow version [[#2476](https://github.com/woodpecker-ci/woodpecker/pull/2476)]
- UI enhancements/fixes [[#2754](https://github.com/woodpecker-ci/woodpecker/pull/2754)]
- Fail on missing secrets [[#2749](https://github.com/woodpecker-ci/woodpecker/pull/2749)]
- Add deprecation warnings [[#2725](https://github.com/woodpecker-ci/woodpecker/pull/2725)]
- Enhance linter and errors [[#1572](https://github.com/woodpecker-ci/woodpecker/pull/1572)]
- Option to change temp dir for local backend [[#2702](https://github.com/woodpecker-ci/woodpecker/pull/2702)]
- Revert breaking pipeline changes [[#2677](https://github.com/woodpecker-ci/woodpecker/pull/2677)]
- Add ports into pipeline backend step model [[#2656](https://github.com/woodpecker-ci/woodpecker/pull/2656)]
- Unregister stateless agents from server on termination [[#2606](https://github.com/woodpecker-ci/woodpecker/pull/2606)]
- Let the backend engine report the current platform [[#2688](https://github.com/woodpecker-ci/woodpecker/pull/2688)]
- Showing the pending pipelines on top [[#1488](https://github.com/woodpecker-ci/woodpecker/pull/1488)]
- Print local backend command logs [[#2678](https://github.com/woodpecker-ci/woodpecker/pull/2678)]
- Report problems with listening to ports and exit [[#2102](https://github.com/woodpecker-ci/woodpecker/pull/2102)]
- Use path.Join for server side path generation [[#2689](https://github.com/woodpecker-ci/woodpecker/pull/2689)]
- Refactor UI dark/bright mode [[#2590](https://github.com/woodpecker-ci/woodpecker/pull/2590)]
- Stop steps after they are done [[#2681](https://github.com/woodpecker-ci/woodpecker/pull/2681)]
- Fix where syntax [[#2676](https://github.com/woodpecker-ci/woodpecker/pull/2676)]
- Add "Repair all" button [[#2642](https://github.com/woodpecker-ci/woodpecker/pull/2642)]
- Use pagination utils [[#2633](https://github.com/woodpecker-ci/woodpecker/pull/2633)]
- Dynamic forge request size [[#2622](https://github.com/woodpecker-ci/woodpecker/pull/2622)]
- Update to docker 23 [[#2577](https://github.com/woodpecker-ci/woodpecker/pull/2577)]
- Refactor/simplify pubsub [[#2554](https://github.com/woodpecker-ci/woodpecker/pull/2554)]
- Refactor pipeline parsing and forge refreshing [[#2527](https://github.com/woodpecker-ci/woodpecker/pull/2527)]
- Fix gitlab hooks and simplify config extension [[#2537](https://github.com/woodpecker-ci/woodpecker/pull/2537)]
- Set home variable in local backend for windows  [[#2323](https://github.com/woodpecker-ci/woodpecker/pull/2323)]
- Some cleanups about host config [[#2490](https://github.com/woodpecker-ci/woodpecker/pull/2490)]
- Fix usage of WOODPECKER_ROOT_PATH [[#2485](https://github.com/woodpecker-ci/woodpecker/pull/2485)]
- Some UI enhancement [[#2468](https://github.com/woodpecker-ci/woodpecker/pull/2468)]
- Harmonize pipeline status information and add a review link to the approval [[#2345](https://github.com/woodpecker-ci/woodpecker/pull/2345)]
- Add Renovate [[#2360](https://github.com/woodpecker-ci/woodpecker/pull/2360)]
- Add option to render button as link [[#2378](https://github.com/woodpecker-ci/woodpecker/pull/2378)]
- Close sidebar on outside clicks [[#2325](https://github.com/woodpecker-ci/woodpecker/pull/2325)]
- Add release helper [[#1976](https://github.com/woodpecker-ci/woodpecker/pull/1976)]
- Use API error helpers and improve response codes [[#2366](https://github.com/woodpecker-ci/woodpecker/pull/2366)]
- Import packages only once [[#2362](https://github.com/woodpecker-ci/woodpecker/pull/2362)]
- Execute `make generate` with new versions [[#2365](https://github.com/woodpecker-ci/woodpecker/pull/2365)]
- Only show commit title [[#2361](https://github.com/woodpecker-ci/woodpecker/pull/2361)]
- Truncate commit message in pipeline log view header [[#2356](https://github.com/woodpecker-ci/woodpecker/pull/2356)]
- Increase header padding again [[#2348](https://github.com/woodpecker-ci/woodpecker/pull/2348)]
- Use full width header on pipeline view and show repo name [[#2327](https://github.com/woodpecker-ci/woodpecker/pull/2327)]
- Use html list for changed files list [[#2346](https://github.com/woodpecker-ci/woodpecker/pull/2346)]
- Show that repo is disabled [[#2340](https://github.com/woodpecker-ci/woodpecker/pull/2340)]
- Add spacing to pipeline feed spinner [[#2326](https://github.com/woodpecker-ci/woodpecker/pull/2326)]
- Autodetect host platform in Makefile [[#2322](https://github.com/woodpecker-ci/woodpecker/pull/2322)]
- Add "plugin" support to local backend [[#2239](https://github.com/woodpecker-ci/woodpecker/pull/2239)]
- Rename grpc pipeline to workflow [[#2173](https://github.com/woodpecker-ci/woodpecker/pull/2173)]
- Pass netrc data to external config service request [[#2310](https://github.com/woodpecker-ci/woodpecker/pull/2310)]
- Create settings-panel vue component and use InputFields [[#2177](https://github.com/woodpecker-ci/woodpecker/pull/2177)]
- Use browser-native tooltips [[#2189](https://github.com/woodpecker-ci/woodpecker/pull/2189)]
- Improve agent rpc retry logic with exponential backoff [[#2205](https://github.com/woodpecker-ci/woodpecker/pull/2205)]
- Skip settings proxy config with WithProxy if its empty [[#2242](https://github.com/woodpecker-ci/woodpecker/pull/2242)]
- Move hook and events-stream routes to use `/api` prefix [[#2212](https://github.com/woodpecker-ci/woodpecker/pull/2212)]
- Add SSH clone URL env var [[#2198](https://github.com/woodpecker-ci/woodpecker/pull/2198)]
- Small improvements to mobile interface [[#2202](https://github.com/woodpecker-ci/woodpecker/pull/2202)]
- Switch to upstream ttlcache [[#2187](https://github.com/woodpecker-ci/woodpecker/pull/2187)]
- Convert EqualStringSlice to generic EqualSliceValues [[#2179](https://github.com/woodpecker-ci/woodpecker/pull/2179)]
- Pass netrc to trusted clone images [[#2163](https://github.com/woodpecker-ci/woodpecker/pull/2163)]
- Use Vue setup directive [[#2165](https://github.com/woodpecker-ci/woodpecker/pull/2165)]
- Release file lock on USR1 signal [[#2151](https://github.com/woodpecker-ci/woodpecker/pull/2151)]
- Use min/max width for pipeline step list [[#2141](https://github.com/woodpecker-ci/woodpecker/pull/2141)]
- Add header to pipeline log and always show buttons [[#2140](https://github.com/woodpecker-ci/woodpecker/pull/2140)]
- Use fix width for pipeline step list [[#2138](https://github.com/woodpecker-ci/woodpecker/pull/2138)]
- Make sure we dont have hidden options for backend and pipeline compiler [[#2123](https://github.com/woodpecker-ci/woodpecker/pull/2123)]
- Enhance local backend [[#2017](https://github.com/woodpecker-ci/woodpecker/pull/2017)]
- Don't show badge without information [[#2130](https://github.com/woodpecker-ci/woodpecker/pull/2130)]
- CLI repo sync: Show `forge-remote-id` [[#2103](https://github.com/woodpecker-ci/woodpecker/pull/2103)]
- Lazy-load TimeAgo locales [[#2094](https://github.com/woodpecker-ci/woodpecker/pull/2094)]
- Improve user settings [[#2087](https://github.com/woodpecker-ci/woodpecker/pull/2087)]
- Allow to disable swagger [[#2093](https://github.com/woodpecker-ci/woodpecker/pull/2093)]
- Use consistent woodpecker color scheme [[#2003](https://github.com/woodpecker-ci/woodpecker/pull/2003)]
- Change master to main [[#2044](https://github.com/woodpecker-ci/woodpecker/pull/2044)]
- Remove default branch fallbacks [[#2065](https://github.com/woodpecker-ci/woodpecker/pull/2065)]
- Remove fallback check for old sqlite file location [[#2046](https://github.com/woodpecker-ci/woodpecker/pull/2046)]
- Include the function name in generic datastore errors  [[#2041](https://github.com/woodpecker-ci/woodpecker/pull/2041)]

### 🐛 Bug Fixes

- Fix plugin URLs [[#2850](https://github.com/woodpecker-ci/woodpecker/pull/2850)]
- Fix env vars and add UI url [[#2811](https://github.com/woodpecker-ci/woodpecker/pull/2811)]
- Fix paths for version check [[#2816](https://github.com/woodpecker-ci/woodpecker/pull/2816)]
- Add `privileged` schema definition [[#2777](https://github.com/woodpecker-ci/woodpecker/pull/2777)]
- Use unique label selector for pod label for kubernetes services [[#2723](https://github.com/woodpecker-ci/woodpecker/pull/2723)]
- Some UI fixes [[#2698](https://github.com/woodpecker-ci/woodpecker/pull/2698)]
- Fix active tab not updating on prop change [[#2712](https://github.com/woodpecker-ci/woodpecker/pull/2712)]
- Unique status for matrix  [[#2695](https://github.com/woodpecker-ci/woodpecker/pull/2695)]
- Fix secret image filter regex [[#2674](https://github.com/woodpecker-ci/woodpecker/pull/2674)]
- local backend ignore errors in commands in between [[#2636](https://github.com/woodpecker-ci/woodpecker/pull/2636)]
- Do not print log level on CLI [[#2638](https://github.com/woodpecker-ci/woodpecker/pull/2638)]
- Fix error when closing logs [[#2637](https://github.com/woodpecker-ci/woodpecker/pull/2637)]
- Fix `CI_WORKSPACE` in local backend [[#2627](https://github.com/woodpecker-ci/woodpecker/pull/2627)]
- Some mobile UI fixes [[#2624](https://github.com/woodpecker-ci/woodpecker/pull/2624)]
- Fix secret priority [[#2599](https://github.com/woodpecker-ci/woodpecker/pull/2599)]
- UI cleanups and improvements [[#2548](https://github.com/woodpecker-ci/woodpecker/pull/2548)]
- Fix PR event trigger and list for bitbucket repos [[#2539](https://github.com/woodpecker-ci/woodpecker/pull/2539)]
- Fix ccmenu endpoint [[#2543](https://github.com/woodpecker-ci/woodpecker/pull/2543)]
- Trim last "/" from WOODPECKER_HOST config [[#2538](https://github.com/woodpecker-ci/woodpecker/pull/2538)]
- Use correct mime type when no content is sent [[#2515](https://github.com/woodpecker-ci/woodpecker/pull/2515)]
- Fix bitbucket branches pagination. [[#2509](https://github.com/woodpecker-ci/woodpecker/pull/2509)]
- fix: change config.config_data column type to longblob in mysql [[#2434](https://github.com/woodpecker-ci/woodpecker/pull/2434)]
- Fix: change tasks.task_data column type to longblob in mysql [[#2418](https://github.com/woodpecker-ci/woodpecker/pull/2418)]
- Do not list archived repos for all forges [[#2374](https://github.com/woodpecker-ci/woodpecker/pull/2374)]
- fix(server/api/repo): Fix repair webhook host [[#2372](https://github.com/woodpecker-ci/woodpecker/pull/2372)]
- Delete repos/secrets on org deletion [[#2367](https://github.com/woodpecker-ci/woodpecker/pull/2367)]
- Fix org fetching [[#2343](https://github.com/woodpecker-ci/woodpecker/pull/2343)]
- Show correct event in pipeline step list [[#2334](https://github.com/woodpecker-ci/woodpecker/pull/2334)]
- Add min height to mobile pipeline view and fix overflow [[#2335](https://github.com/woodpecker-ci/woodpecker/pull/2335)]
- Fix grid column size in pipeline log view [[#2336](https://github.com/woodpecker-ci/woodpecker/pull/2336)]
- Fix mobile login view [[#2332](https://github.com/woodpecker-ci/woodpecker/pull/2332)]
- Fix button loading spinner when activating repos [[#2333](https://github.com/woodpecker-ci/woodpecker/pull/2333)]
- make WOODPECKER_MIGRATIONS_ALLOW_LONG have an actuall effect [[#2251](https://github.com/woodpecker-ci/woodpecker/pull/2251)]
- Docker build dont ignore ci env vars [[#2238](https://github.com/woodpecker-ci/woodpecker/pull/2238)]
- Handle parsed hooks that should be ignored [[#2243](https://github.com/woodpecker-ci/woodpecker/pull/2243)]
- Set correct version for release branch releases [[#2227](https://github.com/woodpecker-ci/woodpecker/pull/2227)]
- Bump default git clone plugin [[#2215](https://github.com/woodpecker-ci/woodpecker/pull/2215)]
- Show all steps [[#2190](https://github.com/woodpecker-ci/woodpecker/pull/2190)]
- Fix pipeline config collapsing [[#2166](https://github.com/woodpecker-ci/woodpecker/pull/2166)]
- Fix 'add-orgs' migration [[#2117](https://github.com/woodpecker-ci/woodpecker/pull/2117)]
- docs: Environment Variable Seems to be `DOCKER_HOST`, not `DOCKER_SOCK` [[#2122](https://github.com/woodpecker-ci/woodpecker/pull/2122)]
- Fix swagger response code [[#2119](https://github.com/woodpecker-ci/woodpecker/pull/2119)]
- Forge Github Org: Use `login` instead of `name` [[#2104](https://github.com/woodpecker-ci/woodpecker/pull/2104)]
- client.go: Fix RepoPost path [[#2091](https://github.com/woodpecker-ci/woodpecker/pull/2091)]
- Fix alt text contrast in code boxes [[#2089](https://github.com/woodpecker-ci/woodpecker/pull/2089)]
- Fix WOODPECKER_GRPC_VERIFY being ignored [[#2077](https://github.com/woodpecker-ci/woodpecker/pull/2077)]
- Handle case where there is no latest pipeline for GetBadge [[#2042](https://github.com/woodpecker-ci/woodpecker/pull/2042)]

### Misc

- Update release-helper [[#2863](https://github.com/woodpecker-ci/woodpecker/pull/2863)]
- Add repo owner test [[#2857](https://github.com/woodpecker-ci/woodpecker/pull/2857)]
- Update woodpeckerci/plugin-ready-release-go Docker tag to v1.0.2 [[#2853](https://github.com/woodpecker-ci/woodpecker/pull/2853)]
- Update golang (packages) [[#2839](https://github.com/woodpecker-ci/woodpecker/pull/2839)]
- Update dependency vite to v5 [[#2836](https://github.com/woodpecker-ci/woodpecker/pull/2836)]
- Lock file maintenance [[#2840](https://github.com/woodpecker-ci/woodpecker/pull/2840)]
- Update postgres Docker tag to v16.1 [[#2842](https://github.com/woodpecker-ci/woodpecker/pull/2842)]
- Update docker.io/golang Docker tag to v1.21.4 [[#2828](https://github.com/woodpecker-ci/woodpecker/pull/2828)]
- Update docker.io/techknowlogick/xgo Docker tag to go-1.21.4 [[#2829](https://github.com/woodpecker-ci/woodpecker/pull/2829)]
- Update golang (packages) [[#2815](https://github.com/woodpecker-ci/woodpecker/pull/2815)]
- Update dependency marked to v10 [[#2810](https://github.com/woodpecker-ci/woodpecker/pull/2810)]
- Update release-helper [[#2801](https://github.com/woodpecker-ci/woodpecker/pull/2801)]
- Remove go versions from .golangci.yml [[#2775](https://github.com/woodpecker-ci/woodpecker/pull/2775)]
- [pre-commit.ci] pre-commit autoupdate [[#2767](https://github.com/woodpecker-ci/woodpecker/pull/2767)]
- Lock file maintenance [[#2755](https://github.com/woodpecker-ci/woodpecker/pull/2755)]
- Update golang (packages) [[#2742](https://github.com/woodpecker-ci/woodpecker/pull/2742)]
- Update woodpeckerci/plugin-ready-release-go Docker tag to v0.7.0 [[#2728](https://github.com/woodpecker-ci/woodpecker/pull/2728)]
- Add grafana dashobard to awesome [[#2710](https://github.com/woodpecker-ci/woodpecker/pull/2710)]
- Pin alpine versions in Dockerfile [[#2649](https://github.com/woodpecker-ci/woodpecker/pull/2649)]
- Use full qualifyer for images [[#2692](https://github.com/woodpecker-ci/woodpecker/pull/2692)]
- chore(deps): lock file maintenance [[#2673](https://github.com/woodpecker-ci/woodpecker/pull/2673)]
- fix(deps): update golang (packages) [[#2671](https://github.com/woodpecker-ci/woodpecker/pull/2671)]
- Use `pre-commit`  [[#2650](https://github.com/woodpecker-ci/woodpecker/pull/2650)]
- fix(deps): update dependency fuse.js to v7 [[#2666](https://github.com/woodpecker-ci/woodpecker/pull/2666)]
- chore(deps): update dependency @types/node to v20 [[#2664](https://github.com/woodpecker-ci/woodpecker/pull/2664)]
- chore(deps): update woodpeckerci/plugin-docker-buildx docker tag to v2.2.0 [[#2663](https://github.com/woodpecker-ci/woodpecker/pull/2663)]
- chore(deps): update mysql docker tag to v8.2.0 [[#2662](https://github.com/woodpecker-ci/woodpecker/pull/2662)]
- Add some tests [[#2652](https://github.com/woodpecker-ci/woodpecker/pull/2652)]
- chore(deps): update docs npm deps non-major [[#2660](https://github.com/woodpecker-ci/woodpecker/pull/2660)]
- chore(deps): update web npm deps non-major [[#2661](https://github.com/woodpecker-ci/woodpecker/pull/2661)]
- Fix codecov plugin version [[#2643](https://github.com/woodpecker-ci/woodpecker/pull/2643)]
- Add prettier [[#2600](https://github.com/woodpecker-ci/woodpecker/pull/2600)]
- Do not run docker prepare steps [[#2626](https://github.com/woodpecker-ci/woodpecker/pull/2626)]
- Fix docker workflow and only run if needed [[#2625](https://github.com/woodpecker-ci/woodpecker/pull/2625)]
- fix(deps): update golang (packages) [[#2614](https://github.com/woodpecker-ci/woodpecker/pull/2614)]
- chore(deps): lock file maintenance [[#2620](https://github.com/woodpecker-ci/woodpecker/pull/2620)]
- chore(deps): update codeberg.org/woodpecker-plugins/trivy docker tag to v1.0.1 [[#2618](https://github.com/woodpecker-ci/woodpecker/pull/2618)]
- chore(deps): update node.js to v21 [[#2615](https://github.com/woodpecker-ci/woodpecker/pull/2615)]
- Only publish PR images when label is set [[#2608](https://github.com/woodpecker-ci/woodpecker/pull/2608)]
- chore(deps): lock file maintenance [[#2595](https://github.com/woodpecker-ci/woodpecker/pull/2595)]
- chore(deps): update postgres docker tag to v16 [[#2588](https://github.com/woodpecker-ci/woodpecker/pull/2588)]
- Update renovate schedule & use central config repo [[#2597](https://github.com/woodpecker-ci/woodpecker/pull/2597)]
- chore(deps): update woodpeckerci/plugin-surge-preview docker tag to v1.2.2 [[#2593](https://github.com/woodpecker-ci/woodpecker/pull/2593)]
- Update README badge link [[#2596](https://github.com/woodpecker-ci/woodpecker/pull/2596)]
- fix(deps): update golang (packages) to v23.0.7+incompatible [[#2586](https://github.com/woodpecker-ci/woodpecker/pull/2586)]
- Fix missing web dist [[#2580](https://github.com/woodpecker-ci/woodpecker/pull/2580)]
- Run tests on `main` branch [[#2576](https://github.com/woodpecker-ci/woodpecker/pull/2576)]
- fix(deps): update module github.com/google/go-github/v55 to v56 [[#2573](https://github.com/woodpecker-ci/woodpecker/pull/2573)]
- Add plugin "NixOS Remote Builder" to plugin index [[#2571](https://github.com/woodpecker-ci/woodpecker/pull/2571)]
- Fix renovate [[#2569](https://github.com/woodpecker-ci/woodpecker/pull/2569)]
- renovate: add `golang` group [[#2567](https://github.com/woodpecker-ci/woodpecker/pull/2567)]
- chore(deps): update golang docker tag to v1.21.3 [[#2564](https://github.com/woodpecker-ci/woodpecker/pull/2564)]
- chore(deps): update techknowlogick/xgo docker tag to go-1.21.3 [[#2565](https://github.com/woodpecker-ci/woodpecker/pull/2565)]
- fix(deps): update golang deps non-major [[#2566](https://github.com/woodpecker-ci/woodpecker/pull/2566)]
- chore(deps): update mstruebing/editorconfig-checker docker tag to v2.7.2 [[#2563](https://github.com/woodpecker-ci/woodpecker/pull/2563)]
- Bump to mysql 8 [[#2559](https://github.com/woodpecker-ci/woodpecker/pull/2559)]
- fix(deps): update module github.com/xanzy/go-gitlab to v0.93.1 [[#2560](https://github.com/woodpecker-ci/woodpecker/pull/2560)]
- Require Go 1.21 [[#2553](https://github.com/woodpecker-ci/woodpecker/pull/2553)]
- chore(deps): update techknowlogick/xgo docker tag to go-1.21.2 [[#2523](https://github.com/woodpecker-ci/woodpecker/pull/2523)]
- Update issue config [[#2353](https://github.com/woodpecker-ci/woodpecker/pull/2353)]
- Add test for handling pipeline error [[#2547](https://github.com/woodpecker-ci/woodpecker/pull/2547)]
- chore(deps): update golang docker tag to v1.21.2 [[#2532](https://github.com/woodpecker-ci/woodpecker/pull/2532)]
- fix(deps): update golang.org/x/exp digest to 7918f67 [[#2535](https://github.com/woodpecker-ci/woodpecker/pull/2535)]
- fix(deps): update golang deps non-major [[#2533](https://github.com/woodpecker-ci/woodpecker/pull/2533)]
- fix(deps): update golang.org/x/exp digest to 3e424a5 [[#2530](https://github.com/woodpecker-ci/woodpecker/pull/2530)]
- Use golangci-lint to lint zerolog [[#2524](https://github.com/woodpecker-ci/woodpecker/pull/2524)]
- Renovate config updates [[#2519](https://github.com/woodpecker-ci/woodpecker/pull/2519)]
- fix(deps): update module github.com/docker/distribution to v2.8.3+incompatible [[#2517](https://github.com/woodpecker-ci/woodpecker/pull/2517)]
- fix(deps): update module github.com/melbahja/goph to v1.4.0 [[#2513](https://github.com/woodpecker-ci/woodpecker/pull/2513)]
- fix(deps): update golang deps non-major [[#2500](https://github.com/woodpecker-ci/woodpecker/pull/2500)]
- chore(deps): lock file maintenance [[#2497](https://github.com/woodpecker-ci/woodpecker/pull/2497)]
- Fix broken link to 3rd party plugin library [[#2494](https://github.com/woodpecker-ci/woodpecker/pull/2494)]
- fix(deps): update golang deps non-major [[#2486](https://github.com/woodpecker-ci/woodpecker/pull/2486)]
- chore(deps): lock file maintenance [[#2469](https://github.com/woodpecker-ci/woodpecker/pull/2469)]
- Add devx lable to compose file PRs [[#2467](https://github.com/woodpecker-ci/woodpecker/pull/2467)]
- chore(deps): update postgres docker tag to v16 [[#2463](https://github.com/woodpecker-ci/woodpecker/pull/2463)]
- Update gitea sdk [[#2464](https://github.com/woodpecker-ci/woodpecker/pull/2464)]
- fix(deps): update golang deps non-major [[#2462](https://github.com/woodpecker-ci/woodpecker/pull/2462)]
- fix(deps): update dependency ansi_up to v6 [[#2431](https://github.com/woodpecker-ci/woodpecker/pull/2431)]
- chore(deps): update web npm deps non-major [[#2461](https://github.com/woodpecker-ci/woodpecker/pull/2461)]
- fix(deps): update module github.com/tevino/abool to v2 [[#2460](https://github.com/woodpecker-ci/woodpecker/pull/2460)]
- fix(deps): update module github.com/google/go-github/v39 to v55 [[#2456](https://github.com/woodpecker-ci/woodpecker/pull/2456)]
- fix(deps): update module github.com/golang-jwt/jwt/v4 to v5 [[#2449](https://github.com/woodpecker-ci/woodpecker/pull/2449)]
- fix(deps): update module github.com/golang-jwt/jwt/v4 to v5 [[#2447](https://github.com/woodpecker-ci/woodpecker/pull/2447)]
- chore(deps): update node.js to v20 [[#2422](https://github.com/woodpecker-ci/woodpecker/pull/2422)]
- Add renovate package rule to apply build label [[#2440](https://github.com/woodpecker-ci/woodpecker/pull/2440)]
- fix(deps): update dependency prism-react-renderer to v2 [[#2436](https://github.com/woodpecker-ci/woodpecker/pull/2436)]
- fix(deps): update dependency node-emoji to v2 [[#2435](https://github.com/woodpecker-ci/woodpecker/pull/2435)]
- Add renovate package rule to apply dependencies label [[#2438](https://github.com/woodpecker-ci/woodpecker/pull/2438)]
- fix(deps): update golang deps non-major [[#2437](https://github.com/woodpecker-ci/woodpecker/pull/2437)]
- chore(deps): update postgres docker tag to v15 [[#2423](https://github.com/woodpecker-ci/woodpecker/pull/2423)]
- fix(deps): update dependency esbuild-loader to v4 [[#2433](https://github.com/woodpecker-ci/woodpecker/pull/2433)]
- fix(deps): update dependency clsx to v2 [[#2432](https://github.com/woodpecker-ci/woodpecker/pull/2432)]
- fix(deps): update dependency @vueuse/core to v10 [[#2430](https://github.com/woodpecker-ci/woodpecker/pull/2430)]
- fix(deps): update dependency @svgr/webpack to v8 [[#2429](https://github.com/woodpecker-ci/woodpecker/pull/2429)]
- fix(deps): update dependency @kyvg/vue3-notification to v3 [[#2427](https://github.com/woodpecker-ci/woodpecker/pull/2427)]
- fix(deps): update dependency @intlify/unplugin-vue-i18n to v1 [[#2426](https://github.com/woodpecker-ci/woodpecker/pull/2426)]
- chore(deps): update typescript-eslint monorepo to v6 (major) [[#2425](https://github.com/woodpecker-ci/woodpecker/pull/2425)]
- chore(deps): update react monorepo to v18 (major) [[#2424](https://github.com/woodpecker-ci/woodpecker/pull/2424)]
- chore(deps): update dependency prettier to v3 [[#2420](https://github.com/woodpecker-ci/woodpecker/pull/2420)]
- chore(deps): update dependency eslint-config-prettier to v9 [[#2415](https://github.com/woodpecker-ci/woodpecker/pull/2415)]
- chore(deps): update dependency @tsconfig/docusaurus to v2 [[#2410](https://github.com/woodpecker-ci/woodpecker/pull/2410)]
- chore(deps): update dependency typescript to v5 [[#2421](https://github.com/woodpecker-ci/woodpecker/pull/2421)]
- chore(deps): update dependency concurrently to v8 [[#2414](https://github.com/woodpecker-ci/woodpecker/pull/2414)]
- Add renovate deps groups [[#2417](https://github.com/woodpecker-ci/woodpecker/pull/2417)]
- fix(deps): update module xorm.io/xorm to v1.3.3 [[#2393](https://github.com/woodpecker-ci/woodpecker/pull/2393)]
- chore(deps): update dependency marked to v9 [[#2419](https://github.com/woodpecker-ci/woodpecker/pull/2419)]
- chore(deps): update dependency @types/marked to v5 [[#2411](https://github.com/woodpecker-ci/woodpecker/pull/2411)]
- fix(deps): update module github.com/rs/zerolog to v1.30.0 [[#2404](https://github.com/woodpecker-ci/woodpecker/pull/2404)]
- fix(deps): update module github.com/jellydator/ttlcache/v3 to v3.1.0 [[#2402](https://github.com/woodpecker-ci/woodpecker/pull/2402)]
- fix(deps): update module github.com/xanzy/go-gitlab to v0.91.1 [[#2405](https://github.com/woodpecker-ci/woodpecker/pull/2405)]
- fix(deps): update module github.com/antonmedv/expr to v1.15.1 [[#2400](https://github.com/woodpecker-ci/woodpecker/pull/2400)]
- chore(deps): update dependency axios to v1 [[#2413](https://github.com/woodpecker-ci/woodpecker/pull/2413)]
- fix(deps): update module github.com/prometheus/client_golang to v1.16.0 [[#2403](https://github.com/woodpecker-ci/woodpecker/pull/2403)]
- fix(deps): update module github.com/urfave/cli/v2 to v2.25.7 [[#2391](https://github.com/woodpecker-ci/woodpecker/pull/2391)]
- fix(deps): update module google.golang.org/protobuf to v1.31.0 [[#2409](https://github.com/woodpecker-ci/woodpecker/pull/2409)]
- fix(deps): update kubernetes packages to v0.28.1 [[#2399](https://github.com/woodpecker-ci/woodpecker/pull/2399)]
- fix(deps): update module github.com/swaggo/swag to v1.16.2 [[#2390](https://github.com/woodpecker-ci/woodpecker/pull/2390)]
- fix(deps): update dependency @easyops-cn/docusaurus-search-local to ^0.36.0 [[#2406](https://github.com/woodpecker-ci/woodpecker/pull/2406)]
- fix(deps): update module github.com/stretchr/testify to v1.8.4 [[#2389](https://github.com/woodpecker-ci/woodpecker/pull/2389)]
- fix(deps): update module github.com/caddyserver/certmagic to v0.19.2 [[#2401](https://github.com/woodpecker-ci/woodpecker/pull/2401)]
- chore(deps): update postgres docker tag to v12.16 [[#2397](https://github.com/woodpecker-ci/woodpecker/pull/2397)]
- fix(deps): update module github.com/mattn/go-sqlite3 to v1.14.17 [[#2387](https://github.com/woodpecker-ci/woodpecker/pull/2387)]
- fix(deps): update module github.com/google/uuid to v1.3.1 [[#2386](https://github.com/woodpecker-ci/woodpecker/pull/2386)]
- chore(deps): update dependency unplugin-vue-components to ^0.25.0 [[#2395](https://github.com/woodpecker-ci/woodpecker/pull/2395)]
- fix(deps): update dependency @intlify/unplugin-vue-i18n to ^0.13.0 [[#2398](https://github.com/woodpecker-ci/woodpecker/pull/2398)]
- chore(deps): update dependency unplugin-icons to ^0.17.0 [[#2394](https://github.com/woodpecker-ci/woodpecker/pull/2394)]
- chore(deps): update golang docker tag [[#2396](https://github.com/woodpecker-ci/woodpecker/pull/2396)]
- fix(deps): update module github.com/moby/moby to v20.10.25+incompatible [[#2388](https://github.com/woodpecker-ci/woodpecker/pull/2388)]
- fix(deps): update module github.com/docker/docker to v20.10.25+incompatible [[#2385](https://github.com/woodpecker-ci/woodpecker/pull/2385)]
- fix(deps): update module github.com/docker/cli to v20.10.25+incompatible [[#2384](https://github.com/woodpecker-ci/woodpecker/pull/2384)]
- fix(deps): update module github.com/alessio/shellescape to v1.4.2 [[#2381](https://github.com/woodpecker-ci/woodpecker/pull/2381)]
- fix(deps): update golang.org/x/exp digest to 9212866 [[#2380](https://github.com/woodpecker-ci/woodpecker/pull/2380)]
- Check for correct license header [[#2137](https://github.com/woodpecker-ci/woodpecker/pull/2137)]
- Add TestCompilerCompile [[#2183](https://github.com/woodpecker-ci/woodpecker/pull/2183)]
- Fix `docs` workflow [[#2128](https://github.com/woodpecker-ci/woodpecker/pull/2128)]
- Add some tests for bitbucket forge  [[#2097](https://github.com/woodpecker-ci/woodpecker/pull/2097)]
- Publish releases and branch tags to quay.io too [[#2069](https://github.com/woodpecker-ci/woodpecker/pull/2069)]

## [1.0.5](https://github.com/woodpecker-ci/woodpecker/releases/tag/v1.0.5) - 2023-11-09

- ENHANCEMENTS
  - Switch to go vanity urls (#2706) (#2773)
- MISC
  - Fix release pipeline for 1.x.x (#2774)

## [1.0.4](https://github.com/woodpecker-ci/woodpecker/releases/tag/v1.0.4) - 2023-11-05

- BUGFIXES
  - Fix secret image filter regex (#2674) (#2686)
  - Fix error when closing logs (#2637) (#2640)

## [1.0.3](https://github.com/woodpecker-ci/woodpecker/releases/tag/v1.0.3) - 2023-10-14

- SECURITY
  - Update dependencies (#2587)
  - Frontend: bump postcss to 8.4.31 (#2541)
  - Check permissions on repo lookup (#2358)
  - Change token logging to trace level (#2247) (#2248)
- BUGFIXES
  - Fix gitlab hooks (#2537) (#2542)
  - Trim last "/" from WOODPECKER_HOST config (#2538) (#2540)
  - Fix(server/api/repo): Fix repair webhook host (#2372) (#2452)
  - Show correct event in pipeline step list (#2448)
  - Make WOODPECKER_MIGRATIONS_ALLOW_LONG have an actuall effect (#2251) (#2309)
  - Docker build dont ignore ci env vars (#2238) (#2246)
  - Handle parsed hooks that should be ignored (#2243) (#2244)
  - Return 204 not 500 on filtered pipeline (#2230)
  - Set correct version for release branch releases (#2227) (#2229)
- MISC
  - Rebuild swagger with latest version (#2455)

## [1.0.2](https://github.com/woodpecker-ci/woodpecker/releases/tag/v1.0.2) - 2023-08-16

- SECURITY
  - Validate webhook before change any data (#2221) (#2222)
- BUGFIXES
  - Bump default git clone plugin (#2215) (#2220)
  - Show all steps (#2190) (#2191)

## [1.0.1](https://github.com/woodpecker-ci/woodpecker/releases/tag/v1.0.1) - 2023-08-08

- SECURITY
  - Fix WOODPECKER_GRPC_VERIFY being ignored (#2077) (#2082)
- BUGFIXES
  - Fix 'add-orgs' migration (#2117) (#2145)
  - Fix UI and backend paths with subpath (#1799) (#2133)
  - Fix swagger response code (#2119) (#2121)
  - Forge Github Org: Use `login` instead of `name` (#2104) (#2106)
  - Client.go: Backport fix RepoPost path (#2100)
  - Fix translation key (#2098)

## [1.0.0](https://github.com/woodpecker-ci/woodpecker/releases/tag/v1.0.0) - 2023-07-29

- BREAKING
  - Use IDs to access organizations (#1873)
  - Drop support for Bitbucket Server (#1994)
  - Rename yaml `pipeline` to `steps` (#1833)
  - Drop ".drone.yml" as default pipeline config (#1795)
  - Build-in Env Vars, use _URL for all links/URLs (#1794)
  - Resolve built-in variables for global when filtered too (#1790)
  - Drop "Gogs" support (#1752)
  - Access repos by their IDs (#1691)
  - Drop "coding" support (#1644)
  - Add queue details UI for admins (#1632)
  - Remove `command:` from steps (#1032)
  - Remove old `build` API routes (#1283)
  - Let single line command be a single command (#1009)
  - Drop deprecated environment vars (#920)
  - Drop Var-Args in steps in favor of settings (#919)
  - Fix branch condition on tags (#917)
  - Use asymmetric key to sign webhooks (#916)
  - Add agent tagging / filtering for pipelines (#902)
  - Delete old fallback for "drone.sqlite" (#791)
  - Migrate to certmagic (#360)
- FEATURES
  - Implement YAML Map Merge, Overrides, and Sequence Merge Support (#1720)
  - Add users UI for admins (#1634)
  - Add agent no-schedule flag (#1567)
  - Change locale in user settings (#1305)
  - Add when evaluate filter (#1213)
  - Store an agents list and add agent heartbeats (#1189)
  - Add ability to trigger manual builds (#1156)
  - Add default event filter (#1140)
  - Add CLI support for global and organization secrets (#1113)
  - Allow multiple when conditions (#1087)
  - Add syntax highlighting for pipeline config (#1082)
  - Add `logs` command to CLI & update forges features docs (#1064)
  - Add method to check organization membership (#1037)
  - Global and organization secrets (#1027)
  - Add pipeline log output download (#1023)
  - Provide global environment variables for pipeline substitution (#968)
  - Add cron jobs (#934)
  - Support localized web UI (#912)
  - Add support to define a custom docker network and enable docker ipv6 (#893)
  - Add SSH backend (#861)
  - Add support for superseding runs (#831)
  - Add support for steps to be a list (instead of dict) (#826)
  - Add editing of secrets and registries (#823)
  - Allow loading sensitive flags from files (#815)
  - Add support for pipeline configuration service (#804)
  - Support all backends for CLI exec (#801)
  - Add support for pipeline root.when conditions (#770)
  - Add support to run pipelines using a local backend (#709)
  - Add initial version of Kubernetes backend (#552)
- SECURITY
  - Fix ignoring server set pipeline max-timeout (#1875)
  - Only grant privileged to plugins (#1646)
  - Only inject netrc to trusted clone plugins (#1352)
  - Support plugin-only secrets (#1344)
  - Fix insecure /tmp usage in local backend (#872)
- BUGFIXES
  - Handle case where there is no latest pipeline for GetBadge (#2042) (#2050)
  - Fix repo gate protection (#1969)
  - Make secrets with "/" in name editable / deletable (#1938)
  - Fix Bitbucket implement missing features (#1887) (#1889)
  - Fix nil pointer in repo repair (#1804)
  - Do not use OAuth client without token (#1803)
  - Correct label argument parsing in agent code (#1717)
  - Fully support `.yaml` (#1713)
  - Consistent status on delete (#1703)
  - Fix Bitbucket Server branches (#1698)
  - Set 'HOME' during local pipeline step (#1686)
  - Pipeline compiler: handle nil entrys in settings list (#1626)
  - Fix: backend auto-detection should be consistent (#1618)
  - Return 404 on badge endpoint for inactive repos (#1600)
  - Ensure the SharedInformerFactory closes eventually (#1585)
  - Deduplicate step docker container volumes (#1571)
  - Don't require secret value on secret edit (#1552) (#1553)
  - Rework status constraint logic for successes (#1515)
  - Don't panic on hook parsing (#1501)
  - Hide not owned repos from sidebar and repo list (#1453)
  - Fix cut of woodpecker animation (#1402)
  - Fix approval on mobile (#1320)
  - Unify buttons, links and improve focus styles (#1317)
  - Fix pipeline manual trigger on web (#1307)
  - Fix SCM visibility if user visibility is private (#1217)
  - Hide log output container if step does not have logs (#1086)
  - Fix to show build pipeline parse error (#1066)
  - Pipeline compiler should not alter specified image (#1005)
  - Gracefully handle non-zero exit code in local backend (#1002)
  - Replace run_on references with runs_on (#965)
  - Set default logging value of CLI to info (#871)
  - Support conditional branch as an array to align with documentation (#836)
  - Fix redirect after login (#824)
- ENHANCEMENTS
  - Add BranchHead implementation for bitbucket forge (#2011)
  - Use global logger for xorm logs and add options (#1997)
  - Let HookParse func explicit ignore events (#1942)
  - Link swagger in navbar (#1984)
  - Add option to read grpc-secret from file (#1972)
  - Let pipeline-compiler export step types (#1958)
  - docker backend use uuid instead of name as identifier (#1967)
  - Kubernetes do not set Pod's Image pull policy if not explicitly set (#1914)
  - Fixed when:evaluate on non-standard (non-CI*) env vars (#1907)
  - Add pull-request implementation for bitbucket forge (#1889)
  - Store agent ID in config file (#1888)
  - Fix bitbucket forge add repo (#1887)
  - Added Woodpecker Host Config used for Webhooks (#1869)
  - Drop old columns (#1838)
  - Remove MSSQL specific code and cleanups (#1796)
  - Remove unused file system API (#1791)
  - Add Forge Metadata to built-in environment variables (#1789)
  - Redirect to new pipeline (#1761)
  - Add reset token button (#1755)
  - Add agent functions to go-sdk (#1754)
  - Always send a status back to forge (#1751)
  - Allow to configure listener port for SSL (#1735)
  - Identify users using their remote ID (#1732)
  - Let agent retry to connecting to server (#1728)
  - Stable sort order for DB lists (#1702)
  - Add backend label to agents (#1692)
  - Web: use i18n-t to avoid v-html directive (#1676)
  - Various UI improvements (#1663)
  - Do not store inactive repos without any resources (#1658)
  - Implement visual display of queue statistics (#1657)
  - Agent check gRPC version against server (#1653)
  - Initiate Pagination Implementation for API and Infinite Scroll in UI (#1651)
  - Add PR pipeline list (#1641)
  - Save agent-id for tasks and add endpoint to get agent tasks (#1631)
  - Return 404 if pipeline not exist and handle 404 errors in WebUI (#1627)
  - UI should confirm secret deletion (#1604)
  - Add collapsable support to panel elements (#1601)
  - Add cancel button on secrets tab (#1599)
  - Allow custom dnsConfig in agent deployment (#1569)
  - Show platform, backend and capacity as badges in agent list (#1568)
  - Define WOODPECKER_FORGE_TIMEOUT server config (#1558)
  - Sort repos by org/name (#1548)
  - Improve button and input contrast in dark mode (#1456)
  - Consistent and more descriptive naming of parameters in index.ts (#1455)
  - Add button in UI to trigger the deployment event (#1415)
  - Use icons for step and workflow states (#1409)
  - Match notification font size to rest of the UI (#1399)
  - Support .yaml as file-ending for workflow config too (#1388)
  - Show workflow state in UI and collapse completed workflows (#1383)
  - Use pipeline wrapper and improve scaffold UI (#1368)
  - Lazy load locales (#1362)
  - Always use rounded quadrat user avatars (#1350)
  - Fix display of long pipeline and job names (#1346)
  - Support changed files for Gitea PRs (#1342)
  - Allow to change directory for steps (#1329)
  - UI use system font stack (#1326)
  - Add pull request labels as environment variable (#1321)
  - Make pipeline workflows collapsable (#1304)
  - Make submit buttons green and add forms (#1302)
  - Add pipeline build number into Pipeline list (#1301)
  - Add title to docs links (#1298)
  - Check if repo exists before creating pipeline (#1297)
  - Use HTML buttons to allow keyboard navigation (#1242)
  - Introduce and use Pagination helper func (#1236)
  - Sort secret lists and events (#1223)
  - Add support sub-settings and secrets in sub-settings (#1221)
  - Add option to ignore failures on steps (#1219)
  - Set a default value for `pipeline-event` flag of `cli exec` command (#1212)
  - Add option for docker runtime to provide default volumes (#1203)
  - Make healthcheck port configurable (#1197)
  - Don't show "changed files" if event can't have them (#1191)
  - Add dedicated DroneCI env compatibility layer (#1185)
  - Only enable debug endpoints if log level is debug or below (#1160)
  - Sort pipelines based on creation date (#1159)
  - Add option to turn on and off log automatic scrolling (#1149)
  - Checkout tags on tag pipeline (#1110)
  - Use fixed version of git clone plugin (#1108)
  - Fetch repositories with remote ID if possible (#1078)
  - Support Docker credential helpers (#1075)
  - Do not show pipeline name if it's a single file (#1069)
  - Remove xterm and use ansi converter for logs (#1067)
  - Update jsonschema and define "services" (#1036)
  - Show forge icons in UI (#987)
  - Make pipeline runtime log with description (#970)
  - Improve UI colors to have more contrast (#943)
  - Add branches support for BitBucket (#907)
  - Auto cancel blocked pipelines (#905)
  - Allow to change forge status messages (#900)
  - Added support for step errors when executing backend (#817)
  - Do not filter on linux/amd64 per default (#805)
- DOCUMENTATION
  - Remove never implemented "tag"-filter and document "ref"-filter to do the same (#1820)
  - Define Glossary (#1800)
  - Add more documentation about branch matching (#1186)
  - Use versioned docs (#1145)
  - Add gitpod setup (#1020)
- MISC
  - Drop tarball release (#1819)
  - Move helm charts to own repo "helm" (#1589)
  - Replace yarn with pnpm (#1240)
  - Publish preview docker images of pulls (#1072)

## [0.15.11](https://github.com/woodpecker-ci/woodpecker/releases/tag/v0.15.11) - 2023-07-12

- SECURITY
  - Update github.com/gin-gonic/gin to 1.9.1 (#1989)
- ENHANCEMENTS
  - Allow gitea dev version (#914) (#1988)

## [0.15.10](https://github.com/woodpecker-ci/woodpecker/releases/tag/v0.15.10) - 2023-07-09

- SECURITY
  - Fix agent auth (#1952) (#1953)
  - Return after error (#1875) (#1876)
  - Update github.com/docker/distribution (#1750)

## [0.15.9](https://github.com/woodpecker-ci/woodpecker/releases/tag/v0.15.9) - 2023-05-11

- SECURITY
  - Backport securitycheck and bump deps where needed (#1745)

## [0.15.8](https://github.com/woodpecker-ci/woodpecker/releases/tag/0.15.8) - 2023-04-29

- BUGFIXES
  - Use codeberg.org/6543/go-yaml2json (#1719)
  - Fix faulty hardlink in release tarball (#1669) (#1671)
  - Persist `DepStatus` of tasks (#1610) (#1625)

## [0.15.7](https://github.com/woodpecker-ci/woodpecker/releases/tag/v0.15.7) - 2023-03-14

- SECURITY
  - Update dependencies golang/x libs (#1612) (#1621)
- BUGFIXES
  - Docker backend should not close 'engine.Tail' result (#1616) (#1620)
  - Force pure Go resolver onto server (#1502) (#1503)
- ENHANCEMENTS
  - SanitizeParamKey "-" to "_" for plugin settings (#1511)
- MISC
  - Bump xgo and go to v1.19.5 (#1538) (#1547)
  - Pin official default clone image (#1526) (#1534)

## [0.15.6](https://github.com/woodpecker-ci/woodpecker/releases/tag/v0.15.6) - 2022-12-23

- SECURITY
  - Update golang.org/x/net (#1494)
  - [**BREAKING**] Disable metrics access if no token is set (#1469) (#1470)
  - Update dep moby (#1263) (#1264)
- BUGFIXES
  - Update json schema for cli lint to cover valid cases (#1384)
  - Add pipeline.step.when.branch string-array type to schema.json (#1380)
  - Display system CA error only if there is an error (#870) (#1286)
- ENHANCEMENTS
  - Bump Frontend Deps and remove unused (#1404)

## [0.15.5](https://github.com/woodpecker-ci/woodpecker/releases/tag/v0.15.5) - 2022-10-13

- BUGFIXES
  - Change build message column type to text (#1252) (#1253)
- ENHANCEMENTS
  - Bump DefaultCloneImage version to v1.6.0 (#1254)
  - On Repo update, keep old "Clone" if update would empty it (#1170) (#1195)

## [0.15.4](https://github.com/woodpecker-ci/woodpecker/releases/tag/v0.15.4) - 2022-09-06

- BUGFIXES
  - Extract commit message from branch creation (#1150) (#1153)
  - Respect WOODPECKER_GITEA_SKIP_VERIFY (#1152) (#1151)
  - update golang.org/x/crypto (#1124)
  - Implement Refresher for GitLab (#1031) (#1120)
  - Make returned proc list to be returned always in correct order (#1060) (#1065)
  - Update type of 'log_data' from blob to longblob (#1050) (#1052)
  - Make ListItem component more accessible by using a button tag when clickable (#1044) (#1046)
- MISC
  - Update base images (#1024) (#1025)

## [0.15.3](https://github.com/woodpecker-ci/woodpecker/releases/tag/v0.15.3) - 2022-06-16

- SECURITY
  - Update github.com/containerd/containerd (#978) (#980)
- BUGFIXES
  - Return to page after clicking login at navbar (#975) (#976)

## [0.15.2](https://github.com/woodpecker-ci/woodpecker/releases/tag/v0.15.2) - 2022-06-14

- BUGFIXES
  - Fix uppercase from_secrets (#842) (#925)
  - Fix key/val format for dind env vars (#889) (#890)
  - Update helm chart releasing (#882) (#888)
- DOCUMENTATION
  - Fix run_on references with runs_on in docs (#965)

## [0.15.1](https://github.com/woodpecker-ci/woodpecker/releases/tag/v0.15.1) - 2022-04-13

- SECURITY
  - Escape html / xml in log view (#879) (#880)
- FEATURES
  - Build multiarch images for server (#821) (#822)
- BUGFIXES
  - Branch list enhancements (#808) (#809)
  - Get Netrc machine from clone url (#800) (#803)

## [v0.15.0](https://github.com/woodpecker-ci/woodpecker/releases/tag/v0.15.0) - 2022-02-24

- BREAKING
  - Change paths to use woodpecker instead of drone (#494)
  - Move plugin config to root.pipeline.[step].settings (#464)
  - Replace debug with log-level flag (#440)
  - Change prometheus metrics from `drone_*` to `woodpecker_*` (#439)
  - Replace DRONE_with CI_ variables in pipeline steps (#427)
  - Enable pull_request hook by default on repository activation (#420)
  - Remote Gitea drop basic auth support (#365)
  - Change pipeline config path resolution (#299)
  - Remove push, tag and deployment webhook filters (#281)
  - Clean up config environment variables for server and agent (#218)
- SECURITY
  - Add linter bidichk to prevent malicious utf8 chars (#516)
- FEATURES
  - Show changed files of pipeline in UI (#650)
  - Show yml config of pipeline in UI (#649)
  - Multiarch build for cli and agent docker images (#634), (#622)
  - Get secrets in settings (#604)
  - Add multi-pipeline support to exec & lint (#568)
  - Add repo branches endpoint (#481)
  - Add repo permission endpoint (#436)
  - Add web-config endpoint (#433)
  - Replace www-path with www-proxy option for development (#248)
- BUGFIXES
  - Make gRPC error "too many keepalive pings" only show up in trace logs (#787)
  - WOODPECKER_ENVIRONMENT: ignore items only containing a key and no value (#781)
  - Fix pipeline timestamps (#730)
  - Remove "panic()" as much as possible from code (#682)
  - Send decline events back to UI (#680)
  - Notice all changed files of all related commits for gitea push webhooks (#675)
  - Use global branch filter only on events containing branch info (#659)
  - API GetRepos() return empty list if no active repos exist (#658)
  - Skip nested GitLab repositories during sync (#656), (#652)
  - Build proc tree function should not depend on sorted procs list (#647)
  - Fix sqlite migration on column drop of abnormal schemas (#629)
  - Fix gRPC incompatibility in helm chart (#627)
  - Fix new pipeline not published to UI if protected repo mode enabled (#619)
  - Dont panic, report error back (#582)
  - Improve status updates (#561)
  - Let normal repo admins change timeout to lower values (#543)
  - Fix registry delete (#532)
  - Fix overflowing commit messages (#528)
  - Fix passing of netrc credentials to clone step (#492)
  - Fix various typos (#416)
  - Append trailing slash to default GH API URL (#411)
  - Fix filter pipeline config files (#279)
- ENHANCEMENTS
  - Return better error if repo was deleted/renamed (#780)
  - Add support to set default clone image via environment variable (#769)
  - Add flag to always authenticate when cloning public repositories from locked down / private only forges (#760)
  - UI: show date time on hover over time items (#756)
  - Add repo-link to badge markdown in UI (#753)
  - Allow specifying dind container in values (#750)
  - Add page to view all projects of a user / group (#741)
  - Let non required migration tasks fail and continue (#729)
  - Improve pipeline compiler (#699)
  - Support ChangedFiles for GitHub & Gitlab PRs and pushes and Gitea pushes (#697)
  - Remove unused flags / options (#693)
  - Automatically determine platform of agent (#690)
  - Build ref link point to commit not compare if only one commit was pushed (#673)
  - Hide multi line secrets from log (#671)
  - Do not exclude repo owner from gated rule (#641)
  - Add field for image list in Secrets Repo Settings (Web UI) (#638)
  - Use Woodpecker theme colors on Safari Tab Bar / Header Bar (#632)
  - Add "woodpeckerci/plugin-docker-buildx" to privileged plugins (#623)
  - Use gitlab generic webhooks instead of drone-ci-service (#620)
  - Calculate build number on creation (#615)
  - Hide gin routes logging on non-debug starts (#603)
  - Let remove be a remove (#593)
  - Add flag to set oauth redirect host in dev mode (#586)
  - Add log-level option to cli (#584)
  - Improve favicons (#576)
  - Show icon and index of a pull request in pipelines triggered by pull requests (#575)
  - Improve secrets tab (#574)
  - Use monospace font for build logs (#527)
  - Show environ in every BuildProc (#526)
  - Drop error only on purpose or else report back or log (#514)
  - Migrate database backend to Xorm (#474)
  - Add backend selection for agent (#463)
  - Switch default git plugin (#449)
  - Add log level API (#444)
  - Move entirely to zerolog (#426)
  - Pass context.Context down (#371)
  - Extend Logging & Report to WebHook Caller back if pulls are disabled (#369)
  - If config is no file assume its a folder (#354)
  - Rename cmd agent and server folders and binaries (#330)
  - Release Helm charts (#302)
  - Add flag for specific grpc server addr (#295)
  - Add option to charts, to pass in topology pod constraints (#262)
  - Use server-host as source for public links and warn if it is set to localhost (#251)
  - Rewrite of UI (#245)
- REFACTOR
  - Remove github.com/kr/pretty in favor of assert.EqualValues() (#564)
  - Simplify web router code (#541)
  - Server obtain remote from glob config not from context (#540)
  - Serve index.html directly without template (#539)
  - Add linter revive, unused, ineffassign, varcheck, structcheck, staticcheck, whitespace, misspell (#550), (#551), (#554), (#538), (#537), (#535), (#531), (#530)
  - Rename struct field and add new types into server/model's (#523)
  - Update database in one transaction on syncing user repositories (#513)
  - Format code with 'simplify' flag and check via CI (#509)
  - Use Goblin Assert as intended (#501)
  - Embedding libcompose types for yaml parsing (#495)
  - Use std method to get SystemCertPool (#488)
  - Upgrade urfave/cli to v2 (#483)
  - Remove some wrapper and make code more readable (#478)
  - More logging and refactor (#457)
  - Simplify routes (#437)
  - Move api-routes to separate file (#434)
  - Rename drone-go to woodpecker-go (#390)
  - Remove ghodss/yaml (#384)
  - Move model/ to server/model/ (#366)
  - Use moby definitions for docker pipeline backend (#364)
  - Rewrite Gitlab Remote (#358)
  - Update Generated Proto Code (#351)
  - Remove legacy/unused code + misc cleanups (#331)
  - CLI use version from version/version.go (#329)
  - Move cli/drone/ to cli/ (#329)
  - Cleanup Code (#348)
  - Move cncd/pipeline/pipeline/ to pipeline/ (#347)
  - Move cncd/{logging,pubsub,queue}/ to server/{logging,pubsub,queue}/ (#346)
  - Move remote/ to server/remote/ (#344)
  - Move plugins/ to server/plugins/ (#343)
  - Move store/ to server/store/ (#341)
  - Move router/ to server/router/ (#339)
  - Create agent/ package for backend agnostic logic (#338)
  - Reorganize into server/{api,grpc,shared} packages (#337)
- TESTING
  - Add tests framework for storage migration (#630)
  - Add more golangci-lint linters & sort them (#499) (#502)
  - Compile on pull too (#287)
- DOCUMENTATION
  - Add note about Gitlab & Gitea internal connections to docs (#711)
  - Add registries docs (#679)
  - Add documentation of all agent configuration options (#667)
  - Add `repo` to `when` block (#642)
  - Add development docs (#610)
  - Clarify Docs on Docker for new users in intro (#606)
  - Update Documentation (fix diffs and add settings) (#569)
  - Add notice of supported YAML versions in docs (#556)
  - Update Agent and Pipeline syntax documentation (#506)
  - Update docs about selecting agent based on platform (#470)
  - Add plugin marketplace (for official plugins) (#451)
  - Add search to docs (#448)
  - Add image migration docs (#406)
  - Add security policy (#396)
  - Explain open registration setting (#361)
  - Add json schema and cli lint command (#342)
  - Improve docs deployment (#333)
  - Improve plugin docs (#313)
  - Add Support section to README (#310)
  - Community Guide (#296)
  - Migrate docs framework to Docusaurus (#282)
  - Use woodpecker env variable instead of drone in docker-compose (#264)
- MISC
  - Add support for building in docker (#759)
  - Compile for more platforms on release (#703)
  - Build agent for multiple platforms (arm, arm64, amd64, linux, windows, darwin) (#408)
  - Release deb, rpm bundles (#405)
  - Release cli images (#404)
  - Publish alpine container (#398)
  - Migrate go-docker to docker/docker (#363)
  - Use go's vendoring (#284)

## [v0.14.4](https://github.com/woodpecker-ci/woodpecker/releases/tag/v0.14.4) - 2022-01-31

- BUGFIXES
  - Docker Images use golang image for ca-certificates (#608)

## [v0.14.3](https://github.com/woodpecker-ci/woodpecker/releases/tag/v0.14.3) - 2021-10-30

- BUGFIXES
  - Add flag for not fetching permissions (FlatPermissions) (#491)
  - Gitea use default branch (#480) (#482)
  - Fix repo access (#476) (#477)
- ENHANCEMENTS
  - Use go embed for web files and remove httptreemux (#382) (#489)

## [v0.14.2](https://github.com/woodpecker-ci/woodpecker/releases/tag/v0.14.2) - 2021-10-19

- BUGFIXES
  - Fix sanitizePath (#326) (aa4fa9aab3)
  - Fix json tag for `Pos` at struct `Line` (#422) (#424)
  - Fix channel buffer used with signal.Notify (#421) (#423)
- ENHANCEMENTS
  - Support recursive glob for path conditions (#327) (#412)
- TESTING
  - Add TestPipelineName to procBuilder_test.go (#461) (#455)

## [v0.14.1](https://github.com/woodpecker-ci/woodpecker/releases/tag/v0.14.1) - 2021-09-21

- SECURITY
  - Migrate jwt token lib (#332)
- BUGFIXES
  - Increase allowed length for user token in db (#328)
  - Fix cli matrix filter (#311)
  - Fix ignore pushes to tags for gitea (#289)
  - Fix use custom config path to sanitize build names (#280)

## [v0.14.0](https://github.com/woodpecker-ci/woodpecker/releases/tag/v0.14.0) - 2021-08-01

- FEATURES
  - Add OAuth2 Support for Gitea Remote (#226)
  - Add support for path-prefix condition (#174)
- BUGFIXES
  - Allow multi pipeline file to be named .drone.yml (#250)
  - Fix release-server make target by build server with correct option (#237)
  - Fix Gitea unable to login on 0.12.0+ with error "cannot authenticate user. 403 Forbidden" (#221)
- ENHANCEMENTS
  - Update / Remove drone dependencies (#236)
  - Add support to gitea remote for path-prefix condition (#235)
  - Enable go vet for ci (#230)
  - Enforce code format (#228)
  - Add multi-pipeline to Gitea (#225)
  - Move flag definitions into extra files (#215)
  - Remove unused code in server (#213)
  - Docs URL configuration (#206)
  - Filter main branch (#205)
  - Fix multi pipeline bug when a pipeline depends on two other pipelines (#201)
  - Using configured server URL instead of obtained from request (#175)
- DOCUMENTATION
  - Switch in docs to new docker hub image repo (#227)
  - Use WOODPECKER_ env vars in docs (#211)
  - Also show WOODPECKER_HOST and WOODPECKER_SERVER_HOST environment variables in log messages (#208)
  - Move woodpecker to dedicated organisation on github (#202)
- MISC
  - Add chart for installing woodpecker server and agent (#199)
