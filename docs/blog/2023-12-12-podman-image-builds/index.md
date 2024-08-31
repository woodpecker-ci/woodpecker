---
title: '[Community] Podman-in-Podman image builds'
description: Build images in Podman with buildah
slug: podman-image-builds
authors:
  - name: handlebargh
    url: https://github.com/handlebargh
    image_url: https://github.com/handlebargh.png
hide_table_of_contents: true
tags: [community, image, podman]
---

<!-- cspell:ignore buildah Containerfile roundcube -->

I run Woodpecker CI with podman backend instead of docker and just figured out how to build images with buildah. Since I couldn't find this anywhere documented, I thought I might as well just share it here.

It's actually pretty straight forward. Here's what my repository structure looks like:

```bash
.
├── roundcube
│   ├── Containerfile
│   ├── docker-entrypoint.sh
│   └── php.ini
└── .woodpecker
    └── .build_roundcube.yml
```

As you can see I'm building a roundcube mail image.

This is the `.woodpecker/.build_roundcube.yaml`

```yaml
when:
  event: [cron, manual]
  cron: build_roundcube

steps:
  build-image:
    image: quay.io/buildah/stable:latest
    pull: true
    privileged: true
    commands:
      - echo $REGISTRY_LOGIN_TOKEN | buildah login -u <username> --password-stdin registry.gitlab.com
      - cd roundcube
      - buildah build --tag registry.gitlab.com/<namespace>/<repository_name>/roundcube:latest .
      - buildah push registry.gitlab.com/<namespace>/<repository_name>/roundcube:latest

    secrets: [registry_login_token]
```

As you can see, I'm using this workflow over at gitlab.com. It should work with GitHub as well, with adjusting the registry login.

You may have to adjust the `when:` to your needs. Furthermore, you must check the `trusted` checkbox in project settings. Therefore, be sure to run trusted code only in this setup.

This seems to work fine so far. I wonder if anybody else made this work a different way.

EDIT: Removed the additional step that would run buildah in a podman container. I didn't know it could be that easy to be honest.
