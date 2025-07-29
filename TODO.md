# TODO List for GoEmqutiti MQTT Client

This document tracks the tasks and goals for developing the GoEmqutiti MQTT client using Bubble Tea.

---

## **1. Responsive Layout**
- [ ] Modularize UI components into separate files:
  - [ ] `header.go` (Header section: title, connection status)
  - [ ] `connections.go` (Broker manager: list, CRUD operations)
  - [ ] `topic.go` (Topic input section)
  - [ ] `message.go` (Message input section)
  - [ ] `messages_log.go` (Messages log section)
  - [ ] `payloads.go` (Stored payloads section)
  - [ ] `status.go` (Status bar: focus, shortcuts, etc.)
  - [x] UI code split into multiple files for easier maintenance
  - [ ] Implement responsiveness using `tea.WindowSizeMsg`.
  - [x] Use `lipgloss` for styling and layout constraints.
- [ ] Ensure sections stack vertically and resize gracefully.

---

## **2. Connection Manager**
### **Secure Storage**
- [x] Integrate Linux keyring using [`zalando/go-keyring`](https://github.com/zalando/go-keyring).
- [x] Store sensitive fields (e.g., passwords) in the keyring.
 - [x] Update `config.toml` to reference keyring entries (e.g., `password = "keyring:<service>/<username>"`).
- [x] Prompt user to unlock the keyring on application startup (handled
  automatically by the OS keyring).
- [x] Handle cases where the keyring is unavailable or inaccessible
  using the `MQTT_PASSWORD` environment variable (one connection only for
  now).

### **CRUD Operations**
- [x] Add new brokers with full MQTT configuration options.
- [x] Edit existing brokers.
- [x] Delete brokers.
- [x] Load brokers from the configuration file and keyring on startup.
- [x] Save brokers to the configuration file and keyring when modified.

### **UI Components**
- [x] Display a selectable list of brokers using `bubbles/list`.
- [x] Provide a menu option to open the broker manager.
 - [x] Highlight the active connection in the list.
- [x] Show connection status (connected/disconnected).
- [x] Provide a form for adding/editing brokers using `bubbles/textinput`.
- [x] Support advanced connection options (keep alive, clean session, TLS, LWT, etc.).
- [x] Option to append a random client ID suffix.

---

## **3. General Features**
 - [x] Implement keyboard shortcuts for navigation and actions. (Ctrl+S or Ctrl+Enter to publish messages)
- [x] Add error handling for failed connections or invalid inputs.
- [x] Display configuration load warnings in the history log.
 - [x] Persist brokers, topics, payloads, and messages between runs using BadgerDB.
- [x] Support dynamic updates from the MQTT broker (real-time message logging).
- [x] Import and publish messages from CSV or XLS files using command-line flags.

---

## **4. Testing and Debugging**
 - [ ] Test responsiveness across different terminal sizes.
 - [x] Test secure storage integration with the Linux keyring.
 - [x] Debug any issues with loading/saving connections.
 - [x] Add unit tests for config parsing and state persistence.

---

## **5. Future Enhancements**
 - [x] Add advanced filtering or search functionality for large logs. History view now accepts `topic`, `start`, `end`, and payload filters.
 - [ ] Implement TLS/SSL certificate management (consider storing certificate paths securely).
 - [x] Add support for Last Will and Testament (LWT) settings.
 - [ ] Persist import wizard settings for reuse.

---

## **Notes**
- Prioritize modularization and secure storage as the foundation for the application.
- Regularly update this TODO list as tasks are completed or new requirements emerge.
- Keep the documentation in `README.md`, `TODO.md`, and `AGENTS.md` aligned with the current implementation.
- Ensure `Ctrl+D` quits the TUI from every screen.
- The importer code is part of the main package and runs via the `--import`/`-i` flag.

