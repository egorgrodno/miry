package main

import (
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"strings"
	"time"
)

const (
	hiddenCursor       = "\033[?25l"
	visibleCursor      = "\033[?25h"
	progressComplete   = "\u2588"
	progressIncomplete = "\u2591"
	moduleSeparator    = " | "
)

func startProgress(done <-chan int, total int) {
	timeSpend := time.Duration(0)
	eta := time.Duration(-1)
	startedAt := time.Now()
	d := 0
	base1percentage := 0.0

	termID := int(os.Stdout.Fd())
	width, _, err := terminal.GetSize(termID)
	assertNoErr(err)

	hideCursor()
	defer showCursor()

	printProgress(d, total, int(base1percentage), width, eta)

	for range done {
		d++

		base1percentage = float64(d) / float64(total)
		timeSpend = time.Since(startedAt)
		eta = time.Duration((1 - base1percentage) / base1percentage * float64(timeSpend))

		printProgress(d, total, int(base1percentage*100), width, eta.Round(time.Second))

		if d == total {
			break
		}
	}
}

func printProgress(done, total, percentage, width int, eta time.Duration) {
	percentageModule := fmt.Sprintf("%v%%", percentage)
	etaModule := createEtaModule(eta)
	countModule := fmt.Sprintf("%v/%v", done, total)
	modules := strings.Join([]string{percentageModule, etaModule, countModule}, moduleSeparator)
	barWidth := width - len(modules) - 2
	bar := createBar(barWidth, percentage)
	fmt.Print("\r", bar, " ", modules, " ")
}

func createEtaModule(eta time.Duration) string {
	if eta == -1 {
		return "ETA: unknown"
	}

	return fmt.Sprintf("ETA: %v", eta)
}

func createBar(barWidth, percentage int) string {
	bar := ""
	completedBars := barWidth * percentage / 100

	for i := 0; i < barWidth; i++ {
		if i <= completedBars {
			bar += progressComplete
		} else {
			bar += progressIncomplete
		}
	}

	return bar
}

func hideCursor() {
	fmt.Print(hiddenCursor)
}

func showCursor() {
	fmt.Print(visibleCursor)
}
