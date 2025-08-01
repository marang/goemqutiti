# GoEmqutiti

GoEmqutiti is a polished MQTT client for the terminal built on
[Bubble Tea](https://github.com/charmbracelet/bubbletea). Profiles live in
`~/.emqutiti/config.toml` so you can switch brokers with a few key presses. The
short demo below shows the app in action.

## Features

- Slick interface for publishing and subscribing
- Manage multiple brokers with one config file
- Credentials stored securely via the OS keyring
- Import CSV or XLS files with a friendly wizard
- Persistent history and trace recording, even headless

## Installation
### From Source
```bash
go install github.com/marang/goemqutiti@latest
```

### Arch Linux
```bash
yay -S emqutiti
```

## Usage

```bash
emqutiti
```

### Importing from CSV or XLS

Launch `emqutiti --import data.csv -p local` to map columns to JSON and publish them. The wizard supports dry runs and will remember settings in future versions.

Press `Ctrl+R` in the UI to manage recorded traces.

### Headless tracing

Use `emqutiti --trace myrun --topics "sensors/#" -p local` to capture messages without the UI. Traces are stored under `~/.emqutiti/data/<profile>/traces` and can be viewed in the application (run `emqutiti` and press `CTRL+R` in the app to view traces)

## Configuration
stored in `~/.emqutiti/config.toml` describing broker profiles. You can also create connections within the UI.

Minimal config example:

```toml
default_profile = "local"

[[profiles]]
name     = "local"
schema   = "tcp"
host     = "localhost"
port     = 1883
username = "user"
password = "keyring:emqutiti-local/user"
```

Tips:
- More options like TLS and session settings are available; see the `config` package for details.
- Set `random_id_suffix = true` for unique client IDs.
- Enable **Load from env** to read variables such as `GOEMQUTITI_LOCAL_BROKER_PASSWORD`.

### Shortcuts

| Action | Key |
| --- | --- |
| Open broker manager | `Ctrl+B` |
| Publish message | `Ctrl+S` or `Ctrl+Enter` |
| Manage topics | `Ctrl+T` |
| Manage payloads | `Ctrl+P` |
| Manage traces | `Ctrl+R` |
| Clear history filters | `Ctrl+F` |
| Copy selected entry | `Ctrl+C` |
| Scroll view | `Ctrl+Up` / `Ctrl+Down` |
| Exit the program | `Ctrl+D` |
| Resize panels | `Ctrl+Shift+Up` / `Ctrl+Shift+Down` |

Other keys: `Tab` and `Shift+Tab` cycle focus, `Enter` subscribes to the typed topic, `x` disconnects in the broker manager and `Esc` navigates back. Use ↑/↓ or `j`/`k` to move through lists, hold `Shift` for range selection in history. Press `/` in the history view for a filter dialog with topic suggestions (Tab or arrows cycle matches; `Enter` or space accepts the highlight). Start and end default to the last hour on first open and stay blank once removed, letting you search all time. Active filters show above the history list and `Ctrl+F` clears them. `Ctrl+Up`/`Ctrl+Down` scrolls the view. All `Ctrl` shortcuts work even when an input is active.

## License

This project is licensed under the terms of the MIT License. See [LICENSE](LICENSE) for details.

## Testing

Unit tests can be run with `go test ./...`. The example `ExampleSet_manual` in
`keyring_util_test.go` interacts with the real system keyring and is excluded
from automated runs. Execute it manually when a keyring is available.
Tests also cover configuration parsing and saved state persistence.

Before sending a pull request run `go vet ./...` along with the tests to catch
common mistakes.

Additional notes for repository contributors are available in [AGENTS.md](AGENTS.md).

## Development

### Creating documentation

#### Recording Demos for new features and howtos

Build the provided `Dockerfile.cast` and run the helper inside
an interactive container. The file is named this way so it won't
clash with other `Dockerfile`s in the project. Use `-it` to allocate
a TTY so asciinema can capture the session:

```bash
docker build -f docs/scripts/Dockerfile.cast -t emqutiti-caster .
docker run --rm -it \
  --network=host \
  -v "$PWD/..:/app/docs" \
  -v "$PWD/../scripts:/app/scripts" \
  emqutiti-caster \
  ./../scripts/record_casts.sh
```
You'll interact with the TUI inside the container just like running it locally.
