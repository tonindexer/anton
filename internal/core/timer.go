package core

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

func Timer(start time.Time, fun string, args ...any) {
	elapsed := float64(time.Since(start)) / 1e9
	if elapsed < 0.1 {
		return
	}
	log.Debug().Str("func", fmt.Sprintf(fun, args...)).Float64("elapsed", elapsed).Msg("timer")
}
