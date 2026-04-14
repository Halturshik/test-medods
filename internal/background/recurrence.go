package background

import (
	"context"
	"log/slog"
	"time"

	"example.com/taskservice/internal/usecase/task"
)

func StartRecurrenceWorker(ctx context.Context, uc task.Usecase, logger *slog.Logger) {
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()

	logger.Info("recurrence worker started", "interval", "6h")

	for {
		select {
		case <-ctx.Done():
			logger.Info("recurrence worker stopped")
			return

		case <-ticker.C:
			start := time.Now()
			if err := uc.GenerateMissingInstances(ctx, 60); err != nil {
				logger.Error("recurrence generator failed", "error", err, "duration", time.Since(start))
			} else {
				logger.Info("recurrence generation completed", "duration", time.Since(start))
			}
		}
	}
}
