package traces

import (
	"sort"

	"github.com/charmbracelet/bubbles/list"
)

// Init prepares the initial tracing state.
func Init() State {
	traceList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	traceList.DisableQuitKeybindings()
	traceList.SetShowTitle(false)
	tracesCfg := loadTraces()
	var traceItems []list.Item
	var traceData []*traceItem
	keys := make([]string, 0, len(tracesCfg))
	for k := range tracesCfg {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		it := &traceItem{key: k, cfg: tracesCfg[k]}
		traceItems = append(traceItems, it)
		traceData = append(traceData, it)
	}
	traceList.SetItems(traceItems)
	ts := State{
		list:  traceList,
		items: traceData,
		form:  nil,
	}
	return ts
}
