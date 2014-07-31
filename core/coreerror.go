// core error
package core

type CoreError interface {
	Error() string
	InnerError() error
	Stack() []byte
}
