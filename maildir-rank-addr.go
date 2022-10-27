package main

import ()

func main() {
	config := loadConfig()
	data := walkMaildir(config.maildir, config.addresses, config.customFilters)
	classeddata := calculateRanks(data)
	saveData(classeddata, config.outputpath, config.template)
}
