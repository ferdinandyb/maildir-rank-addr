package main

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestE2EAddressbookOverride(t *testing.T) {
	addressbook := map[string]string{
		"foo@bar.com":           "override FOO",
		"something@example.com": "override EXAMPLE",
	}
	data := walkMaildirs([]string{"./testdata/endtoend"}, nil, nil, addressbook)
	classeddata := calculateRanks(data)

	tests := []struct {
		testname string
		class    int
		address  string
		want     string
	}{
		{"override 1", 2, "foo@bar.com", "override FOO"},
		{"override 2", 2, "something@example.com", "override EXAMPLE"},
	}

	for _, tt := range tests {
		t.Run(tt.testname, func(t *testing.T) {
			class, ok := classeddata[tt.class]
			addr, ok := class[tt.address]
			assert.True(t, ok)
			assert.Equal(t, tt.want, addr.Name)
		})
	}
}

func TestE2EClass(t *testing.T) {
	data := walkMaildirs(
		[]string{"./testdata/endtoend"},
		[]*regexp.Regexp{regexp.MustCompile(".+@myself.me")},
		nil,
		nil,
	)
	classeddata := calculateRanks(data)

	tests := []struct {
		testname string
		class    int
		address  string
	}{
		{"class 2 check 1", 2, "friend1@friends.com"},
		{"class 2 check 2", 2, "friend3@friends.com"},
		{"class 2 check 3", 2, "friend4@friends.com"},
		{"class 1 check 1", 1, "friend2@friends.com"},
		{"class 0 check 0", 0, "me@myself.me"},
		{"class 0 check 1", 0, "foo@bar.com"},
	}

	for _, tt := range tests {
		t.Run(tt.testname, func(t *testing.T) {
			assert.Contains(t, classeddata[tt.class], tt.address)
		})
	}
}

func TestE2ERankingRecency(t *testing.T) {
	data := walkMaildirs(
		[]string{"./testdata/endtoend"},
		[]*regexp.Regexp{regexp.MustCompile(".+@myself.me")},
		nil,
		nil,
	)
	classeddata := calculateRanks(data)

	tests := []struct {
		testname string
		lower    string
		higher   string
	}{
		{"class 2 check 1", "friend1@friends.com", "friend3@friends.com"},
	}

	for _, tt := range tests {
		t.Run(tt.testname, func(t *testing.T) {
			assert.Less(
				t,
				classeddata[2][tt.lower].RecencyRank,
				classeddata[2][tt.higher].RecencyRank,
			)
		})
	}
}
