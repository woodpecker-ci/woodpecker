package autoscaler

import "context"

type Autoscaler interface {
	Start(ctx context.Context)
}
