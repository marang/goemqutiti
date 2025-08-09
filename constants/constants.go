package constants

// AppMode defines application modes.
type AppMode int

const (
	ModeClient AppMode = iota
	ModeConnections
	ModeEditConnection
	ModeConfirmDelete
	ModeTopics
	ModePayloads
	ModeTracer
	ModeEditTrace
	ModeViewTrace
	ModeTraceFilter
	ModeImporter
	ModeHistoryFilter
	ModeHistoryDetail
	ModeHelp
	ModeLogs
)

// ID constants for shared elements.
const (
	IDConnList = "conn-list"
)
