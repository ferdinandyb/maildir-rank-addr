package main

import (
	"regexp"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestE2EAddressbookOverride(t *testing.T) {
	addressbook := map[string]string{
		"foo@bar.com":           "override FOO",
		"something@example.com": "override EXAMPLE",
	}
	data := walkMaildirs([]string{"./testdata/endtoend"}, nil, nil)
	classeddata := calculateRanks(data, addressbook, nil)

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

func TestE2ENormalization(t *testing.T) {
	data := walkMaildirs([]string{"./testdata/endtoend"}, nil, nil)
	classeddata := calculateRanks(data, nil, nil)

	tests := []struct {
		testname string
		class    int
		address  string
		want     string
	}{
		{"normalize 1", 2, "diacritics@hungary.hu", "ouooueau"},
	}

	for _, tt := range tests {
		t.Run(tt.testname, func(t *testing.T) {
			class, ok := classeddata[tt.class]
			addr, ok := class[tt.address]
			assert.True(t, ok)
			assert.Equal(t, tt.want, addr.NormalizedName)
		})
	}
}

func TestE2EClass(t *testing.T) {
	data := walkMaildirs(
		[]string{"./testdata/endtoend"},
		[]*regexp.Regexp{regexp.MustCompile(".+@myself.me")},
		nil,
	)
	classeddata := calculateRanks(data, nil, nil)

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
	)
	classeddata := calculateRanks(data, nil, nil)

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

func TestE2ERankingFrequency(t *testing.T) {
	data := walkMaildirs(
		[]string{"./testdata/endtoend"},
		[]*regexp.Regexp{regexp.MustCompile(".+@myself.me")},
		nil,
	)
	classeddata := calculateRanks(data, nil, nil)

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
				classeddata[2][tt.lower].FrequencyRank,
				classeddata[2][tt.higher].FrequencyRank,
			)
		})
	}
}

func TestE2EListTemplate(t *testing.T) {
	data := walkMaildirs(
		[]string{"./testdata/endtoend"},
		[]*regexp.Regexp{regexp.MustCompile(".+@myself.me")},
		nil,
	)
	listtmpl, _ := template.New("listtemplate").Parse("{{.ListName}}")
	classeddata := calculateRanks(data, nil, listtmpl)

	tests := []struct {
		testname string
		class    int
		normaddr string
		wanted   string
	}{
		{"standard", 0, "somelist-devel@lists.sourceforge.net", "All-purpose somelist list"},
		{"empty", 0, "empty@ietf.org", ""},
		{"verylong", 0, "tls@ietf.org", "This is the mailing list for the Transport Layer Security working group of the IETF."},
	}

	for _, tt := range tests {
		t.Run(tt.testname, func(t *testing.T) {
			assert.Equal(
				t,
				classeddata[tt.class][tt.normaddr].Name,
				tt.wanted,
			)
		})
	}
}

func TestE2EListTemplateDisable(t *testing.T) {
	data := walkMaildirs(
		[]string{"./testdata/endtoend"},
		[]*regexp.Regexp{regexp.MustCompile(".+@myself.me")},
		nil,
	)
	listtmpl, _ := template.New("listtemplate").Parse("DISABLELIST")
	classeddata := calculateRanks(data, nil, listtmpl)

	tests := []struct {
		testname string
		class    int
		normaddr string
		wanted   string
	}{
		{"standard", 0, "somelist-devel@lists.sourceforge.net", "Project Maintainer via somelist-devel"},
	}

	for _, tt := range tests {
		t.Run(tt.testname, func(t *testing.T) {
			assert.Equal(
				t,
				classeddata[tt.class][tt.normaddr].Name,
				tt.wanted,
			)
		})
	}
}
