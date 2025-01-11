package main

func main() {
	config := loadConfig()
	addressbook := parseAddressbook(config.addressbookLookupCommand)
	data := walkMaildirs(
		config.maildirs,
		config.useraddresses,
		config.customFilters,
	)
	classeddata := calculateRanks(
		data,
		addressbook,
		config.listtemplate,
	)
	saveData(classeddata, config.outputpath, config.template, addressbook, config.addressbookAddUnmatched)
}
