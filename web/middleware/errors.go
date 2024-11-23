package middleware

import (
	"context"
	"github.com/jamius19/postbranch/web/responseerror"
	"net/http"
)

func requestError(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestErrors := &responseerror.ResponseErrors{}
		ctx := context.WithValue(r.Context(), responseerror.ErrorsContextKey, requestErrors)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
