# SSH backend

The local backend will execute the pipelines using SSH on a remote system without any isolation of any kind.

Since the code run directly on the SSH machine, a malicious pipeline could access and edit files the SSH user has access to. Always restrict the user as far as possible!

It is recommended to use this backend only for private setup where the code and pipeline can be trusted. You shouldn't
use it for a public facing CI where anyone can submit code or add new repositories.

The backend will use a random directory in $TMPDIR to store the clone code and execute commands.
