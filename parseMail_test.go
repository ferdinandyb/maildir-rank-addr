package main

import (
	"testing"
	"regexp"
)


func TestAssignClass(t *testing.T) {

	var useraddresses = []*regexp.Regexp {
		regexp.MustCompile("foo@bar.com"),
		regexp.MustCompile("name-.+@example.com"),
	}

	var tests = []struct {
		testname string
		in_field string
		in_sender string
		in_useraddresses []*regexp.Regexp
		want  int
	}{
		{"No useraddresses", "something", "foo@bar.com", []*regexp.Regexp {}, 2},
		{"Field is from", "from", "foo@bar.com", useraddresses, 0},
		{"Explicit matching in to", "to", "foo@bar.com", useraddresses, 2},
		{"Explicit matching in cc", "cc", "foo@bar.com", useraddresses, 1},
		{"Regex match in to", "to", "name-foo@example.com", useraddresses, 2},
		{"No matches", "to", "something@example.com", useraddresses, 0},
	}

	for _, tt := range tests {
	    t.Run(tt.testname, func(t *testing.T) {
		ans := assignClass(tt.in_field, tt.in_sender, tt.in_useraddresses)
		if ans != tt.want {
		    t.Errorf("got %d, want %d", ans, tt.want)
		}
	    })
	}
}
