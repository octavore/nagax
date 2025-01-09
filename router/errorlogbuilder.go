package router

import (
	"fmt"
	"net/http"
	"path"

	"github.com/go-errors/errors"

	"github.com/octavore/nagax/router/httperror"
)

// handleErrorLogBuilder is a helper to generate a log line for HandleError
type handleErrorLogBuilder struct {
	path   string
	status int
	action string
	detail *string
	err    error
}

func newHandlerErrorLogBuilder(req *http.Request, status int) *handleErrorLogBuilder {
	path := "<unknown>"
	if req != nil && req.URL != nil {
		path = req.URL.Path
	}
	return &handleErrorLogBuilder{
		path:   path,
		status: status,
	}
}

func (b *handleErrorLogBuilder) WithAction(action string) *handleErrorLogBuilder {
	b.action = action
	return b
}

func (b *handleErrorLogBuilder) WithDetail(detail string) *handleErrorLogBuilder {
	b.detail = &detail
	return b
}

func (b *handleErrorLogBuilder) WithError(err error) *handleErrorLogBuilder {
	b.err = err
	return b
}

func (b *handleErrorLogBuilder) String() string {
	msg := fmt.Sprintf("[%d] %s %s", b.status, b.path, b.action)
	if b.detail != nil {
		msg += fmt.Sprintf(" detail:%q", *b.detail)
	}

	loggedError := httperror.UnwrapAll(b.err)
	var httpErr *httperror.HTTPError
	if errors.As(b.err, &httpErr) && httpErr.BaseError == nil {
		msg += fmt.Sprintf(" error:<nil>")
	} else if b.err != nil {
		msg += fmt.Sprintf(" error:%q error-type:%T", loggedError, loggedError)
	}

	var errWithStack *errors.Error
	if errors.As(b.err, &errWithStack) {
		// get the file name and line number of file where error ocurred
		s := errWithStack.StackFrames()[0]
		f := path.Base(s.File)
		msg += fmt.Sprintf(" loc:%s/%s|%d", s.Package, f, s.LineNumber)
	}
	return msg
}
