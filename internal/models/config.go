package models

import "time"

type YAMLConfig struct {
	Checks     []Check `yaml:"checks"`
	Containerd struct {
		Socket    string `yaml:"socket" env-description:"containerd socket path" env-default:"/run/containerd/containerd.sock"`
		Namespace string `yaml:"namespace" env-description:"containerd namespace" env-default:"services.linuxkit"`
	} `yaml:"containerd"`
}

type ServerConfig struct {
	Env  string
	Addr string
}

type Check struct {
	ContainerTask   string        `yaml:"container_task" env-description:"containerd task name" env-required:"true"`
	Timeout         time.Duration `yaml:"timeout" env-description:"Timeout (in seconds) used for the HTTP request"`
	ExecutionPeriod time.Duration `yaml:"execution_period" env-description:"health check interval in seconds"`
	InitialDelay    time.Duration `yaml:"initial_delay" env-description:"Time to delay first check execution in seconds"`
	RestartDelay    time.Duration `yaml:"restart_delay" env-description:"Time to sleep after restarting a containerd task in seconds"`
	Threshold       int64         `yaml:"threshold" env-description:"The number of consecute health check failures required before considering a target unhealthy"`
	HTTP            struct {
		URL            string `yaml:"url" env-description:"URL to be called by the check" env-required:"true"`
		Method         string `yaml:"method" env-description:"HTTP method"`
		ExpectedBody   string `yaml:"expected_body" env-description:"Operates as a basic 'body should contain <string>'"`
		ExpectedStatus int    `yaml:"espected_status" env-description:"Expected response status code"`
	}
}
