package monitoring

import (
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	mTaskRestart = stats.Int64("restart", "Number of restarts per containerd task ", "restarts")
	keyTask, _   = tag.NewKey("task")

	ViewRestart = &view.View{
		Name:        "containerd_task_restarts_total",
		Measure:     mTaskRestart,
		Description: "Number of restarts per containerd tasks",
		TagKeys:     []tag.Key{keyTask},
		Aggregation: view.Count(),
	}
)
