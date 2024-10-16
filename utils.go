package blocksyncer

import "time"

func IntervalOrDefault(duration time.Duration, defaultInterval time.Duration) time.Duration {
	if duration <= 0 {
		return defaultInterval
	}

	return duration
}
