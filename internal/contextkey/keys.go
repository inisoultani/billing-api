package contextkey

// shareable context key that will be use by logger, middleware, etc
type Key string

const (
	IdempotencyKey Key = "idempotency_key"
	RequestIDKey   Key = "request_id"
)
