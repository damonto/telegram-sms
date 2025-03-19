package util

type IfValue interface {
	~int | ~int8 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint32 | ~uint64 |
		~float32 | ~float64 | ~string | ~bool |
		~[]byte | ~[]rune
}

func If[T IfValue](condition bool, trueVal T, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}

func When(condition bool, f func() error) error {
	if condition {
		return f()
	}
	return nil
}
