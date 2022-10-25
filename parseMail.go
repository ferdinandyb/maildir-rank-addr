package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	_ "github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
)

// RFC 1123Z regexp
var dateRe = regexp.MustCompile(`(((Mon|Tue|Wed|Thu|Fri|Sat|Sun))[,]?\s[0-9]{1,2})\s` +
	`(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\s` +
	`([0-9]{4})\s([0-9]{2}):([0-9]{2})(:([0-9]{2}))?\s([\+|\-][0-9]{4})`)

func parseDate(h *mail.Header) (time.Time, error) {
	t, err := h.Date()
	if err == nil {
		return t, nil
	}
	text, err := h.Text("date")
	// sometimes, no error occurs but the date is empty.
	// In this case, guess time from received header field
	if err != nil || text == "" {
		t, err := parseReceivedHeader(h)
		if err == nil {
			return t, nil
		}
	}
	layouts := []string{
		// X-Mailer: EarthLink Zoo Mail 1.0
		"Mon, _2 Jan 2006 15:04:05 -0700 (GMT-07:00)",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, text); err == nil {
			return t, nil
		}
	}
	// still no success, try the received header as a last resort
	t, err = parseReceivedHeader(h)
	if err != nil {
		return time.Time{}, fmt.Errorf("unrecognized date format: %s", text)
	}
	return t, nil
}

func parseReceivedHeader(h *mail.Header) (time.Time, error) {
	guess, err := h.Text("received")
	if err != nil {
		return time.Time{}, fmt.Errorf("received header not parseable: %w",
			err)
	}
	return time.Parse(time.RFC1123Z, dateRe.FindString(guess))
}

func parseMessage(
	path string,
	headers chan<- *mail.Header,
	semaphore chan struct{},
	wg *sync.WaitGroup,
) {
	semaphore <- struct{}{}
	defer func() {
		<-semaphore
		wg.Done()
	}()
	f, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		headers <- &mail.Header{}
		return
	}
	r, err := mail.CreateReader(f)
	if err != nil {
		fmt.Println(err)
		headers <- &mail.Header{}
		return
	}
	h := &mail.Header{Header: r.Header.Header}
	f.Close()
	headers <- h
	return
}

func assignClass(
	field string,
	sender string,
	addresses []string,
) int {
	if field == "from" {
		return 0
	}

	for _, addr := range addresses {
		if addr == sender {
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

func filterAddress(address string) bool {
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
	return false

}

func processHeaders(
	headers chan *mail.Header,
	retvalchan chan map[string]AddressData,
	addresses []string,
) {
	count := 0
	shouldAssignClass := len(addresses) > 0
	retval := make(map[string]AddressData)
	fields := [4]string{"to", "cc", "bcc", "from"}
	for h := range headers {
		count++
		time, err := h.Date()
		senderaddress, err := h.AddressList("from")
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
			header, err := h.AddressList(field)
			if err != nil {
				continue
			}
			for _, address := range header {
				normaddr := strings.ToLower(address.Address)
				if filterAddress(normaddr) {
					continue
				}
				if aD, ok := retval[normaddr]; ok {
					aD.Names = append(aD.Names, address.Name)
					if shouldAssignClass {
						aC := assignClass(
							field,
							sender,
							addresses,
						)
						if aD.Class < aC {
							aD.Class = aC
						}
					}
					if aD.Date < time.Unix() {
						aD.Date = time.Unix()
					}
					retval[normaddr] = aD
				} else {
					aD := AddressData{}
					aD.Names = append(aD.Names, address.Name)
					aD.Date = time.Unix()
					aD.Address = normaddr
					if shouldAssignClass {
						aD.Class = assignClass(
							field,
							sender,
							addresses,
						)
					} else {
						aD.Class = 2
					}
					retval[normaddr] = aD
				}

			}

		}

	}
	fmt.Println("Read ", count, " messages")
	retvalchan <- retval
}

func walkMaildir(path string, addresses []string) map[string]AddressData {
	headers := make(chan *mail.Header)
	concurrent := runtime.GOMAXPROCS(2 * runtime.NumCPU())
	semaphore := make(chan struct{}, concurrent)
	retvalchan := make(chan map[string]AddressData)
	var wg sync.WaitGroup
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
		case "new", "tmp", "cur":
			wg.Add(1)
			go parseMessage(path, headers, semaphore, &wg)
		default:
			return nil
		}
		return nil
	})
	go processHeaders(headers, retvalchan, addresses)
	wg.Wait()
	close(headers)
	close(semaphore)
	retval := <-retvalchan
	close(retvalchan)
	return retval

}
