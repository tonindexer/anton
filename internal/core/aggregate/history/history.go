package history

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/uptrace/bun/extra/bunbig"

	"github.com/iam047801/tonidx/internal/core"
)

type ReqParams struct {
	From     time.Time `form:"from"`
	To       time.Time `form:"to"`
	Interval time.Duration
}

type CountRes []struct {
	Value     int
	Timestamp time.Time
}

type BigIntRes []struct {
	Value     *bunbig.Int `ch:"type:UInt256"`
	Timestamp time.Time
}

func GetRoundingFunction(interval time.Duration) (string, error) {
	var unitInterval, unit string

	switch sec := int(interval.Seconds()); sec {
	case 15 * 60:
		unitInterval, unit = "15", "minute"
	case 60 * 60:
		unitInterval, unit = "1", "hour"
	case 4 * 60 * 60:
		unitInterval, unit = "4", "hour"
	case 8 * 60 * 60:
		unitInterval, unit = "8", "hour"
	case 24 * 60 * 60:
		unitInterval, unit = "1", "day"
	case 7 * 24 * 60 * 60:
		unitInterval, unit = "1", "week"
	default:
		return "", errors.Wrapf(core.ErrInvalidArg, "unsupported interval %d seconds", sec)
	}

	return fmt.Sprintf("toStartOfInterval(%s, INTERVAL %s %s)", "%s", unitInterval, unit), nil
}
