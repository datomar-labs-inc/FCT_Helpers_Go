package prettylogger

import (
	"bytes"
	"time"

	"go.uber.org/zap/zapcore"
)

// Test interface conformance
var _ zapcore.ArrayEncoder = (*prettyloggerEncoder)(nil)

func (e *prettyloggerEncoder) AppendComplex64(v complex64)   { e.appendComplex(complex128(v), 32) }
func (e *prettyloggerEncoder) AppendComplex128(v complex128) { e.appendComplex(v, 64) }
func (e *prettyloggerEncoder) AppendFloat32(v float32)       { e.appendFloat(float64(v), 32) }
func (e *prettyloggerEncoder) AppendFloat64(v float64)       { e.appendFloat(v, 64) }
func (e *prettyloggerEncoder) AppendInt(v int)               { e.AppendInt64(int64(v)) }
func (e *prettyloggerEncoder) AppendInt32(v int32)           { e.AppendInt64(int64(v)) }
func (e *prettyloggerEncoder) AppendInt16(v int16)           { e.AppendInt64(int64(v)) }
func (e *prettyloggerEncoder) AppendInt8(v int8)             { e.AppendInt64(int64(v)) }
func (e *prettyloggerEncoder) AppendUint(v uint)             { e.AppendUint64(uint64(v)) }
func (e *prettyloggerEncoder) AppendUint32(v uint32)         { e.AppendUint64(uint64(v)) }
func (e *prettyloggerEncoder) AppendUint16(v uint16)         { e.AppendUint64(uint64(v)) }
func (e *prettyloggerEncoder) AppendUint8(v uint8)           { e.AppendUint64(uint64(v)) }
func (e *prettyloggerEncoder) AppendUintptr(v uintptr)       { e.AppendUint64(uint64(v)) }

func (e *prettyloggerEncoder) AppendBool(b bool) {
	e.addSeparator()
	e.buf.AppendBool(b)

	e.inList = true
	e.listSep = e._listSepComma
}

func (e *prettyloggerEncoder) AppendByteString(bytes []byte) {
	e.addSeparator()
	e.appendSafeByte(bytes)

	e.inList = true
	e.listSep = e._listSepComma
}

func (e *prettyloggerEncoder) appendComplex(c complex128, precision int) {
	e.addSeparator()
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
	e.listSep = e._listSepComma
}

func (e *prettyloggerEncoder) appendFloat(f float64, precision int) {
	e.addSeparator()
	e.buf.AppendFloat(f, precision)

	e.inList = true
	e.listSep = e._listSepComma
}

func (e *prettyloggerEncoder) AppendInt64(i int64) {
	e.addSeparator()
	e.buf.AppendInt(i)

	e.inList = true
	e.listSep = e._listSepComma
}

func (e *prettyloggerEncoder) AppendString(s string) {
	e.addSeparator()
	e.addSafeString(s)

	e.inList = true
	e.listSep = e._listSepComma
}

func (e *prettyloggerEncoder) AppendUint64(u uint64) {
	e.addSeparator()
	e.buf.AppendUint(u)

	e.inList = true
	e.listSep = e._listSepComma
}

func (e *prettyloggerEncoder) AppendDuration(duration time.Duration) {
	e.addSeparator()
	cur := e.buf.Len()
	if durationEncoder := e.cfg.EncodeDuration; e != nil {
		durationEncoder(duration, e)
	}
	if cur == e.buf.Len() {
		// User-supplied EncodeDuration is a no-op. Fall back to nanoseconds to keep
		// JSON valid.
		e.buf.AppendInt(int64(duration))
	}

	e.inList = true
	e.listSep = e._listSepComma
}

func (e *prettyloggerEncoder) AppendTime(t time.Time) {
	e.addSeparator()
	cur := e.buf.Len()
	if timeEncoder := e.cfg.EncodeTime; e != nil {
		timeEncoder(t, e)
	}
	if cur == e.buf.Len() {
		// User-supplied EncodeTime is a no-op. Fall back to RFC3339
		e.buf.AppendTime(t, time.RFC3339)
	}

	e.inList = true
	e.listSep = e._listSepComma
}

func (e *prettyloggerEncoder) AppendArray(marshaler zapcore.ArrayMarshaler) error {
	e.addSeparator()
	enc := e.clone()
	enc.OpenNamespace("")
	enc.colorizeAtLevel("[")
	enc.inList = false
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
	e.listSep = e._listSepComma
	return nil
}

func (e *prettyloggerEncoder) AppendObject(marshaler zapcore.ObjectMarshaler) error {
	e.addSeparator()
	enc := e.clone()
	enc.OpenNamespace("")
	enc.colorizeAtLevel("{")
	enc.inList = false
	enc.keyPrefix = ""
	l := enc.buf.Len()

	if err := marshaler.MarshalLogObject(enc); err != nil {
		return err
	}
	if bytes.ContainsRune(enc.buf.Bytes()[l:], '\n') {
		enc.buf.AppendString(e.cfg.LineEnding)
		for ii := 0; ii < enc.namespaceIndent-1; ii++ {
			enc.buf.AppendByte(' ')
		}
	}
	enc.colorizeAtLevel("}")

	_, _ = e.buf.Write(enc.buf.Bytes())
	putPrettyloggerEncoder(enc)

	e.inList = true
	e.listSep = e._listSepComma
	return nil
}

func (e *prettyloggerEncoder) AppendReflected(value interface{}) error {
	e.addSeparator()
	enc := e.clone()
	enc.OpenNamespace("")
	enc.keyPrefix = ""
	enc.inList = false
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
	e.listSep = e._listSepComma
	return nil
}
