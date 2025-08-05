package traces

// FileStore implements Store using on-disk state and the tracer message
// database. It provides the default application behaviour but can be
// replaced for testing.
type FileStore struct{}

func (FileStore) LoadTraces() map[string]TracerConfig { return loadTraces() }
func (FileStore) SaveTraces(data map[string]TracerConfig) error {
	return saveTraces(data)
}
func (FileStore) AddTrace(cfg TracerConfig) error { return addTrace(cfg) }
func (FileStore) RemoveTrace(key string) error    { return removeTrace(key) }
func (FileStore) Messages(profile, key string) ([]TracerMessage, error) {
	return tracerMessages(profile, key)
}
func (FileStore) HasData(profile, key string) (bool, error) {
	return tracerHasData(profile, key)
}
func (FileStore) ClearData(profile, key string) error {
	return tracerClearData(profile, key)
}
func (FileStore) LoadCounts(profile, key string, topics []string) (map[string]int, error) {
	return tracerLoadCounts(profile, key, topics)
}
