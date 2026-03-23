package middleware

// Redactor is an interface that can be implemented by types containing sensitive
// information. When logged, the Redact() method will be called to return a safe
// representation of the data.
type Redactor interface {
	Redact() any
}
