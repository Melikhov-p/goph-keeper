package secret

import "context"

type Repository interface {
	Create(ctx context.Context, secret *Secret)
	CreatePassword(ctx context.Context, secret *Secret)
}
