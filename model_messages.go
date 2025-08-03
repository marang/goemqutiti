package emqutiti

type payloadItem struct{ topic, payload string }

func (p payloadItem) FilterValue() string { return p.topic }
func (p payloadItem) Title() string       { return p.topic }
func (p payloadItem) Description() string { return p.payload }
