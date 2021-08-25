package server

import (
	"containerdhealthcheck/internal/models"
	"containerdhealthcheck/internal/monitoring"
	"net/http"

	gosundheit "github.com/AppsFlyer/go-sundheit"
	"github.com/sirupsen/logrus"
)

// App defines a struct to hold the dependencies and configuration settings for the web application
type App struct {
	Server       *http.Server
	ServerConfig models.ServerConfig
	YAMLConfig   models.YAMLConfig
	BuildInfo    models.BuildInfo
	Collector    monitoring.Collector
	HealthCheck  gosundheit.Health
	Logger       *logrus.Logger
}
