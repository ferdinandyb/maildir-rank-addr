package main

import (
	"sort"
	"strings"
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

func normalizeAddressNames(aD AddressData) AddressData {
	aD.Name = getMostFrequent(aD.Names)
	return aD
}

func calculateRanks(data map[string]AddressData) map[int]map[string]AddressData {
	type KeyValue struct {
		Key   string
		Value AddressData
	}
	classedData := map[int]map[string]AddressData{
		2: {},
		1: {},
		0: {},
	}
	for addr, value := range data {
		classedData[value.Class][addr] = normalizeAddressNames(value)
	}

	for class := 2; class >= 0; class-- {
		thisclass, _ := classedData[class]
		s := make([]KeyValue, 0, len(thisclass))
		for k, v := range thisclass {
			s = append(s, KeyValue{k, v})
		}
		sort.SliceStable(s, func(i, j int) bool {
			if s[i].Value.ClassCount[class] == s[j].Value.ClassCount[class] {
				return s[i].Value.Address < s[j].Value.Address
			} else {
				return s[i].Value.ClassCount[class] > s[j].Value.ClassCount[class]
			}
		})
		for rank, kv := range s {
			thisval, _ := thisclass[kv.Key]
			thisval.FrequencyRank = rank
			thisclass[kv.Key] = thisval
		}
		sort.SliceStable(s, func(i, j int) bool {
			if s[i].Value.ClassDate[class] == s[j].Value.ClassDate[class] {
				return s[i].Value.Address < s[j].Value.Address
			} else {
				return s[i].Value.ClassDate[class] > s[j].Value.ClassDate[class]
			}
		})
		for rank, kv := range s {
			thisval, _ := thisclass[kv.Key]
			thisval.RecencyRank = rank
			thisclass[kv.Key] = thisval
		}
		for k, v := range thisclass {
			v.TotalRank = v.FrequencyRank + v.RecencyRank
			thisclass[k] = v
		}

		classedData[class] = thisclass
	}
	return classedData
}
