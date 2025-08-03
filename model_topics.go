package emqutiti

type topicItem struct {
	title      string
	subscribed bool
}

func (t topicItem) FilterValue() string { return t.title }
func (t topicItem) Title() string       { return t.title }
func (t topicItem) Description() string {
	if t.subscribed {
		return "subscribed"
	}
	return "unsubscribed"
}

type chipBound struct {
	xPos, yPos    int
	width, height int
}

type paneManager interface{ SetActivePane(int) }

type paneState struct {
	sel     int
	page    int
	m       paneManager
	index   int
	focused bool
}

func (p *paneState) Focus() {
	p.focused = true
	if p.m != nil {
		p.m.SetActivePane(p.index)
	}
}

func (p *paneState) Blur() { p.focused = false }

func (p paneState) IsFocused() bool { return p.focused }

func (p paneState) View() string { return "" }

type topicsPanes struct {
	subscribed   paneState
	unsubscribed paneState
	active       int
}
