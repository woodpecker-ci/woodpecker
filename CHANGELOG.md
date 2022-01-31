## [0.14.4](https://github.com/woodpecker-ci/woodpecker/releases/tag/v0.14.4) - 2022-01-31

* BUGFIXES
  * Docker Images use golang image for ca-certificates (#608)

## [0.14.3](https://github.com/woodpecker-ci/woodpecker/releases/tag/v0.14.3) - 2021-10-30

* BUGFIXES
  * Add flag for not fetching permissions (FlatPermissions) (#491)
  * Gitea use default branch (#480) (#482)
  * Fix repo access (#476) (#477)
* ENHANCEMENTS
  * Use go embed for web files and remove httptreemux (#382) (#489)

## [0.14.2](https://github.com/woodpecker-ci/woodpecker/releases/tag/v0.14.2) - 2021-10-19

* BUGFIXES
  * Fix sanitziePath (#326) (aa4fa9aab3)
  * Fix json tag for `Pos` at struct `Line` (#422) (#424)
  * Fix channel buffer used with signal.Notify (#421) (#423)
* ENHANCEMENTS
  * Support recursive glob for path conditions (#327) (#412)
* TESTING
  * Add TestPipelineName to procBuilder_test.go (#461) (#455)

## [0.14.1](https://github.com/woodpecker-ci/woodpecker/releases/tag/v0.14.1) - 2021-09-21

* SECURITY
  * Migrate jwt token lib (#332)
* BUGFIXES
  * Increase allowed length for user token in db (#328)
  * Fix cli matrix filter (#311)
  * Fix ignore pushes to tags for gitea (#289)
  * Fix use custom config path to sanitize build names (#280)

## [0.14.0](https://github.com/woodpecker-ci/woodpecker/releases/tag/v0.14.0) - 2021-08-01

* FEATURES
  * Add OAuth2 Support for Gitea Remote (#226)
  * Add support for path-prefix condition (#174)
* BUGFIXES
  * Allow multi pipeline file to be named .drone.yml (#250)
  * Fix release-server make target by build server with correct option (#237)
  * Fix Gitea unable to login on 0.12.0+ with error "cannot authenticate user. 403 Forbidden" (#221)
* ENHANCEMENTS
  * Update / Remove drone dependencies (#236)
  * Add support to gitea remote for path-prefix condition (#235)
  * Enable go vet for ci (#230)
  * Enforce code format (#228)
  * Add mutli-pipeline to Gitea (#225)
  * Move flag definitions into extra files (#215)
  * Remove unused code in server (#213)
  * Docs URL configuration (#206)
  * Filter main branch (#205)
  * Fix multi pipeline bug when a pipeline depends on two other pipelines (#201)
  * Using configured server URL instead of obtained from request (#175)
* DOCUMENTATION
  * Switch in docs to new docker hub image repo (#227)
  * Use WOODPECKER_ env vars in docs (#211)
  * Also show WOODPECKER_HOST and WOODPECKER_SERVER_HOST environment variables in log messages (#208)
  * Move woodpecker to dedicated organisation on github (#202)
* MISC
  * Add chart for installing woodpecker server and agent (#199)
