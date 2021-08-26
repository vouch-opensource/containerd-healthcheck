package server

import (
	"containerdhealthcheck/internal/containerd"
	"containerdhealthcheck/internal/models"
	"containerdhealthcheck/internal/monitoring"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"contrib.go.opencensus.io/exporter/prometheus"
	gosundheit "github.com/AppsFlyer/go-sundheit"
	"github.com/AppsFlyer/go-sundheit/checks"
	"github.com/AppsFlyer/go-sundheit/opencensus"
	"github.com/sirupsen/logrus"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	MTaskRestart = stats.Int64("restart", "Number of restarts per containerd task ", "restarts")
	KeyTask, _   = tag.NewKey("task")
)

type checkEventsLogger struct {
	Context    context.Context
	Containerd *containerd.Containerd
	Checks     []models.Check
	Logger     *logrus.Logger
}

func (l checkEventsLogger) OnCheckRegistered(name string, res gosundheit.Result) {
	l.Logger.WithFields(logrus.Fields{
		"Name":    name,
		"Details": res.Details,
	}).Info("Check registered")
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
		}).Printf("Task restarted. Waiting %s before the next health check", check.RestartDelay*time.Second)

		monitoring.RecordRestartTask(l.Context, name)

		time.Sleep(check.RestartDelay * time.Second)

	}

}

// NewApp provides a new service with prometheus http handler
func NewApp(serverConfig models.ServerConfig, yamlConfig models.YAMLConfig, buildInfo models.BuildInfo, logger *logrus.Logger) (*App, error) {

	ctx := context.Background()
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

	oc := opencensus.NewMetricsListener()
	h := gosundheit.New(gosundheit.WithCheckListeners(oc, checkEventsLogger{
		Context:    ctx,
		Containerd: containerd,
		Logger:     logger,
		Checks:     hchecks,
	}), gosundheit.WithHealthListeners(oc))

	view.Register(opencensus.DefaultHealthViews...)
	view.Register(monitoring.ViewRestart)

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
			logger.Error(err)
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
		Logger:       logger,
		HealthCheck:  h,
	}, nil

}

// Run listens and serves http server using gin
func (app *App) Run() {

	exporter, _ := prometheus.NewExporter(prometheus.Options{})

	http.Handle("/metrics", exporter)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})

	app.Logger.WithFields(logrus.Fields{
		"Address":     app.ServerConfig.Addr,
		"Environment": app.ServerConfig.Env,
		"Version":     app.BuildInfo.Version,
	}).Info("HTTP server started")

	app.Logger.Fatal(http.ListenAndServe(app.ServerConfig.Addr, nil))

}
