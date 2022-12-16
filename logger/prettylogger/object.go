package prettylogger

import (
	"bytes"
	"encoding/base64"
	"strings"
	"time"

	"go.uber.org/zap/zapcore"
)

// Test interface conformance
var _ zapcore.ObjectEncoder = (*prettyloggerEncoder)(nil)

func (e *prettyloggerEncoder) AddFloat32(k string, v float32) { e.addFloat(k, float64(v), 32) }
func (e *prettyloggerEncoder) AddFloat64(k string, v float64) { e.addFloat(k, v, 64) }
func (e *prettyloggerEncoder) AddInt(k string, v int)         { e.AddInt64(k, int64(v)) }
func (e *prettyloggerEncoder) AddInt32(k string, v int32)     { e.AddInt64(k, int64(v)) }
func (e *prettyloggerEncoder) AddInt16(k string, v int16)     { e.AddInt64(k, int64(v)) }
func (e *prettyloggerEncoder) AddInt8(k string, v int8)       { e.AddInt64(k, int64(v)) }
func (e *prettyloggerEncoder) AddUint(k string, v uint)       { e.AddUint64(k, uint64(v)) }
func (e *prettyloggerEncoder) AddUint32(k string, v uint32)   { e.AddUint64(k, uint64(v)) }
func (e *prettyloggerEncoder) AddUint16(k string, v uint16)   { e.AddUint64(k, uint64(v)) }
func (e *prettyloggerEncoder) AddUint8(k string, v uint8)     { e.AddUint64(k, uint64(v)) }
func (e *prettyloggerEncoder) AddUintptr(k string, v uintptr) { e.AddUint64(k, uint64(v)) }
func (e *prettyloggerEncoder) AddBinary(key string, value []byte) {
	e.AddString(key, base64.StdEncoding.EncodeToString(value))
}
func (e *prettyloggerEncoder) AddComplex64(k string, v complex64) {
	e.addComplex(k, complex128(v), 32)
}
func (e *prettyloggerEncoder) AddComplex128(k string, v complex128) {
	e.addComplex(k, v, 64)
}

func (e *prettyloggerEncoder) OpenNamespace(key string) {
	if e.namespaceIndent == 0 {
		e.buf.AppendString(e.cfg.LineEnding)
		e.colorizeAtLevel("  ↳ " + key)
		e.namespaceIndent = 4 + len(key)
	} else {
		if e.inList {
			e.buf.AppendString(e.cfg.LineEnding)
			for ii := 0; ii < e.namespaceIndent; ii++ {
				e.buf.AppendByte(' ')
			}
		}
		if len(key) > 0 {
			e.colorizeAtLevel(e.keyPrefix + key)
		}
		e.namespaceIndent += 1 + len(key)
	}
	e.inList = false
	e.listSep = e._listSepSpace
	e.keyPrefix = "."
}

func (e *prettyloggerEncoder) AddObject(key string, marshaler zapcore.ObjectMarshaler) error {
	enc := e.clone()
	enc.OpenNamespace(key)

	if err := marshaler.MarshalLogObject(enc); err != nil {
		return err
	}

	_, _ = e.buf.Write(enc.buf.Bytes())
	putPrettyloggerEncoder(enc)

	e.inList = true
	e.listSep = e.cfg.LineEnding + strings.Repeat(" ", e.namespaceIndent)
	return nil
}

func (e *prettyloggerEncoder) AddArray(key string, marshaler zapcore.ArrayMarshaler) error {
	enc := e.clone()
	enc.OpenNamespace(key)

	enc.colorizeAtLevel("=[")
	enc.namespaceIndent += 2
	l := enc.buf.Len()

	if err := marshaler.MarshalLogArray(enc); err != nil {
		return err
	}
	if bytes.ContainsRune(enc.buf.Bytes()[l:], '\n') {
		enc.buf.AppendString(e.cfg.LineEnding)
		for ii := 0; ii < enc.namespaceIndent-1; ii++ {
			enc.buf.AppendByte(' ')
		}
	}
	enc.colorizeAtLevel("]")

	_, _ = e.buf.Write(enc.buf.Bytes())
	putPrettyloggerEncoder(enc)

	e.inList = true
	e.listSep = e.cfg.LineEnding + strings.Repeat(" ", e.namespaceIndent)
	return nil
}

func (e *prettyloggerEncoder) AddReflected(key string, value interface{}) error {
	enc := e.clone()
	enc.OpenNamespace(key)

	enc.colorizeAtLevel("=")
	enc.namespaceIndent += 1
	l := enc.buf.Len()
	iw := indentingWriter{
		buf:        enc.buf,
		indent:     enc.namespaceIndent,
		lineEnding: []byte(e.cfg.LineEnding),
	}

	if reflectedEncoder := e.cfg.NewReflectedEncoder(iw); e != nil {
		if err := reflectedEncoder.Encode(value); err != nil {
			return err
		}
	}
	if l-enc.buf.Len() == 0 {
		// User-supplied reflectedEncoder is a no-op. Fall back to dd
		if err := defaultReflectedEncoder(iw).Encode(value); err != nil {
			return err
		}
	}

	_, _ = e.buf.Write(enc.buf.Bytes())
	putPrettyloggerEncoder(enc)

	e.inList = true
	e.listSep = e.cfg.LineEnding + strings.Repeat(" ", e.namespaceIndent)
	return nil
}

func (e *prettyloggerEncoder) AddByteString(key string, value []byte) {
	e.addSeparator()
	e.addKey(key)
	e.appendSafeByte(value)

	e.inList = true
	e.listSep = e._listSepSpace
}

func (e *prettyloggerEncoder) AddBool(key string, value bool) {
	e.addSeparator()
	e.addKey(key)
	e.buf.AppendBool(value)

	e.inList = true
	e.listSep = e._listSepSpace
}

func (e *prettyloggerEncoder) addComplex(key string, c complex128, precision int) {
	e.addSeparator()
	e.addKey(key)
	// Cast to a platform-independent, fixed-size type.
	r, i := real(c), imag(c)
	// Because we're always in a quoted string, we can use strconv without
	// special-casing NaN and +/-Inf.
	e.buf.AppendFloat(r, precision)
	// If imaginary part is less than 0, minus (-) sign is added by default
	// by AppendFloat.
	if i >= 0 {
		e.buf.AppendByte('+')
	}
	e.buf.AppendFloat(i, precision)
	e.buf.AppendByte('i')

	e.inList = true
	e.listSep = e._listSepSpace
}

func (e *prettyloggerEncoder) AddDuration(key string, value time.Duration) {
	e.addSeparator()
	e.addKey(key)
	cur := e.buf.Len()
	// Both of these append, and we're at the first element of the sublist
	e.inList = false
	if durationEncoder := e.cfg.EncodeDuration; e != nil {
		durationEncoder(value, e)
	}
	if cur == e.buf.Len() {
		// User-supplied EncodeDuration is a no-op. Fall back to Go format
		e.buf.AppendString(value.String())
	}

	e.inList = true
	e.listSep = e._listSepSpace
}

func (e *prettyloggerEncoder) addFloat(key string, value float64, precision int) {
	e.addSeparator()
	e.addKey(key)
	e.buf.AppendFloat(value, precision)

	e.inList = true
	e.listSep = e._listSepSpace
}

func (e *prettyloggerEncoder) AddInt64(key string, value int64) {
	e.addSeparator()
	e.addKey(key)
	e.buf.AppendInt(value)

	e.inList = true
	e.listSep = e._listSepSpace
}

func (e *prettyloggerEncoder) AddString(key, value string) {
	e.addSeparator()
	e.addKey(key)
	e.addSafeString(value)

	e.inList = true
	e.listSep = e._listSepSpace
}

// addIndentedString appends a string, replacing any newlines with the
// current indent.
func (e *prettyloggerEncoder) addIndentedString(key string, s string) {
	e.addSeparator()
	e.addKey(key)
	iw := indentingWriter{
		buf:        e.buf,
		indent:     e.namespaceIndent,
		lineEnding: []byte(e.cfg.LineEnding),
	}
	_, _ = iw.Write([]byte(s))

	e.inList = true
	e.listSep = e._listSepSpace
}

func (e *prettyloggerEncoder) AddTime(key string, value time.Time) {
	e.addSeparator()
	e.addKey(key)
	// Don't use configured time encoder as it's been customized to display the
	// log's time, .e.g, this will be coloured dark grey in time.Kitchen
	e.buf.AppendTime(value, time.RFC3339)

	e.inList = true
	e.listSep = e._listSepSpace
}

func (e *prettyloggerEncoder) AddUint64(key string, value uint64) {
	e.addSeparator()
	e.addKey(key)
	e.buf.AppendUint(value)

	e.inList = true
	e.listSep = e._listSepSpace
}
