package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRarityScorecards(t *testing.T) {
	getTokenStub := func(tid int, colUrl string) (Token, error) {
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
		url:   "2_hats",
	}, getTokenStub)

	assert.Len(t, scorecards, 2)
	assert.Equal(t, scorecards[0].rarity, 0.25)
	assert.Equal(t, scorecards[1].rarity, 0.25)
}

func TestGetRarityScorecardsNoResponse(t *testing.T) {
	getTokenStub := func(tid int, colUrl string) (Token, error) {
		return Token{}, nil
	}

	scorecards := getRarityScorecards(Collection{
		count: 2,
		url:   "empty",
	}, getTokenStub)

	assert.Len(t, scorecards, 0)
}
