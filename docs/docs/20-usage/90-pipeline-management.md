# Advanced pipeline management

## Using variables
Once your pipeline starts to grow in size, it will become important to keep it DRY ("Don't Repeat Yourself") by using variables and environment variables. Depending on your specific need, there are a number of options.

### YAML extensions
As described in [Advanced YAML syntax](/docs/docs/20-usage/35-advanced-yaml-syntax.md).
```yml
variables:
  - &golang_image 'golang:1.18'

 steps:
   build:
     image: *golang_image
     commands: build
```
Note that the `golang_image` alias cannot be used with string interpolation. But this is otherwise a good option for most cases.

### YAML extensions (alternate form)
Another approach using YAML extensions:
```yml
variables:
  - global_env: &global_env
    - BASH_VERSION=1.2.3
    - PATH_SRC=src/
    - PATH_TEST=test/
    - FOO=something

steps:
  build:
    image: bash:${BASH_VERSION}
    directory: ${PATH_SRC}
    commands:
      - make ${FOO} -o ${PATH_TEST}
    environment: *global_env

  test:
    image: bash:${BASH_VERSION}
    commands:
      - test ${PATH_TEST}
    environment:
      - <<:*global_env
      - ADDITIONAL_LOCAL="var value"
```

### Persisting environment data between steps
One can create a file containing environment variables, and then source it in each step that needs them.
```yml
steps:
  init:
    image: bash
    commands:
      echo "FOO=hello" >> envvars
      echo "BAR=world" >> envvars

  debug:
    image: bash
    commands:
      - source envvars
      - echo $FOO
```

### Declaring global variables in `docker-compose.yml`
As described in [Global environment variables](/docs/docs/20-usage/50-environment.md#global-environment-variables), one can define global variables:
```yml
services:
  woodpecker-server:
    # ...
    environment:
      - WOODPECKER_ENVIRONMENT=first_var:value1,second_var:value2
      # ...
```
Note that this tightly couples the server and app configurations (where the app is a completely separate application). But this is a good option for truly global variables which should apply to all steps in all pipelines for all apps.
