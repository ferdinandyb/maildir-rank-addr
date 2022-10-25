package main

import (
	"fmt"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
)

func loadConfig() Config {
	pflag.String("maildir", "", "path to maildir folder")
	pflag.String("outputpath", "", "path to output file")
	pflag.String("template", "", "output template")
	pflag.StringSlice("addresses", []string{}, "comma separated list of your email addresses")
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
	config := Config{
		maildir:    maildir,
		outputpath: outputpath,
		addresses:  viper.GetStringSlice("addresses"),
		template:   viper.GetString("template"),
	}
	return config
}
