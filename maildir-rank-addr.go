package main

func main() {
	config := loadConfig()
	addressbook := parseAddressbook(config.addressbookLookupCommand)
	data := walkMaildirs(config.maildirs, config.useraddresses, config.customFilters, addressbook)
	classeddata := calculateRanks(data)
	saveData(classeddata, config.outputpath, config.template, addressbook, config.addressbookAddUnmatched)
}
