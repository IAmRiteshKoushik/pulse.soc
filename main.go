package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	cmd "github.com/IAmRiteshKoushik/pulse/cmd"
	c "github.com/IAmRiteshKoushik/pulse/controllers"
	mw "github.com/IAmRiteshKoushik/pulse/middleware"
	"github.com/IAmRiteshKoushik/pulse/pkg"
	"github.com/IAmRiteshKoushik/pulse/sse"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func init() {
	prometheus.MustRegister(mw.RequestCount, mw.RequestLatency)
}

func StartApp() {
	failMsg := "Could not initialize app\n%w"

	// Initialize configuration
	config, err := cmd.LoadConfig()
	if err != nil {
		panic(fmt.Errorf(failMsg, err))
	}
	cmd.AppConfig = config
	log.Println("[OK]: Configuration loaded successfully")

	// Initialize logger
	f, err := os.OpenFile("app.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Errorf(failMsg, err))
	}
	defer f.Close()
	cmd.Log = cmd.NewLoggerService(cmd.AppConfig.Environment, f)
	cmd.Log.Info("[ACTIVE]: Logging service configured successfully.")

	// Initialize database connection pool
	cmd.DBPool, err = cmd.InitDB()
	if err != nil {
		panic(fmt.Errorf(failMsg, err))
	}
	cmd.Log.Info("[ACTIVE]: DB initialized successfully.")

	// Initialize oAuth
	cmd.OAuthInit()
	cmd.Log.Info("[ACTIVE]: GitHub oAuth configuration loaded successfully.")

	// Initialize Valkey
	client, err := cmd.InitValkey()
	if err != nil {
		return
	}
	pkg.Valkey = client
	cmd.Log.Info("[ACTIVE]: Valkey service is online.")

	// Initialize LiveServer
	sse.Live = sse.NewLiveServer()
	mw.RecoverRoutine("LiveServer", sse.Live.Start)
	cmd.Log.Info("[ACTIVE]: LiveServer is online.")

	// Initialize the Valkey ReadStream
	mw.RecoverRoutine("readValkeyStream", func() {
		pkg.ReadStream(sse.Live)
	})
	cmd.Log.Info("[ACTIVE]: Valkey-Stream online.")

	// Starting the server
	ginLogs, err := os.Create("gin.log")
	if err != nil {
		cmd.Log.Fatal("[CRASH]: Error creating log file for Gin", err)
	}
	defer ginLogs.Close()
	multiWriter := io.MultiWriter(os.Stdout, ginLogs)
	gin.DefaultWriter = multiWriter
	gin.DefaultErrorWriter = multiWriter
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(mw.RecoveryMiddleware)
	router.Use(gin.Logger())
	// router.Use(mw.MonitorMiddleware)
	router.Use(cors.New(cors.Config{
		// Webhook domain hardcoded for now, move to array based environment
		// variable later
		AllowOrigins:     []string{cmd.AppConfig.FrontendURL, "https://webhooksoc.vercel.app"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Server is LIVE",
		})
		cmd.Log.Info(fmt.Sprintf(
			"[SUCCESS]: Processed request at %s %s",
			c.Request.Method, c.FullPath(),
		))
	})

	// Prometheus endpoint, not exposed externally via Nginx
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	v1 := router.Group("/api/v1")

	v1.GET("/auth/github", c.InitiateGitHubOAuth)
	v1.GET("/auth/github/callback", c.CompleteGitHubOAuth)
	v1.POST("/auth/register", c.RegisterUserAccount)
	v1.POST("/auth/register/otp/verify", mw.Auth, c.RegisterUserOtpVerify)
	v1.GET("/auth/register/otp/resend", mw.Auth, c.RegisterUserOtpResend)
	v1.GET("/auth/refresh", c.RegenerateToken)

	v1.GET("/profile", mw.Auth, c.FetchUserAccount)

	v1.GET("/leaderboard", c.FetchLeaderboard)
	v1.GET("/registrations", c.FetchRegistrationBoard)
	v1.GET("/projects", c.FetchProjects)
	v1.GET("/issues/:projectId", c.FetchIssues)
	v1.GET("/updates/latest", c.FetchLatestUpdates)
	v1.GET("/updates/live", c.SetupLiveUpdates)
	v1.GET("/hof", c.FetchHallOfFame)

	port := strconv.Itoa(cmd.AppConfig.Port)
	cmd.Log.Info("[ACTIVE]: Server configured and starting on PORT " + port)
	err = router.Run(":" + port)
	if err != nil {
		cmd.Log.Fatal("[CRASH]: Server failed to start", err)
		panic(err)
	}

	cmd.CloseValkey(pkg.Valkey)
	cmd.Log.Info("[DEACTIVE]: Valkey offline.")
	cmd.Log.Info("[DEACTIVE]: Logging service offline.")
}

func main() {
	StartApp()
}
