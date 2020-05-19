package autoscaler

import "time"

type Provider interface {
	SetCapacity(int, time.Duration) error
}
