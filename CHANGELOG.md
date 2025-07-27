# Changelog

## [3.9.0](https://github.com/woodpecker-ci/woodpecker/releases/tag/v3.9.0) - 2025-07-27

### ❤️ Thanks to all contributors! ❤️

@6543, @anbraten, @henkka, @hhamalai, @hrfee, @ivaltryek, @qwerty287, @wgroeneveld

### 🔒 Security

- Prevent secrets from leaking to Kubernetes API Server logs [[#5305](https://github.com/woodpecker-ci/woodpecker/pull/5305)]

### ✨ Features

- Add and edit additional forges in UI [[#5328](https://github.com/woodpecker-ci/woodpecker/pull/5328)]

### 📚 Documentation

- Add ASCII JUnit Test Report plugin [[#5355](https://github.com/woodpecker-ci/woodpecker/pull/5355)]
- fix(deps): update docs npm deps non-major [[#5340](https://github.com/woodpecker-ci/woodpecker/pull/5340)]
- chore(deps): update docs npm deps non-major [[#5316](https://github.com/woodpecker-ci/woodpecker/pull/5316)]

### 🐛 Bug Fixes

- Correct OpenApi LookupOrg router path [[#5351](https://github.com/woodpecker-ci/woodpecker/pull/5351)]
- fix(agent): handle context cancellation [[#5323](https://github.com/woodpecker-ci/woodpecker/pull/5323)]
- woodpecker-go/types: fix time-related struct field tags [[#5343](https://github.com/woodpecker-ci/woodpecker/pull/5343)]
- Reload repo on hook [[#5324](https://github.com/woodpecker-ci/woodpecker/pull/5324)]
- Fix loading icons and add missing loading indicators [[#5329](https://github.com/woodpecker-ci/woodpecker/pull/5329)]
- Use correct parameter for forge selection on login [[#5325](https://github.com/woodpecker-ci/woodpecker/pull/5325)]

### 📈 Enhancement

- feat(k8s): Kubernetes namespace per organization [[#5309](https://github.com/woodpecker-ci/woodpecker/pull/5309)]
- Simplify backend types [[#5299](https://github.com/woodpecker-ci/woodpecker/pull/5299)]

### 📦️ Dependency

- fix(deps): update module github.com/bmatcuk/doublestar/v4 to v4.9.1 [[#5365](https://github.com/woodpecker-ci/woodpecker/pull/5365)]
- fix(deps): update module github.com/google/go-github/v73 to v74 [[#5363](https://github.com/woodpecker-ci/woodpecker/pull/5363)]
- chore(deps): update dependency @antfu/eslint-config to v5 [[#5362](https://github.com/woodpecker-ci/woodpecker/pull/5362)]
- chore(deps): update web npm deps non-major [[#5361](https://github.com/woodpecker-ci/woodpecker/pull/5361)]
- chore(deps): update docker.io/mysql docker tag to v9.4.0 [[#5359](https://github.com/woodpecker-ci/woodpecker/pull/5359)]
- chore(deps): update pre-commit hook golangci/golangci-lint to v2.3.0 [[#5360](https://github.com/woodpecker-ci/woodpecker/pull/5360)]
- fix(deps): update golang-packages [[#5356](https://github.com/woodpecker-ci/woodpecker/pull/5356)]
- 📦 update web dependencies [[#5352](https://github.com/woodpecker-ci/woodpecker/pull/5352)]
- chore(config): migrate renovate config - autoclosed [[#5350](https://github.com/woodpecker-ci/woodpecker/pull/5350)]
- chore(deps): lock file maintenance [[#5348](https://github.com/woodpecker-ci/woodpecker/pull/5348)]
- fix(deps): update golang-packages [[#5347](https://github.com/woodpecker-ci/woodpecker/pull/5347)]
- fix(deps): update golang-packages [[#5336](https://github.com/woodpecker-ci/woodpecker/pull/5336)]
- chore(deps): lock file maintenance [[#5344](https://github.com/woodpecker-ci/woodpecker/pull/5344)]
- fix(deps): update dependency simple-icons to v15.7.0 [[#5342](https://github.com/woodpecker-ci/woodpecker/pull/5342)]
- fix(deps): update web npm deps non-major [[#5341](https://github.com/woodpecker-ci/woodpecker/pull/5341)]
- fix(deps): update dependency vue-i18n to v11.1.10 [security] [[#5335](https://github.com/woodpecker-ci/woodpecker/pull/5335)]
- fix(deps): update golang-packages [[#5333](https://github.com/woodpecker-ci/woodpecker/pull/5333)]
- chore(deps): lock file maintenance [[#5320](https://github.com/woodpecker-ci/woodpecker/pull/5320)]
- fix(deps): update dependency simple-icons to v15.6.0 [[#5319](https://github.com/woodpecker-ci/woodpecker/pull/5319)]
- fix(deps): update web npm deps non-major [[#5317](https://github.com/woodpecker-ci/woodpecker/pull/5317)]
- fix(deps): update module github.com/bmatcuk/doublestar/v4 to v4.9.0 [[#5318](https://github.com/woodpecker-ci/woodpecker/pull/5318)]
- chore(deps): update pre-commit hook golangci/golangci-lint to v2.2.2 [[#5315](https://github.com/woodpecker-ci/woodpecker/pull/5315)]
- chore(deps): update dependency golang to v1.24.5 [[#5314](https://github.com/woodpecker-ci/woodpecker/pull/5314)]
- fix(deps): update golang-packages [[#5313](https://github.com/woodpecker-ci/woodpecker/pull/5313)]
- fix(deps): update golang-packages [[#5311](https://github.com/woodpecker-ci/woodpecker/pull/5311)]
- fix(deps): update module gitlab.com/gitlab-org/api/client-go to v0.134.0 [[#5308](https://github.com/woodpecker-ci/woodpecker/pull/5308)]
- chore(deps): lock file maintenance [[#5307](https://github.com/woodpecker-ci/woodpecker/pull/5307)]
- fix(deps): update dependency simple-icons to v15.5.0 [[#5306](https://github.com/woodpecker-ci/woodpecker/pull/5306)]

### Misc

- 🧑‍💻 Add support for proxying to existing woodpecker server [[#5354](https://github.com/woodpecker-ci/woodpecker/pull/5354)]
- Update and improve nix flake [[#5349](https://github.com/woodpecker-ci/woodpecker/pull/5349)]
- Update issue number for link checker [[#5327](https://github.com/woodpecker-ci/woodpecker/pull/5327)]

## [3.8.0](https://github.com/woodpecker-ci/woodpecker/releases/tag/v3.8.0) - 2025-07-05

### ❤️ Thanks to all contributors! ❤️

@OCram85, @henkka, @johanvdw, @mmatous, @qwerty287

### 📚 Documentation

- chore(deps): lock file maintenance [[#5302](https://github.com/woodpecker-ci/woodpecker/pull/5302)]
- chore(deps): update dependency @types/node to v22.15.34 [[#5280](https://github.com/woodpecker-ci/woodpecker/pull/5280)]
- chore(deps): update dependency @types/node to v22.15.33 [[#5277](https://github.com/woodpecker-ci/woodpecker/pull/5277)]
- fix(deps): update docs npm deps non-major [[#5267](https://github.com/woodpecker-ci/woodpecker/pull/5267)]
- add Peckify plugin [[#5260](https://github.com/woodpecker-ci/woodpecker/pull/5260)]
- fix(deps): update docs npm deps non-major [[#5252](https://github.com/woodpecker-ci/woodpecker/pull/5252)]
- fix(deps): update docs npm deps non-major [[#5226](https://github.com/woodpecker-ci/woodpecker/pull/5226)]

### 🐛 Bug Fixes

- Fix gitlab MR fetching [[#5287](https://github.com/woodpecker-ci/woodpecker/pull/5287)]
- Use pipeline number in title [[#5275](https://github.com/woodpecker-ci/woodpecker/pull/5275)]
- Adjust documentation urls [[#5273](https://github.com/woodpecker-ci/woodpecker/pull/5273)]
- Fix doc links in agent settings [[#5251](https://github.com/woodpecker-ci/woodpecker/pull/5251)]

### 📈 Enhancement

- Add pipeline author and avatar env vars [[#5227](https://github.com/woodpecker-ci/woodpecker/pull/5227)]
- Support for pull request file changes in bitbucketdatacenter [[#5205](https://github.com/woodpecker-ci/woodpecker/pull/5205)]

### 📦️ Dependency

- chore(deps): update dependency vue-tsc to v3 [[#5301](https://github.com/woodpecker-ci/woodpecker/pull/5301)]
- chore(deps): update web npm deps non-major [[#5300](https://github.com/woodpecker-ci/woodpecker/pull/5300)]
- chore(deps): update docker.io/woodpeckerci/plugin-ready-release-go docker tag to v3.3.0 [[#5298](https://github.com/woodpecker-ci/woodpecker/pull/5298)]
- chore(deps): update docker.io/woodpeckerci/plugin-trivy docker tag to v1.4.1 [[#5297](https://github.com/woodpecker-ci/woodpecker/pull/5297)]
- chore(deps): update docker.io/woodpeckerci/plugin-docker-buildx docker tag to v6.0.2 [[#5295](https://github.com/woodpecker-ci/woodpecker/pull/5295)]
- chore(deps): update docker.io/woodpeckerci/plugin-editorconfig-checker docker tag to v0.3.1 [[#5296](https://github.com/woodpecker-ci/woodpecker/pull/5296)]
- chore(deps): lock file maintenance [[#5289](https://github.com/woodpecker-ci/woodpecker/pull/5289)]
- fix(deps): update web npm deps non-major [[#5281](https://github.com/woodpecker-ci/woodpecker/pull/5281)]
- fix(deps): update golang-packages [[#5291](https://github.com/woodpecker-ci/woodpecker/pull/5291)]
- chore(deps): update pre-commit hook golangci/golangci-lint to v2.2.1 [[#5288](https://github.com/woodpecker-ci/woodpecker/pull/5288)]
- fix(deps): update dependency marked to v16 [[#5284](https://github.com/woodpecker-ci/woodpecker/pull/5284)]
- chore(deps): update dependency @vitejs/plugin-vue to v6 [[#5282](https://github.com/woodpecker-ci/woodpecker/pull/5282)]
- chore(deps): update pre-commit hook golangci/golangci-lint to v2.2.0 [[#5286](https://github.com/woodpecker-ci/woodpecker/pull/5286)]
- chore(deps): update dependency vite to v7 [[#5283](https://github.com/woodpecker-ci/woodpecker/pull/5283)]
- fix(deps): update module github.com/google/go-github/v72 to v73 [[#5285](https://github.com/woodpecker-ci/woodpecker/pull/5285)]
- chore(deps): update pre-commit hook rbubley/mirrors-prettier to v3.6.2 [[#5278](https://github.com/woodpecker-ci/woodpecker/pull/5278)]
- fix(deps): update golang-packages to v28.3.0+incompatible [[#5274](https://github.com/woodpecker-ci/woodpecker/pull/5274)]
- chore(deps): lock file maintenance [[#5271](https://github.com/woodpecker-ci/woodpecker/pull/5271)]
- fix(deps): update dependency vue-i18n to v11.1.7 [[#5270](https://github.com/woodpecker-ci/woodpecker/pull/5270)]
- fix(deps): update dependency simple-icons to v15.3.0 [[#5269](https://github.com/woodpecker-ci/woodpecker/pull/5269)]
- fix(deps): update web npm deps non-major [[#5268](https://github.com/woodpecker-ci/woodpecker/pull/5268)]
- fix(deps): update golang-packages to v0.33.2 [[#5265](https://github.com/woodpecker-ci/woodpecker/pull/5265)]
- fix(deps): update golang-packages [[#5261](https://github.com/woodpecker-ci/woodpecker/pull/5261)]
- fix(deps): update module github.com/go-viper/mapstructure/v2 to v2.3.0 [[#5259](https://github.com/woodpecker-ci/woodpecker/pull/5259)]
- chore(deps): lock file maintenance [[#5257](https://github.com/woodpecker-ci/woodpecker/pull/5257)]
- fix(deps): update dependency simple-icons to v15.2.0 [[#5256](https://github.com/woodpecker-ci/woodpecker/pull/5256)]
- fix(deps): update web npm deps non-major [[#5254](https://github.com/woodpecker-ci/woodpecker/pull/5254)]
- chore(deps): update gitea/gitea docker tag to v1.24 [[#5253](https://github.com/woodpecker-ci/woodpecker/pull/5253)]
- fix(deps): update golang-packages [[#5250](https://github.com/woodpecker-ci/woodpecker/pull/5250)]
- chore(deps): lock file maintenance [[#5233](https://github.com/woodpecker-ci/woodpecker/pull/5233)]
- fix(deps): update dependency simple-icons to v15.1.0 [[#5246](https://github.com/woodpecker-ci/woodpecker/pull/5246)]
- fix(deps): update web npm deps non-major [[#5244](https://github.com/woodpecker-ci/woodpecker/pull/5244)]
- fix(deps): update golang-packages [[#5242](https://github.com/woodpecker-ci/woodpecker/pull/5242)]
- chore(deps): update dependency golang to v1.24.4 [[#5241](https://github.com/woodpecker-ci/woodpecker/pull/5241)]

### Misc

- Disable package name linting [[#5294](https://github.com/woodpecker-ci/woodpecker/pull/5294)]

## [3.7.0](https://github.com/woodpecker-ci/woodpecker/releases/tag/v3.7.0) - 2025-06-06

### ❤️ Thanks to all contributors! ❤️

@6543, @Epsilon02, @Levy-Tal, @OCram85, @Spiffyk, @SuperSandro2000, @deltamualpha, @qwerty287, @rruzicic, @sebastinez, @xoxys

### 📚 Documentation

- update docs-link for todo checker [[#5236](https://github.com/woodpecker-ci/woodpecker/pull/5236)]
- Add `sccache` plugin [[#5234](https://github.com/woodpecker-ci/woodpecker/pull/5234)]
- fix(deps): update dependency redocusaurus to v2.3.0 [[#5203](https://github.com/woodpecker-ci/woodpecker/pull/5203)]
- chore(deps): update docs npm deps non-major [[#5197](https://github.com/woodpecker-ci/woodpecker/pull/5197)]
- Add reference to woodpecker-community plugin org [[#5186](https://github.com/woodpecker-ci/woodpecker/pull/5186)]
- fix(deps): update docs npm deps non-major [[#5183](https://github.com/woodpecker-ci/woodpecker/pull/5183)]
- Move `gitea-package` plugin to codeberg [[#5175](https://github.com/woodpecker-ci/woodpecker/pull/5175)]
- add Portainer Service Update plugin [[#5172](https://github.com/woodpecker-ci/woodpecker/pull/5172)]
- Split 'pull' option docs from 'image' docs [[#5161](https://github.com/woodpecker-ci/woodpecker/pull/5161)]
- chore(deps): update docs npm deps non-major [[#5164](https://github.com/woodpecker-ci/woodpecker/pull/5164)]

### 📈 Enhancement

- Move forge webhook fixtures into own files [[#5216](https://github.com/woodpecker-ci/woodpecker/pull/5216)]
- Treat no available route in grpc as fatal error [[#5192](https://github.com/woodpecker-ci/woodpecker/pull/5192)]

### 🐛 Bug Fixes

- Always collect metrics (reverts #4667) [[#5213](https://github.com/woodpecker-ci/woodpecker/pull/5213)]
- fix(bitbucketDC): manual event has broken commit link [[#5160](https://github.com/woodpecker-ci/woodpecker/pull/5160)]
- fix(bitbucketdc): build status gets incorrectly reported on multi workflow builds [[#5178](https://github.com/woodpecker-ci/woodpecker/pull/5178)]
- fix(bitbucketdc): build status not reported on PR builds [[#5162](https://github.com/woodpecker-ci/woodpecker/pull/5162)]

### 📦️ Dependency

- fix(deps): update golang-packages to v28.2.1+incompatible [[#5217](https://github.com/woodpecker-ci/woodpecker/pull/5217)]
- fix(deps): update dependency simple-icons to v15 [[#5232](https://github.com/woodpecker-ci/woodpecker/pull/5232)]
- chore(deps): update woodpeckerci/plugin-git docker tag to v2.6.5 [[#5230](https://github.com/woodpecker-ci/woodpecker/pull/5230)]
- fix(deps): update web npm deps non-major [[#5228](https://github.com/woodpecker-ci/woodpecker/pull/5228)]
- chore(deps): update docker.io/woodpeckerci/plugin-surge-preview docker tag to v1.4.0 [[#5225](https://github.com/woodpecker-ci/woodpecker/pull/5225)]
- chore(deps): update docker.io/alpine docker tag to v3.22 [[#5224](https://github.com/woodpecker-ci/woodpecker/pull/5224)]
- fix(deps): update golang-packages [[#5209](https://github.com/woodpecker-ci/woodpecker/pull/5209)]
- chore(deps): lock file maintenance [[#5204](https://github.com/woodpecker-ci/woodpecker/pull/5204)]
- fix(deps): update dependency simple-icons to v14.15.0 [[#5202](https://github.com/woodpecker-ci/woodpecker/pull/5202)]
- fix(deps): update dependency vue-i18n to v11.1.4 [[#5201](https://github.com/woodpecker-ci/woodpecker/pull/5201)]
- chore(deps): update docker.io/woodpeckerci/plugin-surge-preview docker tag to v1.3.6 [[#5200](https://github.com/woodpecker-ci/woodpecker/pull/5200)]
- fix(deps): update web npm deps non-major [[#5198](https://github.com/woodpecker-ci/woodpecker/pull/5198)]
- fix(deps): update module github.com/oklog/ulid/v2 to v2.1.1 [[#5194](https://github.com/woodpecker-ci/woodpecker/pull/5194)]
- fix(deps): update module github.com/gin-gonic/gin to v1.10.1 [[#5193](https://github.com/woodpecker-ci/woodpecker/pull/5193)]
- fix(deps): update module gitlab.com/gitlab-org/api/client-go to v0.129.0 [[#5190](https://github.com/woodpecker-ci/woodpecker/pull/5190)]
- chore(deps): lock file maintenance [[#5189](https://github.com/woodpecker-ci/woodpecker/pull/5189)]
- chore(deps): update pre-commit hook igorshubovych/markdownlint-cli to v0.45.0 [[#5187](https://github.com/woodpecker-ci/woodpecker/pull/5187)]
- fix(deps): update dependency simple-icons to v14.14.0 [[#5188](https://github.com/woodpecker-ci/woodpecker/pull/5188)]
- fix(deps): update web npm deps non-major [[#5185](https://github.com/woodpecker-ci/woodpecker/pull/5185)]
- fix(deps): update golang-packages to v0.33.1 [[#5184](https://github.com/woodpecker-ci/woodpecker/pull/5184)]
- fix(deps): update golang-packages [[#5180](https://github.com/woodpecker-ci/woodpecker/pull/5180)]
- chore(deps): lock file maintenance [[#5171](https://github.com/woodpecker-ci/woodpecker/pull/5171)]
- fix(deps): update module github.com/google/go-github/v71 to v72 [[#5167](https://github.com/woodpecker-ci/woodpecker/pull/5167)]
- fix(deps): update dependency simple-icons to v14.13.0 [[#5170](https://github.com/woodpecker-ci/woodpecker/pull/5170)]
- fix(deps): update module github.com/urfave/cli/v3 to v3.3.3 [[#5169](https://github.com/woodpecker-ci/woodpecker/pull/5169)]
- fix(deps): update web npm deps non-major [[#5166](https://github.com/woodpecker-ci/woodpecker/pull/5166)]
- chore(deps): update postgres docker tag to v17.5 [[#5165](https://github.com/woodpecker-ci/woodpecker/pull/5165)]
- chore(deps): update dependency golang to v1.24.3 [[#5163](https://github.com/woodpecker-ci/woodpecker/pull/5163)]

### Misc

- Ignore direnv config and folder [[#5235](https://github.com/woodpecker-ci/woodpecker/pull/5235)]
- flake.lock: Update [[#5206](https://github.com/woodpecker-ci/woodpecker/pull/5206)]
- Add Bluesky post plugin [[#5156](https://github.com/woodpecker-ci/woodpecker/pull/5156)]

## [3.6.0](https://github.com/woodpecker-ci/woodpecker/releases/tag/v3.6.0) - 2025-05-06

### ❤️ Thanks to all contributors! ❤️

@Spiffyk, @SuperSandro2000, @gsaslis, @joshuachp, @lukashass, @maurerle, @pat-s, @qwerty287, @renich, @sp1thas, @xoxys

### ✨ Features

- Use docker go client directly [[#5134](https://github.com/woodpecker-ci/woodpecker/pull/5134)]

### 📚 Documentation

- Simplify NixOS docs [[#5120](https://github.com/woodpecker-ci/woodpecker/pull/5120)]
- chore(deps): lock file maintenance [[#5150](https://github.com/woodpecker-ci/woodpecker/pull/5150)]
- plugins: Add SSH/SCP plugin [[#4871](https://github.com/woodpecker-ci/woodpecker/pull/4871)]
- chore(deps): update dependency @types/node to v22.15.3 [[#5142](https://github.com/woodpecker-ci/woodpecker/pull/5142)]
- chore(deps): lock file maintenance [[#5136](https://github.com/woodpecker-ci/woodpecker/pull/5136)]
- Explain tasks [[#5129](https://github.com/woodpecker-ci/woodpecker/pull/5129)]
- Mention named volumes [[#5130](https://github.com/woodpecker-ci/woodpecker/pull/5130)]
- chore(deps): update docs npm deps non-major [[#5128](https://github.com/woodpecker-ci/woodpecker/pull/5128)]
- Fix link to agent configuration in `v3.5` docs [[#5122](https://github.com/woodpecker-ci/woodpecker/pull/5122)]
- Fix link to agent configuration in `next` docs [[#5119](https://github.com/woodpecker-ci/woodpecker/pull/5119)]
- Move `plugin-s3` to Codeberg [[#5118](https://github.com/woodpecker-ci/woodpecker/pull/5118)]
- Use slugified plugin urls in docs [[#5116](https://github.com/woodpecker-ci/woodpecker/pull/5116)]
- Fix example value for `WOODPECKER_GRPC_ADDR` in autoscaler docs [[#5102](https://github.com/woodpecker-ci/woodpecker/pull/5102)]
- .deb and .rpm installation commands fixed [[#5087](https://github.com/woodpecker-ci/woodpecker/pull/5087)]
- chore(deps): update dependency @types/react to v19.1.2 [[#5107](https://github.com/woodpecker-ci/woodpecker/pull/5107)]
- Slugify plugin names used for urls [[#5098](https://github.com/woodpecker-ci/woodpecker/pull/5098)]
- Mention `backend_options` in workflow syntax docs [[#5096](https://github.com/woodpecker-ci/woodpecker/pull/5096)]
- Document rootless container requirements for skip-clone [[#5056](https://github.com/woodpecker-ci/woodpecker/pull/5056)]

### 📈 Enhancement

- View full pipeline duration in tooltip [[#5123](https://github.com/woodpecker-ci/woodpecker/pull/5123)]
- Set dynamic page titles [[#5104](https://github.com/woodpecker-ci/woodpecker/pull/5104)]
- Use centrally typed inject provide in Vue [[#5113](https://github.com/woodpecker-ci/woodpecker/pull/5113)]
- Scroll to selected pipeline step [[#5103](https://github.com/woodpecker-ci/woodpecker/pull/5103)]

### 🐛 Bug Fixes

- Fix args docs for admin secrets [[#5127](https://github.com/woodpecker-ci/woodpecker/pull/5127)]
- Add name flag to admin secret add [[#5101](https://github.com/woodpecker-ci/woodpecker/pull/5101)]

### 📦️ Dependency

- fix(deps): update golang-packages [[#5152](https://github.com/woodpecker-ci/woodpecker/pull/5152)]
- chore(deps): update pre-commit hook golangci/golangci-lint to v2.1.6 [[#5149](https://github.com/woodpecker-ci/woodpecker/pull/5149)]
- chore(deps): update docker.io/woodpeckerci/plugin-docker-buildx docker tag to v6.0.1 [[#5147](https://github.com/woodpecker-ci/woodpecker/pull/5147)]
- chore(deps): update pre-commit hook adrienverge/yamllint to v1.37.1 [[#5148](https://github.com/woodpecker-ci/woodpecker/pull/5148)]
- chore(deps): update docker.io/woodpeckerci/plugin-docker-buildx docker tag to v6 [[#5144](https://github.com/woodpecker-ci/woodpecker/pull/5144)]
- fix(deps): update web npm deps non-major [[#5143](https://github.com/woodpecker-ci/woodpecker/pull/5143)]
- fix(deps): update module github.com/getkin/kin-openapi to v0.132.0 [[#5141](https://github.com/woodpecker-ci/woodpecker/pull/5141)]
- chore(deps): update dependency vite to v6.3.4 [security] [[#5139](https://github.com/woodpecker-ci/woodpecker/pull/5139)]
- fix(deps): update module github.com/urfave/cli/v3 to v3.3.2 [[#5137](https://github.com/woodpecker-ci/woodpecker/pull/5137)]
- fix(deps): update module github.com/urfave/cli/v3 to v3.3.1 [[#5135](https://github.com/woodpecker-ci/woodpecker/pull/5135)]
- fix(deps): update module github.com/docker/docker to v28 [[#5132](https://github.com/woodpecker-ci/woodpecker/pull/5132)]
- fix(deps): update module github.com/docker/cli to v28 [[#5131](https://github.com/woodpecker-ci/woodpecker/pull/5131)]
- fix(deps): update dependency vue-router to v4.5.1 [[#5126](https://github.com/woodpecker-ci/woodpecker/pull/5126)]
- chore(deps): update pre-commit hook golangci/golangci-lint to v2.1.5 [[#5125](https://github.com/woodpecker-ci/woodpecker/pull/5125)]
- fix(deps): update web npm deps non-major [[#5077](https://github.com/woodpecker-ci/woodpecker/pull/5077)]
- fix(deps): update golang-packages [[#5121](https://github.com/woodpecker-ci/woodpecker/pull/5121)]
- fix(deps): update golang-packages [[#5111](https://github.com/woodpecker-ci/woodpecker/pull/5111)]
- chore(deps): lock file maintenance [[#5112](https://github.com/woodpecker-ci/woodpecker/pull/5112)]
- chore(deps): update docker.io/mysql docker tag to v9.3.0 [[#5109](https://github.com/woodpecker-ci/woodpecker/pull/5109)]
- chore(deps): update docker.io/woodpeckerci/plugin-ready-release-go docker tag to v3.2.0 [[#5110](https://github.com/woodpecker-ci/woodpecker/pull/5110)]
- chore(deps): update pre-commit hook golangci/golangci-lint to v2.1.2 [[#5108](https://github.com/woodpecker-ci/woodpecker/pull/5108)]
- fix(deps): update golang-packages [[#5097](https://github.com/woodpecker-ci/woodpecker/pull/5097)]

### Misc

- Add pre-commit plugin [[#5146](https://github.com/woodpecker-ci/woodpecker/pull/5146)]
- Fix gitpod golang version [[#5093](https://github.com/woodpecker-ci/woodpecker/pull/5093)]

## [3.5.2](https://github.com/woodpecker-ci/woodpecker/releases/tag/v3.5.2) - 2025-04-15

### ❤️ Thanks to all contributors! ❤️

@xoxys

### 📚 Documentation

- chore(deps): lock file maintenance [[#5092](https://github.com/woodpecker-ci/woodpecker/pull/5092)]
- fix(deps): update docs npm deps non-major [[#5089](https://github.com/woodpecker-ci/woodpecker/pull/5089)]
- Move plugin-surge docs to codeberg [[#5086](https://github.com/woodpecker-ci/woodpecker/pull/5086)]
- chore(deps): lock file maintenance [[#5080](https://github.com/woodpecker-ci/woodpecker/pull/5080)]
- chore(deps): update docs npm deps non-major [[#5075](https://github.com/woodpecker-ci/woodpecker/pull/5075)]

### 🐛 Bug Fixes

- Avoid db errors while executing migrations check [[#5072](https://github.com/woodpecker-ci/woodpecker/pull/5072)]

### 📦️ Dependency

- fix(deps): update module github.com/google/go-github/v70 to v71 [[#5090](https://github.com/woodpecker-ci/woodpecker/pull/5090)]
- chore(deps): update pre-commit hook golangci/golangci-lint to v2.1.1 [[#5091](https://github.com/woodpecker-ci/woodpecker/pull/5091)]
- chore(deps): update dependency vite to v6.2.6 [security] [[#5088](https://github.com/woodpecker-ci/woodpecker/pull/5088)]
- fix(deps): update module github.com/prometheus/client_golang to v1.22.0 [[#5084](https://github.com/woodpecker-ci/woodpecker/pull/5084)]
- fix(deps): update golang-packages [[#5083](https://github.com/woodpecker-ci/woodpecker/pull/5083)]
- fix(deps): update module golang.org/x/crypto to v0.37.0 [[#5079](https://github.com/woodpecker-ci/woodpecker/pull/5079)]
- fix(deps): update golang-packages [[#5078](https://github.com/woodpecker-ci/woodpecker/pull/5078)]
- fix(deps): update module github.com/fsnotify/fsnotify to v1.9.0 [[#5076](https://github.com/woodpecker-ci/woodpecker/pull/5076)]
- chore(deps): update dependency vite to v6.2.5 [security] [[#5074](https://github.com/woodpecker-ci/woodpecker/pull/5074)]

### Misc

- Add markdown template for release umbrella issues [[#5055](https://github.com/woodpecker-ci/woodpecker/pull/5055)]

## [3.5.1](https://github.com/woodpecker-ci/woodpecker/releases/tag/v3.5.1) - 2025-04-04

### ❤️ Thanks to all contributors! ❤️

@xoxys

### 🐛 Bug Fixes

- Add missing icon for changes files tab [[#5068](https://github.com/woodpecker-ci/woodpecker/pull/5068)]
- Improve CLI info text and remove markdown [[#5069](https://github.com/woodpecker-ci/woodpecker/pull/5069)]
- Fix cli format flag fallback [[#5057](https://github.com/woodpecker-ci/woodpecker/pull/5057)]

### 📚 Documentation

- chore(deps): update docs npm deps non-major [[#5060](https://github.com/woodpecker-ci/woodpecker/pull/5060)]

### 📦️ Dependency

- fix(deps): update module code.gitea.io/sdk/gitea to v0.21.0 [[#5067](https://github.com/woodpecker-ci/woodpecker/pull/5067)]
- chore(deps): lock file maintenance [[#5062](https://github.com/woodpecker-ci/woodpecker/pull/5062)]
- fix(deps): update module github.com/mattn/go-sqlite3 to v1.14.27 [[#5058](https://github.com/woodpecker-ci/woodpecker/pull/5058)]

## [3.5.0](https://github.com/woodpecker-ci/woodpecker/releases/tag/v3.5.0) - 2025-04-02

### ❤️ Thanks to all contributors! ❤️

@6543, @Levy-Tal, @anbraten, @jenrik, @nekowinston, @qwerty287, @rhafer, @xoxys

### 🐛 Bug Fixes

- BitbucketDC: add event pull request opened [[#5048](https://github.com/woodpecker-ci/woodpecker/pull/5048)]
- Fix exclude path constraint behavior [[#5042](https://github.com/woodpecker-ci/woodpecker/pull/5042)]
- Use pointer cursor for icon buttons [[#5002](https://github.com/woodpecker-ci/woodpecker/pull/5002)]
- Add back cursor-pointer to pipeline step list buttons [[#4982](https://github.com/woodpecker-ci/woodpecker/pull/4982)]

### 📚 Documentation

- chore(deps): lock file maintenance [[#5044](https://github.com/woodpecker-ci/woodpecker/pull/5044)]
- chore(deps): lock file maintenance [[#5032](https://github.com/woodpecker-ci/woodpecker/pull/5032)]
- Print at which file docs parsing failed [[#5040](https://github.com/woodpecker-ci/woodpecker/pull/5040)]
- fix(deps): update dependency yaml to v2.7.1 [[#5029](https://github.com/woodpecker-ci/woodpecker/pull/5029)]
- fix(deps): update docs npm deps non-major [[#5026](https://github.com/woodpecker-ci/woodpecker/pull/5026)]
- Revert manual changes to changelog [[#5007](https://github.com/woodpecker-ci/woodpecker/pull/5007)]
- Add missing docs for 3.x minor versions [[#4992](https://github.com/woodpecker-ci/woodpecker/pull/4992)]
- chore(deps): lock file maintenance [[#5000](https://github.com/woodpecker-ci/woodpecker/pull/5000)]
- fix(deps): update dependency redocusaurus to v2.2.2 [[#4998](https://github.com/woodpecker-ci/woodpecker/pull/4998)]
- Add missing links to 3.x docs [[#4991](https://github.com/woodpecker-ci/woodpecker/pull/4991)]
- chore(deps): update docs npm deps non-major [[#4987](https://github.com/woodpecker-ci/woodpecker/pull/4987)]
- Rework secrets docs and document multiline secrets [[#4974](https://github.com/woodpecker-ci/woodpecker/pull/4974)]
- Add documentation for WOODPECKER_EXPERT env vars [[#4972](https://github.com/woodpecker-ci/woodpecker/pull/4972)]

### 📈 Enhancement

- add nushell support to local backend [[#5043](https://github.com/woodpecker-ci/woodpecker/pull/5043)]
- Style navbar login button as navbar-link [[#5033](https://github.com/woodpecker-ci/woodpecker/pull/5033)]
- Use xorm quoter for feed query [[#5018](https://github.com/woodpecker-ci/woodpecker/pull/5018)]
- Use badge value instead of label for single values [[#5010](https://github.com/woodpecker-ci/woodpecker/pull/5010)]
- Add icons to all tabs [[#4421](https://github.com/woodpecker-ci/woodpecker/pull/4421)]
- Tag pipeline with source information [[#4796](https://github.com/woodpecker-ci/woodpecker/pull/4796)]
- Add titles and descriptions to repos page [[#4981](https://github.com/woodpecker-ci/woodpecker/pull/4981)]

### 📦️ Dependency

- fix(deps): update golang-packages [[#5046](https://github.com/woodpecker-ci/woodpecker/pull/5046)]
- fix(deps): update module github.com/urfave/cli/v3 to v3.1.0 [[#5039](https://github.com/woodpecker-ci/woodpecker/pull/5039)]
- chore(deps): update dependency vite to v6.2.4 [security] [[#5036](https://github.com/woodpecker-ci/woodpecker/pull/5036)]
- fix(deps): update dependency simple-icons to v14.12.0 [[#5030](https://github.com/woodpecker-ci/woodpecker/pull/5030)]
- chore(deps): update pre-commit hook golangci/golangci-lint to v2 [[#5028](https://github.com/woodpecker-ci/woodpecker/pull/5028)]
- fix(deps): update web npm deps non-major [[#5027](https://github.com/woodpecker-ci/woodpecker/pull/5027)]
- chore(deps): update docker.io/woodpeckerci/plugin-ready-release-go docker tag to v3.1.4 [[#5025](https://github.com/woodpecker-ci/woodpecker/pull/5025)]
- fix(deps): update module golang.org/x/net to v0.38.0 [[#5024](https://github.com/woodpecker-ci/woodpecker/pull/5024)]
- chore(deps): update woodpeckerci/plugin-git docker tag to v2.6.3 [[#5021](https://github.com/woodpecker-ci/woodpecker/pull/5021)]
- chore(deps): update dependency vite to v6.2.3 [security] [[#5014](https://github.com/woodpecker-ci/woodpecker/pull/5014)]
- fix(deps): update golang-packages [[#5012](https://github.com/woodpecker-ci/woodpecker/pull/5012)]
- chore(deps): update docker.io/woodpeckerci/plugin-docker-buildx docker tag to v5.2.2 [[#4997](https://github.com/woodpecker-ci/woodpecker/pull/4997)]
- fix(deps): update dependency simple-icons to v14.11.1 [[#4999](https://github.com/woodpecker-ci/woodpecker/pull/4999)]
- chore(deps): update pre-commit hook adrienverge/yamllint to v1.37.0 [[#4996](https://github.com/woodpecker-ci/woodpecker/pull/4996)]
- fix(deps): update module github.com/rs/zerolog to v1.34.0 [[#4995](https://github.com/woodpecker-ci/woodpecker/pull/4995)]
- chore(deps): update dependency @antfu/eslint-config to v4.11.0 [[#4994](https://github.com/woodpecker-ci/woodpecker/pull/4994)]
- chore(deps): update woodpeckerci/plugin-release docker tag to v0.2.5 [[#4993](https://github.com/woodpecker-ci/woodpecker/pull/4993)]
- fix(deps): update module github.com/google/go-github/v69 to v70 [[#4990](https://github.com/woodpecker-ci/woodpecker/pull/4990)]
- fix(deps): update web npm deps non-major [[#4989](https://github.com/woodpecker-ci/woodpecker/pull/4989)]
- chore(deps): update pre-commit non-major [[#4988](https://github.com/woodpecker-ci/woodpecker/pull/4988)]
- fix(deps): update module github.com/golang-jwt/jwt/v5 to v5.2.2 [security] [[#4986](https://github.com/woodpecker-ci/woodpecker/pull/4986)]
- fix(deps): update module github.com/go-sql-driver/mysql to v1.9.1 [[#4985](https://github.com/woodpecker-ci/woodpecker/pull/4985)]
- fix(deps): update module github.com/getkin/kin-openapi to v0.131.0 [[#4984](https://github.com/woodpecker-ci/woodpecker/pull/4984)]
- fix(deps): update module github.com/expr-lang/expr to v1.17.1 [[#4983](https://github.com/woodpecker-ci/woodpecker/pull/4983)]
- fix(deps): update module gitlab.com/gitlab-org/api/client-go to v0.126.0 [[#4976](https://github.com/woodpecker-ci/woodpecker/pull/4976)]

### Misc

- Bump golangci-lint to v2 [[#5034](https://github.com/woodpecker-ci/woodpecker/pull/5034)]
- Update flake development environment [[#5022](https://github.com/woodpecker-ci/woodpecker/pull/5022)]

## [3.4.0](https://github.com/woodpecker-ci/woodpecker/releases/tag/v3.4.0) - 2025-03-17

### ❤️ Thanks to all contributors! ❤️

@qwerty287, @xoxys

### 📈 Enhancement

- Remove woodpecker prefix from env var title in docs [[#4968](https://github.com/woodpecker-ci/woodpecker/pull/4968)]
- Add backoff retry for store setup [[#4964](https://github.com/woodpecker-ci/woodpecker/pull/4964)]
- Migrate repo output format to customizable output [[#4888](https://github.com/woodpecker-ci/woodpecker/pull/4888)]

### 📚 Documentation

- chore(deps): lock file maintenance [[#4970](https://github.com/woodpecker-ci/woodpecker/pull/4970)]
- fix(deps): update docs npm deps non-major [[#4958](https://github.com/woodpecker-ci/woodpecker/pull/4958)]
- Add global var note [[#4956](https://github.com/woodpecker-ci/woodpecker/pull/4956)]
- chore(deps): lock file maintenance [[#4948](https://github.com/woodpecker-ci/woodpecker/pull/4948)]
- chore(deps): update dependency @types/node to v22.13.10 [[#4944](https://github.com/woodpecker-ci/woodpecker/pull/4944)]
- chore(deps): update dependency axios to v1.8.2 [security] [[#4941](https://github.com/woodpecker-ci/woodpecker/pull/4941)]
- Fix dockerhub links in docs [[#4931](https://github.com/woodpecker-ci/woodpecker/pull/4931)]

### 🐛 Bug Fixes

- Fix fs owner in scratch-based container images [[#4961](https://github.com/woodpecker-ci/woodpecker/pull/4961)]

### 📦️ Dependency

- fix(deps): update module github.com/expr-lang/expr to v1.17.0 [[#4969](https://github.com/woodpecker-ci/woodpecker/pull/4969)]
- fix(deps): update dependency simple-icons to v14.11.0 [[#4966](https://github.com/woodpecker-ci/woodpecker/pull/4966)]
- fix(deps): update golang-packages [[#4963](https://github.com/woodpecker-ci/woodpecker/pull/4963)]
- chore(deps): update pre-commit hook adrienverge/yamllint to v1.36.1 [[#4962](https://github.com/woodpecker-ci/woodpecker/pull/4962)]
- fix(deps): update dependency @vueuse/core to v13 [[#4960](https://github.com/woodpecker-ci/woodpecker/pull/4960)]
- fix(deps): update web npm deps non-major [[#4959](https://github.com/woodpecker-ci/woodpecker/pull/4959)]
- chore(deps): update pre-commit non-major [[#4957](https://github.com/woodpecker-ci/woodpecker/pull/4957)]
- fix(deps): update golang-packages to v0.32.3 [[#4953](https://github.com/woodpecker-ci/woodpecker/pull/4953)]
- fix(deps): update dependency prismjs to v1.30.0 [security] [[#4951](https://github.com/woodpecker-ci/woodpecker/pull/4951)]
- chore(deps): update dependency @intlify/eslint-plugin-vue-i18n to v4 [[#4943](https://github.com/woodpecker-ci/woodpecker/pull/4943)]
- fix(deps): update module al.essio.dev/pkg/shellescape to v1.6.0 [[#4947](https://github.com/woodpecker-ci/woodpecker/pull/4947)]
- fix(deps): update dependency simple-icons to v14.10.0 [[#4946](https://github.com/woodpecker-ci/woodpecker/pull/4946)]
- chore(deps): update dependency @types/node to v22.13.10 [[#4945](https://github.com/woodpecker-ci/woodpecker/pull/4945)]
- fix(deps): update web npm deps non-major [[#4942](https://github.com/woodpecker-ci/woodpecker/pull/4942)]
- fix(deps): update dependency vue-i18n to v11.1.2 [security] [[#4940](https://github.com/woodpecker-ci/woodpecker/pull/4940)]
- fix(deps): update golang-packages [[#4936](https://github.com/woodpecker-ci/woodpecker/pull/4936)]
- chore(deps): lock file maintenance [[#4933](https://github.com/woodpecker-ci/woodpecker/pull/4933)]
- fix(deps): update golang-packages [[#4929](https://github.com/woodpecker-ci/woodpecker/pull/4929)]

## [3.3.0](https://github.com/woodpecker-ci/woodpecker/releases/tag/v3.3.0) - 2025-03-04

### ❤️ Thanks to all contributors! ❤️

@Levy-Tal, @qwerty287, @xoxys

### 📚 Documentation

- Refactor admin docs [[#4899](https://github.com/woodpecker-ci/woodpecker/pull/4899)]
- chore(deps): lock file maintenance [[#4928](https://github.com/woodpecker-ci/woodpecker/pull/4928)]
- chore(deps): update dependency @types/node to v22.13.9 [[#4925](https://github.com/woodpecker-ci/woodpecker/pull/4925)]
- chore(deps): lock file maintenance [[#4922](https://github.com/woodpecker-ci/woodpecker/pull/4922)]
- Add some blog posts [[#4921](https://github.com/woodpecker-ci/woodpecker/pull/4921)]
- chore(deps): update dependency @types/node to v22.13.8 [[#4915](https://github.com/woodpecker-ci/woodpecker/pull/4915)]
- Remove Slack plugin from examples [[#4914](https://github.com/woodpecker-ci/woodpecker/pull/4914)]
- chore(deps): update docs npm deps non-major [[#4911](https://github.com/woodpecker-ci/woodpecker/pull/4911)]

### 🐛 Bug Fixes

- Add migration to fix zero forge_id in orgs table [[#4924](https://github.com/woodpecker-ci/woodpecker/pull/4924)]
- Fix unique constraint for orgs [[#4923](https://github.com/woodpecker-ci/woodpecker/pull/4923)]

### 📈 Enhancement

- BitbucketDC: optimize repository search [[#4919](https://github.com/woodpecker-ci/woodpecker/pull/4919)]
- Include forge type in netrc [[#4908](https://github.com/woodpecker-ci/woodpecker/pull/4908)]

### 📦️ Dependency

- chore(deps): update dependency @types/node to v22.13.9 [[#4926](https://github.com/woodpecker-ci/woodpecker/pull/4926)]
- chore(deps): update pre-commit non-major [[#4927](https://github.com/woodpecker-ci/woodpecker/pull/4927)]
- chore(deps): update dependency @antfu/eslint-config to v4.4.0 [[#4917](https://github.com/woodpecker-ci/woodpecker/pull/4917)]
- fix(deps): update module gitlab.com/gitlab-org/api/client-go to v0.124.0 [[#4920](https://github.com/woodpecker-ci/woodpecker/pull/4920)]
- chore(deps): update dependency @types/node to v22.13.8 [[#4916](https://github.com/woodpecker-ci/woodpecker/pull/4916)]
- chore(deps): update dependency @types/lodash to v4.17.16 [[#4913](https://github.com/woodpecker-ci/woodpecker/pull/4913)]
- chore(deps): update web npm deps non-major [[#4912](https://github.com/woodpecker-ci/woodpecker/pull/4912)]

## [3.2.0](https://github.com/woodpecker-ci/woodpecker/releases/tag/v3.2.0) - 2025-02-26

### ❤️ Thanks to all contributors! ❤️

@DHandspikerWade, @anbraten, @arthurpro, @hhomar, @jenrik, @jpgleeson, @mark-pitblado, @maurerle, @qwerty287, @xoxys

### 🔒 Security

- Fix approval requirement if PR is closed [[#4902](https://github.com/woodpecker-ci/woodpecker/pull/4902)]

### 📚 Documentation

- chore(deps): lock file maintenance [[#4906](https://github.com/woodpecker-ci/woodpecker/pull/4906)]
- chore(deps): update dependency axios to v1.8.1 [[#4905](https://github.com/woodpecker-ci/woodpecker/pull/4905)]
- Fix typo on forgejo/gitea documentation [[#4898](https://github.com/woodpecker-ci/woodpecker/pull/4898)]
- chore(deps): update docs npm deps non-major [[#4878](https://github.com/woodpecker-ci/woodpecker/pull/4878)]
- plugins: add Hugo plugin for woodpecker [[#4870](https://github.com/woodpecker-ci/woodpecker/pull/4870)]
- Add Microsoft Teams Notification (Advanced) plugin [[#4868](https://github.com/woodpecker-ci/woodpecker/pull/4868)]
- chore(deps): update dependency @types/react to v19.0.9 [[#4864](https://github.com/woodpecker-ci/woodpecker/pull/4864)]
- Drop versioned docs for v1 [[#4844](https://github.com/woodpecker-ci/woodpecker/pull/4844)]
- Add a Home Assistant notification plugin  [[#4841](https://github.com/woodpecker-ci/woodpecker/pull/4841)]

### 🐛 Bug Fixes

- Use forge IDs for hook tokens [[#4897](https://github.com/woodpecker-ci/woodpecker/pull/4897)]
- Fix nil dereference in Bitbucket webhook handling [[#4896](https://github.com/woodpecker-ci/woodpecker/pull/4896)]
- Fix org assign on login [[#4817](https://github.com/woodpecker-ci/woodpecker/pull/4817)]
- Directly fetch directory contents [[#4842](https://github.com/woodpecker-ci/woodpecker/pull/4842)]

### 📈 Enhancement

- Remove eslint types [[#4893](https://github.com/woodpecker-ci/woodpecker/pull/4893)]
- Add default option for allowing pull requests on repositories [[#4873](https://github.com/woodpecker-ci/woodpecker/pull/4873)]
- Replace deprecated linter [[#4843](https://github.com/woodpecker-ci/woodpecker/pull/4843)]

### 📦️ Dependency

- chore(deps): update woodpeckerci/plugin-git docker tag to v2.6.2 [[#4903](https://github.com/woodpecker-ci/woodpecker/pull/4903)]
- fix(deps): update web npm deps non-major [[#4904](https://github.com/woodpecker-ci/woodpecker/pull/4904)]
- fix(deps): update golang-packages [[#4900](https://github.com/woodpecker-ci/woodpecker/pull/4900)]
- chore(deps): lock file maintenance [[#4895](https://github.com/woodpecker-ci/woodpecker/pull/4895)]
- chore(deps): update dependency vue-tsc to v2.2.4 [[#4894](https://github.com/woodpecker-ci/woodpecker/pull/4894)]
- fix(deps): update dependency simple-icons to v14.8.0 [[#4891](https://github.com/woodpecker-ci/woodpecker/pull/4891)]
- fix(deps): update golang-packages [[#4890](https://github.com/woodpecker-ci/woodpecker/pull/4890)]
- chore(deps): update dependency @types/eslint__js to v9 [[#4884](https://github.com/woodpecker-ci/woodpecker/pull/4884)]
- chore(deps): update pre-commit hook rbubley/mirrors-prettier to v3.5.2 [[#4883](https://github.com/woodpecker-ci/woodpecker/pull/4883)]
- fix(deps): update module codeberg.org/mvdkleijn/forgejo-sdk/forgejo to v2 [[#4858](https://github.com/woodpecker-ci/woodpecker/pull/4858)]
- fix(deps): update web npm deps non-major [[#4882](https://github.com/woodpecker-ci/woodpecker/pull/4882)]
- chore(deps): update postgres docker tag to v17.4 [[#4881](https://github.com/woodpecker-ci/woodpecker/pull/4881)]
- chore(deps): update woodpeckerci/plugin-git docker tag to v2.6.1 [[#4879](https://github.com/woodpecker-ci/woodpecker/pull/4879)]
- chore(deps): update docker.io/woodpeckerci/plugin-editorconfig-checker docker tag to v0.3.0 [[#4880](https://github.com/woodpecker-ci/woodpecker/pull/4880)]
- chore(deps): update docker.io/woodpeckerci/plugin-surge-preview docker tag to v1.3.5 [[#4877](https://github.com/woodpecker-ci/woodpecker/pull/4877)]
- fix(deps): update module github.com/prometheus/client_golang to v1.21.0 [[#4874](https://github.com/woodpecker-ci/woodpecker/pull/4874)]
- fix(deps): update module github.com/go-sql-driver/mysql to v1.9.0 [[#4872](https://github.com/woodpecker-ci/woodpecker/pull/4872)]
- fix(deps): update module github.com/google/go-github/v69 to v69.2.0 [[#4869](https://github.com/woodpecker-ci/woodpecker/pull/4869)]
- chore(deps): lock file maintenance [[#4866](https://github.com/woodpecker-ci/woodpecker/pull/4866)]
- chore(deps): update docker.io/woodpeckerci/plugin-trivy docker tag to v1.4.0 [[#4865](https://github.com/woodpecker-ci/woodpecker/pull/4865)]
- fix(deps): update dependency simple-icons to v14.7.0 [[#4862](https://github.com/woodpecker-ci/woodpecker/pull/4862)]
- fix(deps): update dependency pinia to v3 [[#4856](https://github.com/woodpecker-ci/woodpecker/pull/4856)]
- fix(deps): update module gitlab.com/gitlab-org/api/client-go to v0.123.0 [[#4860](https://github.com/woodpecker-ci/woodpecker/pull/4860)]
- chore(deps): update dependency vue-tsc to v2.2.2 [[#4859](https://github.com/woodpecker-ci/woodpecker/pull/4859)]
- fix(deps): update web npm deps non-major [[#4857](https://github.com/woodpecker-ci/woodpecker/pull/4857)]
- chore(deps): update pre-commit non-major [[#4855](https://github.com/woodpecker-ci/woodpecker/pull/4855)]
- chore(deps): update postgres docker tag to v17.3 [[#4854](https://github.com/woodpecker-ci/woodpecker/pull/4854)]
- chore(deps): update docker.io/techknowlogick/xgo docker tag to go-1.24.x [[#4853](https://github.com/woodpecker-ci/woodpecker/pull/4853)]
- chore(deps): update docker.io/golang docker tag to v1.24 [[#4852](https://github.com/woodpecker-ci/woodpecker/pull/4852)]
- chore(deps): update woodpeckerci/plugin-release docker tag to v0.2.4 [[#4851](https://github.com/woodpecker-ci/woodpecker/pull/4851)]
- fix(deps): update dependency @tailwindcss/vite to v4.0.6 [[#4846](https://github.com/woodpecker-ci/woodpecker/pull/4846)]
- chore(deps): lock file maintenance [[#4845](https://github.com/woodpecker-ci/woodpecker/pull/4845)]
- fix(deps): update dependency tailwindcss to v4 [[#4778](https://github.com/woodpecker-ci/woodpecker/pull/4778)]
- fix(deps): update golang-packages [[#4839](https://github.com/woodpecker-ci/woodpecker/pull/4839)]

### Misc

- kubernetes: create service for detached steps [[#4892](https://github.com/woodpecker-ci/woodpecker/pull/4892)]
- docs: remove latest from docker compose example [[#4849](https://github.com/woodpecker-ci/woodpecker/pull/4849)]

## [3.1.0](https://github.com/woodpecker-ci/woodpecker/releases/tag/v3.1.0) - 2025-02-12

### ❤️ Thanks to all contributors! ❤️

@Levy-Tal, @anbraten, @cduchenoy, @damuzhi0810, @lafriks, @mzampetakis, @pat-s, @qwerty287, @xoxys

### ✨ Features

- Add allow list for approvals [[#4768](https://github.com/woodpecker-ci/woodpecker/pull/4768)]

### 🐛 Bug Fixes

- Unsanitize user and org names in DB [[#4762](https://github.com/woodpecker-ci/woodpecker/pull/4762)]
- Store/delete repos after forge communication [[#4827](https://github.com/woodpecker-ci/woodpecker/pull/4827)]
- Fix k8s secret schema [[#4819](https://github.com/woodpecker-ci/woodpecker/pull/4819)]
- Move section description to the top [[#4773](https://github.com/woodpecker-ci/woodpecker/pull/4773)]

### 📚 Documentation

- Docs: Add Radicle forge addon [[#4833](https://github.com/woodpecker-ci/woodpecker/pull/4833)]
- fix(deps): update docs npm deps non-major [[#4823](https://github.com/woodpecker-ci/woodpecker/pull/4823)]
- chore(deps): update dependency isomorphic-dompurify to v2.21.0 [[#4805](https://github.com/woodpecker-ci/woodpecker/pull/4805)]
- chore(deps): update dependency @types/node to v22.13.0 [[#4799](https://github.com/woodpecker-ci/woodpecker/pull/4799)]
- Add bluesky post plugin [[#4549](https://github.com/woodpecker-ci/woodpecker/pull/4549)]
- Various docs improvements [[#4772](https://github.com/woodpecker-ci/woodpecker/pull/4772)]
- fix(deps): update docs npm deps non-major [[#4774](https://github.com/woodpecker-ci/woodpecker/pull/4774)]
- Add git basic changelog [[#4755](https://github.com/woodpecker-ci/woodpecker/pull/4755)]

### 📈 Enhancement

- Optimize repository list loading to return also latest pipeline info [[#4814](https://github.com/woodpecker-ci/woodpecker/pull/4814)]
- Add Git Ref To Build Status in BitbucketDatacenter [[#4724](https://github.com/woodpecker-ci/woodpecker/pull/4724)]

### 📦️ Dependency

- fix(deps): update golang-packages [[#4834](https://github.com/woodpecker-ci/woodpecker/pull/4834)]
- fix(deps): update web npm deps non-major [[#4831](https://github.com/woodpecker-ci/woodpecker/pull/4831)]
- fix(deps): update dependency simple-icons to v14.6.0 [[#4830](https://github.com/woodpecker-ci/woodpecker/pull/4830)]
- fix(deps): update golang-packages [[#4829](https://github.com/woodpecker-ci/woodpecker/pull/4829)]
- fix(deps): update web npm deps non-major to v4.0.5 [[#4828](https://github.com/woodpecker-ci/woodpecker/pull/4828)]
- chore(deps): update docker.io/woodpeckerci/plugin-docker-buildx docker tag to v5.2.1 [[#4822](https://github.com/woodpecker-ci/woodpecker/pull/4822)]
- fix(deps): update module github.com/google/go-github/v68 to v69 [[#4826](https://github.com/woodpecker-ci/woodpecker/pull/4826)]
- fix(deps): update web npm deps non-major [[#4825](https://github.com/woodpecker-ci/woodpecker/pull/4825)]
- fix(deps): update golang-packages [[#4812](https://github.com/woodpecker-ci/woodpecker/pull/4812)]
- chore(deps): update dependency vitest to v3.0.5 [security] [[#4810](https://github.com/woodpecker-ci/woodpecker/pull/4810)]
- chore(deps): lock file maintenance [[#4808](https://github.com/woodpecker-ci/woodpecker/pull/4808)]
- chore(deps): update dependency @antfu/eslint-config to v4.1.1 [[#4806](https://github.com/woodpecker-ci/woodpecker/pull/4806)]
- fix(deps): update module gitlab.com/gitlab-org/api/client-go to v0.121.0 [[#4804](https://github.com/woodpecker-ci/woodpecker/pull/4804)]
- fix(deps): update dependency simple-icons to v14.5.0 [[#4803](https://github.com/woodpecker-ci/woodpecker/pull/4803)]
- fix(deps): update web npm deps non-major to v4.0.3 [[#4802](https://github.com/woodpecker-ci/woodpecker/pull/4802)]
- fix(deps): update web npm deps non-major [[#4798](https://github.com/woodpecker-ci/woodpecker/pull/4798)]
- fix(deps): update module github.com/getkin/kin-openapi to v0.129.0 [[#4790](https://github.com/woodpecker-ci/woodpecker/pull/4790)]
- chore(deps): lock file maintenance [[#4783](https://github.com/woodpecker-ci/woodpecker/pull/4783)]
- chore(deps): update dependency @antfu/eslint-config to v4.1.0 [[#4780](https://github.com/woodpecker-ci/woodpecker/pull/4780)]
- fix(deps): update module github.com/bmatcuk/doublestar/v4 to v4.8.1 [[#4781](https://github.com/woodpecker-ci/woodpecker/pull/4781)]
- chore(deps): update dependency @antfu/eslint-config to v4 [[#4779](https://github.com/woodpecker-ci/woodpecker/pull/4779)]
- fix(deps): update web npm deps non-major [[#4777](https://github.com/woodpecker-ci/woodpecker/pull/4777)]
- chore(deps): update pre-commit hook igorshubovych/markdownlint-cli to v0.44.0 [[#4776](https://github.com/woodpecker-ci/woodpecker/pull/4776)]
- fix(deps): update module google.golang.org/protobuf to v1.36.4 [[#4775](https://github.com/woodpecker-ci/woodpecker/pull/4775)]
- fix(deps): update module google.golang.org/grpc to v1.70.0 [[#4770](https://github.com/woodpecker-ci/woodpecker/pull/4770)]
- chore(deps): update docker.io/woodpeckerci/plugin-docker-buildx docker tag to v5.2.0 [[#4767](https://github.com/woodpecker-ci/woodpecker/pull/4767)]
- chore(deps): update docker.io/mysql docker tag to v9.2.0 [[#4766](https://github.com/woodpecker-ci/woodpecker/pull/4766)]
- fix(deps): update module github.com/hashicorp/go-plugin to v1.6.3 [[#4765](https://github.com/woodpecker-ci/woodpecker/pull/4765)]
- chore(deps): update docker.io/woodpeckerci/plugin-ready-release-go docker tag to v3.1.3 [[#4764](https://github.com/woodpecker-ci/woodpecker/pull/4764)]
- fix(deps): update docker to v27.5.1+incompatible [[#4761](https://github.com/woodpecker-ci/woodpecker/pull/4761)]
- chore(deps): update dependency vite to v6.0.9 [security] [[#4757](https://github.com/woodpecker-ci/woodpecker/pull/4757)]

### Misc

- chore: fix some function names in comment [[#4769](https://github.com/woodpecker-ci/woodpecker/pull/4769)]

## [3.0.1](https://github.com/woodpecker-ci/woodpecker/releases/tag/v3.0.1) - 2025-01-20

### ❤️ Thanks to all contributors! ❤️

@pat-s, @qwerty287, @xoxys

### 🐛 Bug Fixes

- Only show visited repos and hide at all if less than 4 repos [[#4753](https://github.com/woodpecker-ci/woodpecker/pull/4753)]
- Fix sql identifier escaping in datastore feed [[#4746](https://github.com/woodpecker-ci/woodpecker/pull/4746)]
- Fix log folder permissions [[#4749](https://github.com/woodpecker-ci/woodpecker/pull/4749)]
- Add missing error message for org_access_denied [[#4744](https://github.com/woodpecker-ci/woodpecker/pull/4744)]
- Fix package configs [[#4741](https://github.com/woodpecker-ci/woodpecker/pull/4741)]

### 📚 Documentation

- chore(deps): lock file maintenance [[#4751](https://github.com/woodpecker-ci/woodpecker/pull/4751)]

### 📦️ Dependency

- fix(deps): update golang-packages [[#4750](https://github.com/woodpecker-ci/woodpecker/pull/4750)]
- fix(deps): update dependency simple-icons to v14.3.0 [[#4739](https://github.com/woodpecker-ci/woodpecker/pull/4739)]
- chore(deps): update dependency vitest to v3 [[#4736](https://github.com/woodpecker-ci/woodpecker/pull/4736)]

### Misc

- fix minor tag creation for server scratch image [[#4748](https://github.com/woodpecker-ci/woodpecker/pull/4748)]
- use v3 woodpecker libs [[#4742](https://github.com/woodpecker-ci/woodpecker/pull/4742)]

## [3.0.0](https://github.com/woodpecker-ci/woodpecker/releases/tag/v3.0.0) - 2025-01-18

### ❤️ Thanks to all contributors! ❤️

@6543, @Fishbowler, @Levy-Tal, @M0Rf30, @anbraten, @cduchenoy, @cevatkerim, @fernandrone, @gedankenstuecke, @gnowland, @greenaar, @hg, @j04n-f, @jenrik, @johanneskastl, @jolheiser, @lafriks, @lukashass, @meln5674, @not-my-profile, @pat-s, @plafue, @qwerty287, @smainz, @stevapple, @tori-27, @tsufeki, @xoxys, @xtexChooser, @zc-devs

### 💥 Breaking changes

- Add rootless (alpine) images [[#4617](https://github.com/woodpecker-ci/woodpecker/pull/4617)]
- Unify CLI bin name [[#4673](https://github.com/woodpecker-ci/woodpecker/pull/4673)]
- Support Git as only VCS [[#4346](https://github.com/woodpecker-ci/woodpecker/pull/4346)]
- Add rolling semver tags, remove `latest` tag [[#4600](https://github.com/woodpecker-ci/woodpecker/pull/4600)]
- Drop native Let's Encrypt support [[#4541](https://github.com/woodpecker-ci/woodpecker/pull/4541)]
- Require approval for prs from public repos by default [[#4456](https://github.com/woodpecker-ci/woodpecker/pull/4456)]
- Do not set empty environment variables [[#4193](https://github.com/woodpecker-ci/woodpecker/pull/4193)]
- Unify cli commands and flags [[#4481](https://github.com/woodpecker-ci/woodpecker/pull/4481)]
- Move pipeline logs command [[#4480](https://github.com/woodpecker-ci/woodpecker/pull/4480)]
- Fix woodpecker-go repo model to match server [[#4479](https://github.com/woodpecker-ci/woodpecker/pull/4479)]
- Restructure cli commands [[#4467](https://github.com/woodpecker-ci/woodpecker/pull/4467)]
- Add pagination options to all supported endpoints in sdk [[#4463](https://github.com/woodpecker-ci/woodpecker/pull/4463)]
- Allow to set custom trusted clone plugins [[#4352](https://github.com/woodpecker-ci/woodpecker/pull/4352)]
- Add PipelineListsOptions to woodpecker-go [[#3652](https://github.com/woodpecker-ci/woodpecker/pull/3652)]
- Remove `secrets` in favor of `from_secret` [[#4363](https://github.com/woodpecker-ci/woodpecker/pull/4363)]
- Kubernetes | Docker: Add support for rootless images [[#4151](https://github.com/woodpecker-ci/woodpecker/pull/4151)]
- Split repo trusted setting [[#4025](https://github.com/woodpecker-ci/woodpecker/pull/4025)]
- Move docker resource limit settings from server to agent [[#3174](https://github.com/woodpecker-ci/woodpecker/pull/3174)]
- Set `/woodpecker` as default workdir for the **woodpecker-cli** container [[#4130](https://github.com/woodpecker-ci/woodpecker/pull/4130)]
- Require upgrade from 2.x [[#4112](https://github.com/woodpecker-ci/woodpecker/pull/4112)]
- Don't expose task data via api [[#4108](https://github.com/woodpecker-ci/woodpecker/pull/4108)]
- Remove some ci environment variables [[#3846](https://github.com/woodpecker-ci/woodpecker/pull/3846)]
- Remove all default privileged plugins [[#4053](https://github.com/woodpecker-ci/woodpecker/pull/4053)]
- Add option to filter secrets by plugins with specific tags [[#4069](https://github.com/woodpecker-ci/woodpecker/pull/4069)]
- Remove old pipeline options [[#4016](https://github.com/woodpecker-ci/woodpecker/pull/4016)]
- Remove various deprecations [[#4017](https://github.com/woodpecker-ci/woodpecker/pull/4017)]
- Drop repo name fallback for hooks [[#4013](https://github.com/woodpecker-ci/woodpecker/pull/4013)]
- Improve local backend detection [[#4006](https://github.com/woodpecker-ci/woodpecker/pull/4006)]
- Refactor JSON and SDK fields [[#3968](https://github.com/woodpecker-ci/woodpecker/pull/3968)]
- Migrate to maintained cron lib and remove seconds [[#3785](https://github.com/woodpecker-ci/woodpecker/pull/3785)]
- Switch to profile-based AppArmor configuration [[#4008](https://github.com/woodpecker-ci/woodpecker/pull/4008)]
- Remove Kubernetes default image pull secret name `regcred` [[#4005](https://github.com/woodpecker-ci/woodpecker/pull/4005)]
- Drop "WOODPECKER_WEBHOOK_HOST" env var and adjust docs [[#3969](https://github.com/woodpecker-ci/woodpecker/pull/3969)]
- Drop version in schema [[#3970](https://github.com/woodpecker-ci/woodpecker/pull/3970)]
- Update docker to v27 [[#3972](https://github.com/woodpecker-ci/woodpecker/pull/3972)]
- Require gitlab 12.4 [[#3966](https://github.com/woodpecker-ci/woodpecker/pull/3966)]
- Migrate to maintained httpsign library [[#3839](https://github.com/woodpecker-ci/woodpecker/pull/3839)]
- Remove `WOODPECKER_DEV_OAUTH_HOST` and `WOODPECKER_DEV_GITEA_OAUTH_URL` [[#3961](https://github.com/woodpecker-ci/woodpecker/pull/3961)]
- Remove deprecated pipeline keywords: `pipeline:`, `platform:`, `branches:` [[#3916](https://github.com/woodpecker-ci/woodpecker/pull/3916)]
- server: remove old unused routes [[#3845](https://github.com/woodpecker-ci/woodpecker/pull/3845)]
- CLI: remove step-id and add step-number as option to logs [[#3927](https://github.com/woodpecker-ci/woodpecker/pull/3927)]

### 🔒 Security

- Don't log DB passwords [[#4583](https://github.com/woodpecker-ci/woodpecker/pull/4583)]
- Do not log forge tokens [[#4551](https://github.com/woodpecker-ci/woodpecker/pull/4551)]
- Add server config to disable user registered agents [[#4206](https://github.com/woodpecker-ci/woodpecker/pull/4206)]
- chore: fix `http-proxy-middleware` CVE [[#4257](https://github.com/woodpecker-ci/woodpecker/pull/4257)]
- Allow altering trusted clone plugins and filter them via tag [[#4074](https://github.com/woodpecker-ci/woodpecker/pull/4074)]
- Update gitea sdk [[#4012](https://github.com/woodpecker-ci/woodpecker/pull/4012)]
- Update Forgejo SDK [[#3948](https://github.com/woodpecker-ci/woodpecker/pull/3948)]

### ✨ Features

- Add user as docker backend_option [[#4526](https://github.com/woodpecker-ci/woodpecker/pull/4526)]
- Add dns config option to official feature set [[#4418](https://github.com/woodpecker-ci/woodpecker/pull/4418)]
- Implement org/user agents [[#3539](https://github.com/woodpecker-ci/woodpecker/pull/3539)]
- Replay pipeline using `cli exec` by downloading metadata [[#4103](https://github.com/woodpecker-ci/woodpecker/pull/4103)]
- Update clone plugin to support sha256 [[#4136](https://github.com/woodpecker-ci/woodpecker/pull/4136)]

### 📚 Documentation

- Improve 3.0.0 migration notes [[#4737](https://github.com/woodpecker-ci/woodpecker/pull/4737)]
- Add docs for 3.0 [[#4705](https://github.com/woodpecker-ci/woodpecker/pull/4705)]
- fix(deps): update docs npm deps non-major [[#4733](https://github.com/woodpecker-ci/woodpecker/pull/4733)]
- chore(deps): update dependency @types/react to v19.0.5 [[#4714](https://github.com/woodpecker-ci/woodpecker/pull/4714)]
- fix(deps): update docs npm deps non-major [[#4702](https://github.com/woodpecker-ci/woodpecker/pull/4702)]
- fix(deps): update react monorepo to v19 (major) [[#4529](https://github.com/woodpecker-ci/woodpecker/pull/4529)]
- Refactor `secrets` page in docs [[#4644](https://github.com/woodpecker-ci/woodpecker/pull/4644)]
- fix(deps): update docs npm deps non-major [[#4661](https://github.com/woodpecker-ci/woodpecker/pull/4661)]
- chore(deps): lock file maintenance [[#4647](https://github.com/woodpecker-ci/woodpecker/pull/4647)]
- chore(deps): update dependency concurrently to v9.1.1 [[#4631](https://github.com/woodpecker-ci/woodpecker/pull/4631)]
- Add docker in docker example to advanced usage in docs [[#4620](https://github.com/woodpecker-ci/woodpecker/pull/4620)]
- fixed a typo [[#4621](https://github.com/woodpecker-ci/woodpecker/pull/4621)]
- Fix misleading example in Workflow syntax/Privileged mode [[#4613](https://github.com/woodpecker-ci/woodpecker/pull/4613)]
- Update docs section about "Custom clone plugins" [[#4618](https://github.com/woodpecker-ci/woodpecker/pull/4618)]
- Search in plugin tags [[#4604](https://github.com/woodpecker-ci/woodpecker/pull/4604)]
- chore(deps): update dependency @types/react to v18.3.18 [[#4599](https://github.com/woodpecker-ci/woodpecker/pull/4599)]
- Update About [[#4555](https://github.com/woodpecker-ci/woodpecker/pull/4555)]
- chore(deps): update dependency marked to v15.0.4 [[#4570](https://github.com/woodpecker-ci/woodpecker/pull/4570)]
- Expand docs around the deprecation of former secret syntax [[#4561](https://github.com/woodpecker-ci/woodpecker/pull/4561)]
- fix(deps): update docs npm deps non-major [[#4568](https://github.com/woodpecker-ci/woodpecker/pull/4568)]
- Show client flags [[#4542](https://github.com/woodpecker-ci/woodpecker/pull/4542)]
- chore(deps): update react monorepo to v19 (major) [[#4523](https://github.com/woodpecker-ci/woodpecker/pull/4523)]
- chore(deps): update docs npm deps non-major [[#4519](https://github.com/woodpecker-ci/woodpecker/pull/4519)]
- chore(deps): update dependency isomorphic-dompurify to v2.18.0 [[#4493](https://github.com/woodpecker-ci/woodpecker/pull/4493)]
- fix(deps): update docs npm deps non-major [[#4484](https://github.com/woodpecker-ci/woodpecker/pull/4484)]
- Add migration notes for restructured cli commands [[#4476](https://github.com/woodpecker-ci/woodpecker/pull/4476)]
- Various fixes for `awesome.md` [[#4448](https://github.com/woodpecker-ci/woodpecker/pull/4448)]
- chore(deps): update dependency isomorphic-dompurify to v2.17.0 [[#4449](https://github.com/woodpecker-ci/woodpecker/pull/4449)]
- fix(deps): update docs npm deps non-major [[#4441](https://github.com/woodpecker-ci/woodpecker/pull/4441)]
- chore(deps): update dependency @docusaurus/tsconfig to v3.6.2 [[#4433](https://github.com/woodpecker-ci/woodpecker/pull/4433)]
- Bump minimum nodejs to v20 [[#4417](https://github.com/woodpecker-ci/woodpecker/pull/4417)]
- Add microsoft teams plugin [[#4400](https://github.com/woodpecker-ci/woodpecker/pull/4400)]
- fix(deps): update docs npm deps non-major [[#4394](https://github.com/woodpecker-ci/woodpecker/pull/4394)]
- chore(deps): update dependency @types/node to v22 [[#4395](https://github.com/woodpecker-ci/woodpecker/pull/4395)]
- chore(deps): update dependency marked to v15 [[#4396](https://github.com/woodpecker-ci/woodpecker/pull/4396)]
- Kubernetes documentation enhancements [[#4374](https://github.com/woodpecker-ci/woodpecker/pull/4374)]
- Podman is not (official) supported [[#4367](https://github.com/woodpecker-ci/woodpecker/pull/4367)]
- Add EditorConfig-Checker Plugin to docs [[#4371](https://github.com/woodpecker-ci/woodpecker/pull/4371)]
- Update netrc option description [[#4342](https://github.com/woodpecker-ci/woodpecker/pull/4342)]
- Fix deployment event note [[#4283](https://github.com/woodpecker-ci/woodpecker/pull/4283)]
- Improve migration notes [[#4291](https://github.com/woodpecker-ci/woodpecker/pull/4291)]
- Add instructions how to build images locally [[#4277](https://github.com/woodpecker-ci/woodpecker/pull/4277)]
- chore(deps): update docs npm deps non-major [[#4238](https://github.com/woodpecker-ci/woodpecker/pull/4238)]
- Correct spelling [[#4279](https://github.com/woodpecker-ci/woodpecker/pull/4279)]
- Add Telegram plugin [[#4229](https://github.com/woodpecker-ci/woodpecker/pull/4229)]
- Remove archived plugin [[#4227](https://github.com/woodpecker-ci/woodpecker/pull/4227)]
- Use "Woodpecker Authors" as copyright on website [[#4225](https://github.com/woodpecker-ci/woodpecker/pull/4225)]
- chore(deps): update dependency cookie to v1 [[#4224](https://github.com/woodpecker-ci/woodpecker/pull/4224)]
- fix(deps): update docs npm deps non-major [[#4221](https://github.com/woodpecker-ci/woodpecker/pull/4221)]
- Fix errant apostrophe in docker-compose documentation [[#4203](https://github.com/woodpecker-ci/woodpecker/pull/4203)]
- chore(deps): update dependency concurrently to v9 [[#4176](https://github.com/woodpecker-ci/woodpecker/pull/4176)]
- chore(deps): update docs npm deps non-major [[#4164](https://github.com/woodpecker-ci/woodpecker/pull/4164)]
- Update image filter error message [[#4143](https://github.com/woodpecker-ci/woodpecker/pull/4143)]
- Docs: reference to built-in docker compose and remove deprecated version from compose examples [[#4123](https://github.com/woodpecker-ci/woodpecker/pull/4123)]
- directory key is allowed for services [[#4127](https://github.com/woodpecker-ci/woodpecker/pull/4127)]
- [docs] Removes dot prefix from pipeline configuration filenames [[#4105](https://github.com/woodpecker-ci/woodpecker/pull/4105)]
- Use kaniko plugin in docs as example [[#4072](https://github.com/woodpecker-ci/woodpecker/pull/4072)]
- Add some posts and videos [[#4070](https://github.com/woodpecker-ci/woodpecker/pull/4070)]
- Move event type descriptions from Terminology to Workflow Syntax [[#4062](https://github.com/woodpecker-ci/woodpecker/pull/4062)]
- Add community posts from discussions [[#4058](https://github.com/woodpecker-ci/woodpecker/pull/4058)]
- Add a pull request template with some basic guidelines [[#4055](https://github.com/woodpecker-ci/woodpecker/pull/4055)]
- Add examples of CI environment variable values [[#4009](https://github.com/woodpecker-ci/woodpecker/pull/4009)]
- Fix inline author warning [[#4040](https://github.com/woodpecker-ci/woodpecker/pull/4040)]
- Updated Secrets image filter docs [[#4028](https://github.com/woodpecker-ci/woodpecker/pull/4028)]
- Update dependency marked to v14 [[#4036](https://github.com/woodpecker-ci/woodpecker/pull/4036)]
- Update docs npm deps non-major [[#4033](https://github.com/woodpecker-ci/woodpecker/pull/4033)]
- Overhaul README [[#3995](https://github.com/woodpecker-ci/woodpecker/pull/3995)]
- fix(deps): update docs npm deps non-major [[#3990](https://github.com/woodpecker-ci/woodpecker/pull/3990)]
- Add spellchecking for docs [[#3787](https://github.com/woodpecker-ci/woodpecker/pull/3787)]

### 🐛 Bug Fixes

- Check organization first [[#4723](https://github.com/woodpecker-ci/woodpecker/pull/4723)]
- Fix mobile view of the popup [[#4717](https://github.com/woodpecker-ci/woodpecker/pull/4717)]
- Apply changed files filter to closed PR [[#4711](https://github.com/woodpecker-ci/woodpecker/pull/4711)]
- Add margins to moving WP svg logo [[#4697](https://github.com/woodpecker-ci/woodpecker/pull/4697)]
- Add hosts for detached steps [[#4674](https://github.com/woodpecker-ci/woodpecker/pull/4674)]
- Fix addon `nil` values [[#4666](https://github.com/woodpecker-ci/woodpecker/pull/4666)]
- fix cli exec statement in debug tab [[#4643](https://github.com/woodpecker-ci/woodpecker/pull/4643)]
- Fix misaligned step list indentation [[#4609](https://github.com/woodpecker-ci/woodpecker/pull/4609)]
- Ignore blocked pipelines for badge rendering [[#4582](https://github.com/woodpecker-ci/woodpecker/pull/4582)]
- Remove related pipeline logs during pipeline deletion [[#4572](https://github.com/woodpecker-ci/woodpecker/pull/4572)]
- Weakly decode backend options [[#4577](https://github.com/woodpecker-ci/woodpecker/pull/4577)]
- Add client error to sdk and fix purge cli [[#4574](https://github.com/woodpecker-ci/woodpecker/pull/4574)]
- Fix pipeline purge cli command [[#4569](https://github.com/woodpecker-ci/woodpecker/pull/4569)]
- Fix BB ambiguous commit status key [[#4544](https://github.com/woodpecker-ci/woodpecker/pull/4544)]
- fix: addon JSON pointers [[#4508](https://github.com/woodpecker-ci/woodpecker/pull/4508)]
- Fix apparmorProfile being ignored when it's the only field [[#4507](https://github.com/woodpecker-ci/woodpecker/pull/4507)]
- Sanitize strings in table output [[#4466](https://github.com/woodpecker-ci/woodpecker/pull/4466)]
- Cleanup openapi generation [[#4331](https://github.com/woodpecker-ci/woodpecker/pull/4331)]
- Support github refresh tokens [[#3811](https://github.com/woodpecker-ci/woodpecker/pull/3811)]
- Fix not working overflow on repo list message [[#4420](https://github.com/woodpecker-ci/woodpecker/pull/4420)]
- fix `error="io: read/write on closed pipe"` on k8s backend [[#4281](https://github.com/woodpecker-ci/woodpecker/pull/4281)]
- Move update notifier dot into settings button [[#4334](https://github.com/woodpecker-ci/woodpecker/pull/4334)]
- gitea: add check if pull_request webhook is missing pull info [[#4305](https://github.com/woodpecker-ci/woodpecker/pull/4305)]
- Refresh token before loading branches [[#4284](https://github.com/woodpecker-ci/woodpecker/pull/4284)]
- Delete GitLab webhooks with partial URL match [[#4259](https://github.com/woodpecker-ci/woodpecker/pull/4259)]
- Increase `WOODPECKER_FORGE_TIMEOUT` to fix config fetching for GitLab [[#4262](https://github.com/woodpecker-ci/woodpecker/pull/4262)]
- Ensure cli exec has by default not the same prefix [[#4132](https://github.com/woodpecker-ci/woodpecker/pull/4132)]
- Fix repo add loading spinner [[#4135](https://github.com/woodpecker-ci/woodpecker/pull/4135)]
- Fix migration registries table [[#4111](https://github.com/woodpecker-ci/woodpecker/pull/4111)]
- Wait for tracer to be done before finishing workflow [[#4068](https://github.com/woodpecker-ci/woodpecker/pull/4068)]
- Fix schema with detached steps [[#4066](https://github.com/woodpecker-ci/woodpecker/pull/4066)]
- Fix schema with commands and entrypoint [[#4065](https://github.com/woodpecker-ci/woodpecker/pull/4065)]
- Read long log lines from file storage correctly [[#4048](https://github.com/woodpecker-ci/woodpecker/pull/4048)]
- Set refspec for gitlab MR [[#4021](https://github.com/woodpecker-ci/woodpecker/pull/4021)]
- Set `CI_PREV_COMMIT_{SOURCE,TARGET}_BRANCH` as mentioned in the documentation [[#4001](https://github.com/woodpecker-ci/woodpecker/pull/4001)]
- [Bitbucket Datacenter] Return empty list instead of null [[#4010](https://github.com/woodpecker-ci/woodpecker/pull/4010)]
- Fix BB PR pipeline ref [[#3985](https://github.com/woodpecker-ci/woodpecker/pull/3985)]
- Change Bitbucket PR hook to point the source branch, commit & ref [[#3965](https://github.com/woodpecker-ci/woodpecker/pull/3965)]
- Add updated, merged and declined events to bb webhook activation [[#3963](https://github.com/woodpecker-ci/woodpecker/pull/3963)]
- Fix login via navbar [[#3962](https://github.com/woodpecker-ci/woodpecker/pull/3962)]
- Truncate creation in list [[#3952](https://github.com/woodpecker-ci/woodpecker/pull/3952)]
- Fix panic if forge is unreachable [[#3944](https://github.com/woodpecker-ci/woodpecker/pull/3944)]

### 📈 Enhancement

- Harmonize en texts [[#4716](https://github.com/woodpecker-ci/woodpecker/pull/4716)]
- feat: add linter support for step-level `depends_on` existence [[#4657](https://github.com/woodpecker-ci/woodpecker/pull/4657)]
- Reduce version redundancy [[#4707](https://github.com/woodpecker-ci/woodpecker/pull/4707)]
- Add priority menu to tabs [[#4641](https://github.com/woodpecker-ci/woodpecker/pull/4641)]
- feat(bitbucketdatacenter): Add support for fetching and converting projects to teams [[#4663](https://github.com/woodpecker-ci/woodpecker/pull/4663)]
- Migrate from Windi to Tailwind [[#4614](https://github.com/woodpecker-ci/woodpecker/pull/4614)]
- Do not start metrics collector if metrics are disabled [[#4667](https://github.com/woodpecker-ci/woodpecker/pull/4667)]
- Improve badge coloring [[#4447](https://github.com/woodpecker-ci/woodpecker/pull/4447)]
- Inline web helpers [[#4639](https://github.com/woodpecker-ci/woodpecker/pull/4639)]
- Use filled status icons and harmonize contextually [[#4584](https://github.com/woodpecker-ci/woodpecker/pull/4584)]
- Two row layout for title and context of pipeline list [[#4625](https://github.com/woodpecker-ci/woodpecker/pull/4625)]
- Remove workflow-level volumes and networks [[#4636](https://github.com/woodpecker-ci/woodpecker/pull/4636)]
- Migrate away from goblin [[#4624](https://github.com/woodpecker-ci/woodpecker/pull/4624)]
- Use lighter red shades for error messages [[#4611](https://github.com/woodpecker-ci/woodpecker/pull/4611)]
- Avoid usage of inline css style [[#4629](https://github.com/woodpecker-ci/woodpecker/pull/4629)]
- Use icon sizes relative to font size [[#4575](https://github.com/woodpecker-ci/woodpecker/pull/4575)]
- Use docusaurus faster [[#4528](https://github.com/woodpecker-ci/woodpecker/pull/4528)]
- Add settings title action [[#4499](https://github.com/woodpecker-ci/woodpecker/pull/4499)]
- Use pagination helper to list pipelines in cli [[#4478](https://github.com/woodpecker-ci/woodpecker/pull/4478)]
- Some UI improvements [[#4497](https://github.com/woodpecker-ci/woodpecker/pull/4497)]
- Add status filter to list pipeline API [[#4494](https://github.com/woodpecker-ci/woodpecker/pull/4494)]
- Use JS-native date/time formatting [[#4488](https://github.com/woodpecker-ci/woodpecker/pull/4488)]
- Add pipeline purge command to cli [[#4470](https://github.com/woodpecker-ci/woodpecker/pull/4470)]
- Add option to limit the resultset returned by paginate helper [[#4475](https://github.com/woodpecker-ci/woodpecker/pull/4475)]
- Add filter to list repository pipelines API [[#4416](https://github.com/woodpecker-ci/woodpecker/pull/4416)]
- Increase log level when failing to fetch YAML [[#4107](https://github.com/woodpecker-ci/woodpecker/pull/4107)]
- Trim space to all config flags that allow to read value from file [[#4468](https://github.com/woodpecker-ci/woodpecker/pull/4468)]
- Change default icon size to 20 [[#4458](https://github.com/woodpecker-ci/woodpecker/pull/4458)]
- Use same default sort for repo and org repo list [[#4461](https://github.com/woodpecker-ci/woodpecker/pull/4461)]
- Add visibility icon to repo list [[#4460](https://github.com/woodpecker-ci/woodpecker/pull/4460)]
- Improve tab layout and add hover effect [[#4431](https://github.com/woodpecker-ci/woodpecker/pull/4431)]
- Unify pipeline status icons [[#4414](https://github.com/woodpecker-ci/woodpecker/pull/4414)]
- Improve project settings descriptions [[#4410](https://github.com/woodpecker-ci/woodpecker/pull/4410)]
- Add count badge to visualize counters in tab title [[#4419](https://github.com/woodpecker-ci/woodpecker/pull/4419)]
- Redesign repo list and include last pipeline [[#4386](https://github.com/woodpecker-ci/woodpecker/pull/4386)]
- Use KeyValueEditor for DeployPipelinePopup too [[#4412](https://github.com/woodpecker-ci/woodpecker/pull/4412)]
- Use separate routes instead of anchors [[#4285](https://github.com/woodpecker-ci/woodpecker/pull/4285)]
- Untangle settings / header slots [[#4403](https://github.com/woodpecker-ci/woodpecker/pull/4403)]
- Fix responsiveness of the settings template [[#4383](https://github.com/woodpecker-ci/woodpecker/pull/4383)]
- Use squared spinner for active pipelines [[#4379](https://github.com/woodpecker-ci/woodpecker/pull/4379)]
- Add server configuration option to add default set of labels for workflows that has no labels specified [[#4326](https://github.com/woodpecker-ci/woodpecker/pull/4326)]
- Add `cli lint` option to treat warnings as errors [[#4373](https://github.com/woodpecker-ci/woodpecker/pull/4373)]
- Improve error message for wrong secrets / environment config [[#4359](https://github.com/woodpecker-ci/woodpecker/pull/4359)]
- Improve linter messages in UI [[#4351](https://github.com/woodpecker-ci/woodpecker/pull/4351)]
- Pass settings to services [[#4338](https://github.com/woodpecker-ci/woodpecker/pull/4338)]
- Inline model types for migrations [[#4293](https://github.com/woodpecker-ci/woodpecker/pull/4293)]
- Add options to control the database connections (open,idle,timeout) [[#4212](https://github.com/woodpecker-ci/woodpecker/pull/4212)]
- Move Queue creation behind new func that evaluates queue type [[#4252](https://github.com/woodpecker-ci/woodpecker/pull/4252)]
- Add additional error message on swagger v2 to v3 convert [[#4254](https://github.com/woodpecker-ci/woodpecker/pull/4254)]
- Fix wording for privileged plugins linter error [[#4280](https://github.com/woodpecker-ci/woodpecker/pull/4280)]
- Deprecate `secrets` [[#4235](https://github.com/woodpecker-ci/woodpecker/pull/4235)]
- Agent edit/detail view: change the help url based on the backend [[#4219](https://github.com/woodpecker-ci/woodpecker/pull/4219)]
- Use middleware to load org [[#4208](https://github.com/woodpecker-ci/woodpecker/pull/4208)]
- Assign workflows to agents with the best label matches [[#4201](https://github.com/woodpecker-ci/woodpecker/pull/4201)]
- Report custom labels set by agent admins back [[#4141](https://github.com/woodpecker-ci/woodpecker/pull/4141)]
- Highlight invalid entries in manual pipeline trigger [[#4153](https://github.com/woodpecker-ci/woodpecker/pull/4153)]
- Print agent labels in debug mode [[#4155](https://github.com/woodpecker-ci/woodpecker/pull/4155)]
- Implement registries for Kubernetes backend [[#4092](https://github.com/woodpecker-ci/woodpecker/pull/4092)]
- Correct cli exec flags and remove ineffective ones [[#4129](https://github.com/woodpecker-ci/woodpecker/pull/4129)]
- Set repo user to repairing user when old user is missing [[#4128](https://github.com/woodpecker-ci/woodpecker/pull/4128)]
- Restart tasks on dead agents sooner [[#4114](https://github.com/woodpecker-ci/woodpecker/pull/4114)]
- Adjust cli exec metadata structure to equal server metadata [[#4119](https://github.com/woodpecker-ci/woodpecker/pull/4119)]
- Allow to restart declined pipelines [[#4109](https://github.com/woodpecker-ci/woodpecker/pull/4109)]
- Add indices to repo table [[#4087](https://github.com/woodpecker-ci/woodpecker/pull/4087)]
- Add systemd unit files to the RPM/DEB packages [[#3986](https://github.com/woodpecker-ci/woodpecker/pull/3986)]
- Duplicate key `workflow_id` in the agent logs [[#4046](https://github.com/woodpecker-ci/woodpecker/pull/4046)]
- Improve error on config loading [[#4024](https://github.com/woodpecker-ci/woodpecker/pull/4024)]
- Show error if secret name is missing [[#4014](https://github.com/woodpecker-ci/woodpecker/pull/4014)]
- Show error returned from API [[#3980](https://github.com/woodpecker-ci/woodpecker/pull/3980)]
- Move manual popup to own page [[#3981](https://github.com/woodpecker-ci/woodpecker/pull/3981)]
- Fail on InvalidImageName [[#4007](https://github.com/woodpecker-ci/woodpecker/pull/4007)]
- Use Bitbucket PR title for pipeline message [[#3984](https://github.com/woodpecker-ci/woodpecker/pull/3984)]
- Show logs if step has error [[#3979](https://github.com/woodpecker-ci/woodpecker/pull/3979)]
- Refactor docker backend and add more test coverage [[#2700](https://github.com/woodpecker-ci/woodpecker/pull/2700)]
- Make cli plugin log purge recognize steps by name [[#3953](https://github.com/woodpecker-ci/woodpecker/pull/3953)]
- Pin page size [[#3946](https://github.com/woodpecker-ci/woodpecker/pull/3946)]
- Improve cron list [[#3947](https://github.com/woodpecker-ci/woodpecker/pull/3947)]
- Add PULLREQUEST_DRONE_PULL_REQUEST drone env [[#3939](https://github.com/woodpecker-ci/woodpecker/pull/3939)]
- Make agent gRPC errors distinguishable [[#3936](https://github.com/woodpecker-ci/woodpecker/pull/3936)]

### 📦️ Dependency

- fix(deps): update web npm deps non-major [[#4735](https://github.com/woodpecker-ci/woodpecker/pull/4735)]
- chore(deps): update woodpeckerci/plugin-release docker tag to v0.2.3 [[#4734](https://github.com/woodpecker-ci/woodpecker/pull/4734)]
- chore(deps): update docker.io/woodpeckerci/plugin-surge-preview docker tag to v1.3.4 [[#4732](https://github.com/woodpecker-ci/woodpecker/pull/4732)]
- fix(deps): update golang-packages to v0.32.1 [[#4727](https://github.com/woodpecker-ci/woodpecker/pull/4727)]
- fix(deps): update module google.golang.org/protobuf to v1.36.3 [[#4726](https://github.com/woodpecker-ci/woodpecker/pull/4726)]
- fix(deps): update golang-packages [[#4725](https://github.com/woodpecker-ci/woodpecker/pull/4725)]
- chore(deps): lock file maintenance [[#4721](https://github.com/woodpecker-ci/woodpecker/pull/4721)]
- fix(deps): update module code.gitea.io/sdk/gitea to v0.20.0 [[#4710](https://github.com/woodpecker-ci/woodpecker/pull/4710)]
- fix(deps): update dependency simple-icons to v14.2.0 [[#4709](https://github.com/woodpecker-ci/woodpecker/pull/4709)]
- chore(deps): update dependency jsdom to v26 [[#4704](https://github.com/woodpecker-ci/woodpecker/pull/4704)]
- fix(deps): update web npm deps non-major [[#4703](https://github.com/woodpecker-ci/woodpecker/pull/4703)]
- chore(deps): update gitea/gitea docker tag to v1.23 [[#4701](https://github.com/woodpecker-ci/woodpecker/pull/4701)]
- fix(deps): update golang-packages [[#4688](https://github.com/woodpecker-ci/woodpecker/pull/4688)]
- fix(deps): update golang-packages [[#4678](https://github.com/woodpecker-ci/woodpecker/pull/4678)]
- fix(deps): update module golang.org/x/term to v0.28.0 [[#4671](https://github.com/woodpecker-ci/woodpecker/pull/4671)]
- chore(deps): lock file maintenance [[#4672](https://github.com/woodpecker-ci/woodpecker/pull/4672)]
- fix(deps): update dependency simple-icons to v14.1.0 [[#4668](https://github.com/woodpecker-ci/woodpecker/pull/4668)]
- fix(deps): update module golang.org/x/oauth2 to v0.25.0 [[#4665](https://github.com/woodpecker-ci/woodpecker/pull/4665)]
- chore(deps): update pre-commit hook golangci/golangci-lint to v1.63.4 [[#4660](https://github.com/woodpecker-ci/woodpecker/pull/4660)]
- fix(deps): update module github.com/moby/term to v0.5.2 [[#4658](https://github.com/woodpecker-ci/woodpecker/pull/4658)]
- fix(deps): update web npm deps non-major [[#4659](https://github.com/woodpecker-ci/woodpecker/pull/4659)]
- chore(deps): update docker.io/woodpeckerci/plugin-ready-release-go docker tag to v3.1.1 [[#4642](https://github.com/woodpecker-ci/woodpecker/pull/4642)]
- fix(deps): update dependency simple-icons to v14.0.1 [[#4640](https://github.com/woodpecker-ci/woodpecker/pull/4640)]
- fix(deps): update module github.com/google/go-github/v67 to v68 [[#4635](https://github.com/woodpecker-ci/woodpecker/pull/4635)]
- fix(deps): update dependency vue-i18n to v11 [[#4634](https://github.com/woodpecker-ci/woodpecker/pull/4634)]
- fix(deps): update dependency simple-icons to v14 [[#4633](https://github.com/woodpecker-ci/woodpecker/pull/4633)]
- chore(deps): update dependency vite to v6.0.6 [[#4632](https://github.com/woodpecker-ci/woodpecker/pull/4632)]
- fix(deps): update github.com/getkin/kin-openapi digest to cea0a13 [[#4630](https://github.com/woodpecker-ci/woodpecker/pull/4630)]
- chore(deps): lock file maintenance [[#4540](https://github.com/woodpecker-ci/woodpecker/pull/4540)]
- fix(deps): update web npm deps non-major [[#4440](https://github.com/woodpecker-ci/woodpecker/pull/4440)]
- fix(deps): update golang-packages [[#4615](https://github.com/woodpecker-ci/woodpecker/pull/4615)]
- fix(deps): update module gitlab.com/gitlab-org/api/client-go to v0.118.0 [[#4606](https://github.com/woodpecker-ci/woodpecker/pull/4606)]
- fix(deps): update module github.com/cenkalti/backoff/v4 to v5 [[#4601](https://github.com/woodpecker-ci/woodpecker/pull/4601)]
- fix(deps): update golang-packages [[#4586](https://github.com/woodpecker-ci/woodpecker/pull/4586)]
- fix(deps): update module golang.org/x/net to v0.33.0 [security] [[#4585](https://github.com/woodpecker-ci/woodpecker/pull/4585)]
- fix(deps): update golang-packages [[#4579](https://github.com/woodpecker-ci/woodpecker/pull/4579)]
- Replace discontinued mitchellh/mapstructure by maintained fork [[#4573](https://github.com/woodpecker-ci/woodpecker/pull/4573)]
- chore(deps): update docker.io/woodpeckerci/plugin-codecov docker tag to v2.1.6 [[#4566](https://github.com/woodpecker-ci/woodpecker/pull/4566)]
- fix(deps): update github.com/muesli/termenv digest to 8c990cd [[#4565](https://github.com/woodpecker-ci/woodpecker/pull/4565)]
- fix(deps): update module google.golang.org/grpc to v1.69.0 [[#4563](https://github.com/woodpecker-ci/woodpecker/pull/4563)]
- fix(deps): update golang-packages [[#4553](https://github.com/woodpecker-ci/woodpecker/pull/4553)]
- Update kin-openapi [[#4560](https://github.com/woodpecker-ci/woodpecker/pull/4560)]
- fix(deps): update module golang.org/x/crypto to v0.31.0 [security] [[#4557](https://github.com/woodpecker-ci/woodpecker/pull/4557)]
- fix(deps): update golang-packages [[#4546](https://github.com/woodpecker-ci/woodpecker/pull/4546)]
- chore(deps): update docker.io/woodpeckerci/plugin-ready-release-go docker tag to v3.1.0 [[#4536](https://github.com/woodpecker-ci/woodpecker/pull/4536)]
- chore(deps): update docker.io/curlimages/curl docker tag to v8.11.0 [[#4530](https://github.com/woodpecker-ci/woodpecker/pull/4530)]
- fix(deps): update golang-packages [[#4496](https://github.com/woodpecker-ci/woodpecker/pull/4496)]
- chore(deps): update docker.io/woodpeckerci/plugin-docker-buildx docker tag to v5.1.0 [[#4524](https://github.com/woodpecker-ci/woodpecker/pull/4524)]
- chore(deps): update docker.io/woodpeckerci/plugin-prettier docker tag to v1 [[#4522](https://github.com/woodpecker-ci/woodpecker/pull/4522)]
- chore(deps): update docker.io/alpine docker tag to v3.21 [[#4520](https://github.com/woodpecker-ci/woodpecker/pull/4520)]
- chore(deps): update dependency vite to v6 [[#4485](https://github.com/woodpecker-ci/woodpecker/pull/4485)]
- chore(deps): update docker.io/woodpeckerci/plugin-ready-release-go docker tag to v3 [[#4506](https://github.com/woodpecker-ci/woodpecker/pull/4506)]
- chore(deps): lock file maintenance [[#4502](https://github.com/woodpecker-ci/woodpecker/pull/4502)]
- chore(deps): lock file maintenance [[#4501](https://github.com/woodpecker-ci/woodpecker/pull/4501)]
- chore(deps): update docker.io/woodpeckerci/plugin-surge-preview docker tag to v1.3.3 [[#4495](https://github.com/woodpecker-ci/woodpecker/pull/4495)]
- fix(deps): update golang-packages [[#4477](https://github.com/woodpecker-ci/woodpecker/pull/4477)]
- fix(deps): update dependency @vueuse/core to v12 [[#4486](https://github.com/woodpecker-ci/woodpecker/pull/4486)]
- fix(deps): update module github.com/google/go-github/v66 to v67 [[#4487](https://github.com/woodpecker-ci/woodpecker/pull/4487)]
- chore(deps): update woodpeckerci/plugin-release docker tag to v0.2.2 [[#4483](https://github.com/woodpecker-ci/woodpecker/pull/4483)]
- chore(deps): update pre-commit hook golangci/golangci-lint to v1.62.2 [[#4482](https://github.com/woodpecker-ci/woodpecker/pull/4482)]
- fix(deps): update golang-packages [[#4452](https://github.com/woodpecker-ci/woodpecker/pull/4452)]
- chore(deps): lock file maintenance [[#4453](https://github.com/woodpecker-ci/woodpecker/pull/4453)]
- fix(deps): update golang-packages [[#4411](https://github.com/woodpecker-ci/woodpecker/pull/4411)]
- chore(deps): update pre-commit hook igorshubovych/markdownlint-cli to v0.43.0 [[#4443](https://github.com/woodpecker-ci/woodpecker/pull/4443)]
- chore(deps): update postgres docker tag to v17.2 [[#4442](https://github.com/woodpecker-ci/woodpecker/pull/4442)]
- chore(deps): lock file maintenance [[#4435](https://github.com/woodpecker-ci/woodpecker/pull/4435)]
- chore(deps): update docker.io/woodpeckerci/plugin-trivy docker tag to v1.3.0 [[#4434](https://github.com/woodpecker-ci/woodpecker/pull/4434)]
- chore(deps): update web npm deps non-major [[#4432](https://github.com/woodpecker-ci/woodpecker/pull/4432)]
- fix(deps): update golang-packages [[#4401](https://github.com/woodpecker-ci/woodpecker/pull/4401)]
- chore(deps): lock file maintenance [[#4402](https://github.com/woodpecker-ci/woodpecker/pull/4402)]
- chore(deps): update web npm deps non-major [[#4391](https://github.com/woodpecker-ci/woodpecker/pull/4391)]
- fix(deps): update dependency @intlify/unplugin-vue-i18n to v6 [[#4397](https://github.com/woodpecker-ci/woodpecker/pull/4397)]
- chore(deps): update pre-commit hook golangci/golangci-lint to v1.62.0 [[#4390](https://github.com/woodpecker-ci/woodpecker/pull/4390)]
- chore(deps): update postgres docker tag to v17.1 [[#4389](https://github.com/woodpecker-ci/woodpecker/pull/4389)]
- chore(deps): update docker.io/techknowlogick/xgo docker tag to go-1.23.x [[#4388](https://github.com/woodpecker-ci/woodpecker/pull/4388)]
- chore(config): migrate renovate config [[#4296](https://github.com/woodpecker-ci/woodpecker/pull/4296)]
- chore(deps): update docker.io/woodpeckerci/plugin-trivy docker tag to v1.2.0 [[#4289](https://github.com/woodpecker-ci/woodpecker/pull/4289)]
- chore(deps): update docker.io/techknowlogick/xgo docker tag to go-1.23.x [[#4282](https://github.com/woodpecker-ci/woodpecker/pull/4282)]
- fix(deps): update golang-packages [[#4251](https://github.com/woodpecker-ci/woodpecker/pull/4251)]
- fix(deps): update web npm deps non-major [[#4258](https://github.com/woodpecker-ci/woodpecker/pull/4258)]
- chore(deps): update web npm deps non-major [[#4250](https://github.com/woodpecker-ci/woodpecker/pull/4250)]
- chore(deps): update node.js to v23 [[#4239](https://github.com/woodpecker-ci/woodpecker/pull/4239)]
- chore(deps): update web npm deps non-major [[#4237](https://github.com/woodpecker-ci/woodpecker/pull/4237)]
- chore(deps): update docker.io/mysql docker tag to v9.1.0 [[#4236](https://github.com/woodpecker-ci/woodpecker/pull/4236)]
- fix(deps): update dependency simple-icons to v13.14.0 [[#4226](https://github.com/woodpecker-ci/woodpecker/pull/4226)]
- fix(deps): update web npm deps non-major [[#4223](https://github.com/woodpecker-ci/woodpecker/pull/4223)]
- fix(deps): update golang-packages [[#4215](https://github.com/woodpecker-ci/woodpecker/pull/4215)]
- fix(deps): update golang-packages [[#4210](https://github.com/woodpecker-ci/woodpecker/pull/4210)]
- fix(deps): update module github.com/google/go-github/v65 to v66 [[#4205](https://github.com/woodpecker-ci/woodpecker/pull/4205)]
- fix(deps): update dependency vue-i18n to v10.0.4 [[#4200](https://github.com/woodpecker-ci/woodpecker/pull/4200)]
- chore(deps): update pre-commit hook pre-commit/pre-commit-hooks to v5 [[#4192](https://github.com/woodpecker-ci/woodpecker/pull/4192)]
- fix(deps): update dependency simple-icons to v13.13.0 [[#4196](https://github.com/woodpecker-ci/woodpecker/pull/4196)]
- chore(deps): lock file maintenance [[#4186](https://github.com/woodpecker-ci/woodpecker/pull/4186)]
- chore(deps): update web npm deps non-major [[#4174](https://github.com/woodpecker-ci/woodpecker/pull/4174)]
- chore(deps): update docker.io/postgres docker tag to v17 [[#4179](https://github.com/woodpecker-ci/woodpecker/pull/4179)]
- fix(deps): update dependency @intlify/unplugin-vue-i18n to v5 [[#4183](https://github.com/woodpecker-ci/woodpecker/pull/4183)]
- fix(deps): update dependency @vueuse/core to v11 [[#4184](https://github.com/woodpecker-ci/woodpecker/pull/4184)]
- chore(deps): update docker.io/woodpeckerci/plugin-codecov docker tag to v2.1.5 [[#4167](https://github.com/woodpecker-ci/woodpecker/pull/4167)]
- fix(deps): update module github.com/google/go-github/v64 to v65 [[#4185](https://github.com/woodpecker-ci/woodpecker/pull/4185)]
- chore(deps): update docker.io/mysql docker tag to v9 [[#4178](https://github.com/woodpecker-ci/woodpecker/pull/4178)]
- chore(deps): update docker.io/alpine docker tag to v3.20 [[#4169](https://github.com/woodpecker-ci/woodpecker/pull/4169)]
- fix(deps): update github.com/urfave/cli/v3 digest to 20ef97b [[#4166](https://github.com/woodpecker-ci/woodpecker/pull/4166)]
- chore(deps): update docker.io/woodpeckerci/plugin-surge-preview docker tag to v1.3.2 [[#4168](https://github.com/woodpecker-ci/woodpecker/pull/4168)]
- chore(deps): update woodpeckerci/plugin-release docker tag to v0.2.1 [[#4175](https://github.com/woodpecker-ci/woodpecker/pull/4175)]
- chore(deps): update woodpeckerci/plugin-ready-release-go docker tag to v2 [[#4182](https://github.com/woodpecker-ci/woodpecker/pull/4182)]
- fix(deps): update github.com/muesli/termenv digest to 82936c5 [[#4165](https://github.com/woodpecker-ci/woodpecker/pull/4165)]
- chore(deps): update postgres docker tag to v17 [[#4181](https://github.com/woodpecker-ci/woodpecker/pull/4181)]
- chore(deps): update pre-commit non-major [[#4173](https://github.com/woodpecker-ci/woodpecker/pull/4173)]
- chore(deps): update docker.io/golang docker tag to v1.23 [[#4170](https://github.com/woodpecker-ci/woodpecker/pull/4170)]
- chore(deps): update node.js to v22 [[#4180](https://github.com/woodpecker-ci/woodpecker/pull/4180)]
- fix(deps): update golang-packages [[#4161](https://github.com/woodpecker-ci/woodpecker/pull/4161)]
- chore(deps): update dependency @antfu/eslint-config to v3 [[#4095](https://github.com/woodpecker-ci/woodpecker/pull/4095)]
- chore(deps): update dependency jsdom to v25 [[#4094](https://github.com/woodpecker-ci/woodpecker/pull/4094)]
- chore(deps): update docker.io/golang docker tag to v1.23 [[#4081](https://github.com/woodpecker-ci/woodpecker/pull/4081)]
- chore(deps): update docker.io/woodpeckerci/plugin-prettier docker tag to v0.2.0 [[#4082](https://github.com/woodpecker-ci/woodpecker/pull/4082)]
- fix(deps): update module github.com/google/go-github/v63 to v64 [[#4073](https://github.com/woodpecker-ci/woodpecker/pull/4073)]
- fix(deps): update golang-packages [[#4059](https://github.com/woodpecker-ci/woodpecker/pull/4059)]
- Update github.com/urfave/cli/v3 digest to fc07a8c [[#4043](https://github.com/woodpecker-ci/woodpecker/pull/4043)]
- Update woodpeckerci/plugin-git Docker tag to v2.5.2 [[#4041](https://github.com/woodpecker-ci/woodpecker/pull/4041)]
- Update web npm deps non-major [[#4034](https://github.com/woodpecker-ci/woodpecker/pull/4034)]
- Update dependency simple-icons to v13 [[#4037](https://github.com/woodpecker-ci/woodpecker/pull/4037)]
- chore(deps): lock file maintenance [[#3991](https://github.com/woodpecker-ci/woodpecker/pull/3991)]
- fix(deps): update golang-packages [[#3958](https://github.com/woodpecker-ci/woodpecker/pull/3958)]

### Misc

- Use mirror.gcr.io as `trivy` registry [[#4729](https://github.com/woodpecker-ci/woodpecker/pull/4729)]
- Add docs-dependencies target to makefile [[#4719](https://github.com/woodpecker-ci/woodpecker/pull/4719)]
- Move link checks into cron-curated issue dashboard [[#4515](https://github.com/woodpecker-ci/woodpecker/pull/4515)]
- Remove `renovate` branch triggers [[#4437](https://github.com/woodpecker-ci/woodpecker/pull/4437)]
- Dont run pipeline on push events to renovate branches [[#4406](https://github.com/woodpecker-ci/woodpecker/pull/4406)]
- Harden and correct fifo task queue tests [[#4377](https://github.com/woodpecker-ci/woodpecker/pull/4377)]
- Use release-helper for release/* branches [[#4301](https://github.com/woodpecker-ci/woodpecker/pull/4301)]
- Fix renovate support for `xgo` [[#4276](https://github.com/woodpecker-ci/woodpecker/pull/4276)]
- Improve nix development environment [[#4256](https://github.com/woodpecker-ci/woodpecker/pull/4256)]
- [pre-commit.ci] pre-commit autoupdate [[#4209](https://github.com/woodpecker-ci/woodpecker/pull/4209)]
- Add `.lycheeignore` [[#4154](https://github.com/woodpecker-ci/woodpecker/pull/4154)]
- Add eslint-plugin-promise back [[#4022](https://github.com/woodpecker-ci/woodpecker/pull/4022)]
- Improve wording [[#3951](https://github.com/woodpecker-ci/woodpecker/pull/3951)]
- Fix typos and optimize wording [[#3940](https://github.com/woodpecker-ci/woodpecker/pull/3940)]

## [2.7.2](https://github.com/woodpecker-ci/woodpecker/releases/tag/v2.7.2) - 2024-11-03

### Important

To secure your instance, set `WOODPECKER_PLUGINS_PRIVILEGED` to only allow specific versions of the `woodpeckerci/plugin-docker-buildx` plugin, use version 5.0.0 or above. This prevents older, potentially unstable versions from being privileged.

For example, to allow only version 5.0.0, use:

```bash
WOODPECKER_PLUGINS_PRIVILEGED=woodpeckerci/plugin-docker-buildx:5.0.0
```

To allow multiple versions, you can separate them with commas:

```bash
WOODPECKER_PLUGINS_PRIVILEGED=woodpeckerci/plugin-docker-buildx:5.0.0,woodpeckerci/plugin-docker-buildx:5.1.0
```

This setup ensures only specified, stable plugin versions are given privileged access.

Read more about it in [#4213](https://github.com/woodpecker-ci/woodpecker/pull/4213)

### ❤️ Thanks to all contributors! ❤️

@6543, @anbraten, @j04n-f, @pat-s, @qwerty287

### 🔒 Security

- Chore(deps): update dependency vite to v5.4.6 [security] ([#4163](https://github.com/woodpecker-ci/woodpecker/pull/4163)) [[#4187](https://github.com/woodpecker-ci/woodpecker/pull/4187)]

### 🐛 Bug Fixes

- Don't parse forge config files multiple times if no error occured ([#4272](https://github.com/woodpecker-ci/woodpecker/pull/4272)) [[#4273](https://github.com/woodpecker-ci/woodpecker/pull/4273)]
- Fix repo/owner parsing for gitlab ([#4255](https://github.com/woodpecker-ci/woodpecker/pull/4255)) [[#4261](https://github.com/woodpecker-ci/woodpecker/pull/4261)]
- Run queue.process() in background [[#4115](https://github.com/woodpecker-ci/woodpecker/pull/4115)]
- Only update agent.LastWork if not done recently ([#4031](https://github.com/woodpecker-ci/woodpecker/pull/4031)) [[#4100](https://github.com/woodpecker-ci/woodpecker/pull/4100)]

### Misc

- Backport JS dependency updates [[#4189](https://github.com/woodpecker-ci/woodpecker/pull/4189)]

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
- Enhance pipeline list [[#3898](https://github.com/woodpecker-ci/woodpecker/pull/3898)]
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
