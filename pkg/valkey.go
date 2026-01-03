package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/IAmRiteshKoushik/pulse/cmd"
	"github.com/IAmRiteshKoushik/pulse/types"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var Valkey *redis.Client

const (
	Leaderboard = "leaderboard-sset"
	CppRank     = "cpp-ranking-sset"
	JavaRank    = "java-ranking-sset"
	PyRank      = "python-ranking-sset"
	JsRank      = "javascript-ranking-sset"
	GoRank      = "go-ranking-sset"
	RustRank    = "rust-ranking-sset"
	ZigRank     = "zig-ranking-sset"
	FlutterRank = "flutter-ranking-sset"
	KotlinRank  = "kotlin-ranking-sset"
	HaskellRank = "haskell-ranking-sset"

	DocSet  = "doc-set"
	BugSet  = "bug-hunter-set"
	TestSet = "testing-set"
	FeatSet = "feature-suggestion-set"

	LiveUpdateStream = "live-update-stream"
)

func GetParticipantRank(username string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// RevRank returns position in descending order (0-based). If there is no
	// rank available then return -1. If there is an error then also return -1.
	// Incase of unavailability the error is redis.Nil which is bubbled up
	rank, err := Valkey.ZRevRank(ctx, Leaderboard, username).Result()
	if err != nil {
		return -1, err
	}
	return rank + 1, nil
}

// The ParticipantGlobal struct captures username and bounty for the leaderboard
// from the individual sorted sets for each language and returns only the top 2
type ParticipantGlobal struct {
	Username string `json:"github_username"`
	Bounty   string `json:"bounty"`
	Count    string `json:"pull_requests_merged"`
}

func GetLeaderboard() ([]ParticipantGlobal, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	results := []ParticipantGlobal{}
	participants, err := Valkey.ZRevRangeByScoreWithScores(ctx, Leaderboard,
		&redis.ZRangeBy{
			Min: "1",
			Max: "+inf",
		}).Result()
	if err != nil {
		return results, err
	}

	for _, participant := range participants {
		// The leaderboard is fetched from a Redis sorted set where the score
		// is a floating-point number.
		// eg: 100.002
		// Bounty = 100
		// pull_request_count = 2

		// The precision is 3 with the assumption that no one would cross
		// 999 pull-requests in the entire duration of the program (4 months)
		score := participant.Score
		bounty := int(score)
		// To get the count, we subtract the integer part, multiply by 1000
		// and round to handle precision issues.
		count := int(math.Round((score - float64(bounty)) * 1000))

		results = append(results, ParticipantGlobal{
			Username: fmt.Sprintf("%v", participant.Member),
			Bounty:   fmt.Sprintf("%d", bounty),
			Count:    fmt.Sprintf("%d", count),
		})
	}
	return results, nil
}

// The Participant struct captures participant username and pull_request_count
// from the individual sorted sets for each language and returns only the top 2
type Participant struct {
	Username string `json:"github_username"`
	Score    string `json:"pull_request_merged"`
}

func GetTopParticipants() (map[string]map[string]Participant, error) {
	ssets := []string{
		CppRank, JavaRank,
		PyRank, JsRank,
		GoRank, RustRank,
		ZigRank, FlutterRank,
		KotlinRank, HaskellRank,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	results := make(map[string]map[string]Participant)
	var err error
	for _, sset := range ssets {
		name := strings.Split(sset, "-")[0] // extracting the language name
		results[name] = map[string]Participant{
			"first_place": {
				Username: "-",
				Score:    "-1",
			},
			"second_place": {
				Username: "-",
				Score:    "-1",
			},
		}

		members, err := Valkey.ZRevRangeByScoreWithScores(ctx, sset,
			&redis.ZRangeBy{
				Min:    "1",
				Max:    "+inf",
				Offset: 0,
				Count:  2,
			}).Result()

		// If unable to find the sorted-set due to Redis errors
		if err != nil {
			cmd.Log.Error("Error while trying to query sorted-set:"+sset, err)
			return nil, err
		}
		// There can be an edge case where only one participant has made PRs
		// in a particular language. Need to handle that case where there isn't
		// two entries
		if len(members) > 0 {
			results[name]["first_place"] = Participant{
				Username: fmt.Sprintf("%v", members[0].Member),
				Score:    fmt.Sprintf("%d", int(members[0].Score)),
			}
		}
		if len(members) > 1 {
			results[name]["second_place"] = Participant{
				Username: fmt.Sprintf("%v", members[1].Member),
				Score:    fmt.Sprintf("%d", int(members[1].Score)),
			}
		}
	}
	return results, err
}

func GetLatestLiveEvents(c *gin.Context) ([]types.LiveUpdate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	events, err := Valkey.XRevRangeN(ctx, LiveUpdateStream, "+", "-", 5).Result()
	if err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return []types.LiveUpdate{}, nil
	}

	result := []types.LiveUpdate{}
	for _, event := range events {
		data, ok := event.Values["data"].(string)
		if !ok {
			cmd.Log.Fatal("Malformed events in Live-Update Stream.",
				fmt.Errorf("event entry is not in string format"))
			continue // skipping malformed events (should not happen though)
		}
		var entry types.LiveUpdate
		if err := json.Unmarshal([]byte(data), &entry); err != nil {
			cmd.Log.Fatal("Malformed events in Live-Update Stream.", err)
			continue // skipping malformed events (shouldn't happen)
		}
		result = append(result, entry)
	}

	return result, nil
}

func GetParticipantDocCount(username string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	count, err := Valkey.HGet(ctx, DocSet, username).Int()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil && err != redis.Nil {
		return 0, err
	}
	return count, nil
}

func GetParticipantTestCount(username string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	count, err := Valkey.HGet(ctx, TestSet, username).Int()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil && err != redis.Nil {
		return 0, err
	}
	return count, nil
}

func GetParticipantFeatCount(username string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	count, err := Valkey.HGet(ctx, FeatSet, username).Int()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil && err != redis.Nil {
		return 0, err
	}
	return count, nil
}

func GetParticipantBugCount(username string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	count, err := Valkey.HGet(ctx, BugSet, username).Int()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil && err != redis.Nil {
		return 0, err
	}
	return count, nil
}
