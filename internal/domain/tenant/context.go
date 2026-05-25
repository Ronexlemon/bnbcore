
package tenant

import "context"

type contextKey struct{}

func NewContext(ctx context.Context, t *Tenant) context.Context {
    return context.WithValue(ctx, contextKey{}, t)
}

func FromContext(ctx context.Context) *Tenant {
    t, _ := ctx.Value(contextKey{}).(*Tenant)
    return t
}