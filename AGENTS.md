# Repo Guidelines

- Always run `gofmt -w` on modified Go files.
- Run `go vet ./...` and attempt `go test ./...` before committing.

## Agent Notes
The TUI runs fullscreen with colorful borders. Press `m` to open the connection manager to add, edit, or delete MQTT profiles. Passwords are stored securely using the system keyring.

## Test Info
`ExampleSet_manual` in `keyring_util_test.go` requires a real keyring. It does not run during `go test ./...` and can be executed manually if needed.
