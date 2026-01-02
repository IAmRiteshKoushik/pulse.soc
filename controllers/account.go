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
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
)

func FetchUserAccount(c *gin.Context) {
	username := c.Query("user")
	if username == "" {
		cmd.Log.Warn(
			fmt.Sprintf(
				"Query param 'user' not provided at %s %s",
				c.Request.Method,
				c.FullPath(),
			),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Query parameter 'user' is required.",
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
	user, err := q.FetchProfileQuery(ctx, conn, pgtype.Text{String: username, Valid: true})
	if err != nil {
		pkg.DbError(c, err)
		return
	}

	rank, err := pkg.GetParticipantRank(username)
	if err != nil {
		if err != redis.Nil {
			cmd.Log.Error(
				fmt.Sprintf("Failed to fetch rank at %s %s",
					c.Request.Method,
					c.FullPath(),
				),
				err,
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Oops! Something happened. Please try again",
			})
			return
		}
		cmd.Log.Warn(
			fmt.Sprintf("Rank does not exist yet at %s %s",
				c.Request.Method, c.FullPath()),
		)
	}

	doc, err := pkg.GetParticipantDocCount(username)
	if err != nil {
		cmd.Log.Error(
			fmt.Sprintf("Failed to fetch documentation_count at %s %s",
				c.Request.Method,
				c.FullPath(),
			),
			err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again",
		})
		return
	}

	test, err := pkg.GetParticipantTestCount(username)
	if err != nil {
		cmd.Log.Error(
			fmt.Sprintf("Failed to fetch impact_count at %s %s",
				c.Request.Method,
				c.FullPath(),
			),
			err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again",
		})
		return
	}

	feat, err := pkg.GetParticipantFeatCount(username)
	if err != nil {
		cmd.Log.Error(
			fmt.Sprintf("Failed to fetch feature_count at %s %s",
				c.Request.Method,
				c.FullPath(),
			),
			err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again",
		})
		return
	}

	bugs, err := pkg.GetParticipantBugCount(username)
	if err != nil {
		cmd.Log.Error(
			fmt.Sprintf("Failed to fetch bug_report_count at %s %s",
				c.Request.Method,
				c.FullPath(),
			),
			err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":               "User profile retrived successfully",
		"github_username":       user.Ghusername,
		"first_name":            user.FirstName,
		"middle_name":           user.MiddleName.String,
		"last_name":             user.LastName,
		"bounty":                user.Bounty,
		"pull_request_count":    user.SolutionsAccepted + user.SolutionsPending,
		"pull_request_merged":   user.SolutionsAccepted,
		"pull_request_unmerged": user.SolutionsPending,
		"pending_issue_count":   user.ActiveClaims,
		"rank":                  rank,
		"documentation_count":   doc,
		"bug_report_count":      bugs,
		"feature_count":         feat,
		"test_count":            test,
		"badges":                user.Badges,
	})
	cmd.Log.Info(
		fmt.Sprintf("Successfully retrived user profile at %s %s",
			c.Request.Method, c.FullPath(),
		),
	)
}
