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

The client expects a configuration file at `~/.emqutiti/config.toml` describing broker profiles. A sample configuration with all options looks like:

```toml
default_profile = "local"

[[profiles]]
name = "local"
schema = "tcp"
host = "localhost"
port = 1883
client_id = "goemqutiti"
random_id_suffix = false       # append nano timestamp to the client ID
username = "user"
password = "keyring:emqutiti-local/user"
ssl_tls = false                # enable TLS/SSL encryption
mqtt_version = ""              # protocol version 3, 4, or 5
connect_timeout = 0            # seconds to wait when connecting
keep_alive = 0                 # keep-alive interval in seconds
qos = 0                        # default Quality of Service level
auto_reconnect = false         # automatically reconnect when the link drops
reconnect_period = 0           # seconds between reconnect attempts
clean_start = false            # start a clean session (MQTT v5)
session_expiry_interval = 0    # session expiry time in seconds
receive_maximum = 0            # maximum inflight messages
maximum_packet_size = 0        # limit the packet size in bytes
topic_alias_maximum = 0        # maximum number of topic aliases
request_response_info = false  # request response information from the broker
request_problem_info = false   # request problem information from the broker
last_will_enabled = false      # enable a Last Will and Testament message
last_will_topic = ""
last_will_qos = 0
last_will_retain = false
last_will_payload = ""
```

Setting `random_id_suffix` to `true` appends a nano timestamp to the client ID
so multiple connections remain unique.

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

Additional notes for repository contributors are available in [AGENTS.md](AGENTS.md).
