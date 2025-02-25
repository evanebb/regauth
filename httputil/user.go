package httputil

import (
	"context"
	"github.com/evanebb/regauth/user"
)

type userCtxKey struct{}

// WithUser sets the given user.User in the context.
// Use UserFromContext to retrieve the user.
func WithUser(ctx context.Context, u user.User) context.Context {
	return context.WithValue(ctx, userCtxKey{}, u)
}

// UserFromContext parses the current user.User from the given request context.
// This requires the user to have been previously set in the context by WithUser.
func UserFromContext(ctx context.Context) (user.User, bool) {
	u, ok := ctx.Value(userCtxKey{}).(user.User)
	return u, ok
}
