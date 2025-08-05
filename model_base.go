package emqutiti

import (
	"github.com/marang/emqutiti/constants"
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
	idTopicsSubscribed   = "topics-subscribed"   // subscribed topics pane
	idTopicsUnsubscribed = "topics-unsubscribed" // unsubscribed topics pane
	idPayloadList        = payloads.IDList       // payload manager list
	idHelp               = help.ID               // help icon
)

var focusByMode = map[constants.AppMode][]string{
	constants.ModeClient:         {idTopics, idTopic, idMessage, idHistory, idHelp},
	constants.ModeConnections:    {constants.IDConnList, idHelp},
	constants.ModeEditConnection: {constants.IDConnList, idHelp},
	constants.ModeConfirmDelete:  {},
	constants.ModeTopics:         {idTopicsSubscribed, idTopicsUnsubscribed, idHelp},
	constants.ModePayloads:       {idPayloadList, idHelp},
	constants.ModeTracer:         {traces.IDList, idHelp},
	constants.ModeEditTrace:      {traces.IDForm, idHelp},
	constants.ModeViewTrace:      {idHelp},
	constants.ModeImporter:       {idHelp},
	constants.ModeHistoryFilter:  {idHelp},
	constants.ModeHistoryDetail:  {idHelp},
	constants.ModeHelp:           {idHelp},
}
