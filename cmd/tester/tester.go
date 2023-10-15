package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/bluguard/dnshield/internal/dns/client/udp"
)

func main() {
	target := flag.String("target", "127.0.0.1:53", "host:port of the targeted server")
	domainList := flag.String("domainlist", "", "file containing the list of the domains to test the target")
	runners := flag.Int("runners", 1, "number of runners to load the server")
	loop := flag.Int("loop", 1, "number to duplicate requests")
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

	domainChan := make(chan string, 100)

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go readDomains(mainContext, wg, *domainList, domainChan, *loop)

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

func readDomains(ctx context.Context, wg *sync.WaitGroup, domainsFile string, domainChan chan<- string, loop int) {
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
			return
		default:

			if !scanner.Scan() {
				close(domainChan)
				return
			}
			domain := scanner.Text()
			for i := 0; i < loop; i++ {
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
				log.Println(err)
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

	for {
		select {
		case <-ctx.Done():
			printStats(durations)
			return
		case d, ok := <-durChan:
			if !ok {
				printStats(durations)
				return
			}
			durations = append(durations, d)
		}
	}

}

func printStats(durations []time.Duration) {
	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})
	mean := computeMean(durations)
	normalizedMean := computeMean(durations[1 : len(durations)-1])
	min, max := computeMinMax(durations)

	_, _ = fmt.Printf("\nmin: %s\tavg: %s\tmax: %s\tnormAvg: %s\ttotal requests:%d\n", min.String(), mean.String(), max.String(), normalizedMean.String(), len(durations))
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
