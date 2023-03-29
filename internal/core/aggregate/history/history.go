package history

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/uptrace/bun/extra/bunbig"

	"github.com/tonindexer/anton/internal/core"
)

type ReqParams struct {
	From     time.Time     `form:"from"`
	To       time.Time     `form:"to"`
	Interval time.Duration `form:"interval"`
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
	funcFormat := "toStartOfInterval(%s, INTERVAL %d %s)"

	sec := int(interval.Seconds())

	min := sec / 60
	if min < 5 {
		return "", errors.Wrapf(core.ErrInvalidArg, "unsupported interval %d seconds", sec)
	}
	if min < 60 {
		return fmt.Sprintf(funcFormat, "%s", min, "minute"), nil
	}

	hour := min / 60
	if hour < 24 {
		return fmt.Sprintf(funcFormat, "%s", hour, "hour"), nil
	}

	days := hour / 24
	if days < 7 {
		return fmt.Sprintf(funcFormat, "%s", days, "day"), nil
	}

	return "", errors.Wrapf(core.ErrInvalidArg, "unsupported interval %d hours", days)
}
