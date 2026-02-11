package interfaces

import "net/http"

type HeaderProcessor interface {
	//Process the headers that are returned to a certain request
	ProcessHeader(req *http.Request, headerName string, headerValues []string)
}
