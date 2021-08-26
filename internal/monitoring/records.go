package monitoring

import (
	"context"

	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

func RecordRestartTask(ctx context.Context, name string) {
	stats.RecordWithTags(ctx, []tag.Mutator{tag.Upsert(keyTask, name)}, mTaskRestart.M(1))
}
