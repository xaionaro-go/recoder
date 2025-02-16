package recoder

type CustomOption = any
type CustomOptions []CustomOption

func GetCustomOption[T any](in CustomOptions) (T, bool) {
	for _, item := range in {
		v, ok := item.(T)
		if ok {
			return v, ok
		}
	}

	var zeroValue T
	return zeroValue, false
}
