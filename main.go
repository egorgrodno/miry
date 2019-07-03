package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var (
	jobs      int
	limit     int
	timeout   int
	output    string
	outputAbs string
	backup    bool
	progress  bool
)

func main() {
	parseFlags()
	validateAndSanitizeFlags()

	mirrorData := fetchMirrorData()
	mirrors := filterMirrors(&mirrorData.Mirrors, func(m *Mirror) bool {
		return m.Active && m.Protocol == "http" || m.Protocol == "https"
	})
	nTotal := len(mirrors)
	nToBeRated := nTotal

	if limit != -1 && len(mirrors) > limit {
		mirrors = mirrors[:limit]
		nToBeRated = limit
	}

	fmt.Printf("Rating %d (out of %d) mirrors, %d concurrent jobs\n", nToBeRated, nTotal, jobs)

	rateAll(&mirrors)
	mirrors, erroredMirrors := splitMirrors(&mirrors, func(m *Mirror) bool {
		return m.Rating != -1
	})
	nErrored := len(erroredMirrors)

	if nErrored > 0 {
		if len(mirrors) == 0 {
			fmt.Println("All the mirrors failed rating, aborting")
			os.Exit(1)
		}

		fmt.Printf("%d mirrors failed rating\n", nErrored)
	}

	sortMirrorsByRateDesc(&mirrors)
	if backup {
		backupOutput()
	}
	writeMirrorlist(&mirrors)

	fmt.Printf("Successfully written %d mirrors\n", len(mirrors))
}

func parseFlags() {
	flag.IntVar(&jobs, "jobs", 4, "Number of concurrently running rating jobs")
	flag.IntVar(&limit, "limit", -1, "Number of mirrors to rate, -1 means no limit")
	flag.IntVar(&timeout, "timeout", 5000, "Number of milliseconds passed before cancelling mirror rating")
	flag.StringVar(&output, "output", "/etc/pacman.d/mirrorlist", "Output file")
	flag.BoolVar(&backup, "backup", true, "Enable output file backup")
	flag.BoolVar(&progress, "progress", true, "Print rating progress")
	flag.Parse()
}

func validateAndSanitizeFlags() {
	if jobs < 1 {
		printUsageErrorAndExit("Jobs can't be less than 1")
	} else if limit != -1 && limit < 1 {
		printUsageErrorAndExit("Limit can't be less than 1")
	} else if tail := flag.Args(); len(tail) > 0 {
		err := fmt.Sprintf("%s not recognized", tail[0])
		printUsageErrorAndExit(err)
	}

	var err error
	outputAbs, err = filepath.Abs(output)
	assertNoErr(err)
}

func printUsageErrorAndExit(err string) {
	fmt.Fprintln(os.Stderr, err)
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func assertNoErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
