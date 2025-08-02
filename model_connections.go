package main

type connectionItem struct {
	title  string
	status string
	detail string
}

func (c connectionItem) FilterValue() string { return c.title }
func (c connectionItem) Title() string       { return c.title }
func (c connectionItem) Description() string {
	if c.detail != "" {
		return c.status + "\n" + c.detail
	}
	return c.status
}

type connectionData struct {
	Topics   []topicItem
	Payloads []payloadItem
}

type connectionsState struct {
	connection  string
	active      string
	manager     Connections
	form        *connectionForm
	deleteIndex int
	statusChan  chan string
	saved       map[string]connectionData
}
