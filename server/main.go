package main

import (
	"log"
	"os"

	. "github.com/seatsurfing/seatsurfing/server/app"
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
	a.InitializeSingleOrgSettings()
	a.InitializePlugins()
	a.InitializeRouter()
	a.InitializeTimers()

	res, err := GetCache().Get("blubb")
	log.Println("Cache result:", string(res), "Error:", err)

	log.Println(GetCache().Set("blubb", []byte("blabb"), 60))

	res, err = GetCache().Get("blubb")
	log.Println("Cache result:", string(res), "Error:", err)

	a.Run()
	db.Close()
	os.Exit(0)
}
