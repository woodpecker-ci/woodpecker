# Local backend

:::danger
The local backend will execute the pipelines on the local system without any isolation of any kind.
:::

Since the code run directly in the same context as the agent (same user, same filesystem), a malicious pipeline could 
be used to access the agent configuration especially the `WOODPECKER_AGENT_SECRET` variable.

It is recommended to use this backend only for private setup where the code and pipeline can be trusted. You shouldn't
use it for a public facing CI where anyone can submit code or add new repositories.

The backend will use a random directory in $TMPDIR to store the clone code and execute commands.
