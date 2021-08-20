package server

import (
	"containerdhealthcheck/internal/models"
	"containerdhealthcheck/internal/monitoring"
	"net/http"

	"github.com/sirupsen/logrus"
)

// App defines a struct to hold the dependencies and configuration settings for the web application
type App struct {
	Server       *http.Server
	ServerConfig models.ServerConfig
	YAMLConfig   models.YAMLConfig
	BuildInfo    models.BuildInfo
	Collector    monitoring.Collector
	Logger       *logrus.Logger
}
