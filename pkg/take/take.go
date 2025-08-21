package take

func Left[T, U any](a T, b U) T {
	return a
}

func Right[T, U any](a T, b U) U {
	return b
}
