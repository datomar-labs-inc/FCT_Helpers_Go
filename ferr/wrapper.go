package ferr

import (
	"fmt"
	"runtime"
)

type ErrorSummary struct {
	Cause string          `json:"cause"`
	Stack []*SummaryFrame `json:"frames"`
}

func (es *ErrorSummary) String() string {
	str := fmt.Sprintf("cause: %s\n", es.Cause)

	for _, frame := range es.Stack {
		str += fmt.Sprintf("\t%s\n\t\t%s#%d\n", frame.Func, frame.File, frame.Line)

		if frame.message != "" {
			str += fmt.Sprintf("\t\t%s\n", frame.message)
		}
	}

	return str
}

type SummaryFrame struct {
	WrappedFrame
	message string
	stack   string
}

type Wrapper struct {
	cause        error
	extraMessage string
	stack        string
	frame        WrappedFrame
}

func (w *Wrapper) Error() string {
	if w.cause != nil {
		return w.Summarize().String()
	}

	return "cause: nil"
}

func (w *Wrapper) Unwrap() error {
	return w.cause
}

func (w *Wrapper) Cause() error {
	return w.cause
}

func Summarize(err error) *ErrorSummary {
	if werr, ok := err.(*Wrapper); ok {
		return werr.Summarize()
	}

	return &ErrorSummary{
		Cause: err.Error(),
		Stack: []*SummaryFrame{},
	}
}

func (w *Wrapper) Summarize() *ErrorSummary {
	if unwrapped, ok := w.cause.(*Wrapper); ok {
		es := unwrapped.Summarize()

		return &ErrorSummary{
			Cause: es.Cause,
			Stack: append([]*SummaryFrame{w.getSummaryFrame()}, es.Stack...),
		}
	}

	var cause string

	if w.Cause() != nil {
		cause = w.Cause().Error()
	}

	return &ErrorSummary{
		Cause: cause,
		Stack: []*SummaryFrame{w.getSummaryFrame()},
	}
}

func (w *Wrapper) getSummaryFrame() *SummaryFrame {
	return &SummaryFrame{
		WrappedFrame: w.frame,
		message:      w.extraMessage,
		stack:        w.stack,
	}
}

func (w *Wrapper) WithStack() *Wrapper {
	stackBuf := make([]byte, 4086)
	stackSize := runtime.Stack(stackBuf, false)
	w.stack = string(stackBuf[:stackSize])

	return w
}

type WrappedFrame struct {
	File string `json:"file"`
	Line int    `json:"line"`
	Func string `json:"func"`
}

func WrapWithOffset(err error, offset int) *Wrapper {
	pc, file, line, ok := runtime.Caller(1 + offset)

	if !ok {
		panic(err) // TODO probably don't do this
	}

	var fnName string

	fn := runtime.FuncForPC(pc)

	if fn != nil {
		fnName = fn.Name()
	}

	frame := WrappedFrame{
		File: file,
		Line: line,
		Func: fnName,
	}

	return &Wrapper{
		cause:        err,
		extraMessage: "",
		frame:        frame,
	}
}

func Wrap(err error) *Wrapper {
	if err != nil {
		return WrapWithOffset(err, 1)
	} else {
		return nil
	}
}

func Wrapf(err error, message string, args ...any) *Wrapper {
	wrapped := WrapWithOffset(err, 1)
	wrapped.extraMessage = fmt.Sprintf(message, args...)
	return wrapped
}
