package middleware

import (
	"context"
	"net/http"
)

func shutdownContext(rootCtx context.Context) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithCancel(r.Context())

			go func() {
				select {
				case <-rootCtx.Done():
					cancel()
				case <-ctx.Done():
				}
			}()
			
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
