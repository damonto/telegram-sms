package util

func If[T any](condition bool, trueVal T, falseVal T) T {
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
