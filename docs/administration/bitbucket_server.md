Drone comes with experimental support for Bitbucket Server, formerly known as Atlassian Stash. To enable Bitbucket Server you should configure the Drone container using the following environment variables:

```diff
version: '2'

services:
  drone-server:
    image: drone/drone:{{% version %}}
    ports:
      - 80:8000
      - 9000
    volumes:
      - /var/lib/drone:/var/lib/drone/
    restart: always
    environment:
      - DRONE_OPEN=true
      - DRONE_HOST=${DRONE_HOST}
+     - DRONE_STASH=true
+     - DRONE_STASH_GIT_USERNAME=foo
+     - DRONE_STASH_GIT_PASSWORD=bar
+     - DRONE_STASH_CONSUMER_KEY=95c0282573633eb25e82
+     - DRONE_STASH_CONSUMER_RSA=/etc/bitbucket/key.pem
+     - DRONE_STASH_URL=http://stash.mycompany.com
      - DRONE_SECRET=${DRONE_SECRET}
    volumes:
+     - /path/to/key.pem:/path/to/key.pem

  drone-agent:
    image: drone/agent:{{% version %}}
    restart: always
    depends_on:
      - drone-server
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - DRONE_SERVER=drone-server:9000
      - DRONE_SECRET=${DRONE_SECRET}
```

# Private Key File

The OAuth process in Bitbucket server requires a private and a public RSA certificate. This is how you create the private RSA certificate.

```nohighlight
openssl genrsa -out /etc/bitbucket/key.pem 1024
```

This stores the private RSA certificate in `key.pem`. The next command generates the public RSA certificate and stores it in `key.pub`.

```nohighlight
openssl rsa -in /etc/bitbucket/key.pem -pubout >> /etc/bitbucket/key.pub
```

Please note that the private key file can be mounted into your Drone conatiner at runtime or as an environment variable

Private key file mounted into your Drone container at runtime as a volume.

```diff
version: '2'

services:
  drone-server:
    image: drone/drone:{{% version %}}
    environment:
    - DRONE_OPEN=true
    - DRONE_HOST=${DRONE_HOST}
      - DRONE_STASH=true
      - DRONE_STASH_GIT_USERNAME=foo
      - DRONE_STASH_GIT_PASSWORD=bar
      - DRONE_STASH_CONSUMER_KEY=95c0282573633eb25e82
+     - DRONE_STASH_CONSUMER_RSA=/etc/bitbucket/key.pem
      - DRONE_STASH_URL=http://stash.mycompany.com
      - DRONE_SECRET=${DRONE_SECRET}
+  volumes:
+     - /etc/bitbucket/key.pem:/etc/bitbucket/key.pem
```

Private key as environment variable

```diff
version: '2'

services:
  drone-server:
    image: drone/drone:{{% version %}}
    environment:
    - DRONE_OPEN=true
    - DRONE_HOST=${DRONE_HOST}
      - DRONE_STASH=true
      - DRONE_STASH_GIT_USERNAME=foo
      - DRONE_STASH_GIT_PASSWORD=bar
      - DRONE_STASH_CONSUMER_KEY=95c0282573633eb25e82
+     - DRONE_STASH_CONSUMER_RSA_STRING=contentOfPemKeyAsString
      - DRONE_STASH_URL=http://stash.mycompany.com
      - DRONE_SECRET=${DRONE_SECRET}
```

# Service Account

Drone uses `git+https` to clone repositories, however, Bitbucket Server does not currently support cloning repositories with oauth token. To work around this limitation, you must create a service account and provide the username and password to Drone. This service account will be used to authenticate and clone private repositories.

# Registration

You must register your application with Bitbucket Server in order to generate a consumer key. Navigate to your account settings and choose Applications from the menu, and click Register new application. Now copy & paste the text value from `/etc/bitbucket/key.pub` into the `Public Key` in the incoming link part of the application registration.

Please use http://drone.mycompany.com/authorize as the Authorization callback URL.


# Configuration

This is a full list of configuration options. Please note that many of these options use default configuration values that should work for the majority of installations.


DRONE_STASH=true
: Set to true to enable the Bitbucket Server (Stash) driver.

DRONE_STASH_URL
: Bitbucket Server address.

DRONE_STASH_CONSUMER_KEY
: Bitbucket Server oauth1 consumer key

DRONE_STASH_CONSUMER_RSA
: Bitbucket Server oauth1 private key file

DRONE_STASH_CONSUMER_RSA_STRING
: Bibucket Server oauth1 private key as a string

DRONE_STASH_GIT_USERNAME
: Machine account username used to clone repositories.

DRONE_STASH_GIT_PASSWORD
: Machine account password used to clone repositories.
