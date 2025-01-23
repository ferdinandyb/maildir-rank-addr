package main

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/emersion/go-mbox"
	"github.com/emersion/go-message"
	_ "github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
)

func mboxParser(path string, headers chan<- *mail.Header) error {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return err
	}
	mbr := mbox.NewReader(f)
	for {
		msg, err := mbr.NextMessage()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}
		entity, err := message.Read(msg)
		h := &mail.Header{Header: entity.Header}
		headers <- h
	}
	return nil
}

func emlParser(path string, headers chan<- *mail.Header) error {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	r, err := mail.CreateReader(f)
	if err != nil {
		return err
	}
	h := &mail.Header{Header: r.Header.Header}
	headers <- h
	return nil
}

func messageParser(
	paths chan string,
	headers chan<- *mail.Header,
) {
	for path := range paths {
		err := emlParser(path, headers)
		if err != nil {
			mboxerr := mboxParser(path, headers)
			if mboxerr == nil {
				// do nothing
			} else if utf8.ValidString(err.Error()) {
				fmt.Fprintln(os.Stderr, path, err)
			} else {
				fmt.Fprintln(os.Stderr, path, "mail reader error, probably tried reading binary")
			}
		}
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

func filterAddress(
	address string,
	customFilters []*regexp.Regexp,
) bool {
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

func processEnvelope(
	envelope *mail.Header,
	addressmap map[string]AddressData,
	useraddresses []*regexp.Regexp,
	customFilters []*regexp.Regexp,
) error {
	addressheaders := [6]string{"to", "cc", "bcc", "from", "sender", "reply-to"}
	time, err := envelope.Date()
	if err != nil {
		return err
	}

	listidheader := envelope.Get("list-id")
	listidpattern := regexp.MustCompile(`(.*)\s*<(.+)>`)
	listname := strings.Trim(
		strings.Trim(
			listidpattern.ReplaceAllString(listidheader, "$1"),
			" ",
		),
		"\"",
	)
	listid := listidpattern.ReplaceAllString(listidheader, "$2")

	senderaddress, err := envelope.AddressList("from")
	if err != nil {
		return err
	}
	var sender string

	if len(senderaddress) > 0 {
		sender = strings.ToLower(senderaddress[0].Address)
	} else {
		sender = ""
	}

	for _, field := range addressheaders {
		header, err := envelope.AddressList(field)
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
			dec := new(mime.WordDecoder)
			name, err := dec.DecodeHeader(address.Name)
			if err != nil {
				continue
			}
			if addressdata, ok := addressmap[normaddr]; ok {
				if (strings.ToLower(name) != normaddr) && (strings.ToLower(name) != "") {
					addressdata.Names = append(addressdata.Names, name)
				}
				if addressdata.Class < class {
					addressdata.Class = class
				}
				if addressdata.ClassDate[class] < time.Unix() {
					addressdata.ClassDate[class] = time.Unix()
				}
				addressdata.ClassCount[class]++
				addressmap[normaddr] = addressdata
			} else {
				addressdata := AddressData{}
				if (strings.ToLower(name) != normaddr) && (strings.ToLower(name) != "") {
					addressdata.Names = append(addressdata.Names, name)
				}
				if len(listid) > 0 && (strings.Join(strings.Split(normaddr, "@"), ".") == listid) {
					addressdata.ListName = listname
					addressdata.ListId = listid
				}
				addressdata.Address = normaddr
				addressdata.Class = class
				addressdata.ClassDate = [3]int64{0, 0, 0}
				addressdata.ClassDate[class] = time.Unix()
				addressdata.ClassCount = [3]int{0, 0, 0}
				addressdata.ClassCount[class] = 1
				addressmap[normaddr] = addressdata
			}
		}

	}
	return nil
}

func processEnvelopeChan(
	envelopechan <-chan *mail.Header,
	retvalchan chan map[string]AddressData,
	useraddresses []*regexp.Regexp,
	customFilters []*regexp.Regexp,
) {
	count := 0
	errcount := 0
	addressmap := make(map[string]AddressData)
	for envelope := range envelopechan {
		err := processEnvelope(
			envelope,
			addressmap,
			useraddresses,
			customFilters,
		)
		if err != nil {
			errcount++
		} else {
			count++
		}

	}
	fmt.Println("Read", count+errcount, "files of which", count, "could be parsed.")
	retvalchan <- addressmap
	close(retvalchan)
}

func walkMaildir(
	path string,
	useraddresses []*regexp.Regexp,
	customFilters []*regexp.Regexp,
) map[string]AddressData {
	envelopechan := make(chan *mail.Header)
	messagePaths := make(chan string, 4096)

	var wg sync.WaitGroup
	for i := 0; i < 2*runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			messageParser(messagePaths, envelopechan)
		}()
	}

	retvalchan := make(chan map[string]AddressData)
	go processEnvelopeChan(envelopechan, retvalchan, useraddresses, customFilters)

	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}

		if info.IsDir() {
			switch filepath.Base(filepath.Dir(path)) {
			case "tmp", ".notmuch":
				return filepath.SkipDir
			}
			return nil
		}
		messagePaths <- path
		return nil
	})
	close(messagePaths)

	wg.Wait()
	close(envelopechan)

	return <-retvalchan
}
