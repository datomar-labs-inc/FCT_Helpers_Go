package prettylogger

import (
	"sync"

	"go.uber.org/zap/buffer"
)

var (
	_prettyloggerPool = sync.Pool{New: func() interface{} {
		return &prettyloggerEncoder{}
	}}
	_recorderPool = sync.Pool{New: func() interface{} {
		return &recordingEncoder{}
	}}
	_bufferPool    = buffer.NewPool()
	_bufferPoolGet = _bufferPool.Get
)

func getRecordingEncoder() *recordingEncoder {
	return _recorderPool.Get().(*recordingEncoder)
}

func putRecordingEncoder() *recordingEncoder {
	return _recorderPool.Get().(*recordingEncoder)
}

func getPrettyloggerEncoder() *prettyloggerEncoder {
	return _prettyloggerPool.Get().(*prettyloggerEncoder)
}

func putPrettyloggerEncoder(e *prettyloggerEncoder) {
	e.cfg = nil
	if e.buf != nil {
		putBuffer(e.buf)
	}
	e.buf = nil

	e.namespaceIndent = 0
	e.inList = false
	e.listSep = ""
	e._listSepSpace = ""
	e._listSepComma = ""

	_prettyloggerPool.Put(e)
}

func getBuffer() *buffer.Buffer {
	return _bufferPool.Get()
}

func putBuffer(b *buffer.Buffer) {
	b.Free()
}
