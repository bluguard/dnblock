package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/bluguard/dnshield/internal/dns/server"
	"github.com/bluguard/dnshield/internal/dns/server/configuration"
)

func main() {

	confFile := flag.String("conf", "./conf", "configuration file, will be created if not exists")
	flag.Parse()

	file, err := os.Open(*confFile)
	if err != nil {
		if os.IsNotExist(err) {
			createDefault(confFile)
			return
		}
	}

	var conf configuration.ServerConf
	json.NewDecoder(file).Decode(&conf)

	s := server.Server{}

	s.Start(conf).Wait()
}

func createDefault(confFile *string) {
	log.Println("creating default configuration")
	file, err := os.Create(*confFile)
	if err != nil {
		panic(err)
	}
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(configuration.Default())
	if err != nil {
		panic(err)
	}
}
