package main

import (
	"log"
	"os"

	. "github.com/seatsurfing/seatsurfing/server/app"
	. "github.com/seatsurfing/seatsurfing/server/config"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

func main() {
	log.Println("Starting...")
	log.Println("Seatsurfing Backend Version " + GetProductVersion())
	db := GetDatabase()
	a := GetApp()
	a.InitializeDatabases()
	a.InitializeDefaultOrg()
	a.InitializeRouter()
	a.InitializeTimers()
	if GetConfig().PrintConfig {
		GetConfig().Print()
	}
	a.Run()
	db.Close()
	os.Exit(0)
}
