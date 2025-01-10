package main

import (
	"sort"
	"strings"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type KeyValue struct {
	normaddr string
	addrdata AddressData
}

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

func sortByFrequency(s []KeyValue, class int) {
	sort.SliceStable(s, func(i, j int) bool {
		if s[i].addrdata.ClassCount[class] == s[j].addrdata.ClassCount[class] {
			return s[i].addrdata.Address < s[j].addrdata.Address
		} else {
			return s[i].addrdata.ClassCount[class] > s[j].addrdata.ClassCount[class]
		}
	})
}

func sortByRecency(s []KeyValue, class int) {
	sort.SliceStable(s, func(i, j int) bool {
		if s[i].addrdata.ClassDate[class] == s[j].addrdata.ClassDate[class] {
			return s[i].addrdata.Address < s[j].addrdata.Address
		} else {
			return s[i].addrdata.ClassDate[class] > s[j].addrdata.ClassDate[class]
		}
	})
}

func getClassRanks(addrmap map[string]AddressData, class int) map[string]AddressData {
	s := make([]KeyValue, 0, len(addrmap))
	for normaddr, addrdata := range addrmap {
		s = append(s, KeyValue{normaddr, addrdata})
	}

	sortByFrequency(s, class)
	for rank, kv := range s {
		addrdata, _ := addrmap[kv.normaddr]
		addrdata.FrequencyRank = rank
		addrmap[kv.normaddr] = addrdata
	}

	sortByRecency(s, class)
	for rank, kv := range s {
		addrdata, _ := addrmap[kv.normaddr]
		addrdata.RecencyRank = rank
		addrmap[kv.normaddr] = addrdata
	}

	for normaddr, addrdata := range addrmap {
		addrdata.TotalRank = addrdata.FrequencyRank + addrdata.RecencyRank
		addrmap[normaddr] = addrdata
	}

	return addrmap
}

func calculateRanks(data map[string]AddressData) map[int]map[string]AddressData {
	classedData := map[int]map[string]AddressData{
		2: {},
		1: {},
		0: {},
	}
	for normaddr, aD := range data {
		classedData[aD.Class][normaddr] = normalizeAddressNames(aD)
	}

	for class := 2; class >= 0; class-- {
		classedData[class] = getClassRanks(classedData[class], class)
	}
	return classedData
}
