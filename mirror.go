package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"sync"
	"time"
)

const (
	mirrorListUrl = "https://www.archlinux.org/mirrors/status/json/"
	dbPath        = "core/os/x86_64/core.db"
)

type MirrorData struct {
	CheckFrequency int      `json:"check_frequency"`
	Cutoff         int      `json:"cutoff"`
	LastCheck      string   `json:"last_check"`
	NumChecks      int      `json:"num_checks"`
	Mirrors        []Mirror `json:"urls"`
	Version        int      `json:"version"`
}

type Mirror struct {
	Url            string  `json:"url"`
	Protocol       string  `json:"protocol"`
	LastSync       string  `json:"last_sync"`
	CompletionPct  float64 `json:"completion_pct"`
	Delay          int     `json:"delay"`
	DurationAvg    float64 `json:"duration_avg"`
	DurationStddev float64 `json:"duration_stddev"`
	Score          float64 `json:"score"`
	Active         bool    `json:"active"`
	Country        string  `json:"country"`
	CountryCode    string  `json:"country_code"`
	Isos           bool    `json:"isos"`
	Ipv4           bool    `json:"ipv4"`
	Ipv6           bool    `json:"ipv6"`
	Details        string  `json:"details"`
	Rating         float64
}

func fetchMirrorData() *MirrorData {
	var data MirrorData

	resp, err := http.Get(mirrorListUrl)
	assertNoErr(err)
	defer func() {
		assertNoErr(resp.Body.Close())
	}()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	assertNoErr(err)
	err = json.Unmarshal(bodyBytes, &data)
	assertNoErr(err)

	return &data
}

func filterMirrors(mirrors *[]Mirror, cmp func(*Mirror) bool) []Mirror {
	m := make([]Mirror, 0, len(*mirrors))

	for _, mirror := range *mirrors {
		if cmp(&mirror) {
			m = append(m, mirror)
		}
	}

	return m
}

func sortMirrorsByRateDesc(m *[]Mirror) {
	sort.Slice(*m, func(i, j int) bool {
		return (*m)[i].Rating > (*m)[j].Rating
	})
}

func rateAll(mirrors *[]Mirror) {
	mirrorsLen := len(*mirrors)
	sem := make(chan int, jobs)
	done := make(chan int)
	wg := sync.WaitGroup{}
	wg.Add(mirrorsLen)

	if progress {
		wg.Add(1)
		go func() {
			defer wg.Done()
			startProgress(done, mirrorsLen)
		}()
	}

	for i := 0; i < mirrorsLen; i++ {
		go func(j int) {
			defer wg.Done()
			sem <- 1
			rate(&(*mirrors)[j])
			<-sem
			done <- 1
		}(i)
	}

	wg.Wait()
	close(sem)
	close(done)
}

func rate(mirror *Mirror) {
	dbUrl := fmt.Sprintf("%s%s", mirror.Url, dbPath)
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(timeout))

	req, err := http.NewRequest(http.MethodGet, dbUrl, nil)
	assertNoErr(err)

	req = req.WithContext(ctx)
	client := &http.Client{}

	ratingChan := make(chan float64)
	errChan := make(chan error)

	go func() {
		startedAt := time.Now()
		resp, err := client.Do(req)
		finishedAt := time.Since(startedAt)

		if err != nil {
			errChan <- err
			return
		}

		size := resp.ContentLength
		if size == -1 {
			defer func() {
				assertNoErr(resp.Body.Close())
			}()
			bodyBytes, err := ioutil.ReadAll(resp.Body)

			if err != nil {
				errChan <- err
				return
			}

			size = int64(len(bodyBytes))
		}

		ratingChan <- float64(size) / finishedAt.Seconds()
	}()

	select {
	case <-ctx.Done():
		cancel()
		<-errChan
		mirror.Rating = -1
	case <-errChan:
		mirror.Rating = -1
	case rating := <-ratingChan:
		mirror.Rating = rating
	}
}

func splitMirrors(mirrors *[]Mirror, cmp func(*Mirror) bool) ([]Mirror, []Mirror) {
	result1 := make([]Mirror, 0, len(*mirrors))
	result2 := make([]Mirror, 0, len(*mirrors))

	for _, mirror := range *mirrors {
		if cmp(&mirror) {
			result1 = append(result1, mirror)
		} else {
			result2 = append(result2, mirror)
		}
	}

	return result1, result2
}
