package main

import (
	"bufio"
	"context"
	"embed"
	"errors"
	"flag"
	"log"
	"math"
	"os"
	"os/signal"
	"slices"
	"sort"
	"sync"
	"syscall"
	"text/template"
	"time"

	"github.com/bluguard/dnshield/internal/dns/client/udp"
)

var (
	target     = flag.String("target", "127.0.0.1:53", "host:port of the targeted server")
	domainList = flag.String("domainlist", "", "file containing the list of the domains to test the target")
	runners    = flag.Int("runners", 1, "number of runners to load the server")
	loop       = flag.Int("loop", 1, "number opf run for the given file")
	duplicates = flag.Int("duplicates", 1, "number of duplicate requests per loop")
)

func main() {
	flag.Parse()

	mainContext, cancel := context.WithCancel(context.Background())
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		if cancel != nil {
			cancel()
		}
	}()

	for i := 0; i < *loop; i++ {

		domainChan := make(chan string, 100)

		wg := &sync.WaitGroup{}

		wg.Add(1)
		go readDomains(mainContext, wg, *domainList, domainChan, *duplicates)

		client := udp.NewUDPClient(*target)

		durChan := make(chan time.Duration, 100)

		wg.Add(*runners)
		for i := 0; i < *runners; i++ {
			go runner(mainContext, wg, client, domainChan, durChan)
		}

		wg2 := &sync.WaitGroup{}
		wg2.Add(1)
		go timeAnalyzer(mainContext, wg2, durChan)

		wg.Wait()

		close(durChan)

		wg2.Wait()
	}
}

func readDomains(ctx context.Context, wg *sync.WaitGroup, domainsFile string, domainChan chan<- string, duplicates int) {
	defer wg.Done()

	file, err := os.Open(domainsFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for {
		select {
		case <-ctx.Done():
			close(domainChan)
			return
		default:
			if !scanner.Scan() {
				close(domainChan)
				return
			}
			domain := scanner.Text()
			for i := 0; i < duplicates; i++ {
				domainChan <- domain
			}
		}
	}

}

func runner(ctx context.Context, wg *sync.WaitGroup, client *udp.UDPClient, domainChan <-chan string, durChan chan<- time.Duration) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			log.Println("closing runner, context Done")
			return
		case <-time.NewTicker(1 * time.Second).C:
			log.Println("closing runner, timeout")
			return
		case domain, ok := <-domainChan:
			if !ok {
				log.Println("closing runner, task completed")
				return
			}
			start := time.Now()
			_, err := client.ResolveV4(domain)
			if err != nil && !errors.Is(err, &udp.NoResponse{}) {
				durChan <- time.Duration(-1)
				continue
			}
			dur := time.Since(start)
			durChan <- dur
		}
	}
}

func timeAnalyzer(ctx context.Context, wg *sync.WaitGroup, durChan <-chan time.Duration) {
	defer wg.Done()

	durations := make([]time.Duration, 0, 100)
	drops := 0

	for {
		select {
		case <-ctx.Done():
			printStats(durations, drops)
			return
		case d, ok := <-durChan:
			if !ok {
				printStats(durations, drops)
				return
			}
			if d < 0 {
				drops++
			} else {
				durations = append(durations, d)
			}
		}
	}

}

func printStats(durations []time.Duration, drops int) {
	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})
	avg := computeMean(durations)
	normalizedavg := computeMean(durations[1 : len(durations)-1])
	min, max := computeMinMax(durations)

	printReport(computeReportData(durations, min, max, avg, normalizedavg, drops))
}

func computeMean(durations []time.Duration) time.Duration {
	mean := time.Duration(0)
	for _, d := range durations {
		mean += d
	}
	mean /= time.Duration(len(durations))
	return mean
}

func computeMinMax(durations []time.Duration) (min time.Duration, max time.Duration) {
	min = math.MaxInt64
	for _, d := range durations {
		if d < min {
			min = d
		}
		if d > max {
			max = d
		}
	}
	return min, max
}

func containsAll[T ~int64](a []T, b []T) bool {
	for _, v := range a {
		if !slices.Contains(b, v) {
			return false
		}
	}
	return true
}

type ReportData struct {
	Target  string     `json:"target,omitempty"`
	Date    string     `json:"date,omitempty"`
	Rep     []RepEntry `json:"rep,omitempty"`
	Min     string     `json:"min,omitempty"`
	Max     string     `json:"max,omitempty"`
	Avg     string     `json:"avg,omitempty"`
	NormAvg string     `json:"normAvg,omitempty"`
	Drops   int        `json:"drops,omitempty"`
}

type RepEntry struct {
	Label string `json:"label,omitempty"`
	Count int    `json:"count,omitempty"`
}

func computeReportData(durations []time.Duration, min, max, avg, normAvg time.Duration, drops int) ReportData {
	res := ReportData{
		Target:  *target,
		Date:    time.Now().Format("2006-01-02T15:04:05"),
		Min:     min.String(),
		Max:     max.String(),
		Avg:     avg.String(),
		Drops:   drops,
		NormAvg: normAvg.String(),
		Rep:     computeRep(durations, 1*time.Millisecond, 10*time.Millisecond, 100*time.Millisecond, 500*time.Millisecond, 1*time.Second),
	}

	log.Println(res)
	return res
}

func computeRep(durations []time.Duration, slice ...time.Duration) []RepEntry {
	res := make([]RepEntry, 0, len(slice)+1)

	i := 0
	for _, d := range slice {
		entry := RepEntry{Label: "-" + d.String(), Count: 0}
		for ; i < len(durations) && durations[i] < d; i++ {
			entry.Count++
		}
		res = append(res, entry)
	}

	entry := RepEntry{Label: "+" + slice[len(slice)-1].String(), Count: 0}
	for ; i < len(durations); i++ {
		entry.Count++
	}
	res = append(res, entry)

	return res
}

//go:embed report.tmpl
var f embed.FS

func printReport(data ReportData) {
	t, err := template.ParseFS(f, "report.tmpl")
	if err != nil {
		panic(err)
	}
	if err = t.Execute(os.Stdout, data); err != nil {
		panic(err)
	}
}
