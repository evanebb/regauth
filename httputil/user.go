package httputil

import (
	"context"
	"github.com/evanebb/regauth/user"
)

type currentUserCtxKey struct{}

// WithLoggedInUser sets the currently logged-in user in the given context.
// Use LoggedInUserFromContext to retrieve the currently logged-in user.
func WithLoggedInUser(ctx context.Context, u user.User) context.Context {
	return context.WithValue(ctx, currentUserCtxKey{}, u)
}

// LoggedInUserFromContext parses the currently logged-in user from the given request context.
// This requires the user to have been previously set in the context by WithLoggedInUser.
func LoggedInUserFromContext(ctx context.Context) (user.User, bool) {
	u, ok := ctx.Value(currentUserCtxKey{}).(user.User)
	return u, ok
}
