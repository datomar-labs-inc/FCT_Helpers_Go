package iter

import "github.com/datomar-labs-inc/FCT_Helpers_Go/maybe"

type ReduceFunc[Ri any] func(accumulator Ri, item Ri) Ri

func Reduce[Ri any](sourceIter Iterator[Ri], reduceFunc ReduceFunc[Ri]) (maybe.Maybe[Ri], error) {
	if !sourceIter.HasNext() {
		return maybe.Empty[Ri](), nil
	}

	accumulator, err := sourceIter.Next()
	if err == ErrNoMoreElements {
		return maybe.Empty[Ri](), nil
	} else if err != nil {
		return maybe.Empty[Ri](), err
	}

	for {
		item, err := sourceIter.Next()
		if err == ErrNoMoreElements {
			return maybe.WithValue(accumulator), nil
		} else if err != nil {
			return maybe.Empty[Ri](), err
		}

		accumulator = reduceFunc(accumulator, item)
	}
}

type FoldFunc[Si any, Ri any] func(accumulator Ri, item Si) Ri

func Fold[Si any, Ri any](sourceIter Iterator[Si], accumulator Ri, foldFunc FoldFunc[Si, Ri]) (maybe.Maybe[Ri], error) {
	if !sourceIter.HasNext() {
		return maybe.Empty[Ri](), nil
	}

	for {
		item, err := sourceIter.Next()
		if err == ErrNoMoreElements {
			return maybe.WithValue(accumulator), nil
		} else if err != nil {
			return maybe.Empty[Ri](), err
		}

		accumulator = foldFunc(accumulator, item)
	}
}
