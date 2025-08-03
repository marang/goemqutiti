package emqutiti

// MessageModel defines the dependencies messageComponent requires from the root model.
type MessageModel interface {
	Width() int
	MessageHeight() int
	FocusedID() string
	OverlayHelp(view string) string
}

var _ MessageModel = (*model)(nil)
