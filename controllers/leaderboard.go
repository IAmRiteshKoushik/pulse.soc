package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/IAmRiteshKoushik/pulse/cmd"
	db "github.com/IAmRiteshKoushik/pulse/db/gen"
	"github.com/IAmRiteshKoushik/pulse/pkg"
	"github.com/gin-gonic/gin"
)

func FetchRegistrationBoard(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := cmd.DBPool.Acquire(ctx)
	if err != nil {
		pkg.DbError(c, err)
		return
	}
	defer conn.Release()
	q := db.New()

	profiles, err := q.FetchParticipantListQuery(ctx, conn)
	if err != nil {
		pkg.DbError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "All registered participants fetched successfully",
		"profiles": profiles,
	})
	cmd.Log.Info(fmt.Sprintf(
		"[SUCCESS]: Processed request at %s %s",
		c.Request.Method, c.FullPath(),
	))
}

func FetchLeaderboard(c *gin.Context) {
	leaderboard, err := pkg.GetLeaderboard()
	if err != nil {
		cmd.Log.Error(fmt.Sprintf(
			"[FAILURE]: Failed to fetch leaderboard at %s %s",
			c.Request.Method, c.FullPath(),
		), err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later.",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := cmd.DBPool.Acquire(ctx)
	if err != nil {
		pkg.DbError(c, err)
		return
	}
	defer conn.Release()

	q := db.New()
	zeroes, err := q.FetchUsersWithNoContributionsQuery(ctx, conn)
	if err != nil {
		pkg.DbError(c, err)
		return
	}

	for _, z := range zeroes {
		leaderboard = append(leaderboard, pkg.ParticipantGlobal{
			Username: z.GithubUsername.String,
			Bounty:   fmt.Sprintf("%d", z.Bounty),
			Count:    fmt.Sprintf("%d", z.Solutions),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Leaderboard fetched successfully",
		"leaderboard": leaderboard,
	})
	cmd.Log.Info(fmt.Sprintf(
		"[SUCCESS]: Processed request at %s %s",
		c.Request.Method, c.FullPath(),
	))
}

// Hall-Of-Fame is the language-wise leaderboard where the first two
// participants of each language are returned
func FetchHallOfFame(c *gin.Context) {
	leaderboard, err := pkg.GetTopParticipants()
	if err != nil {
		cmd.Log.Error(fmt.Sprintf(
			"[FAILURE]: Failed to fetch language-wise leaderboard at %s %s",
			c.Request.Method, c.FullPath(),
		), err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Language-wise leaderboard fetched successfully",
		"leaderboards": leaderboard,
	})
	cmd.Log.Info(fmt.Sprintf(
		"[SUCCESS]: Processed request at %s %s",
		c.Request.Method, c.FullPath(),
	))

}
