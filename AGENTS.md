# Repo Guidelines

- Always run `gofmt -w` on modified Go files.
- Run `go vet ./...` and attempt `go test ./...` before committing.

## Agent Notes
The TUI runs fullscreen with colorful borders. Press `Ctrl+M` to open the connection manager to add, edit, or delete MQTT profiles. Passwords are stored securely using the system keyring.

### Recent Experience
- Keyboard shortcuts bound to plain letters can interfere with text entry. Use `Ctrl` combinations for global actions.
- Provide both keyboard and mouse interaction for lists and chips to keep the UI consistent.
- Favor multi-line text areas where users might paste formatted data.
- Always consider usability and look for ways to improve it.

## Test Info
`ExampleSet_manual` in `keyring_util_test.go` requires a real keyring. It does not run during `go test ./...` and can be executed manually if needed.
