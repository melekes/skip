package main

import (
	"errors"
	"fmt"
	"testing"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
)

func TestGetRarityScorecards(t *testing.T) {
	getTokenStub := func(tid int, colUrl string) (*Token, error) {
		switch tid {
		case 1:
			return &Token{id: 1, attrs: map[string]string{
				"hat":     "green beret",
				"earring": "gold",
			}}, nil
		case 2:
			return &Token{id: 2, attrs: map[string]string{
				"hat":     "green beret",
				"earring": "silver",
			}}, nil
		case 3:
			return &Token{id: 3, attrs: map[string]string{
				"hat":     "green beret",
				"earring": "silver",
			}}, nil
		default:
			panic("unexpected token id")
		}
	}

	scorecards := getRarityScorecards(Collection{
		count: 3,
		url:   "3_hats",
	}, getTokenStub)

	assert.Len(t, scorecards, 3)
	assert.Equal(t, scorecards[0].rarity, 0.8333333333333333)
	assert.Equal(t, scorecards[1].rarity, 0.5833333333333333)
	assert.Equal(t, scorecards[2].rarity, 0.5833333333333333)
}

func TestGetRarityScorecardsNoResponse(t *testing.T) {
	getTokenStub := func(tid int, colUrl string) (*Token, error) {
		return nil, errors.New("no response")
	}

	scorecards := getRarityScorecards(Collection{
		count: 2,
		url:   "empty",
	}, getTokenStub)

	assert.Len(t, scorecards, 0)
}

func TestGetToken(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution

	gock.New(URL).
		Get(fmt.Sprintf("test/%d.json", 1)).
		Reply(200).
		JSON(map[string]string{"hat": "green beret"})

	token, err := getToken(1, "test")
	assert.NoError(t, err)
	assert.Equal(t, token.id, 1)
	assert.Equal(t, token.attrs, map[string]string{"hat": "green beret"})
}
