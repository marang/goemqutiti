# GoEmqutiti

GoEmqutiti is a terminal based MQTT client built with [Bubble Tea](https://github.com/charmbracelet/bubbletea). It loads connection profiles from `~/.emqutiti/config.toml` and lets you choose which broker to connect to at runtime.

## Installation

1. Clone this repository.

```bash
git clone <repo-url>
cd goemqutiti
```

2. Build the application:

```bash
go build
```

This will produce a `goemqutiti` binary in the current directory.

## Usage

Run the built binary (or use `go run .`) to start the TUI application. On startup the connection manager is shown so you can select which profile to use. The manager can also be opened at any time with the `m` key. Profiles expose all common connection options inspired by the EMQX MQTT client:

```bash
./goemqutiti
```

The client expects a configuration file at `~/.emqutiti/config.toml` describing connection profiles. A minimal configuration looks like:

```toml
default_profile = "local"

[[profiles]]
name = "local"
broker = "tcp://localhost:1883"
client_id = "goemqutiti"
username = "user"
password = "keyring:emqutiti-local/user"
```

Passwords can be stored securely using the operating system keyring. You may also set the `MQTT_PASSWORD` environment variable to override the stored password at runtime.

In the interface:

- **Tab** switches between the topic and message fields.
- **Enter** subscribes to a topic the first time and publishes messages afterwards.
- **Ctrl+M** opens the connection manager where you can add, edit or delete MQTT profiles.
- **Ctrl+T** manages subscribed topics.
- **Ctrl+P** manages stored payloads.
- **Ctrl+C** copies the currently selected history entry.
- **Ctrl+D** exits the program.

## License

This project is licensed under the terms of the MIT License. See [LICENSE](LICENSE) for details.

## Testing

Unit tests can be run with `go test ./...`. The example `ExampleSet_manual` in
`keyring_util_test.go` interacts with the real system keyring and is excluded
from automated runs. Execute it manually when a keyring is available.

Additional notes for repository contributors are available in [Agent.md](Agent.md).
