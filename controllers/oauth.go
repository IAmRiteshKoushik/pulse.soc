package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/IAmRiteshKoushik/pulse/cmd"
	db "github.com/IAmRiteshKoushik/pulse/db/gen"
	"github.com/IAmRiteshKoushik/pulse/pkg"
	"github.com/IAmRiteshKoushik/pulse/types"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

func InitiateGitHubOAuth(c *gin.Context) {
	email := c.GetString("email")
	if email == "" {
		cmd.Log.Warn("Failed to email from context")
		c.JSON(http.StatusUnauthorized, gin.H{"message": "You must be logged in to link your GitHub account."})
		return
	}

	stateToken, err := pkg.CreateToken("", email, "temp_token")
	if err != nil {
		cmd.Log.Error("Failed to generate state token", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
		return
	}

	url := cmd.GithubOAuthConfig.AuthCodeURL(stateToken)

	c.JSON(http.StatusOK, gin.H{
		"url": url,
	})

	cmd.Log.Info(fmt.Sprintf("Initiated GitHub Link for %s", email))
}

func CompleteGitHubOAuth(c *gin.Context) {
	redirectUrl := cmd.AppConfig.FrontendURL

	// Extract code from github oauth callback URL
	code := c.Query("code")
	stateToken := c.Query("state")
	if code == "" || stateToken == "" {
		cmd.Log.Warn(
			fmt.Sprintf("Missing authorization code or state token in github oauth callback at %s %s",
				c.Request.Method, c.FullPath()))
		c.Redirect(http.StatusTemporaryRedirect, cmd.AppConfig.FrontendURL)
		return
	}

	claims, err := pkg.VerifyToken(stateToken)
	if err != nil || claims.TokenType != "temp_token" {
		cmd.Log.Warn(fmt.Sprintf("Invalid state token: %v", err))
		c.Redirect(http.StatusTemporaryRedirect, redirectUrl)
		return
	}
	userEmail := claims.Email

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Fetching the github user
	token, err := cmd.GithubOAuthConfig.Exchange(ctx, code)
	if err != nil {
		fmt.Printf("%v", cmd.GithubOAuthConfig)
		cmd.Log.Error(
			fmt.Sprintf("Failed to exchange code for token at %s %s",
				c.Request.Method, c.FullPath()), err)
		c.Redirect(http.StatusTemporaryRedirect, cmd.AppConfig.FrontendURL)
		return
	}

	client := cmd.GithubOAuthConfig.Client(ctx, token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		cmd.Log.Warn(
			fmt.Sprintf("Failed to fetch user info from GitHub at %s %s",
				c.Request.Method, c.FullPath()))
		c.Redirect(http.StatusTemporaryRedirect, cmd.AppConfig.FrontendURL)
		// c.JSON(http.StatusInternalServerError, gin.H{
		// 	"message": "Oops! Something happened. Please try again later",
		// })
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		cmd.Log.Warn(fmt.Sprintf("Failed to unmarshal github user info at %s %s",
			c.Request.Method, c.FullPath()))
		c.Redirect(http.StatusTemporaryRedirect, cmd.AppConfig.FrontendURL)
		// c.JSON(http.StatusInternalServerError, gin.H{
		// 	"message": "Oops! Something happened. Please try again later",
		// })
		return
	}
	// Extracting the github user
	var ghUser types.GithubUser
	if err := json.Unmarshal(body, &ghUser); err != nil {
		cmd.Log.Error(
			fmt.Sprintf("Failed to parse github user info at %s %s",
				c.Request.Method, c.FullPath()), err)
		c.Redirect(http.StatusTemporaryRedirect, cmd.AppConfig.FrontendURL)
		// c.JSON(http.StatusInternalServerError, gin.H{
		// 	"message": "Oops! Something happened. Please try again later",
		// })
		return
	}

	// Verifying the github account's presence against database to validate
	// post registration
	tx, err := cmd.DBPool.Begin(ctx)
	if err != nil {
		// pkg.DbError(c, err)
		c.Redirect(http.StatusTemporaryRedirect, cmd.AppConfig.FrontendURL)
		return
	}
	defer tx.Rollback(ctx)

	q := db.New()

	err = q.UpdateUserGithubUsername(ctx, tx, db.UpdateUserGithubUsernameParams{
		Ghusername: pgtype.Text{String: ghUser.Username, Valid: true},
		Email:      userEmail,
	})

	if err != nil {
		cmd.Log.Error(fmt.Sprintf("Failed to link GitHub account %s to %s", ghUser.Username, userEmail), err)
		c.Redirect(http.StatusTemporaryRedirect, redirectUrl+"?error=account_already_linked")
		return
	}

	userExist, err := q.RetriveExistingUserQuery(ctx, tx, pgtype.Text{String: ghUser.Username, Valid: true})
	if err != nil {
		// pkg.DbError(c, err)
		c.Redirect(http.StatusTemporaryRedirect, redirectUrl)
		return
	}

	// If the presence is verified, then generate access and refresh token
	// , add them in DB and respond back in request
	accessToken, err := pkg.CreateToken(userExist.Ghusername.String, userExist.Email, "access_token")
	if err != nil {
		cmd.Log.Error(
			fmt.Sprintf("Failed to create access token at %s %s", c.Request.Method, c.FullPath()),
			err)
		c.Redirect(http.StatusTemporaryRedirect, redirectUrl)
		// c.JSON(http.StatusInternalServerError, gin.H{
		// 	"message": "Oops! Something happened. Please try again later",
		// })
		return
	}
	refreshToken, err := pkg.CreateToken(userExist.Ghusername.String, userExist.Email, "refresh_token")
	if err != nil {
		cmd.Log.Error(
			fmt.Sprintf("Failed to create token at %s %s", c.Request.Method, c.FullPath()),
			err)
		c.Redirect(http.StatusTemporaryRedirect, redirectUrl)
		// c.JSON(http.StatusInternalServerError, gin.H{
		// 	"message": "Oops! Something happened. Please try again later",
		// })
		return
	}

	loginUser, err := q.AddRefreshTokenQuery(ctx, tx, db.AddRefreshTokenQueryParams{
		Email:        userExist.Email,
		RefreshToken: pgtype.Text{String: refreshToken, Valid: true},
	})
	if err != nil {
		cmd.Log.Error(
			fmt.Sprintf("Failed to process request at %s %s", c.Request.Method, c.FullPath()), err)
		c.Redirect(http.StatusTemporaryRedirect, redirectUrl)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		cmd.Log.Error(fmt.Sprintf("Failed to process request at %s %s", c.Request.Method, c.FullPath()), err)
		c.Redirect(http.StatusTemporaryRedirect, redirectUrl)
		return
	}

	// Redirect back to frontend with tokens
	frontendURL := fmt.Sprintf("%s/auth/callback", cmd.AppConfig.FrontendURL)
	redirectURL := fmt.Sprintf("%s?access_token=%s&refresh_token=%s&github_username=%s&email=%s&bounty=%d",
		frontendURL,
		accessToken,
		refreshToken,
		loginUser.Ghusername.String,
		loginUser.Email,
		loginUser.Bounty,
	)

	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
	cmd.Log.Info(fmt.Sprintf(
		"[SUCCESS]: Processed request at %s %s",
		c.Request.Method, c.FullPath(),
	))
}

func RegenerateToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		cmd.Log.Warn(
			fmt.Sprintf("RefreshToken not sent as Authorization header at %s %s",
				c.Request.Method, c.FullPath()),
		)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Authorization header is missing in request",
		})
		return
	}

	tokenString := ""
	if len(authHeader) > 7 && authHeader[0:7] == "Bearer " {
		tokenString = authHeader[7:]
	} else {
		cmd.Log.Warn(fmt.Sprintf("Authorization failed at %s %s", c.Request.Method, c.FullPath()))
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "Invalid Authorization header format",
		})
		return
	}

	claims, err := pkg.VerifyToken(tokenString)
	if err != nil {
		cmd.Log.Warn(
			fmt.Sprintf("Invalid refresh token at %s %s",
				c.Request.Method, c.FullPath()),
		)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "The request is malformed",
		})
		return
	}

	validIssuer := claims.Issuer == "api.season-of-code"
	validTokenType := claims.TokenType == "refresh_token"
	if !validIssuer || !validTokenType {
		cmd.Log.Error(fmt.Sprintf("Tampered token at %s %s", c.Request.Method, c.FullPath()), err)
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Server refused to process the request"})
		return
	}

	// Actual controller
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := cmd.DBPool.Acquire(ctx)
	if err != nil {
		pkg.DbError(c, err)
		return
	}
	defer conn.Release()

	q := db.New()
	result, err := q.CheckRefreshTokenQuery(ctx, conn, db.CheckRefreshTokenQueryParams{
		Email:        claims.Email,
		RefreshToken: pgtype.Text{String: tokenString, Valid: true},
	})
	if err != nil {
		pkg.DbError(c, err)
		return
	}

	accessToken, err := pkg.CreateToken(result.Ghusername.String, result.Email, "access_token")
	if err != nil {
		cmd.Log.Error(
			fmt.Sprintf("Could not generate access token at %s %s", c.Request.Method, c.FullPath()),
			err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Token refreshed successfully",
		"accessKey": accessToken,
	})
	cmd.Log.Info(fmt.Sprintf(
		"[SUCCESS]: Processed request at %s %s",
		c.Request.Method, c.FullPath(),
	))
}
