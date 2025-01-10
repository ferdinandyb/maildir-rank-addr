package main

import (
	"sort"
	"strings"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func getMostFrequent(names []string) string {
	counter := make(map[string]int)
	for _, name := range names {
		if _, ok := counter[name]; ok {
			counter[name] = counter[name] + 1
		} else if strings.TrimSpace(name) != "" {
			counter[name] = 1
		}
	}
	maxcount := 0
	lastname := ""
	for name, counter := range counter {
		if counter > maxcount {
			lastname = name
			maxcount = counter
		} else if counter == maxcount {
			if name < lastname {
				lastname = name
			}
		}
	}
	lastname = strings.TrimSpace(lastname)
	lastname = strings.Replace(lastname, "\"", "", -1)
	return lastname
}

func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
}

func normalizeAddressNames(aD AddressData) AddressData {
	if aD.Name != "" {
		return aD
	}
	aD.Name = getMostFrequent(aD.Names)
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	normStr, _, _ := transform.String(t, aD.Name)
	aD.NormalizedName = normStr
	return aD
}

func calculateRanks(data map[string]AddressData) map[int]map[string]AddressData {
	type KeyValue struct {
		normaddr string
		addrdata AddressData
	}
	classedData := map[int]map[string]AddressData{
		2: {},
		1: {},
		0: {},
	}
	for normaddr, aD := range data {
		classedData[aD.Class][normaddr] = normalizeAddressNames(aD)
	}

	for class := 2; class >= 0; class-- {
		thisclass, _ := classedData[class]
		s := make([]KeyValue, 0, len(thisclass))
		for normaddr, addrdata := range thisclass {
			s = append(s, KeyValue{normaddr, addrdata})
		}
		sort.SliceStable(s, func(i, j int) bool {
			if s[i].addrdata.ClassCount[class] == s[j].addrdata.ClassCount[class] {
				return s[i].addrdata.Address < s[j].addrdata.Address
			} else {
				return s[i].addrdata.ClassCount[class] > s[j].addrdata.ClassCount[class]
			}
		})
		for rank, kv := range s {
			thisval, _ := thisclass[kv.normaddr]
			thisval.FrequencyRank = rank
			thisclass[kv.normaddr] = thisval
		}
		sort.SliceStable(s, func(i, j int) bool {
			if s[i].addrdata.ClassDate[class] == s[j].addrdata.ClassDate[class] {
				return s[i].addrdata.Address < s[j].addrdata.Address
			} else {
				return s[i].addrdata.ClassDate[class] > s[j].addrdata.ClassDate[class]
			}
		})
		for rank, kv := range s {
			thisval, _ := thisclass[kv.normaddr]
			thisval.RecencyRank = rank
			thisclass[kv.normaddr] = thisval
		}
		for normaddr, addrdata := range thisclass {
			addrdata.TotalRank = addrdata.FrequencyRank + addrdata.RecencyRank
			thisclass[normaddr] = addrdata
		}

		classedData[class] = thisclass
	}
	return classedData
}
