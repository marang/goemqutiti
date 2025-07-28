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

Run the built binary (or use `go run .`) to start the TUI application. On startup the broker manager is shown so you can select which profile to use. The manager can also be opened at any time with the `Ctrl+B`. Each broker entry shows the connection name on the first line and the current status (e.g. "connected" or "connecting") on the second. Profiles expose all common connection options inspired by the EMQX MQTT client:

```bash
./goemqutiti
```

The client expects a configuration file at `~/.emqutiti/config.toml` describing broker profiles. 
There is also an option to create connections within the client (recommended).

A sample configuration with all options looks like:

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

Each broker profile can optionally load all of its settings from environment variables instead of `config.toml`. Enable **Load from env** when editing a profile and set variables using the pattern `GOEMQUTITI_<NAME>_<FIELD>`. The `<NAME>` portion is derived from the profile name: letters are upper‑cased and any other characters become underscores. For example, a profile named `local broker` would use variables like `GOEMQUTITI_LOCAL_BROKER_PASSWORD`.

In the interface:

- **Tab** cycles focus between the topic input, message editor, and topic chips.
- **Enter** subscribes to a topic when the topic field is focused.
- **Ctrl+S** publishes the message currently in the editor.
- **Ctrl+Enter** also publishes the current message.
- **Ctrl+B** opens the broker manager where you can add, edit or delete MQTT profiles.
- **x** disconnects from the current broker in the broker manager.
- **Ctrl+T** manages subscribed topics.
- **Ctrl+P** manages stored payloads.
- **Ctrl+C** copies the currently selected history entry.
- Press `/` while the history is focused to filter messages. Queries support
  `topic=<list>` comma separated, `start=<RFC3339 time>`, `end=<RFC3339 time>`
  and free text to match payloads. Example:
  `topic=sensors/start start=2024-01-01T00:00:00Z payload=error`.
  The history log is stored in BadgerDB under
  `~/.emqutiti/history/<profile>` so messages remain searchable per profile.
- **Esc** navigates back within menus without quitting.
- **Ctrl+D** exits the program.
- Left-click a topic chip to toggle it and middle-click to remove it.
- Active topics are automatically subscribed when you connect to a broker.
- Connection attempts and any errors are shown at the top of the interface and recorded in the history log.
- Warnings while loading connection profiles also appear in the history log.
- Clicking on any pane or input field focuses it.

All `Ctrl` shortcuts are global, so they work even when an input field is active.

### Importing from CSV or XLS

Run the program with `--import` (or `-i`) to launch an interactive wizard that guides you
through selecting a file, mapping column names, defining the topic template and
publishing the messages. During the mapping step each CSV column appears on the
left with an editable field on the right so you can rename it for the JSON
payload. Leaving a mapping blank keeps the original column name. Providing a
path pre-selects the file in the wizard:

```bash
./goemqutiti -i data.csv -p local
```

Each row becomes a JSON object with properties derived from the mapped column
names. Enter nested names using `parent.child` syntax when mapping columns.
The topic template screen shows the original column names as `{field}`
placeholders and focuses the input automatically so you can start typing right
away.
After showing a preview of the first few messages you can perform a dry run or
publish them to the broker. A dry run lists the resulting topics and JSON
payloads so you can verify them; press `Ctrl+P` from the results screen to go
back. Publishing shows a progress bar along with a random sample of recently
published messages. The sample size grows with the total number of rows
(roughly the square root, capped at twenty) and the progress view stays on
screen when complete so you can see how many messages were sent. The wizard
expands to the full terminal width and shows samples inside the same green box
used for the live history. Use `Ctrl+N` and `Ctrl+P` to move forward and back
through the wizard steps. When many messages are shown you can scroll the
publish and dry run results with the arrow keys. Those results reuse the same
history-style box from the main view so the list has a fixed height and a
scrollable bar on the left.

Future versions may store import settings for quick reuse.

## License

This project is licensed under the terms of the MIT License. See [LICENSE](LICENSE) for details.

## Testing

Unit tests can be run with `go test ./...`. The example `ExampleSet_manual` in
`keyring_util_test.go` interacts with the real system keyring and is excluded
from automated runs. Execute it manually when a keyring is available.
Tests also cover configuration parsing and saved state persistence.

Additional notes for repository contributors are available in [AGENTS.md](AGENTS.md).
