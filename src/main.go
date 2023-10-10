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

func (t *Token) isEmpty() bool {
	return t.id == 0 && len(t.attrs) == 0
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

type getTokenFn = func(tid int, colUrl string) (Token, error)

func getToken(tid int, colUrl string) (Token, error) {
	logger.Println(string(COLOR_GREEN), fmt.Sprintf("Getting token %d", tid), string(COLOR_RESET))

	url := fmt.Sprintf("%s/%s/%d.json", URL, colUrl, tid)
	res, err := http.Get(url)
	if err != nil {
		logger.Println(string(COLOR_RED), fmt.Sprintf("Error getting token %d :", tid), err, string(COLOR_RESET))
		return Token{}, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Println(string(COLOR_RED), fmt.Sprintf("Error reading response for token %d :", tid), err, string(COLOR_RESET))
		return Token{}, err
	}
	attrs := make(map[string]string)
	json.Unmarshal(body, &attrs)
	return Token{
		id:    tid,
		attrs: attrs,
	}, nil
}

func worker(inCh chan tokenInfo, wg *sync.WaitGroup, outCh chan Token, getToken getTokenFn) {
	defer wg.Done()

	for i := range inCh {
		retry := 0
	retry:
		token, err := getToken(i.id, i.colUrl)
		if err != nil && retry < MAX_RETRY {
			retry++
			goto retry
		}
		outCh <- token
	}
}

// Calculates rarity score for a token
func calculateRarity(attrs map[string]string, allAttrs map[string]map[string]int) float64 {
	r := 0.0
	for k, v := range attrs {
		r += float64(allAttrs[k][v] * len(allAttrs[k]))
	}
	return 1.0 / r
}

// Returns rarity scorecards for a collection sorted by rarity in descending order
func getRarityScorecards(col Collection, getToken getTokenFn) []*RarityScorecard {
	outCh := make(chan Token, col.count)

	// Create a pool of workers
	inCh := make(chan tokenInfo)
	wg := new(sync.WaitGroup)
	for i := 0; i < MAX_WORKERS; i++ {
		wg.Add(1)
		go worker(inCh, wg, outCh, getToken)
	}

	// Download all the tokens
	for i := 1; i < col.count; i++ {
		inCh <- tokenInfo{
			id:     i,
			colUrl: col.url,
		}
	}

	// Wait for all the workers to finish
	wg.Wait()
	close(inCh)

	// Accumulate results
	tokens := make([]Token, 0, col.count)
	for i := 0; i < col.count; i++ {
		t := <-outCh

		if !t.isEmpty() { // Filter out empty tokens (failed downloads)
			tokens = append(tokens, t)
		}
	}

	close(outCh)

	// Map of attributes to map of values to count
	// "hat" -> "red" -> 10
	attrs := make(map[string]map[string]int)
	for _, t := range tokens {
		for k, v := range t.attrs {
			if attrs[k] == nil {
				attrs[k] = make(map[string]int)
			}

			if _, ok := attrs[k][v]; !ok {
				attrs[k][v] = 1
			} else {
				attrs[k][v]++
			}
		}
	}

	// Calculate rarity scores
	scorecards := make([]*RarityScorecard, len(tokens))
	for i, t := range tokens {
		// Filter out empty tokens (failed downloads)
		if t.isEmpty() {
			continue
		}

		scorecards[i] = &RarityScorecard{
			rarity: calculateRarity(t.attrs, attrs),
			id:     t.id,
		}
	}

	sort.Slice(scorecards, func(i, j int) bool {
		return scorecards[i].rarity < scorecards[j].rarity
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
	logger.Println(string(COLOR_GREEN), fmt.Sprintf("Top %d tokens:\n%v", max, scorecards[:max]), string(COLOR_RESET))
}
