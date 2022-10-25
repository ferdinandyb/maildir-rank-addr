package main

type AddressData struct {
	Address       string
	Names         []string
	Class         int
	FrequencyRank int
	RecencyRank   int
	TotalRank     int
	ClassCount    map[int]int
	ClassDate     map[int]int64
	Name          string
}

type Config struct {
	maildir    string
	outputpath string
	addresses  []string
	template   string
}
