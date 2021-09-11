# Migrations

Some versions need some changes to the server configuration or the pipeline configuration files.

## 0.15.0

- Default pipeline path changed to `.woodpecker/`

  **Solution:** Set configuration location via [project settings](/docs/usage/project-settings#pipeline-path).
  
  There is still a default fallback mechanism in following order: `.woodpecker/*.yml` -> `.woodpecker.yml` -> `.drone.yml`
- ...

## 0.14.0

No breaking changes
