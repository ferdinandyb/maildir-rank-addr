package main

import (
	"regexp"
)

func mergeSources(data map[string]AddressData, dataNew map[string]AddressData) map[string]AddressData {
	if len(data) == 0 {
		return dataNew
	}
	// Merge dataNew into data
	for str, addr := range dataNew {
		orig, ok := data[str]
		if !ok {
			data[str] = addr
		} else {
			orig.Names = append(orig.Names, addr.Names...)
			if addr.Class > orig.Class {
				orig.Class = addr.Class
			}
			for i := range orig.ClassCount {
				orig.ClassCount[i] += addr.ClassCount[i]
			}
			for i := range orig.ClassDate {
				if addr.ClassDate[i] > orig.ClassDate[i] {
					orig.ClassDate[i] = addr.ClassDate[i]
				}
			}
			data[str] = orig
		}
	}
	return data
}

func walkSources(
	maildirs []string,
	useraddresses []*regexp.Regexp,
	customFilters []*regexp.Regexp,
) map[string]AddressData {
	data := make(map[string]AddressData)
	for _, maildir := range maildirs {
		dataNew := walkMaildir(maildir, useraddresses, customFilters)
		data = mergeSources(data, dataNew)
	}
	return data
}
