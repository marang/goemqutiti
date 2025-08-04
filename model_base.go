package emqutiti

import (
	"github.com/marang/emqutiti/help"
	"github.com/marang/emqutiti/message"
	"github.com/marang/emqutiti/payloads"
	"github.com/marang/emqutiti/traces"
)

const (
	idTopics             = "topics"              // topics chip list
	idTopic              = "topic"               // topic input box
	idMessage            = message.ID            // message input box
	idHistory            = "history"             // history list
	idConnList           = "conn-list"           // broker list
	idTopicsSubscribed   = "topics-subscribed"   // subscribed topics pane
	idTopicsUnsubscribed = "topics-unsubscribed" // unsubscribed topics pane
	idPayloadList        = payloads.IDList       // payload manager list
	idHelp               = help.ID               // help icon
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
	modeTopics:         {idTopicsSubscribed, idTopicsUnsubscribed, idHelp},
	modePayloads:       {idPayloadList, idHelp},
	modeTracer:         {traces.IDList, idHelp},
	modeEditTrace:      {idHelp},
	modeViewTrace:      {idHelp},
	modeImporter:       {idHelp},
	modeHistoryFilter:  {idHelp},
	modeHistoryDetail:  {idHelp},
	modeHelp:           {idHelp},
}
