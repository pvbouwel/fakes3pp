package interfaces

import "net/http"

type HeaderProcessor interface {
	//ProcessHeader is called once per response header returned by a proxy backend; implementations may log to access log
	ProcessHeader(req *http.Request, headerName string, headerValues []string)
}
