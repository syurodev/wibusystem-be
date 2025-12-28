package worker

import (
	"context"
	"sync"
	"time"

	"system/internal/domain"

	"go.uber.org/zap"
)

// RankSnapshotWorker xử lý việc tạo snapshot xếp hạng định kỳ
type RankSnapshotWorker struct {
	analyticsRepo domain.ViewAnalyticsRepository
	logger        *zap.Logger
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// NewRankSnapshotWorker creates a new rank snapshot worker
func NewRankSnapshotWorker(
	analyticsRepo domain.ViewAnalyticsRepository,
	logger *zap.Logger,
) *RankSnapshotWorker {
	return &RankSnapshotWorker{
		analyticsRepo: analyticsRepo,
		logger:        logger,
		stopChan:      make(chan struct{}),
	}
}

// Start bắt đầu background worker
func (w *RankSnapshotWorker) Start() {
	w.logger.Info("Starting RankSnapshotWorker")
	w.wg.Add(1)
	go w.run()
}

// Stop dừng background worker gracefully
func (w *RankSnapshotWorker) Stop() {
	w.logger.Info("Stopping RankSnapshotWorker")
	close(w.stopChan)
	w.wg.Wait()
	w.logger.Info("RankSnapshotWorker stopped")
}

// TriggerNow manually triggers a rank snapshot (for testing/debugging)
func (w *RankSnapshotWorker) TriggerNow(period string, entityType string, limit int) error {
	w.logger.Info("Manually triggering rank snapshot",
		zap.String("period", period),
		zap.String("entity_type", entityType),
		zap.Int("limit", limit),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	err := w.analyticsRepo.CreateRankSnapshot(ctx, time.Now(), period, entityType, limit)
	if err != nil {
		w.logger.Error("Failed to create manual rank snapshot",
			zap.String("period", period),
			zap.String("entity_type", entityType),
			zap.Error(err),
		)
		return err
	}

	w.logger.Info("Successfully created manual rank snapshot",
		zap.String("period", period),
		zap.String("entity_type", entityType),
	)
	return nil
}

func (w *RankSnapshotWorker) run() {
	defer w.wg.Done()

	// Check schedule every minute
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopChan:
			return
		case t := <-ticker.C:
			w.checkSchedule(t)
		}
	}
}

func (w *RankSnapshotWorker) checkSchedule(t time.Time) {
	// Sử dụng UTC hoặc Local time tùy policy, ở đây dùng Local server time

	// 1. Weekly Snapshots (Sunday night / Monday morning)
	// Chạy vào Chủ nhật (Weekday = 0)
	if t.Weekday() == time.Sunday {
		// Genre: 01:00
		if t.Hour() == 1 && t.Minute() == 0 {
			w.triggerSnapshot(t, "week", "genre", 0) // 0 = no limit (all)
		}
		// Org: 01:15
		if t.Hour() == 1 && t.Minute() == 15 {
			w.triggerSnapshot(t, "week", "org", 500)
		}
		// Creator: 01:45
		if t.Hour() == 1 && t.Minute() == 45 {
			w.triggerSnapshot(t, "week", "creator", 1000)
		}
		// Novel: 02:30
		if t.Hour() == 2 && t.Minute() == 30 {
			w.triggerSnapshot(t, "week", "novel", 500)
		}
		// Manga: 02:45
		if t.Hour() == 2 && t.Minute() == 45 {
			w.triggerSnapshot(t, "week", "manga", 500)
		}
		// Anime: 03:00
		if t.Hour() == 3 && t.Minute() == 0 {
			w.triggerSnapshot(t, "week", "anime", 500)
		}
	}

	// 2. Monthly Snapshots (1st of month)
	if t.Day() == 1 {
		// Chạy lúc 02:00 AM để tránh conflict với weekly nếu trùng
		// (Weekly chạy 1:00 - 3:00, Monthly dời sang khung giờ khác hoặc chấp nhận chạy song song nhẹ)
		// Plan đề xuất: 2:00 AM.
		// Để an toàn và tránh trùng giờ cao điểm của tuần, có thể chạy khung giờ khác hoặc xử lý song song.
		// Với 2:00 AM, nó trùng với Novel Weekly nếu mùng 1 rơi vào Chủ Nhật.
		// Tốt nhất dời Monthly sang khung giờ 04:00 - 05:00

		hour := t.Hour()
		minute := t.Minute()

		if hour == 4 {
			switch minute {
			case 0:
				w.triggerSnapshot(t, "month", "genre", 0)
			case 15:
				w.triggerSnapshot(t, "month", "org", 500)
			case 30:
				w.triggerSnapshot(t, "month", "creator", 1000)
			case 45:
				w.triggerSnapshot(t, "month", "novel", 500)
			}
		}
		if hour == 5 {
			switch minute {
			case 0:
				w.triggerSnapshot(t, "month", "manga", 500)
			case 15:
				w.triggerSnapshot(t, "month", "anime", 500)
			}
		}
	}

	// 3. Yearly Snapshots (Jan 1st)
	if t.Month() == time.January && t.Day() == 1 {
		// Chạy lúc 06:00 AM
		hour := t.Hour()
		minute := t.Minute()

		if hour == 6 {
			switch minute {
			case 0:
				w.triggerSnapshot(t, "year", "genre", 0)
			case 15:
				w.triggerSnapshot(t, "year", "org", 500)
			case 30:
				w.triggerSnapshot(t, "year", "creator", 1000)
			case 45:
				w.triggerSnapshot(t, "year", "novel", 500)
			}
		}
		if hour == 7 {
			switch minute {
			case 0:
				w.triggerSnapshot(t, "year", "manga", 500)
			case 15:
				w.triggerSnapshot(t, "year", "anime", 500)
			}
		}
	}
}

func (w *RankSnapshotWorker) triggerSnapshot(t time.Time, period string, entityType string, limit int) {
	w.logger.Info("Triggering rank snapshot",
		zap.String("period", period),
		zap.String("entity_type", entityType),
		zap.Time("time", t),
	)

	// Chạy trong goroutine riêng để không block ticker loop
	// (Mặc dù checkSchedule chạy nhanh, nhưng CreateRankSnapshot có thể lâu)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		// Snapshot date logic:
		// Với week: Ticker chạy vào Chủ nhật (cuối tuần), nhưng logic repository tính start date là Thứ 2.
		// Nếu ta chạy vào Chủ nhật tuần này, snapshot cho tuần NÀY (đang kết thúc/sắp kết thúc).
		// Repository CreateRankSnapshot nhận snapshotDate.
		// Logic Week trong repository: currentSnapshotDate = Start of current week (Monday).
		// Vậy ta truyền t (thời điểm chạy) vào, repository sẽ tự tính start date của tuần chứa t.
		// OK.

		err := w.analyticsRepo.CreateRankSnapshot(ctx, t, period, entityType, limit)
		if err != nil {
			w.logger.Error("Failed to create rank snapshot",
				zap.String("period", period),
				zap.String("entity_type", entityType),
				zap.Error(err),
			)
		} else {
			w.logger.Info("Successfully created rank snapshot",
				zap.String("period", period),
				zap.String("entity_type", entityType),
			)
		}
	}()
}
