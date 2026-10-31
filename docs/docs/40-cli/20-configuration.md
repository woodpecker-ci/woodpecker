# Configuration

## First steps

Connect the CLI to your server with `woodpecker-cli setup`. It opens the Woodpecker UI in your browser, where you confirm the login, and saves the server and your token as a context. If the machine has no browser, pass the personal access token from the **CLI & API** page of your user settings with `--token`. `woodpecker-cli info` then shows the user you are logged in as:

```shell
woodpecker-cli setup --server https://ci.example.test
woodpecker-cli info
```

To work with more than one server, run `woodpecker-cli setup --context <name>` once per server and switch between them with [`woodpecker-cli context use`](./90-reference.md#context-ctx).

Shell completion is available for bash, zsh, fish and PowerShell. To enable it for bash, add this line to your `.bashrc`:

```shell
source <(woodpecker-cli completion bash)
```

Run `woodpecker-cli completion --help` for the other shells.

## Configuration storage

woodpecker-cli stores named contexts in a `contexts.json` file under the user's configuration directory. `XDG_CONFIG_HOME` overrides the base directory. The default paths are:

| Platform                     | Path                                                     |
| ---------------------------- | -------------------------------------------------------- |
| Linux and other Unix systems | `~/.config/woodpecker/contexts.json`                     |
| macOS                        | `~/Library/Application Support/woodpecker/contexts.json` |
| Windows                      | `C:\Users\<user>\AppData\Local\woodpecker\contexts.json` |

`--contexts-file` or `WOODPECKER_CONTEXTS_FILE` selects a different file for all commands, including `setup` and `context`.

The file records the selected context and its non-secret connection settings:

```json
{
  "current_context": "production",
  "contexts": {
    "production": {
      "name": "production",
      "server_url": "https://ci.example.test",
      "log_level": "info"
    }
  }
}
```

`current_context` must name an entry in `contexts`. The `woodpecker-cli context use` command changes that selection. Authentication tokens are never written to this JSON file; they are stored in the operating system's keyring and looked up by server URL.

## Precedence

Command-line flags take precedence over their matching environment variables, and either source takes precedence over stored context values:

| Flag          | Environment variable   |
| ------------- | ---------------------- |
| `--server`    | `WOODPECKER_SERVER`    |
| `--token`     | `WOODPECKER_TOKEN`     |
| `--log-level` | `WOODPECKER_LOG_LEVEL` |

Without a current context, both `--server` and `--token` must be set.
