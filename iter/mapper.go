package iter

type MapFunc[S any, D any] func(item S) D

type MappingIterator[Sit any, Si Iterator[Sit], Rt any] struct {
	sourceIter Si
	mapFunc    MapFunc[Sit, Rt]
}

func (m *MappingIterator[Sit, Si, Rt]) Next() (Rt, error) {
	nextItem, err := m.sourceIter.Next()
	if err != nil {
		var rt Rt
		return rt, err
	}

	return m.mapFunc(nextItem), nil
}

func (m *MappingIterator[Sit, Si, Rt]) HasNext() bool {
	return m.sourceIter.HasNext()
}

func Map[St any, Dt any](sourceIter Iterator[St], mapFunc MapFunc[St, Dt]) *MappingIterator[St, Iterator[St], Dt] {
	return &MappingIterator[St, Iterator[St], Dt]{
		sourceIter: sourceIter,
		mapFunc:    mapFunc,
	}
}
