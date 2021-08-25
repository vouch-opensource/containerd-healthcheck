package server

import (
	"containerdhealthcheck/internal/containerd"
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

type checkEventsLogger struct {
	Containerd *containerd.Containerd
	Checks     []models.Check
	Logger     *logrus.Logger
}

func findContainerTask(a []models.Check, x string) int {
	for i, n := range a {
		if x == n.ContainerTask {
			return i
		}
	}
	return len(a)
}

func (l checkEventsLogger) OnCheckRegistered(name string, res gosundheit.Result) {
	log.Printf("Check %q registered with initial result: %v\n", name, res)
}

func (l checkEventsLogger) OnCheckStarted(name string) {
}

func (l checkEventsLogger) OnCheckCompleted(name string, res gosundheit.Result) {
	l.Logger.WithFields(logrus.Fields{
		"Name":               name,
		"Details":            res.Details,
		"Error":              res.Error,
		"ContiguousFailures": res.ContiguousFailures,
	}).Info("Check completed")

	idx := findContainerTask(l.Checks, name)
	check := l.Checks[idx]

	if res.ContiguousFailures >= check.Threshold {
		err := l.Containerd.RestartTask(name)
		if err != nil {
			l.Logger.Error(err)
		}
		l.Logger.WithFields(logrus.Fields{
			"Name":         name,
			"RestartDelay": check.RestartDelay,
		}).Info("Task restarted")
		time.Sleep(check.RestartDelay * time.Second)
	}

}

// NewApp provides a new service with prometheus http handler
func NewApp(serverConfig models.ServerConfig, yamlConfig models.YAMLConfig, buildInfo models.BuildInfo, logger *logrus.Logger) (*App, error) {

	collector := monitoring.NewCollector()
	hchecks := yamlConfig.Checks

	containerd, err := containerd.NewClient(logger, yamlConfig.Containerd.Socket, yamlConfig.Containerd.Namespace)
	if err != nil {
		return nil, err
	}

	for i := range hchecks {
		if hchecks[i].Timeout == 0 {
			hchecks[i].Timeout = 1
		}
		if hchecks[i].ExecutionPeriod == 0 {
			hchecks[i].ExecutionPeriod = 10
		}
		if hchecks[i].Threshold == 0 {
			hchecks[i].Threshold = 3
		}
		if hchecks[i].HTTP.Method == "" {
			hchecks[i].HTTP.Method = "GET"
		}
		if hchecks[i].HTTP.ExpectedStatus == 0 {
			hchecks[i].HTTP.ExpectedStatus = 200
		}
	}

	h := gosundheit.New(gosundheit.WithCheckListeners(checkEventsLogger{
		Containerd: containerd,
		Logger:     logger,
		Checks:     hchecks,
	}))

	for _, c := range hchecks {

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

	return &App{
		ServerConfig: serverConfig,
		YAMLConfig:   yamlConfig,
		BuildInfo:    buildInfo,
		Collector:    collector,
		Logger:       logger,
		HealthCheck:  h,
	}, nil

}

// Run listens and serves http server using gin
func (app *App) Run() {

	http.Handle("/metrics", promhttp.Handler())

	app.Logger.WithFields(logrus.Fields{
		"Address":     app.ServerConfig.Addr,
		"Environment": app.ServerConfig.Env,
		"Version":     app.BuildInfo.Version,
	}).Info("HTTP server started")

	app.Logger.Fatal(http.ListenAndServe(app.ServerConfig.Addr, nil))

}
