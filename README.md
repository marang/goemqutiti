# GoEmqutiti

GoEmqutiti is a polished MQTT client for the terminal built on
[Bubble Tea](https://github.com/charmbracelet/bubbletea). Profiles live in
`~/.emqutiti/config.toml` so you can switch brokers with a few key presses. The
short demo below shows the app in action.

A demo session is stored in `docs/demo.cast`. Convert it with [asciinema-agg](https://github.com/asciinema/agg) if you want a GIF preview locally. Scripts under `scripts/` automate recording multiple demos; run `scripts/record_casts.sh` in an interactive terminal.
## Features

- Slick interface for publishing and subscribing
- Manage multiple brokers with one config file
- Credentials stored securely via the OS keyring
- Import CSV or XLS files with a friendly wizard
- Persistent history and trace recording, even headless

## Installation

```bash
go install github.com/marang/goemqutiti@latest
```

Clone and `go build` if you prefer working from source. Arch users can simply
install it from the AUR using `pacman -S emqutiti` (or your favorite helper).

## Usage

Run the built binary (or use `go run .`) to start the TUI application. On startup the broker manager is shown so you can select which profile to use. The manager can also be opened at any time with the `Ctrl+B`. Each broker entry shows the connection name on the first line and the current status (e.g. "connected" or "connecting") on the second. Profiles expose all common connection options.

```bash
./emqutiti
```

The client expects a configuration file at `~/.emqutiti/config.toml` describing broker profiles. You can also create connections within the UI.

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
- Secrets are stored in the OS keyring or provided via `MQTT_PASSWORD`.
- Enable **Load from env** to read variables such as `GOEMQUTITI_LOCAL_BROKER_PASSWORD`.

### Shortcuts

| Action | Key |
| --- | --- |
| Open broker manager | `Ctrl+B` |
| Publish message | `Ctrl+S` or `Ctrl+Enter` |
| Manage topics | `Ctrl+T` |
| Manage payloads | `Ctrl+P` |
| Manage traces | `Ctrl+R` |
| Copy selected entry | `Ctrl+C` |
| Exit the program | `Ctrl+D` |
| Resize panels | `Ctrl+Shift+Up` / `Ctrl+Shift+Down` |

Other keys: `Tab` and `Shift+Tab` cycle focus, `Enter` subscribes to the typed topic, `x` disconnects in the broker manager and `Esc` navigates back. Use ↑/↓ or `j`/`k` to move through lists, hold `Shift` for range selection in history. Press `/` in the history view to filter messages. All `Ctrl` shortcuts work even when an input is active.

### Importing from CSV or XLS

Launch `emqutiti --import data.csv -p local` to map columns to JSON and publish them. The wizard supports dry runs and will remember settings in future versions.

Press `Ctrl+R` in the UI to manage recorded traces.

### Headless tracing

Use `emqutiti --trace myrun --topics "sensors/#" -p local` to capture messages without the UI. Traces are stored under `~/.emqutiti/data/<profile>/traces`.

### Recording demos

1. Install [asciinema](https://github.com/asciinema/asciinema) and
   [asciinema-agg](https://github.com/asciinema/agg) (via
   `cargo install --git https://github.com/asciinema/agg`).
   `agg` needs a monospace font such as **DejaVu Sans Mono**. Install
   `fonts-dejavu-core` or a similar package if conversion fails.
2. Run `scripts/record_casts.sh` in an interactive terminal. The helper
   feeds prerecorded input through `script` to automate each demo.
3. Cast files are saved to `docs/`. Convert them to GIFs with `agg` if you want
   previews locally but avoid committing the generated images.

#### Without host dependencies

Build the provided `Dockerfile.cast` and run the helper inside
an interactive container. The file is named this way so it won't
clash with other `Dockerfile`s in the project. Use `-it` to allocate
a TTY so asciinema can capture the session:

```bash
docker build -f Dockerfile.cast -t emqutiti-caster .
docker run --rm -it --network=host -v "$PWD/docs:/app/docs" emqutiti-caster \
  ./scripts/record_casts.sh
```
You'll interact with the TUI inside the container just like running it locally.


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
