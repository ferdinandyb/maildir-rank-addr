package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"text/template"
)

func saveData(
	classedData map[int]map[string]AddressData,
	path string,
	tmpl *template.Template,
	addressbook map[string]string,
	addUnmatched bool,
) {
	os.MkdirAll(filepath.Dir(path), os.ModePerm)
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	type KeyValue struct {
		Key   string
		Value AddressData
	}
	count := 0
	for class := 2; class >= 0; class-- {
		thisclass, _ := classedData[class]
		s := make([]KeyValue, 0, len(thisclass))
		for k, v := range thisclass {
			s = append(s, KeyValue{k, v})
		}
		sort.SliceStable(s, func(i, j int) bool {
			if s[i].Value.TotalRank == s[j].Value.TotalRank {
				return s[i].Value.Address < s[j].Value.Address
			} else {
				return s[i].Value.TotalRank < s[j].Value.TotalRank
			}
		})
		for _, kv := range s {
			count++
			tmpl.Execute(f, kv.Value)
		}
	}
	if addUnmatched {
		for ak, av := range addressbook {
			aD := AddressData{}
			aD.Address = ak
			aD.Name = av
			count++
			tmpl.Execute(f, aD)
		}
	}
	fmt.Println(count, " addresses written to ", path)
}
