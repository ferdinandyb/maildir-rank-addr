package main

import (
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"

	_ "github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
)

func messageParser(
	paths chan string,
	headers chan<- *mail.Header,
) {
	for path := range paths {
		f, err := os.Open(path)
		if err != nil {
			fmt.Println(err)
			continue
		}
		r, err := mail.CreateReader(f)
		if err != nil {
			fmt.Println(err)
			f.Close()
			continue
		}
		h := &mail.Header{Header: r.Header.Header}
		f.Close()
		headers <- h
	}
}

func assignClass(
	field string,
	sender string,
	useraddresses []*regexp.Regexp,
) int {
	if len(useraddresses) == 0 {
		return 2
	}
	if field == "from" {
		return 0
	}
	for _, addr := range useraddresses {
		if addr.MatchString(sender) {
			switch field {
			case "to", "bcc":
				return 2
			case "cc":
				return 1
			}
		}
	}
	return 0
}

func filterAddress(address string, customFilters []*regexp.Regexp) bool {
	_, err := mail.ParseAddress(address)
	if err != nil {
		return true
	}
	FILTERLIST := []string{
		"do-not-reply",
		"donotreply",
		"no-reply",
		"bounce",
		"noreply",
		"no.reply",
		"no_reply",
		"nevalaszolj",
		"nincsvalasz",
	}
	firstpart := strings.Split(address, "@")[0]
	for _, filt := range FILTERLIST {
		if strings.Contains(firstpart, filt) {
			return true
		}
	}
	for _, filt := range customFilters {
		if filt.MatchString(address) {
			return true
		}
	}
	return false
}

func processHeaders(
	headers <-chan *mail.Header,
	retvalchan chan map[string]AddressData,
	useraddresses []*regexp.Regexp,
	customFilters []*regexp.Regexp,
	addressbook map[string]string,
) {
	count := 0
	retval := make(map[string]AddressData)
	fields := [4]string{"to", "cc", "bcc", "from"}
	for header := range headers {
		count++
		time, err := header.Date()
		if err != nil {
			continue
		}

		senderaddress, err := header.AddressList("from")
		var sender string

		if len(senderaddress) > 0 {
			sender = strings.ToLower(senderaddress[0].Address)
		} else {
			sender = ""
		}
		if err != nil {
			continue
		}

		for _, field := range fields {
			header, err := header.AddressList(field)
			if err != nil {
				continue
			}
			for _, address := range header {
				normaddr := strings.ToLower(address.Address)
				if filterAddress(normaddr, customFilters) {
					continue
				}
				class := assignClass(
					field,
					sender,
					useraddresses,
				)
				if addressdata, ok := retval[normaddr]; ok {
					if addressdata.Name == "" {
						dec := new(mime.WordDecoder)
						name, err := dec.DecodeHeader(address.Name)
						if err != nil {
							continue
						}
						if (strings.ToLower(name) != normaddr) && (strings.ToLower(name) != "") {
							addressdata.Names = append(addressdata.Names, name)
						}
					}
					if addressdata.Class < class {
						addressdata.Class = class
					}
					if addressdata.ClassDate[class] < time.Unix() {
						addressdata.ClassDate[class] = time.Unix()
					}
					addressdata.ClassCount[class]++
					retval[normaddr] = addressdata
				} else {
					addressdata := AddressData{}
					addressbookname := addressbook[normaddr]
					if addressbookname == "" {
						dec := new(mime.WordDecoder)
						name, err := dec.DecodeHeader(address.Name)
						if err != nil {
							continue
						}
						if (strings.ToLower(name) != normaddr) && (strings.ToLower(name) != "") {
							addressdata.Names = append(addressdata.Names, name)
						}
					} else {
						addressdata.Name = addressbookname
						delete(addressbook, normaddr)
					}
					addressdata.Address = normaddr
					addressdata.Class = class
					addressdata.ClassDate = [3]int64{0, 0, 0}
					addressdata.ClassDate[class] = time.Unix()
					addressdata.ClassCount = [3]int{0, 0, 0}
					addressdata.ClassCount[class] = 1
					retval[normaddr] = addressdata
				}
			}

		}

	}
	fmt.Println("Read ", count, " messages")
	retvalchan <- retval
	close(retvalchan)
}

func walkMaildir(path string, useraddresses []*regexp.Regexp, customFilters []*regexp.Regexp, addressbook map[string]string) map[string]AddressData {
	headers := make(chan *mail.Header)
	messagePaths := make(chan string, 4096)

	var wg sync.WaitGroup
	for i := 0; i < 2*runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			messageParser(messagePaths, headers)
		}()
	}

	retvalchan := make(chan map[string]AddressData)
	go processHeaders(headers, retvalchan, useraddresses, customFilters, addressbook)

	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}

		if info.IsDir() {
			return nil
		}
		switch filepath.Base(filepath.Dir(path)) {
		case "new", "cur":
			messagePaths <- path
		}
		return nil
	})
	close(messagePaths)

	wg.Wait()
	close(headers)

	return <-retvalchan
}

func walkMaildirs(paths []string, useraddresses []*regexp.Regexp, customFilters []*regexp.Regexp, addressbook map[string]string) map[string]AddressData {
	data := make(map[string]AddressData)
	for _, maildir := range paths {
		dataNew := walkMaildir(maildir, useraddresses, customFilters, addressbook)
		if len(data) == 0 {
			data = dataNew
			continue
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
	}
	return data
}
