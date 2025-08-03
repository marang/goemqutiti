package emqutiti

// PayloadsAPI exposes payload management behavior to the rest of the application.
type PayloadsAPI interface {
	Add(topic, payload string)
	Items() []payloadItem
	SetItems([]payloadItem)
	Snapshot() []PayloadSnapshot
	SetSnapshot([]PayloadSnapshot)
	Clear()
}

// payloadsModel defines the dependencies payloadsComponent requires from the model.
type payloadsModel interface {
	navigator
	FocusedID() string
	ResetElemPos()
	SetElemPos(id string, pos int)
	OverlayHelp(string) string
}

var _ payloadsModel = (*model)(nil)
var _ PayloadsAPI = (*payloadsComponent)(nil)
