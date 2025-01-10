package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddressbookOverride(t *testing.T) {
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
