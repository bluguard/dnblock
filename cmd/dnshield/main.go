package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"runtime/pprof"
	"runtime/trace"

	"github.com/bluguard/dnshield/internal/dns/server"
	"github.com/bluguard/dnshield/internal/dns/server/configuration"
)

func main() {

	memprofile := flag.String("memprofile", "", "memory profile file")
	cpuprofile := flag.String("cpuprofile", "", "cpu profile file")
	traceprofile := flag.String("traceprofile", "", "trace profile file")

	confFile := flag.String("conf", "./conf", "configuration file, will be created if not exists")
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			panic(err)
		}
		defer pprof.StopCPUProfile()
		defer f.Close()
	}

	if *traceprofile != "" {
		f, err := os.Create(*traceprofile)
		if err != nil {
			log.Fatal(err)
		}
		if err := trace.Start(f); err != nil {
			panic(err)
		}
		defer trace.Stop()
		defer f.Close()
	}

	file, err := os.Open(*confFile)
	if err != nil {
		if os.IsNotExist(err) {
			createDefault(confFile)
			return
		}
	}

	var conf configuration.ServerConf
	_ = json.NewDecoder(file).Decode(&conf)

	s := server.Server{}

	s.Start(conf).Wait()

	if *cpuprofile != "" {
		pprof.StopCPUProfile()
	}
	if *cpuprofile != "" {
		trace.Stop()
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal(err)
		}
		_ = pprof.WriteHeapProfile(f)
		_ = f.Close()
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
