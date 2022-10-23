package main

import (
	"fmt"
	"os"
	"regexp"
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
		fmt.Println("reader error")
		headers <- &mail.Header{}
		return
	}
	h := &mail.Header{Header: r.Header.Header}
	f.Close()
	headers <- h
	fmt.Println("finished")
	return
	// fmt.Println(h.Text("to"))
	// fmt.Println(h.Text("from"))
	// fmt.Println(h.Text("cc"))
	// fmt.Println(h.Text("bcc"))
	// t, _ := parseDate(&h)
	// fmt.Println(t.Unix())
}
