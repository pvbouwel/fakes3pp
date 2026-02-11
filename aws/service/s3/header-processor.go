package s3

import (
	"log/slog"
	"net/http"

	"github.com/VITObelgium/fakes3pp/aws/service/s3/interfaces"
	"github.com/VITObelgium/fakes3pp/logging"
	"github.com/VITObelgium/fakes3pp/requestctx"
)

type headerToAccessLog struct {
	headers map[string]bool
}

func NewHeaderProcessor(headers []string) interfaces.HeaderProcessor {
	if len(headers) == 0 {
		return nil
	}
	headerMap := map[string]bool{}
	for _, h := range headers {
		headerMap[h] = true
	}
	return &headerToAccessLog{headers: headerMap}
}

func (h *headerToAccessLog) ProcessHeader(r *http.Request, headerName string, headerValues []string) {
	_, ok := h.headers[headerName]
	if ok {
		if len(headerValues) < 1 {
			slog.DebugContext(r.Context(), "Encountered header with no values", "header", headerName, "values", headerValues)
		} else {
			lattr := slog.Attr{
				Key:   headerName,
				Value: logging.SafeString(headerValues[0]),
			}
			requestctx.AddAccessLogInfo(r, "headers", lattr)
		}
	}
}
