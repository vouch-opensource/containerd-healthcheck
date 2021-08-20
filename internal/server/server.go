package server

import (
	"containerdhealthcheck/internal/models"
	"containerdhealthcheck/internal/monitoring"
	"fmt"
	"log"
	"net/http"
	"time"

	gosundheit "github.com/AppsFlyer/go-sundheit"
	"github.com/AppsFlyer/go-sundheit/checks"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

type checkEventsLogger struct{}

func (l checkEventsLogger) OnCheckRegistered(name string, res gosundheit.Result) {
	log.Printf("Check %q registered with initial result: %v\n", name, res)
}

func (l checkEventsLogger) OnCheckStarted(name string) {
	log.Printf("Check %q started...\n", name)
}

func (l checkEventsLogger) OnCheckCompleted(name string, res gosundheit.Result) {
	log.Printf("Check %q completed with result: %v\n", name, res)
	if res.ContiguousFailures > 4 {
		log.Printf("--->> Oh, check %q has %v continguous failures\n", name, res.ContiguousFailures)
	}
}

// NewApp provides a new service with prometheus http handler
func NewApp(serverConfig models.ServerConfig, yamlConfig models.YAMLConfig, buildInfo models.BuildInfo, logger *logrus.Logger) (*App, error) {

	collector := monitoring.NewCollector()

	return &App{
		ServerConfig: serverConfig,
		YAMLConfig:   yamlConfig,
		BuildInfo:    buildInfo,
		Collector:    collector,
		Logger:       logger,
	}, nil

}

// Run listens and serves http server using gin
func (app *App) Run() {

	http.Handle("/metrics", promhttp.Handler())

	app.Logger.Printf("HTTP Server listening on address '%s' in %s environment", app.ServerConfig.Addr, app.ServerConfig.Env)

	app.Logger.Printf("------")
	app.Logger.Printf("Version: %s", app.BuildInfo.Version)
	app.Logger.Printf("Server Addr: '%s'", app.ServerConfig.Addr)
	app.Logger.Printf("------")
	app.Logger.Printf("")

	h := gosundheit.New(gosundheit.WithCheckListeners(&checkEventsLogger{}))

	for _, c := range app.YAMLConfig.Checks {

		if c.Timeout == 0 {
			c.Timeout = 1
		}

		if c.ExecutionPeriod == 0 {
			c.ExecutionPeriod = 10
		}

		if c.InitialDelay == 0 {
			c.InitialDelay = 1
		}

		httpCheckConf := checks.HTTPCheckConfig{
			CheckName:      c.ContainerTask,
			Timeout:        c.Timeout * time.Second,
			Method:         c.HTTP.Method,
			URL:            c.HTTP.URL,
			ExpectedBody:   c.HTTP.ExpectedBody,
			ExpectedStatus: c.HTTP.ExpectedStatus,
		}

		httpCheck, err := checks.NewHTTPCheck(httpCheckConf)
		if err != nil {
			fmt.Println(err)
		}

		err = h.RegisterCheck(
			httpCheck,
			gosundheit.InitialDelay(c.InitialDelay*time.Second),
			gosundheit.ExecutionPeriod(c.ExecutionPeriod*time.Second),
		)

		if err != nil {
			log.Fatal("Failed to register check: ", err)
		}

	}

	app.Logger.Fatal(http.ListenAndServe(app.ServerConfig.Addr, nil))

}
