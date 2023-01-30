package main

import (
	"os/exec"
	"regexp"
	"text/template"
)

type AddressData struct {
	Address       string
	Names         []string
	Class         int
	FrequencyRank int
	RecencyRank   int
	TotalRank     int
	ClassCount    [3]int
	ClassDate     [3]int64
	Name          string
}

type Config struct {
	maildir                  string
	outputpath               string
	addresses                []*regexp.Regexp
	template                 *template.Template
	customFilters            []*regexp.Regexp
	addressbookLookupCommand *exec.Cmd
	addressbookAddUnmatched  bool
}
