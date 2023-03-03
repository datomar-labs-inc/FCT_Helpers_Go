package iter

type SliceIterator[T any] struct {
	internalSlice []T
	idx           int
}

func (s *SliceIterator[T]) Next() (T, error) {
	if len(s.internalSlice) - 1 < s.idx {
		var t T
		return t, ErrNoMoreElements
	}

	defer func() {
		s.idx++
	}()

	return s.internalSlice[s.idx], nil
}

func (s *SliceIterator[T]) HasNext() bool {
	return len(s.internalSlice) - 1 > s.idx
}

func Slice[T any](sliceToIterate []T) *SliceIterator[T] {
	return &SliceIterator[T]{
		internalSlice: sliceToIterate,
		idx:           0,
	}
}

func ToSlice[Si any](sourceIter Iterator[Si]) ([]Si, error) {
	var resultSlice []Si

	for {
		item, err := sourceIter.Next()
		if err == ErrNoMoreElements {
			return resultSlice, nil
		} else if err != nil {
			return resultSlice, err
		}

		resultSlice = append(resultSlice, item)
	}
}
