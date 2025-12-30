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

// WebhookRequest
type WebhookRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

func WebhookHandler(c *gin.Context) {
	var req WebhookRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		cmd.Log.Warn(
			fmt.Sprintf("[INVALID-PAYLOAD]: Invalid request payload at %s %s",
				c.Request.Method,
				c.FullPath(),
			),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request payload",
			"error":   err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Acquire connection from the pool
	conn, err := cmd.DBPool.Acquire(ctx)
	if err != nil {
		pkg.DbError(c, err)
		return
	}
	defer conn.Release()

	q := db.New()

	_, err = q.CreateUserAccountQuery(ctx, conn, db.CreateUserAccountQueryParams{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  req.Password,
	})
	if err != nil {
		pkg.DbError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Webhook data processed and user created successfully",
	})

	cmd.Log.Info(fmt.Sprintf(
		"[SUCCESS]: Processed webhook request at %s %s",
		c.Request.Method, c.FullPath(),
	))
}
