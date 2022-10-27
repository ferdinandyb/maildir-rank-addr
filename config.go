package main

import (
	"fmt"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"regexp"
	"strings"
	"text/template"
)

func loadConfig() Config {
	pflag.String("maildir", "", "path to maildir folder")
	pflag.String("outputpath", "", "path to output file")
	pflag.String("template", "", "output template")
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

	err := viper.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}
	if viper.Get("maildir") == "" {
		pflag.PrintDefaults()
		os.Exit(1)
	}
	maildir, _ := homedir.Expand(viper.GetString("maildir"))
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

	if !strings.HasSuffix(templateString, "\n") {
		templateString += "\n"
	}
	tmpl, err := template.New("output").Parse(templateString)
	if err != nil {
		panic(fmt.Errorf("bad template"))
	}
	config := Config{
		maildir:       maildir,
		outputpath:    outputpath,
		addresses:     addresses,
		template:      tmpl,
		customFilters: customFilters,
	}
	return config
}
