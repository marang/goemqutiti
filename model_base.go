package emqutiti

const (
	idTopics         = "topics"          // topics chip list
	idTopic          = "topic"           // topic input box
	idMessage        = "message"         // message input box
	idHistory        = "history"         // history list
	idConnList       = "conn-list"       // broker list
	idTopicsEnabled  = "topics-enabled"  // enabled topics pane
	idTopicsDisabled = "topics-disabled" // disabled topics pane
	idPayloadList    = "payload-list"    // payload manager list
	idTraceList      = "trace-list"      // traces manager list
	idHelp           = "help"            // help icon
)

type appMode int

const (
	modeClient appMode = iota
	modeConnections
	modeEditConnection
	modeConfirmDelete
	modeTopics
	modePayloads
	modeTracer
	modeEditTrace
	modeViewTrace
	modeImporter
	modeHistoryFilter
	modeHistoryDetail
	modeHelp
)

var focusByMode = map[appMode][]string{
	modeClient:         {idTopics, idTopic, idMessage, idHistory, idHelp},
	modeConnections:    {idConnList, idHelp},
	modeEditConnection: {idConnList, idHelp},
	modeConfirmDelete:  {},
	modeTopics:         {idTopicsEnabled, idTopicsDisabled, idHelp},
	modePayloads:       {idPayloadList, idHelp},
	modeTracer:         {idTraceList, idHelp},
	modeEditTrace:      {idHelp},
	modeViewTrace:      {idHelp},
	modeImporter:       {idHelp},
	modeHistoryFilter:  {idHelp},
	modeHistoryDetail:  {idHelp},
	modeHelp:           {idHelp},
}
