package main

import (
	"github.com/akamensky/argparse"
	"os"
)

func main() {
	parser := argparse.NewParser("maildir-rank-go", "Generate a ranked list of emails from maildir folders.")
	dir, err := os.UserCacheDir()
	if err != nil {
		dir, _ = os.Getwd()
	}
	path := parser.String("p", "path", &argparse.Options{Required: true, Help: "path to maildir folder"})
	outpath := parser.String("o", "outpath", &argparse.Options{Required: false, Help: "output path", Default: dir + "/maildir-rank-addr/addressbook.tsv"})
	parser.Parse(os.Args)
	data := walkMaildir(*path)
	classeddata := calculateRanks(data)
	saveData(classeddata, *outpath)
}
