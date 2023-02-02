# Prometheus

Woodpecker is compatible with Prometheus and exposes a `/metrics` endpoint. Please note that access to the metrics endpoint is restricted and requires an authorization token with administrative privileges.

```yaml
global:
  scrape_interval: 60s

scrape_configs:
  - job_name: 'woodpecker'
    bearer_token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

    static_configs:
       - targets: ['woodpecker.domain.com']
```

## Authorization

An administrator will need to generate a user API token and configure in the Prometheus configuration file as a bearer token. Please see the following example:

```diff
global:
  scrape_interval: 60s

scrape_configs:
  - job_name: 'woodpecker'
+   bearer_token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

    static_configs:
       - targets: ['woodpecker.domain.com']
```

## Metric Reference

List of Prometheus metrics specific to Woodpecker:

```
# HELP woodpecker_pipeline_count Pipeline count.
# TYPE woodpecker_pipeline_count counter
woodpecker_build_count{branch="master",pipeline="total",repo="woodpecker-ci/woodpecker",status="success"} 3
woodpecker_build_count{branch="mkdocs",pipeline="total",repo="woodpecker-ci/woodpecker",status="success"} 3
# HELP woodpecker_pipeline_time Build time.
# TYPE woodpecker_pipeline_time gauge
woodpecker_build_time{branch="master",pipeline="total",repo="woodpecker-ci/woodpecker",status="success"} 116
woodpecker_build_time{branch="mkdocs",pipeline="total",repo="woodpecker-ci/woodpecker",status="success"} 155
# HELP woodpecker_pipeline_total_count Total number of builds.
# TYPE woodpecker_pipeline_total_count gauge
woodpecker_build_total_count 1025
# HELP woodpecker_pending_steps Total number of pending pipeline steps.
# TYPE woodpecker_pending_steps gauge
woodpecker_pending_steps 0
# HELP woodpecker_repo_count Total number of repos.
# TYPE woodpecker_repo_count gauge
woodpecker_repo_count 9
# HELP woodpecker_running_steps Total number of running pipeline steps.
# TYPE woodpecker_running_steps gauge
woodpecker_running_steps 0
# HELP woodpecker_user_count Total number of users.
# TYPE woodpecker_user_count gauge
woodpecker_user_count 1
# HELP woodpecker_waiting_steps Total number of pipeline waiting on deps.
# TYPE woodpecker_waiting_steps gauge
woodpecker_waiting_steps 0
# HELP woodpecker_worker_count Total number of workers.
# TYPE woodpecker_worker_count gauge
woodpecker_worker_count 4
```
