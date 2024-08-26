---
title: '[Community] Podman image build with sigstore'
description: Build images in Podman with sigstore signature checking and signing
slug: podman-image-build-sigstore
authors:
  - name: handlebargh
    url: https://github.com/handlebargh
    image_url: https://github.com/handlebargh.png
hide_table_of_contents: false
tags: [community, image, podman, sigstore, signature]
---

<!-- cspell:ignore BQVUJ Containerfile cosing distroless fulcio keypair nonroot QVRFLS rekor skopeo -->

This example shows how to build a container image with podman while verifying the base image and signing the resulting image.

The image being pulled uses a keyless signature, while the image being built will be signed by a pre-generated private key.

## Prerequisites

### Generate signing keypair

You can use cosing or skopeo to generate the keypair.

Using skopeo:

```bash
skopeo generate-sigstore-key --output-prefix myKey
```

This command will generate a `myKey.private` and a `myKey.pub` keyfile.

Store the `myKey.private` as secret in Woodpecker. In the example below, the secret is called `sigstore_private_key`

### Configure hosts pulling the resulting image

See [here](https://access.redhat.com/documentation/en-us/red_hat_enterprise_linux/8/html/building_running_and_managing_containers/assembly_signing-container-images_building-running-and-managing-containers#proc_verifying-sigstore-image-signatures-using-a-public-key_assembly_signing-container-images) on how to configure the hosts pulling the built and signed image.

## Repository structure

Consider the `Makefile` having a `build` target that will be used in the following workflow.
This target yields a Go binary with the filename `app` that will be placed in the root directory.

```bash
.
├── Containerfile
├── main.go
├── go.mod
├── go.sum
├── .woodpecker.yml
└── Makefile
```

### Containerfile

The Containerfile refers to the base image that will be verified when pulled.

```dockerfile
FROM gcr.io/distroless/static-debian12:nonroot
COPY app /app
CMD ["/app"]
```

### Woodpecker workflow

```yaml
steps:
  build:
    image: docker.io/library/golang:1.21
    pull: true
    commands:
      - make build

  publish:
    image: quay.io/podman/stable:latest
    # Caution: This image is built daily. It might fill up your image store quickly.
    pull: true
    # Fill in the trusted checkbox in Woodpecker's settings as well
    privileged: true
    commands:
      # Configure podman to use sigstore attachments for both, the registry you pull from and the registry you push to.
      - |
        printf "docker:
          registry.gitlab.com:
            use-sigstore-attachments: true
          gcr.io:
            use-sigstore-attachments: true" >> /etc/containers/registries.d/default.yaml

      # At pull, check the keyless sigstore signature of the distroless image.
      # This is a very strict container policy. It allows pulling from gcr.io/distroless only. Every other registry will be rejected.
      # See https://github.com/containers/image/blob/main/docs/containers-policy.json.5.md for more information.

      # fulcio CA crt obtained from https://github.com/sigstore/sigstore/blob/main/pkg/tuf/repository/targets/fulcio_v1.crt.pem
      # rekor public key obtained from https://github.com/sigstore/sigstore/blob/main/pkg/tuf/repository/targets/rekor.pub
      # crt/key data is base64 encoded. --> echo "$CERT" | base64
      - |
        printf '{
            "default": [
              {
                "type": "reject"
              }
            ],
            "transports": {
              "docker": {
                "gcr.io/distroless": [
                  {
                    "type": "sigstoreSigned",
                    "fulcio": {
                      "caData": "LS0tLS1CRUdJTiBDR...QVRFLS0tLS0K",
                      "oidcIssuer": "https://accounts.google.com",
                      "subjectEmail": "keyless@distroless.iam.gserviceaccount.com"
                    },
                    "rekorPublicKeyData": "LS0tLS1CRUdJTiBQVUJ...lDIEtFWS0tLS0tCg==",
                    "signedIdentity": { "type": "matchRepository" }
                  }
                ]
              },
              "docker-daemon": {
                "": [
                  {
                    "type": "reject"
                  }
                ]
              }
            }
          }' > /etc/containers/policy.json

      # Use this key to sign the built image at push.
      - echo "$SIGSTORE_PRIVATE_KEY" > key.private
      # Login at the registry
      - echo $REGISTRY_LOGIN_TOKEN | podman login -u <username> --password-stdin registry.gitlab.com
      # Build the container image
      - podman build --tag registry.gitlab.com/<namespace>/<repository_name>/<image_name>:latest .
      # Sign and push the image
      - podman push --sign-by-sigstore-private-key ./key.private registry.gitlab.com/<namespace>/<repository_name>/<image_name>:latest

    secrets: [sigstore_private_key, registry_login_token]
```
