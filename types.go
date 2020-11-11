package envconfig

// Value is the interface that any custom types that want to be processed by this package have to adhere to
type Value interface {
	Set(v string) error
}
