package main

func main() {
	config := loadConfig()
	addressbook := parseAddressbook(config.addressbookLookupCommand)
	data := walkMaildir(config.maildir, config.addresses, config.customFilters, addressbook)
	classeddata := calculateRanks(data)
	saveData(classeddata, config.outputpath, config.template, addressbook, config.addressbookAddUnmatched)
}
