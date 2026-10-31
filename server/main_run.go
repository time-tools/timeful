package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"timeful/server/appenv"
	"timeful/server/logger"
	"timeful/server/observability"
	"timeful/server/postgres"
	"timeful/server/routes"
	"timeful/server/services/gcloud"
	"timeful/server/slackbot"
	"timeful/server/utils"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "timeful/server/docs"
)

func main() {
	// Set release flag
	release := flag.Bool("release", false, "Whether this is the release version of the server")
	flag.Parse()

	// Cancel on SIGINT/SIGTERM so the server drains requests and the bounded
	// export queue is flushed before exit.
	runContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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

	// Structured diagnostics ship over OTLP on a bounded background pipeline.
	// Incomplete OpenObserve configuration disables export and leaves the
	// existing file and standard-stream diagnostics authoritative.
	recorder, err := observability.Start(runContext, observability.ConfigFromEnv())
	if err != nil {
		logger.StdErr.Printf("Warning: structured diagnostic export disabled: %s", err)
		recorder = nil
	}
	// Outbound requests made through the default client join their request
	// trace with a bounded, QR-004-safe client span.
	observability.InstrumentDefaultTransport(recorder)

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
		// The access log is local diagnostic output, so token-bearing query
		// parameters and handler error messages pass through redaction too.
		path := observability.Redact(param.Path)
		errorMessage := observability.Redact(param.ErrorMessage)
		return fmt.Sprintf("%v |%s %3d %s| %13v | %15s |%s %-7s %s %#v\n%s",
			param.TimeStamp.Format("2006/01/02 15:04:05"),
			statusColor, param.StatusCode, resetColor,
			param.Latency,
			param.ClientIP,
			methodColor, param.Method, resetColor,
			path,
			errorMessage,
		)
	}))
	router.Use(observability.RequestMiddleware(recorder))
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
	closePostgres := postgres.Init(recorder.TracerProvider())
	defer closePostgres()

	// Readiness shares the PostgreSQL dependency with /api/health and is
	// sampled in the background, so request records never wait on the probe.
	// Every sample also feeds the readiness metric.
	readinessMonitor := observability.NewReadinessMonitor(
		postgres.Ping,
		observability.DefaultReadinessInterval,
		observability.DefaultReadinessTimeout,
	)
	readinessMonitor.SetObserver(recorder.RecordReadiness)
	readinessMonitor.Start(runContext)
	recorder.SetReadiness(readinessMonitor.State)
	recorder.SetPoolStats(func() observability.PoolStats {
		used, idle, max := postgres.ConnectionStats()
		return observability.PoolStats{Used: used, Idle: idle, Max: max}
	})

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
	server := &http.Server{Addr: ":" + port, Handler: router}
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.StdErr.Fatalf("server failed: %v", err)
		}
	case <-runContext.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			logger.StdErr.Printf("server shutdown: %v", err)
		}
	}

	// Flush queued structured records with a bounded wait; export failures are
	// local diagnostics only and never affect shutdown.
	flushContext, cancelFlush := context.WithTimeout(context.Background(), observability.DefaultShutdownTimeout)
	defer cancelFlush()
	if err := recorder.Shutdown(flushContext); err != nil {
		logger.StdErr.Printf("observability shutdown: %s", observability.Redact(err.Error()))
	}
}
