# Repo Guidelines

- Always run `gofmt -w` on modified Go files.
- Run `go vet ./...` and attempt `go test ./...` before committing.
- Run `go mod tidy` whenever dependencies change.
- Avoid committing binary artifacts such as GIFs. Generate them locally from `.cast` files instead.
- Keep commit messages short and descriptive; wrap lines at 72 characters.
- Summarize changes and test results in pull request descriptions.
- Keep documentation snappy. Use bullet lists and short sections so the README
  is easy to skim. Mention the main keyboard shortcuts.

## Agent Notes
The TUI runs fullscreen with colorful borders. Press `Ctrl+B` to open the broker manager to add, edit, or delete MQTT profiles. Passwords are stored securely using the system keyring. Publish messages with `Ctrl+S` or `Ctrl+Enter` when the message field is focused. Use the `--import`/`-i` flag to launch an interactive wizard for CSV or XLS bulk publishing and select a connection with `--profile` or `-p`. The wizard lets you rename columns when mapping them to JSON fields. Leaving a mapping blank keeps the original column name. The importer code lives in the main package and runs via these flags.
Press `Ctrl+D` from any screen to exit the program.
Scroll with `Ctrl+Up`/`Ctrl+Down` or `Ctrl+K`/`Ctrl+J`. In history,
`a` archives messages and `Delete` removes them.

### Recent Experience
- Keyboard shortcuts bound to plain letters can interfere with text entry. Use `Ctrl` combinations for global actions.
- Provide both keyboard and mouse interaction for lists and chips to keep the UI consistent.
- Favor multi-line text areas where users might paste formatted data.
- Always consider usability and look for ways to improve it.

### UI Guidelines
- Use the `LegendBox` helper for all boxed sections.
- The box helpers have been simplified into a single `LegendBox` function that
  accepts a border color and optional height. Use this function directly rather
  than maintaining multiple wrapper variants.
- Highlight the selected box using the focused style (pink).
- Present keyboard shortcuts consistently across views and ensure they behave the same everywhere.
- Keep functions small and comment any exported ones for clarity.

### Form Utilities
- Shared form behavior lives in `ui/form*.go`.
- Embed the `ui.Form` type to manage focus with `CycleFocus` and `ApplyFocus`.
- `ui.NewTextField`, `ui.NewSelectField`, `ui.NewSuggestField`, and `ui.NewCheckField`
  build common inputs.
- These helpers optionally call `setReadOnly` when values come from the
  environment.
- New UI code should reuse this logic instead of writing custom forms.

## Test Info
`ExampleSet_manual` in `keyring_util_test.go` requires a real keyring. It does not run during `go test ./...` and can be executed manually if needed.

## Maintenance
Keep `README.md`, `TODO.md`, `AGENTS.md`, and `docs/help.md` in sync when changes are made to the project or development workflow.

## Dependencies
When adding or updating third-party packages, always consult the latest
documentation for each dependency to ensure deprecated APIs are avoided.
Replace outdated calls with the recommended alternatives before committing
changes.

## Contribution Best Practices
- Create topic branches off `main` and keep pull requests focused.
- Describe the problem and solution clearly in commit messages.
- Keep commits small and avoid mixing unrelated changes.
- Record TUI demos with `asciinema` and keep the `.cast` files under `docs/`.
- Generate GIF previews locally using `asciinema-agg` but do not commit them.
- Run `agg docs/demo.cast docs/demo.gif` to regenerate previews when needed.
- Example `.exp` scripts in `docs/scripts/` automate recording.
- See the README for using `Dockerfile.cast` if you prefer not to install asciinema.
