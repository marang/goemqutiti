package history

import (
	"reflect"
	"testing"
	"time"
)

func TestParseHistoryQuery(t *testing.T) {
	cases := []struct {
		query   string
		topics  []string
		start   string
		end     string
		payload string
	}{
		{
			query:   "",
			topics:  nil,
			start:   "",
			end:     "",
			payload: "",
		},
		{
			query:   "topic=a start=2025-07-28T18:25:21Z hello world",
			topics:  []string{"a"},
			start:   "2025-07-28T18:25:21Z",
			end:     "",
			payload: "hello world",
		},
		{
			query:   "start=2025-07-28T18:25:21Z end=2025-07-28T20:00:00Z payload=foo",
			topics:  nil,
			start:   "2025-07-28T18:25:21Z",
			end:     "2025-07-28T20:00:00Z",
			payload: "foo",
		},
		{
			query:   "foo bar topic=a,b end=2025-07-28T21:00:00Z",
			topics:  []string{"a", "b"},
			start:   "",
			end:     "2025-07-28T21:00:00Z",
			payload: "foo bar",
		},
		{
			query:   "payload=first topic=c start=2025-07-28T18:25:21Z second",
			topics:  []string{"c"},
			start:   "2025-07-28T18:25:21Z",
			end:     "",
			payload: "first second",
		},
	}

	for _, c := range cases {
		gotTopics, gotStart, gotEnd, gotPayload := parseHistoryQuery(c.query)

		if !reflect.DeepEqual(gotTopics, c.topics) {
			t.Errorf("topics mismatch for %q: %v != %v", c.query, gotTopics, c.topics)
		}

		wantStart, _ := time.Parse(time.RFC3339, c.start)
		wantEnd, _ := time.Parse(time.RFC3339, c.end)

		if !gotStart.Equal(wantStart) {
			if !(gotStart.IsZero() && c.start == "") {
				t.Errorf("start time mismatch for %q: %v != %v", c.query, gotStart, wantStart)
			}
		}

		if !gotEnd.Equal(wantEnd) {
			if !(gotEnd.IsZero() && c.end == "") {
				t.Errorf("end time mismatch for %q: %v != %v", c.query, gotEnd, wantEnd)
			}
		}

		if gotPayload != c.payload {
			t.Errorf("payload mismatch for %q: %q != %q", c.query, gotPayload, c.payload)
		}
	}
}
