package bugsnag

import (
	"context"
	"fmt"
	"net/http"

	goerrors "github.com/go-errors/errors"

	"github.com/octavore/nagax/logger"
	"github.com/octavore/nagax/util/errors"
)

type bugsnagLogger struct {
	logger.Logger
	Notify func(error, ...any) // Notify is the module.Notify which wraps bugsnag.Notify with added print
}

func (b *bugsnagLogger) Error(args ...any) {
	if len(args) == 1 {
		if originalErr, ok := args[0].(error); ok {
			// this handles case with only one error arugment
			var req *http.Request
			if re, ok := originalErr.(GetRequestable); ok {
				req = re.GetRequest()
			}
			if err, ok := originalErr.(*goerrors.Error); ok {
				if re, ok := err.Err.(GetRequestable); ok && re.GetRequest() != nil {
					req = re.GetRequest()
				}
			} else {
				originalErr = errors.WrapS(originalErr, 1)
			}
			if req != nil {
				b.Notify(originalErr, req)
			} else {
				b.Notify(originalErr)
			}
			return
		}
	}
	b.Notify(errors.New(fmt.Sprint(args...)))
}

func (b *bugsnagLogger) Errorf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	b.Error(msg)
}

func (b *bugsnagLogger) ErrorCtx(ctx context.Context, args ...any) {
	b.Error(args...)
}
