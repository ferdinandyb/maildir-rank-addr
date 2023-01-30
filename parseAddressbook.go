package main

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func parseAddressbook(
	cmd *exec.Cmd,
) map[string]string {
	if cmd == nil {
		return nil
	}
	addressbook := make(map[string]string)
	out, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		slice := strings.Split(scanner.Text(), "\t")
		if len(slice) < 2 {
			fmt.Println("Couldn't parse ", scanner.Text())
		} else {
			addressbook[strings.ToLower(slice[0])] = slice[1]
		}
	}
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
	return addressbook
}
