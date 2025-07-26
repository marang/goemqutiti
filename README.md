# GoEmqutiti

GoEmqutiti is a terminal based MQTT client built with [Bubble Tea](https://github.com/charmbracelet/bubbletea). It loads broker profiles from `~/.emqutiti/config.toml` and lets you choose which broker to connect to at runtime.

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

Run the built binary (or use `go run .`) to start the TUI application. On startup the broker manager is shown so you can select which profile to use. The manager can also be opened at any time with the `Ctrl+B`. Profiles expose all common connection options inspired by the EMQX MQTT client:

```bash
./goemqutiti
```

The client expects a configuration file at `~/.emqutiti/config.toml` describing broker profiles. A minimal configuration looks like:

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

- **Tab** cycles focus between the topic input, message editor, and topic chips.
- **Enter** subscribes to a topic when the topic field is focused.
- **Ctrl+S** publishes the message currently in the editor.
- **Ctrl+Enter** also publishes the current message.
- **Ctrl+B** opens the broker manager where you can add, edit or delete MQTT profiles.
- **Ctrl+T** manages subscribed topics.
- **Ctrl+P** manages stored payloads.
- **Ctrl+C** copies the currently selected history entry.
- **Esc** navigates back within menus without quitting.
- **Ctrl+D** exits the program.
- Left-click a topic chip to toggle it and middle-click to remove it.
- Clicking on any pane or input field focuses it.

All `Ctrl` shortcuts are global, so they work even when an input field is active.

## License

This project is licensed under the terms of the MIT License. See [LICENSE](LICENSE) for details.

## Testing

Unit tests can be run with `go test ./...`. The example `ExampleSet_manual` in
`keyring_util_test.go` interacts with the real system keyring and is excluded
from automated runs. Execute it manually when a keyring is available.

Additional notes for repository contributors are available in [Agent.md](Agent.md).
