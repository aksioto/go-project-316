package analyzer

import "context"

type Waiter interface {
	Wait(ctx context.Context) error
}
