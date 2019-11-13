# Motivation

I was using Drone for two years with great satisfaction. The container architecture, the speedy backend and UI, the simple plugin system made it a flexible and simple platform. Kudos for the author, Brad to make it such a joy to use.

It wasn't without flaws
- inconsistencies in variables and CLI features
- lack of documentation
- lack of published best practices
- UI/UX issues
- stuck builds

Things that could be circumvented by reading the codebase. Over time however these started to annoy me, also PRs that tried to address these were not merged. Instead the development of Drone headed towards a 1.0 release with features less interesting to me.

1.0 landed and it came with a licence change. Drone has been an open-core project since many prior versions, but the enterprise features were limited to features like autoscaling and secret vaults. 

In the 1.0 line however, Postgresql, Mysql and TLS support along with agent based horizontal scaling were also moved under the enterprise license. Limiting the open source version to single node, hobbyist deployments.

These feature reductions and my long time UX annoyance and general dissatisfaction of the CI space lead to this fork.
