package client

func safeUnref[T any](in *T) T {
	if in == nil {
		var zeroValue T
		return zeroValue
	}
	return *in
}
