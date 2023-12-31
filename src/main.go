package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
)

const URL = "https://go-challenge.skip.money"
const COLLECTION = "azuki"
const COLOR_GREEN = "\033[32m"
const COLOR_RED = "\033[31m"
const COLOR_RESET = "\033[0m"

// Max number of workers who download token info in parallel
const MAX_WORKERS = 250

// Max number of retries for a token download
const MAX_RETRY = 3

var logger *log.Logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)

type Token struct {
	id    int
	attrs map[string]string
}

type RarityScorecard struct {
	rarity float64
	id     int
}

type Collection struct {
	count int
	url   string
}

type tokenInfo struct {
	id     int
	colUrl string
}

type getTokenFn = func(tid int, colUrl string) (*Token, error)

// Returns a token or `nil` along with an error
func getToken(tid int, colUrl string) (*Token, error) {
	logger.Println(string(COLOR_GREEN), fmt.Sprintf("Getting token %d", tid), string(COLOR_RESET))

	url := fmt.Sprintf("%s/%s/%d.json", URL, colUrl, tid)
	res, err := http.Get(url)
	if err != nil {
		logger.Println(string(COLOR_RED), fmt.Sprintf("Error getting token %d :", tid), err, string(COLOR_RESET))
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Println(string(COLOR_RED), fmt.Sprintf("Error reading response for token %d :", tid), err, string(COLOR_RESET))
		return nil, err
	}
	attrs := make(map[string]string)
	json.Unmarshal(body, &attrs)
	return &Token{
		id:    tid,
		attrs: attrs,
	}, nil
}

func worker(inCh chan tokenInfo, wg *sync.WaitGroup, outCh chan *Token, getTokenFn getTokenFn) {
	defer wg.Done()

	for i := range inCh {
		retry := 0
	retry:
		token, err := getTokenFn(i.id, i.colUrl)
		if err != nil && retry < MAX_RETRY {
			retry++
			goto retry
		}
		outCh <- token
	}
}

// Calculates rarity score for a token
func calculateRarity(attrs map[string]string, allAttrs map[string]map[string]int) float64 {
	if len(attrs) == 0 {
		return 0.0
	}

	r := 0.0
	for k, v := range attrs {
		r += 1.0 / float64(allAttrs[k][v]*len(allAttrs[k]))
	}
	return r
}

// Returns rarity scorecards for a collection sorted by rarity in descending order
func getRarityScorecards(col Collection, getTokenFn getTokenFn) []*RarityScorecard {
	outCh := make(chan *Token, col.count)

	// Create a pool of workers
	inCh := make(chan tokenInfo)
	wg := new(sync.WaitGroup)
	for i := 0; i < MAX_WORKERS; i++ {
		wg.Add(1)
		go worker(inCh, wg, outCh, getTokenFn)
	}

	logger.Println(string(COLOR_GREEN), fmt.Sprintf("Downloading all the tokens from %q...", col.url), string(COLOR_RESET))
	for i := 1; i <= col.count; i++ {
		inCh <- tokenInfo{
			id:     i,
			colUrl: col.url,
		}
	}

	close(inCh)
	wg.Wait()

	// Accumulating results
	tokens := make([]*Token, 0, col.count)
	for i := 1; i <= col.count; i++ {
		t := <-outCh

		if t != nil { // Filter out failed downloads
			tokens = append(tokens, t)
		}
	}

	close(outCh)

	// Map of attributes to map of values to count
	// "hat" -> "red" -> 10
	allAttrs := make(map[string]map[string]int)
	for _, t := range tokens {
		for k, v := range t.attrs {
			if allAttrs[k] == nil {
				allAttrs[k] = make(map[string]int)
			}

			if _, ok := allAttrs[k][v]; !ok {
				allAttrs[k][v] = 1
			} else {
				allAttrs[k][v]++
			}
		}
	}

	logger.Println(string(COLOR_GREEN), "Calculating rarity scores...", string(COLOR_RESET))
	scorecards := make([]*RarityScorecard, len(tokens))
	for i, t := range tokens {
		scorecards[i] = &RarityScorecard{
			rarity: calculateRarity(t.attrs, allAttrs),
			id:     t.id,
		}
	}

	sort.Slice(scorecards, func(i, j int) bool {
		return scorecards[i].rarity > scorecards[j].rarity
	})

	return scorecards
}

func main() {
	azuki := Collection{
		count: 10000,
		url:   "azuki1",
	}
	scorecards := getRarityScorecards(azuki, getToken)
	// Print top 5 (at max)
	max := 5
	if len(scorecards) < max {
		max = len(scorecards)
	}
	logger.Println(string(COLOR_GREEN), fmt.Sprintf("Top %d tokens (ID: rarity)", max), string(COLOR_RESET))
	for i := 0; i < max; i++ {
		logger.Println(fmt.Sprintf("%d: %.05f", scorecards[i].id, scorecards[i].rarity))
	}
}
