package di

import (
	"context"
	"net/http"
)

type ContextKey string

const ContextContainerKey ContextKey = "container"

func Middleware(c *Container, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestContext := r.Context()
		scopedContainer, _ := c.NewScope()
		requestContext = context.WithValue(requestContext, ContextContainerKey, scopedContainer)

		next.ServeHTTP(w, r.WithContext(requestContext))
	})
}
