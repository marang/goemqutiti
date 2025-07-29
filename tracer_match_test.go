package main

import "testing"

func TestMatch(t *testing.T) {
	cases := []struct {
		filter string
		topic  string
		want   bool
	}{
		{"foo/bar", "foo/bar", true},
		{"foo/+", "foo/bar", true},
		{"foo/+", "foo/bar/baz", false},
		{"foo/#", "foo/bar/baz", true},
		{"+/bar", "foo/bar", true},
		{"+/bar", "foo/baz", false},
		{"foo/#", "bar/foo", false},
	}
	for _, c := range cases {
		if got := tracerMatch(c.filter, c.topic); got != c.want {
			t.Errorf("Match(%q,%q)=%v want %v", c.filter, c.topic, got, c.want)
		}
	}
}
