# SSH backend

:::danger
The SSH backend will execute the pipelines using SSH on a remote system without any isolation of any kind.
:::

:::note
This backend is still pretty new and can not be treated as stable. Its implementation and configuration can change at any time.
:::
Since the code run directly on the SSH machine, a malicious pipeline could access and edit files the SSH user has access to and execute every command the remote user is allowed to use. Always restrict the user as far as possible!

It is recommended to use this backend only for private setups where the code and pipelines can be trusted. You shouldn't use it for a public facing CI where anyone can submit code or add new repositories.

The backend will use a random directory in $TMPDIR to store the clone code and execute commands.
