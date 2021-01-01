package tasks

import (
	"context"
	"sync"
	"time"
	"visual-feed-aggregator/src/database/models"
	"visual-feed-aggregator/src/database/services"
	"visual-feed-aggregator/src/util/logging"

	"github.com/jmoiron/sqlx"
)

// BackgroundTask ...
type BackgroundTask func(stopSignal <-chan bool, lastRun map[string]time.Time,
	db *sqlx.DB, services *services.ServiceCollection, cutoffDays, refreshRateMinutes int64)
type taskFunc func(dateCutoff *time.Time, loc *time.Location, services *services.ServiceCollection)
type channelTaskFunc func(channel *models.Channel, wg *sync.WaitGroup, dateCutoff *time.Time, loc *time.Location,
	services *services.ServiceCollection)

func runChannelTask(stopSignal <-chan bool, lastRun map[string]time.Time, db *sqlx.DB, srv *services.ServiceCollection, cutoffDays, refreshRateMinutes int64, kind string, channelTask channelTaskFunc) {
	runTask(stopSignal, lastRun, db, srv, cutoffDays, refreshRateMinutes, kind,
		func(dateCutoff *time.Time, loc *time.Location, services *services.ServiceCollection) {
			var wg sync.WaitGroup
			channels, err := services.ChannelService.FindChannelsByKind(kind)
			if err == nil {
				for i := range channels {
					wg.Add(1)
					go channelTask(&channels[i], &wg, dateCutoff, loc, services)
				}
				wg.Wait()
			} else {
				logging.Println(logging.Info, err)
			}
		})
}

func runTask(stopSignal <-chan bool, lastRun map[string]time.Time, db *sqlx.DB, services *services.ServiceCollection, cutoffDays, refreshRateMinutes int64, kind string, task taskFunc) {
	tick := time.Time{}
	for true {
		select {
		case <-stopSignal:
			logging.Println(logging.Info, "terminating", kind, "background task")
			return
		default:
			delta := time.Since(tick)
			if int64(delta.Minutes()) >= refreshRateMinutes {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				if err := db.PingContext(ctx); err != nil {
					time.Sleep(1 * time.Minute) // retry in a minute
				} else {
					startTime := time.Now()
					logging.Println(logging.Info, kind, "updating:", startTime.Format(time.RFC3339))

					dateCutoff := time.Now().AddDate(0, 0, -int(cutoffDays))
					loc := startTime.Location()
					task(&dateCutoff, loc, services)

					endTime := time.Now()
					lastRun[kind] = endTime.UTC()
					logging.Println(logging.Info, kind, "finished updating:", endTime.Format(time.RFC3339), " - it took", endTime.Sub(startTime).Minutes(), "minutes")

					tick = time.Now()
				}
				cancel()
			} else {
				time.Sleep(5 * time.Second)
			}
		}
	}
}
