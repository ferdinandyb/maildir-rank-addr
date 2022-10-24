package main

type AddressData struct {
	Address       string
	Names         []string
	Class         int
	FrequencyRank int
	RecencyRank   int
	TotalRank     int
	Date          int64
	Num           int
	Name          string
}

type Config struct {
	maildir    string
	outputpath string
	addresses  []string
	template   string
}
