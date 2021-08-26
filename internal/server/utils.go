package server

import (
	"containerdhealthcheck/internal/models"
	"time"

	gosundheit "github.com/AppsFlyer/go-sundheit"
	"github.com/AppsFlyer/go-sundheit/checks"
)

func findContainerTask(a []models.Check, x string) int {
	for i, n := range a {
		if x == n.ContainerTask {
			return i
		}
	}
	return len(a)
}

func SetCheckDefaults(checks []models.Check) []models.Check {

	for i := range checks {
		if checks[i].Timeout == 0 {
			checks[i].Timeout = DefaultTimeout
		}
		if checks[i].ExecutionPeriod == 0 {
			checks[i].ExecutionPeriod = DefaultExecutionPeriod
		}
		if checks[i].Threshold == 0 {
			checks[i].Threshold = DefaultThreshold
		}
		if checks[i].HTTP.Method == "" {
			checks[i].HTTP.Method = DefaultHTTPMethod
		}
		if checks[i].HTTP.ExpectedStatus == 0 {
			checks[i].HTTP.ExpectedStatus = DefaultHTTPExpectedStatus
		}
	}

	return checks

}

func RegisterChecks(health gosundheit.Health, hchecks []models.Check) (gosundheit.Health, error) {

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
			return nil, err
		}

		err = health.RegisterCheck(
			httpCheck,
			gosundheit.InitialDelay(c.InitialDelay*time.Second),
			gosundheit.ExecutionPeriod(c.ExecutionPeriod*time.Second),
		)

		if err != nil {
			return nil, err
		}

	}

	return health, nil

}
