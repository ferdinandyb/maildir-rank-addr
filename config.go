package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func loadConfig() Config {
	pflag.String("config", "", "path to config file")
	pflag.StringSlice("maildir", []string{}, "comma separated list of paths to maildir folders")
	pflag.String("outputpath", "", "path to output file")
	pflag.String("template", "", "output template")
	pflag.String("addr-book-cmd", "", "optional command to query addresses from your addressbook")
	pflag.Bool("addr-book-add-unmatched", false, "flag to determine if you want unmatched addressbook contacts to be added to the output")
	pflag.StringSlice("addresses", []string{}, "comma separated list of your email addresses (regex possible)")
	pflag.StringSlice("filters", []string{}, "comma separated list of regexes to filter")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
	dir, direrr := os.UserConfigDir()
	if direrr != nil {
		dir, _ = os.Getwd()
	}
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(dir + "/maildir-rank-addr")
	viper.AddConfigPath(".")
	dir, direrr = os.UserCacheDir()
	if direrr != nil {
		dir, _ = os.Getwd()
	}
	viper.SetDefault("outputpath", dir+"/maildir-rank-addr/addressbook.tsv")
	viper.SetDefault("addresses", []string{})
	viper.SetDefault("filters", []string{})
	viper.SetDefault("template", "{{.Address}}\t{{.Name}}")

	configPath, err := pflag.CommandLine.GetString("config")
	if configPath != "" && err == nil {
		viper.SetConfigFile(configPath)
	}

	err = viper.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok || configPath != "" {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}
	if len(viper.GetStringSlice("maildir")) == 0 {
		pflag.PrintDefaults()
		os.Exit(1)
	}
	maildirInput := viper.GetStringSlice("maildir")
	maildirs := make([]string, len(maildirInput))
	for i, maildir := range maildirInput {
		maildirs[i], _ = homedir.Expand(maildir)
	}
	outputpath, _ := homedir.Expand(viper.GetString("outputpath"))
	filterInput := viper.GetStringSlice("filters")
	customFilters := make([]*regexp.Regexp, len(filterInput))
	for i, filter := range filterInput {
		customFilters[i] = regexp.MustCompile(filter)
	}
	addressesInput := viper.GetStringSlice("addresses")
	addresses := make([]*regexp.Regexp, len(addressesInput))
	for i, filter := range addressesInput {
		addresses[i] = regexp.MustCompile(filter)
	}
	templateString := viper.GetString("template")
	addressbookLookupCommandString := viper.GetString("addr-book-cmd")
	addressbookAddUnmatched := viper.GetBool("addr-book-add-unmatched")
	var addressbookLookupCommand *exec.Cmd
	if addressbookLookupCommandString != "" {
		args := strings.Fields(addressbookLookupCommandString)
		application, arguments := args[0], args[1:]
		addressbookLookupCommand = exec.Command(application, arguments...)
	}

	if !strings.HasSuffix(templateString, "\n") {
		templateString += "\n"
	}
	tmpl, err := template.New("output").Parse(templateString)
	if err != nil {
		panic(fmt.Errorf("bad template"))
	}
	config := Config{
		maildirs:                 maildirs,
		outputpath:               outputpath,
		useraddresses:            addresses,
		template:                 tmpl,
		customFilters:            customFilters,
		addressbookLookupCommand: addressbookLookupCommand,
		addressbookAddUnmatched:  addressbookAddUnmatched,
	}
	return config
}
