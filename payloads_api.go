package emqutiti

// payloadsModel defines the dependencies payloadsComponent requires from the model.
type payloadsModel interface {
	navigator
	FocusedID() string
	ResetElemPos()
	SetElemPos(id string, pos int)
	OverlayHelp(string) string
}

var _ payloadsModel = (*model)(nil)
