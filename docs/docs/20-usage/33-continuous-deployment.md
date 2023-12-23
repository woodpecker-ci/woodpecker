# Continuous Deployment

A typical CI pipeline contains steps such as: _clone_, _build_, _test_, _package_ and _push_. The final build product may be binaries pushed to a git repository or a docker container pushed to a container registry.

When these should be deployed on an app server, the pipeline should include a _deploy_ step, which represents the "CD" in CI/CD - the automatic deployment of a pipeline's final product.

There are various ways to accomplish CD with Woodpecker, depending on your project's specific needs.

## Invoking deploy script via SSH

The final step in your pipeline could SSH into the app server and run a deployment script.

One of the benefits would be that the deployment script's output could be included in the pipeline's log. However in general, this is a complicated option as it tightly couples the CI and app servers.

An SSH step could be written by using an plugin, like [ssh](https://plugins.drone.io/plugins/ssh) or [git push](https://woodpecker-ci.org/plugins/Git%20Push).

## Polling for asset changes

This option completely decouples the CI and app servers, and there is no explicit deploy step in the pipeline.

On the app server, one should create a script or cron job that polls for asset changes (every minute, say). When a new version is detected, the script redeploys the app.

This option is easy to maintain, but the downside is a short delay (one minute) before new assets are detected.

## Using a configuration management tool

If you are using a configuration management tool (e.g. Ansible, Chef, Puppet), then you could setup the last pipeline step to call that tool to perform the redeployment.

A plugin for [Ansible](https://plugins.drone.io/plugins/ansible) exists and could be adapted accordingly.

This option is complex and only suitable in an environment in which you're already using configuration management.

## Using webhooks (recommended)

If your forge (Github, Gitlab, Gitea, etc.) supports webhooks, then you could create a separate listening app that receives a webhook when new assets are available and redeploys your app.

The listening "app" can be something as simple as a PHP script.

Alternatively, there are a number of popular webhook servers that simplify this process, so you only need to write your actual deployment script. For example, [webhook](https://github.com/adnanh/webhook) and [webhookd](https://github.com/ncarlier/webhookd).

This is arguably the simplest and most maintainable solution.
