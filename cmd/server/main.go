package main

import (
	"context"
	"flag"
	"fmt"
	"home-hue-server/internal/service"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amimof/huego"
	"github.com/gorilla/websocket"
)

var stateChannel = make(chan map[string]interface{})

type hueBridge struct {
	ip       string
	username string
	user     string
}

type remoteServerConfig struct {
	url        string        // The URL of the remote server
	authToken  string        // The authentication token for accessing the server
	logLevel   string        // Log level for server communication (e.g., "info", "debug", "error")
	isEnabled  bool          // Indicates if the remote server should be used
	timeout    time.Duration // Timeout in seconds for the server connection
	retryCount int           // Number of times to retry connection in case of failure
}

type config struct {
	remoteServerConfig remoteServerConfig
	hueBridge          hueBridge
	env                string
	port               int
}

type application struct {
	config config
	logger *slog.Logger
	hue    *service.Hue
	groups *[]huego.Group
	wsConn *websocket.Conn
	errCh  chan error
}

var version = "1.0.0"

func main() {
	var cfg config
	var rmtSvrCfg remoteServerConfig
	var discoverBridge bool

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.hueBridge.ip, "hue-ip", "", "IP address of the Hue bridge")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.hueBridge.username, "hue-username", "", "Username for the Hue bridge")
	flag.BoolVar(&discoverBridge, "discover-hue", false, "Discover the IP address of the Hue bridge")
	flag.BoolVar(&rmtSvrCfg.isEnabled, "remote", false, "Use remote server")
	flag.StringVar(&rmtSvrCfg.url, "remote-url", "localhost:4000", "URL of the remote server")
	flag.StringVar(&rmtSvrCfg.authToken, "remote-auth-token", "", "Authorization token for the remote server")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Connect to the Hue bridge if the discoverBridge flag is false.
	var hueService service.Hue
	if discoverBridge {
		err := hueService.DiscoverBridge(cfg.hueBridge.username)
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}
		logger.Info("Hue bridge discovered", "ip", hueService.Address)
	} else {
		err := hueService.ConnectToBridge(cfg.hueBridge.ip, cfg.hueBridge.username)
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}
		logger.Info("Connected to Hue bridge", "ip", hueService.Address)
	}

	logger.Info("Hue bridge user", "username", hueService.Bridge.User, "id", hueService.Bridge.ID)

	groups, err := hueService.Bridge.GetGroups()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	for _, group := range groups {
		logger.Info("Group", "id", group.ID, "name", group.Name, "state", group.State)
		if group.Name == "Office" {
			if group.State.On {
				group.Off()
			} else {
				group.On()
			}
		}
	}

	// Set Remote Server Configuration
	rmtSvrCfg.timeout = 5 * time.Second
	rmtSvrCfg.retryCount = 3

	cfg.remoteServerConfig = rmtSvrCfg

	// Create a new instance of the application struct.
	app := &application{
		config: cfg,
		logger: logger,
		hue:    &hueService,
		groups: &groups,
		errCh:  make(chan error),
	}

	svr := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	if cfg.remoteServerConfig.isEnabled {
		go app.wsConnect()
	}

	// Setup signal handling to gracefully shutdown the application
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("PORT:", cfg.port)
	logger.Info("Starting server", "port", cfg.port)

	// Start the HTTP server in a goroutine
	go func() {
		if err := svr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(err.Error())
			os.Exit(1)
		}
	}()

	// Block until a signal is received
	<-stop
	logger.Info("Shutting down server...")

	// Create a context with a timeout to allow for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt to gracefully shutdown the server
	if err := svr.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown:", err)
	}

	logger.Info("Server exiting")
}
