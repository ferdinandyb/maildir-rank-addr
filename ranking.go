package main

import (
	"sort"
)

func getMostFrequent(names []string) string {
	return names[0]
}

func normalize(aD AddressData) AddressData {
	aD.Name = getMostFrequent(aD.Names)
	aD.Num = len(aD.Names)
	return aD
}

func calculateRanks(data map[string]AddressData) map[int]map[string]AddressData {
	type KeyValue struct {
		Key   string
		Value AddressData
	}
	classedData := map[int]map[string]AddressData{
		2: map[string]AddressData{},
		1: map[string]AddressData{},
		0: map[string]AddressData{},
	}
	for addr, value := range data {
		classedData[value.Class][addr] = normalize(value)
	}

	for class := 2; class >= 0; class-- {
		thisclass, _ := classedData[class]
		s := make([]KeyValue, 0, len(thisclass))
		for k, v := range thisclass {
			s = append(s, KeyValue{k, v})
		}
		sort.SliceStable(s, func(i, j int) bool {
			return s[i].Value.Num > s[j].Value.Num
		})
		for rank, kv := range s {
			thisval, _ := thisclass[kv.Key]
			thisval.FrequencyRank = rank
			thisclass[kv.Key] = thisval
		}
		sort.SliceStable(s, func(i, j int) bool {
			return s[i].Value.Date > s[j].Value.Date
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
