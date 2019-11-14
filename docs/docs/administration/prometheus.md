# Prometheus

Woodpecker is compatible with Prometheus and exposes a `/metrics` endpoint. Please note that access to the metrics endpoint is restricted and requires an authorization token with administrative privileges.

```yaml
global:
  scrape_interval: 60s

scrape_configs:
  - job_name: 'drone'
    bearer_token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

    static_configs:
       - targets: ['woodpecker.domain.com']
```

## Authorization

An administrator will need to generate a user api token and configure in the prometheus configuration file as a bearer token. Please see the following example:

```diff
global:
  scrape_interval: 60s

scrape_configs:
  - job_name: 'drone'
+   bearer_token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

    static_configs:
       - targets: ['woodpecker.domain.com']
```

## Metric Reference

List of prometheus metrics specific to Woodpecker:

```
# HELP drone_build_count Build count.
# TYPE drone_build_count counter
drone_build_count{branch="master",pipeline="total",repo="laszlocph/woodpecker",status="success"} 3
drone_build_count{branch="mkdocs",pipeline="total",repo="laszlocph/woodpecker",status="success"} 3
# HELP drone_build_time Build time.
# TYPE drone_build_time gauge
drone_build_time{branch="master",pipeline="total",repo="laszlocph/woodpecker",status="success"} 116
drone_build_time{branch="mkdocs",pipeline="total",repo="laszlocph/woodpecker",status="success"} 155
# HELP drone_build_total_count Total number of builds.
# TYPE drone_build_total_count gauge
drone_build_total_count 1025
# HELP drone_pending_jobs Total number of pending build processes.
# TYPE drone_pending_jobs gauge
drone_pending_jobs 0
# HELP drone_repo_count Total number of repos.
# TYPE drone_repo_count gauge
drone_repo_count 9
# HELP drone_running_jobs Total number of running build processes.
# TYPE drone_running_jobs gauge
drone_running_jobs 0
# HELP drone_user_count Total number of users.
# TYPE drone_user_count gauge
drone_user_count 1
# HELP drone_waiting_jobs Total number of builds waiting on deps.
# TYPE drone_waiting_jobs gauge
drone_waiting_jobs 0
# HELP drone_worker_count Total number of workers.
# TYPE drone_worker_count gauge
drone_worker_count 4
```
