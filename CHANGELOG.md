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
