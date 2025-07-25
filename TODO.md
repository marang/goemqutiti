# TODO List for GoEmqutiti MQTT Client

This document tracks the tasks and goals for developing the GoEmqutiti MQTT client using Bubble Tea.

---

## **1. Responsive Layout**
- [ ] Modularize UI components into separate files:
  - [ ] `header.go` (Header section: title, connection status)
  - [ ] `connections.go` (Connection manager: list, CRUD operations)
  - [ ] `topic.go` (Topic input section)
  - [ ] `message.go` (Message input section)
  - [ ] `messages_log.go` (Messages log section)
  - [ ] `payloads.go` (Stored payloads section)
  - [ ] `status.go` (Status bar: focus, shortcuts, etc.)
- [ ] Implement responsiveness using `tea.WindowSizeMsg`.
- [ ] Use `lipgloss` for styling and layout constraints.
- [ ] Ensure sections stack vertically and resize gracefully.

---

## **2. Connection Manager**
### **Secure Storage**
- [x] Integrate Linux keyring using [`zalando/go-keyring`](https://github.com/zalando/go-keyring).
- [ ] Store sensitive fields (e.g., passwords) in the keyring.
- [ ] Update `config.toml` to reference keyring entries (e.g., `password = "keyring:<service>/<username>"`).
- [ ] Prompt user to unlock the keyring on application startup.
- [ ] Handle cases where the keyring is unavailable or inaccessible.

### **CRUD Operations**
- [ ] Add new connections with full MQTT configuration options.
- [ ] Edit existing connections.
- [ ] Delete connections.
- [ ] Load connections from the configuration file and keyring on startup.
- [ ] Save connections to the configuration file and keyring when modified.

### **UI Components**
- [ ] Display a selectable list of connections using `bubbles/list`.
- [ ] Highlight the active connection in the list.
- [ ] Show connection status (connected/disconnected).
- [ ] Provide a form for adding/editing connections using `bubbles/textinput`.

---

## **3. General Features**
- [ ] Implement keyboard shortcuts for navigation and actions.
- [ ] Add error handling for failed connections or invalid inputs.
- [ ] Persist all data securely between application runs.
- [ ] Support dynamic updates from the MQTT broker (real-time message logging).

---

## **4. Testing and Debugging**
- [ ] Test responsiveness across different terminal sizes.
- [ ] Test secure storage integration with the Linux keyring.
- [ ] Verify encryption/decryption logic (if still applicable).
- [ ] Debug any issues with loading/saving connections.

---

## **5. Future Enhancements**
- [ ] Add filtering or search functionality for large logs.
- [ ] Implement TLS/SSL certificate management (consider storing certificates in the keyring).
- [ ] Add support for Last Will and Testament (LWT) settings.
- [ ] Explore additional security features (e.g., biometric authentication).

---

## **Notes**
- Prioritize modularization and secure storage as the foundation for the application.
- Regularly update this TODO list as tasks are completed or new requirements emerge.

