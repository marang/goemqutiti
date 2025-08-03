package emqutiti

// TraceStore defines persistence and messaging operations for traces.
type TraceStore interface {
	LoadTraces() map[string]TracerConfig
	SaveTraces(map[string]TracerConfig)
	AddTrace(TracerConfig)
	RemoveTrace(string)
	Messages(profile, key string) ([]TracerMessage, error)
	HasData(profile, key string) (bool, error)
	ClearData(profile, key string) error
	LoadCounts(profile, key string, topics []string) (map[string]int, error)
}

// fileTraceStore implements TraceStore using on-disk state and the
// tracer message database. It provides the default application
// behaviour but can be replaced for testing.
type fileTraceStore struct{}

func (fileTraceStore) LoadTraces() map[string]TracerConfig     { return loadTraces() }
func (fileTraceStore) SaveTraces(data map[string]TracerConfig) { saveTraces(data) }
func (fileTraceStore) AddTrace(cfg TracerConfig)               { addTrace(cfg) }
func (fileTraceStore) RemoveTrace(key string)                  { removeTrace(key) }
func (fileTraceStore) Messages(profile, key string) ([]TracerMessage, error) {
	return tracerMessages(profile, key)
}
func (fileTraceStore) HasData(profile, key string) (bool, error) {
	return tracerHasData(profile, key)
}
func (fileTraceStore) ClearData(profile, key string) error {
	return tracerClearData(profile, key)
}
func (fileTraceStore) LoadCounts(profile, key string, topics []string) (map[string]int, error) {
	return tracerLoadCounts(profile, key, topics)
}

// tracesStore exposes the default TraceStore implementation.
func (m *model) tracesStore() TraceStore { return fileTraceStore{} }
