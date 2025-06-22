package query

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/tonindexer/anton/internal/core/aggregate"
)

const (
	statsRetryDelay         = 15 * time.Second
	statsUpdateDelay        = 5 * time.Minute
	statsLifespan           = 10 * time.Minute
	statsCalculationTimeout = 2 * time.Minute
)

func (s *Service) updateStatsLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		if !s.running() {
			return
		}

		s.mx.RLock()
		lastUpdate := s.statsUpdateTs
		lastTry := s.statsFailTs
		s.mx.RUnlock()

		if time.Since(lastUpdate) > statsUpdateDelay && time.Since(lastTry) > statsRetryDelay {
			s.updateStats()
		}
	}
}

func (s *Service) updateStats() {
	ctx, cancel := context.WithTimeout(context.Background(), statsCalculationTimeout)
	defer cancel()

	stats, err := aggregate.GetStatistics(ctx, s.DB.CH, s.DB.PG)

	s.mx.Lock()
	defer s.mx.Unlock()

	if err != nil {
		log.Error().Err(err).Msg("cannot update stats")
		s.statsFailTs = time.Now()
		return
	}

	s.statsCached = stats
	s.statsUpdateTs = time.Now()
}
