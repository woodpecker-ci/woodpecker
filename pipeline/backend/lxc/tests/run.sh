#!/bin/sh -ex

#
# Cloning self
#

woodpecker-cli exec --backend-engine=lxc --local=false --log-level=trace --repo-clone-url=$(pwd) pipeline/backend/lxc/tests/clone.yml

#
# No cloning
#

woodpecker-cli exec --backend-engine=lxc --local=true --log-level=trace pipeline/backend/lxc/tests/invalid-image.yml > /tmp/out 2>&1 || true
cat /tmp/out
grep 'does not match' /tmp/out

woodpecker-cli exec --backend-engine=lxc --local=true --log-level=trace pipeline/backend/lxc/tests/service.yml

woodpecker-cli exec --backend-engine=lxc --local=true --log-level=trace pipeline/backend/lxc/tests/simple.yml

woodpecker-cli exec --backend-engine=lxc --local=true --log-level=trace pipeline/backend/lxc/tests/workspace.yml
