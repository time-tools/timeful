package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"timeful/server/appenv"
	"timeful/server/logger"
	"timeful/server/postgres"
	"timeful/server/routes"
	"timeful/server/services/gcloud"
	"timeful/server/slackbot"
	"timeful/server/utils"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "timeful/server/docs"
)

const defaultLogPath = "logs/server.log"

var postgresPing = postgres.Ping

// @title Timeful API
// @version 1.0
// @description This is the API for Timeful.

// @host localhost:3002/api

func init() {
	mime.AddExtensionType(".css", "text/css")
	mime.AddExtensionType(".js", "application/javascript")
	mime.AddExtensionType(".svg", "image/svg+xml")
	mime.AddExtensionType(".woff", "font/woff")
	mime.AddExtensionType(".woff2", "font/woff2")
	mime.AddExtensionType(".ttf", "font/ttf")
	mime.AddExtensionType(".json", "application/json")
	mime.AddExtensionType(".map", "application/json")
}

func main() {
	// Set release flag
	release := flag.Bool("release", false, "Whether this is the release version of the server")
	flag.Parse()
	// Init logfile
	logFile, err := openLogFile(defaultLogPath)
	if err != nil {
		log.Fatal(err)
	}
	gin.DefaultWriter = io.MultiWriter(logFile, os.Stdout)

	// Init logger
	logger.Init(logFile)

	currentAppEnv := configureRuntime()
	if *release || shouldRunInReleaseMode(currentAppEnv) {
		os.Setenv("GIN_MODE", "release")
		gin.SetMode(gin.ReleaseMode)
	} else {
		os.Setenv("GIN_MODE", "debug")
	}

	// Init router
	router := gin.New()
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		var statusColor, methodColor, resetColor string
		if param.IsOutputColor() {
			statusColor = param.StatusCodeColor()
			methodColor = param.MethodColor()
			resetColor = param.ResetColor()
		}

		if param.Latency > time.Minute {
			param.Latency = param.Latency.Truncate(time.Second)
		}
		return fmt.Sprintf("%v |%s %3d %s| %13v | %15s |%s %-7s %s %#v\n%s",
			param.TimeStamp.Format("2006/01/02 15:04:05"),
			statusColor, param.StatusCode, resetColor,
			param.Latency,
			param.ClientIP,
			methodColor, param.Method, resetColor,
			param.Path,
			param.ErrorMessage,
		)
	}))
	router.Use(gin.Recovery())

	// The public app origin is always allowed; CORS_ORIGINS adds non-canonical origins.
	corsOrigins, err := utils.CORSOrigins(os.Getenv("CORS_ORIGINS"))
	if err != nil {
		logger.StdErr.Panicln(err)
	}
	router.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins,
		AllowMethods:     []string{"GET", "POST", "PATCH", "PUT", "DELETE"},
		AllowHeaders:     []string{"Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Init database
	closePostgres := postgres.Init()
	defer closePostgres()

	// Init google cloud stuff
	closeTasks := gcloud.InitTasks()
	defer closeTasks()

	// Session
	store := cookie.NewStore([]byte(os.Getenv("SESSION_SECRET")))
	router.Use(sessions.Sessions("session", store))

	// Init routes
	apiRouter := router.Group("/api")
	initHealthRoute(apiRouter)
	routes.InitAuth(apiRouter)
	routes.InitUser(apiRouter)
	routes.InitUsers(apiRouter)
	routes.InitEvents(apiRouter)
	routes.InitAnalytics(apiRouter)
	routes.InitFolders(apiRouter)
	slackbot.InitSlackbot(apiRouter)

	frontendDist := os.Getenv("FRONTEND_DIST")
	if frontendDist == "" {
		frontendDist = "./frontend/dist"
		if _, err := os.Stat(frontendDist); os.IsNotExist(err) {
			frontendDist = "../frontend/dist"
		}
	}

	indexPath := filepath.Join(frontendDist, "index.html")
	hasFrontendIndex := false
	if _, err := os.Stat(indexPath); err == nil {
		router.LoadHTMLFiles(indexPath)
		hasFrontendIndex = true
	} else {
		logger.StdErr.Printf("Warning: index.html not found at %s", indexPath)
	}
	router.GET("/e/:eventId", eventPageHandler(hasFrontendIndex))
	router.NoRoute(noRouteHandler())

	// Init swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	port, err := appenv.ResolvePort(currentAppEnv, os.Getenv("APP_PORT"))
	if err != nil {
		logger.StdErr.Fatal(err)
	}

	// Run server
	router.Run(":" + port)
}

func initHealthRoute(apiRouter *gin.RouterGroup) {
	apiRouter.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	apiRouter.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := postgresPing(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

func shouldRunInReleaseMode(env appenv.Environment) bool {
	return appenv.ShouldUseReleaseMode(os.Getenv("GIN_MODE"), env)
}

func openLogFile(path string) (*os.File, error) {
	logDir := filepath.Dir(path)
	if logDir != "." {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, err
		}
	}

	return os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}

// configureRuntime validates Compose-injected configuration before startup.
func configureRuntime() appenv.Environment {
	currentAppEnv := appenv.Current()

	validateSessionSecret()
	if err := utils.ValidateBaseUrl(); err != nil {
		logger.StdErr.Panicln(err)
	}

	return currentAppEnv
}

// validateSessionSecret ensures SESSION_SECRET is set and meets security requirements
func validateSessionSecret() {
	secret := os.Getenv("SESSION_SECRET")

	if secret == "" {
		logger.StdErr.Panicln("SESSION_SECRET environment variable is required but not set")
	}

	// Minimum 32 characters for adequate security (256 bits)
	if len(secret) < 32 {
		logger.StdErr.Panicln("SESSION_SECRET must be at least 32 characters long")
	}
}

func eventPageHandler(hasFrontendIndex bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !hasFrontendIndex {
			c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
			return
		}

		c.HTML(http.StatusOK, "index.html", gin.H{})
	}
}

func noRouteHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
	}
}
