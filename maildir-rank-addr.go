package main

import (
	"fmt"
	"mime"
	"net/mail"
	"os"
)

func main() {
	r, err := os.Open("/home/fbence/.mail/elte/Inbox/cur/1665525026.1846542_33.mashenka,U=33:2,S")
	if err == nil {
		msg, err := mail.ReadMessage(r)
		if err == nil {
			dec := new(mime.WordDecoder)
			header, err := dec.Decode(msg.Header.Get("Subject"))
			if err == nil {
				fmt.Println("To:", header)
			} else {
				fmt.Println(err)
			}
		}
	}
}
