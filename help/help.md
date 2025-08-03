# Shortcuts

## Global

| Key | Action |
| --- | ------ |
| Ctrl+C | Copy selected entry |
| Ctrl+D | Exit the program |
| Ctrl+P | Manage payloads |
| Ctrl+T | Manage topics |
| Ctrl+R | Manage traces |
| Ctrl+F | Clear history filters |
| Ctrl+B | Open broker manager |
| Ctrl+S / Ctrl+Enter | Publish message |
| Ctrl+Shift+Up / Ctrl+Shift+Down | Resize panels |
| Ctrl+Up/Down or Ctrl+K/J | Scroll view |

## Navigation

- Esc navigates back
- Enter subscribes to the typed topic
- Tab/Shift+Tab cycle focus
- Use arrows or j/k to move through lists
- Ctrl+Up/Down scrolls the view
- Press '/' in history for a filter dialog with topic suggestions shown
  under the field; Tab or arrows cycle matches and Enter or space accepts
  the highlight. A text box lets you match message contents, and start
  and end fields default to the last hour on first open and stay blank
  once removed so you can search all time. Active filters appear above the
  history list and `Ctrl+F` resets them.
- All Ctrl shortcuts work even when an input is active

## Broker Manager

- 'x' disconnects the selected profile

## History View

- Shift selects ranges; Ctrl+A selects all
- 'a' archives selected messages
- Delete removes selected messages
- Ctrl+L toggles archived view
- Press '/' to filter messages

## Tips

- Set `EMQUTITI_DEFAULT_PASSWORD` to override profile passwords when not loading from env.
