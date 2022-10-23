package main

import (
	"fmt"
)

func main() {
	path := "/home/fbence/.mail/elte/Inbox"

	data := walkMaildir(path)
	fmt.Println(data)
	// parseMessage(path)
}
