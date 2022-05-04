package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"runtime/pprof"

	"github.com/bluguard/dnshield/internal/dns/server"
	"github.com/bluguard/dnshield/internal/dns/server/configuration"
)

func main() {

	memprofile := flag.String("memprofile", "", "memory profile file")
	cpuprofile := flag.String("cpuprofile", "", "cpu profile file")

	confFile := flag.String("conf", "./conf", "configuration file, will be created if not exists")
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

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

	if *cpuprofile != "" {
		pprof.StopCPUProfile()
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.WriteHeapProfile(f)
		f.Close()
	}
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
