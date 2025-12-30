package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/IAmRiteshKoushik/pulse/cmd"
	db "github.com/IAmRiteshKoushik/pulse/db/gen"
	"github.com/IAmRiteshKoushik/pulse/pkg"
	"github.com/IAmRiteshKoushik/pulse/types"
	"github.com/gin-gonic/gin"
)

//	func RegisterUserAccount(c *gin.Context) {
//		var body types.RegisterUserRequest
//		if err := c.BindJSON(&body); err != nil {
//			pkg.JSONUnmarshallError(c, err)
//			return
//		}
//		if err := body.Validate(); err != nil {
//			pkg.RequestValidatorError(c, err)
//			return
//		}
//
//		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//		defer cancel()
//
//		otp, err := pkg.GenerateOTP()
//		if err != nil {
//			cmd.Log.Error(
//				fmt.Sprintf("Failed to generate OTP at %s %s", c.Request.Method, c.FullPath()), err)
//			c.JSON(http.StatusInternalServerError, gin.H{
//				"message": "Oops! Something happened. Please try again later.",
//			})
//			return
//		}
//
//		tempToken, err := pkg.CreateToken(body.GhUsername, body.Email, "temp_token")
//		if err != nil {
//			cmd.Log.Fatal(
//				fmt.Sprintf("Failed to generate access token at %s %s.",
//					c.Request.Method, c.FullPath()), err)
//			c.JSON(http.StatusInternalServerError, gin.H{
//				"message": "Oops! Something happened. Please try again later.",
//			})
//			return
//		}
//
//		tx, err := cmd.DBPool.Begin(ctx)
//		if err != nil {
//			pkg.DbError(c, err)
//			return
//		}
//		defer tx.Rollback(ctx)
//
//		q := db.New()
//		ok, err := q.CheckUserExistQuery(ctx, tx, pgtype.Text{String: body.GhUsername, Valid: true})
//		if err != nil {
//			cmd.Log.Debug("Error here")
//			pkg.DbError(c, err)
//			return
//		}
//		if ok {
//			cmd.Log.Warn(
//				fmt.Sprintf("Username exists, registration failed at %s %s",
//					c.Request.Method, c.FullPath()))
//			c.JSON(http.StatusConflict, gin.H{
//				"message": "Username already exists",
//			})
//			return
//		}
//
//		result, err := q.BeginUserRegistrationQuery(ctx, tx,
//			db.BeginUserRegistrationQueryParams{
//				FirstName:  body.FirstName,
//				MiddleName: pgtype.Text{String: body.MiddleName, Valid: true},
//				LastName:   body.LastName,
//				Email:      body.Email,
//				Ghusername: pgtype.Text{String: body.GhUsername, Valid: true},
//				Otp:        otp,
//			})
//		if err != nil {
//			pkg.DbError(c, err)
//			return
//		}
//		if err := tx.Commit(ctx); err != nil {
//			pkg.DbError(c, err)
//			return
//		}
//
//		// Database transaction fails if mail is not sent
//		err = pkg.SendMail([]string{result.Email}, result.Otp)
//		if err != nil {
//			cmd.Log.Error(
//				fmt.Sprintf("Failed to send email at %s %s", c.Request.Method, c.FullPath()),
//				err)
//			c.JSON(http.StatusInternalServerError, gin.H{
//				"message": "Oops! Something happened. Please try again later.",
//			})
//			return
//		}
//
//		c.JSON(http.StatusOK, gin.H{
//			"message":    "User onboarding has been initiated.",
//			"access_key": tempToken,
//		})
//		cmd.Log.Info(fmt.Sprintf(
//			"[SUCCESS]: Processed request at %s %s",
//			c.Request.Method, c.FullPath(),
//		))
//	}
//
//	func RegisterUserOtpVerify(c *gin.Context) {
//		username, ok := pkg.GrabUsername(c)
//		if !ok {
//			cmd.Log.Warn(
//				fmt.Sprintf("Failed to extract username from token at %s %s",
//					c.Request.Method, c.FullPath()))
//			c.JSON(http.StatusInternalServerError, gin.H{
//				"message": "Oops! Something happened. Please try again later.",
//			})
//			return
//		}
//
//		var body types.RegisterUserOtpVerifyRequest
//		if err := c.BindJSON(&body); err != nil {
//			pkg.JSONUnmarshallError(c, err)
//			return
//		}
//		if err := body.Validate(); err != nil {
//			pkg.RequestValidatorError(c, err)
//			return
//		}
//
//		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//		defer cancel()
//
//		tx, err := cmd.DBPool.Begin(ctx)
//		if err != nil {
//			pkg.DbError(c, err)
//			return
//		}
//		defer tx.Rollback(ctx)
//
//		q := db.New()
//		verifiedUser, err := q.VerifyOtpQuery(ctx, tx, db.VerifyOtpQueryParams{
//			Ghusername: pgtype.Text{String: username, Valid: true},
//			Otp:        body.Otp,
//		})
//		if err != nil {
//			pkg.DbError(c, err)
//			return
//		}
//		if verifiedUser.Email == "" {
//			cmd.Log.Warn(
//				fmt.Sprintf("Username grabbed from token not found in DB at %s %s",
//					c.Request.Method, c.FullPath()))
//			c.JSON(http.StatusForbidden, gin.H{
//				"message": "Server refused to process the request",
//			})
//			return
//		}
//
//		onboardGhUsername, err := q.CreateUserAccountQuery(ctx, tx, db.CreateUserAccountQueryParams(verifiedUser))
//		if err != nil {
//			pkg.DbError(c, err)
//			return
//		}
//
//		if onboardGhUsername.String == "" {
//			cmd.Log.Warn(
//				fmt.Sprintf("Failed to onboard user at %s %s", c.Request.Method, c.FullPath()))
//			c.JSON(http.StatusInternalServerError, gin.H{
//				"message": "Oops! Something happened. Please try again later.",
//			})
//			return
//		}
//
//		if err := tx.Commit(ctx); err != nil {
//			pkg.DbError(c, err)
//			return
//		}
//
//		c.JSON(http.StatusOK, gin.H{
//			"message":         "User Registration successful.",
//			"github_username": onboardGhUsername,
//		})
//		cmd.Log.Info(fmt.Sprintf(
//			"[SUCCESS]: Processed request at %s %s",
//			c.Request.Method, c.FullPath()))
//	}
//
//	func RegisterUserOtpResend(c *gin.Context) {
//		username, ok := pkg.GrabUsername(c)
//		if !ok {
//			cmd.Log.Warn(
//				fmt.Sprintf("Failed to extract username from token at %s %s",
//					c.Request.Method, c.FullPath()))
//			c.JSON(http.StatusInternalServerError, gin.H{
//				"message": "Oops! Something happened. Please try again later.",
//			})
//			return
//		}
//
//		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//		defer cancel()
//
//		conn, err := cmd.DBPool.Acquire(ctx)
//		if err != nil {
//			pkg.DbError(c, err)
//			return
//		}
//		defer conn.Release()
//
//		q := db.New()
//		result, err := q.CheckForExistingOtpQuery(ctx, conn, pgtype.Text{String: username, Valid: true})
//		if err != nil {
//			pkg.DbError(c, err)
//			return
//		}
//
//		err = pkg.SendMail([]string{result.Email}, result.Otp)
//		if err != nil {
//			cmd.Log.Error(
//				fmt.Sprintf("Failed to send email at %s %s", c.Request.Method, c.FullPath()),
//				err)
//			c.JSON(http.StatusInternalServerError, gin.H{
//				"message": "Oops! Something happened. Please try again later.",
//			})
//			return
//		}
//
//		c.JSON(http.StatusOK, gin.H{
//			"message": "User OTP resent at specified email address",
//		})
//		cmd.Log.Info(fmt.Sprintf(
//			"[SUCCESS]: Processed request at %s %s",
//			c.Request.Method, c.FullPath(),
//		))
//	}

func LoginUser(c *gin.Context) {
	var body types.LoginUserRequest
	if err := c.BindJSON(&body); err != nil {
		pkg.JSONUnmarshallError(c, err)
		return
	}
	if err := body.Validate(); err != nil {
		pkg.RequestValidatorError(c, err)
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
	user, err := q.GetUserByEmail(ctx, conn, body.Email)
	if err != nil {
		cmd.Log.Warn(fmt.Sprintf("Login failed: Email %s not found or DB error: %v", body.Email, err))
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid email or password"})
		return
	}

	if err := pkg.CompareHash(user.Password, body.Password); err != nil {
		cmd.Log.Warn(fmt.Sprintf("Login failed: Incorrect password for email %s", body.Email))
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid email or password"})
		return
	}

	accessToken, err := pkg.CreateToken(user.Ghusername, user.Email, "access_token")
	if err != nil {
		cmd.Log.Error(fmt.Sprintf("Failed to generate token for %s", user.Email), err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
		return
	}

	refreshToken, err := pkg.CreateToken(user.Ghusername, user.Email, "refresh_token")
	if err != nil {
		cmd.Log.Error(fmt.Sprintf("Failed to generate refresh token for %s", user.Email), err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Login successful",
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"ghUsername":   user.Ghusername,
	})

	cmd.Log.Info(fmt.Sprintf("[SUCCESS]: User logged in: %s", user.Email))
}
