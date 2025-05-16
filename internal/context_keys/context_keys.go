// Package contextkeys пакет с константами для ключей в модифицированном контексте.
package contextkeys

// ContextKey строка названия ключа для значения в модифицированном контексте.
type ContextKey string

const (
	// UserID ключ для значения ID пользователя в контексте.
	UserID ContextKey = "UserID"
)
