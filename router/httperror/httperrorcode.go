package httperror

import "net/http"

type HTTPErrorCode int

func (e HTTPErrorCode) GetCode() int {
	return int(e)
}

func (e HTTPErrorCode) Error() string {
	return http.StatusText(e.GetCode())
}
