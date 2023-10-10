package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRarityScorecards(t *testing.T) {
	getToken := func(tid int, colUrl string) (Token, error) {
		switch tid {
		case 1:
			return Token{id: 1, attrs: map[string]string{
				"hat":     "green beret",
				"earring": "gold",
			}}, nil
		case 2:
			return Token{id: 2, attrs: map[string]string{
				"hat":     "green beret",
				"earring": "silver",
			}}, nil
		default:
			panic("unexpected token id")
		}
	}

	scorecards := getRarityScorecards(Collection{
		count: 2,
		url:   "",
	}, getToken)

	assert.Len(t, scorecards, 2)
	assert.Equal(t, scorecards[0].rarity, 1.0)
	assert.Equal(t, scorecards[1].rarity, 1.0)
}

func TestGetRarityScorecardsNoResponse(t *testing.T) {
	getToken := func(tid int, colUrl string) (Token, error) {
		return Token{}, nil
	}

	scorecards := getRarityScorecards(Collection{
		count: 2,
		url:   "",
	}, getToken)

	assert.Len(t, scorecards, 0)
}

func TestGetToken(t *testing.T) {
	table := []struct {
		tid        int
		colUrl     string
		statusCode int
		body       string
	}{
		{1, "asuki", 200, "My dog used to chase people on a bike a lot. It got so bad I had to take his bike away."},
		{2, "asuki", 404, `Joke with id "173782" not found`},
		{3, "asuki", 400, "Joke ID cannot be empty"},
	}

	for _, v := range table {
		t.Run(fmt.Sprintf("%d", v.tid), func(t *testing.T) {
			httptest.NewRecorder()
			httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s/%d.json", URL, v.colUrl, v.tid), nil)

			token, err := getToken(v.tid, v.colUrl)
			if v.statusCode != 200 {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, token.id, v.tid)
			}
		})
	}
}
