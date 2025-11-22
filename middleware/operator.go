package middleware

import (
	"context"
	"net/http"
	"rgs/sqlc"
)

type key int

const operatorKey key = 1

func ContextWithOperator(ctx context.Context, operator sqlc.Operator) context.Context {
	return context.WithValue(ctx, operatorKey, operator)
}

func OperatorFromContext(ctx context.Context) (sqlc.Operator, bool) {
	operator, ok := ctx.Value(operatorKey).(sqlc.Operator)
	return operator, ok
}

type OperatorMiddleware struct {
	queries *sqlc.Queries
}

func NewOperatorMiddleware(q *sqlc.Queries) *OperatorMiddleware {
	return &OperatorMiddleware{queries: q}
}

func (m *OperatorMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-Operator-Key")
		if apiKey == "" {
			http.Error(w, "missing X-Operator-Key", http.StatusUnauthorized)
			return
		}

		operator, err := m.queries.GetOperatorByApiKey(r.Context(), apiKey)
		if err != nil {
			http.Error(w, "invalid operator api key", http.StatusUnauthorized)
			return
		}

		ctx := ContextWithOperator(r.Context(), operator)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
