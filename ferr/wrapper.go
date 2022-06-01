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
		return w.cause.Error()
	} else {
		return "cause: nil"
	}
}

func Summarize(err error) *ErrorSummary {
	if werr, ok := err.(*Wrapper); ok {
		return werr.Summarize()
	} else {
		return &ErrorSummary{
			Cause: err.Error(),
			Stack: []*SummaryFrame{},
		}
	}
}

func (w *Wrapper) Summarize() *ErrorSummary {
	if unwrapped, ok := w.cause.(*Wrapper); ok {
		es := unwrapped.Summarize()

		return &ErrorSummary{
			Cause: es.Cause,
			Stack: append([]*SummaryFrame{w.summarize()}, es.Stack...),
		}
	} else {
		return &ErrorSummary{
			Cause: w.cause.Error(),
			Stack: []*SummaryFrame{w.summarize()},
		}
	}
}

func (w *Wrapper) summarize() *SummaryFrame {
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
	return WrapWithOffset(err, 1)
}

func Wrapf(err error, message string, args ...any) *Wrapper {
	wrapped := WrapWithOffset(err, 1)
	wrapped.extraMessage = fmt.Sprintf(message, args...)
	return wrapped
}