package controllers

import (
	"context"
	"net/http"
	"strings"

	"bookurrroom/internal/auth"
	"bookurrroom/internal/models"
)

type principalContextKey struct{}

func principalFromContext(ctx context.Context) (auth.Principal, bool) {
	principal, ok := ctx.Value(principalContextKey{}).(auth.Principal)
	return principal, ok
}

func requireAuth(tokens *auth.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
			if authHeader == "" {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
				return
			}

			const BearerPrefix = "Bearer "
			if strings.HasPrefix(authHeader, BearerPrefix) == false {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
				return
			}

			token := strings.TrimSpace(strings.TrimPrefix(authHeader, BearerPrefix))
			if token == "" {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
				return
			}

			principal, err := tokens.Parse(token)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
				return
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), principalContextKey{}, principal)))
		})
	}
}

func requireRoles(roles ...models.Role) func(http.Handler) http.Handler {
	allowed := make(map[models.Role]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := principalFromContext(r.Context())
			if ok == false {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
				return
			}

			if _, exists := allowed[principal.Role]; exists == false {
				writeError(w, http.StatusForbidden, "FORBIDDEN", "forbidden")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
