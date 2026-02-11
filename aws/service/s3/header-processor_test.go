package s3_test

import (
	"net/http"
	"testing"

	"github.com/VITObelgium/fakes3pp/aws/service/s3"
	"github.com/VITObelgium/fakes3pp/requestctx"
	"github.com/google/go-cmp/cmp"
)

func TestHeaderProcessor(t *testing.T) {
	var tests = []struct {
		scenario              string
		providedStrings       []string
		responseHeaders       map[string][]string
		expectedLoggedHeaders map[string]string
	}{
		{
			"Simple single header matching",
			[]string{"a", "b"},
			map[string][]string{
				"a": {"ok"},
			},
			map[string]string{
				"a": "ok",
			},
		},
		{
			"Simple two header matching",
			[]string{"a", "b"},
			map[string][]string{
				"a": {"ok"},
				"b": {"wow"},
			},
			map[string]string{
				"a": "ok",
				"b": "wow",
			},
		},
		{
			"Simple two header only 1 matching",
			[]string{"a"},
			map[string][]string{
				"a": {"ok"},
				"b": {"wow"},
			},
			map[string]string{
				"a": "ok",
			},
		},
		{
			"Simple two header none matching",
			[]string{"c", "d"},
			map[string][]string{
				"a": {"ok"},
				"b": {"wow"},
			},
			map[string]string{},
		},
		{
			"Simple two header no matcher",
			[]string{},
			map[string][]string{
				"a": {"ok"},
				"b": {"wow"},
			},
			map[string]string{},
		},
	}

	for _, tt := range tests {

		r, err := http.NewRequest("GET", "http://localhost", nil)
		if err != nil {
			t.Errorf("Encountered error %q", err)
			t.FailNow()
		}

		r = r.WithContext(requestctx.NewContextFromHttpRequest(r))

		hp := s3.NewHeaderProcessor(tt.providedStrings)
		for hk, hv := range tt.responseHeaders {
			if hp != nil {
				hp.ProcessHeader(r, hk, hv)
			}
		}

		for ek, ev := range tt.expectedLoggedHeaders {
			v := requestctx.GetAccessLogStringInfo(r, "headers", ek)
			if v != ev {
				t.Errorf("Got %q Expected %q", v, ev)
			}
		}

		rc, ok := requestctx.FromContext(r.Context())
		if !ok {
			t.Error("Was not able to get requestcontext")
		}
		loggedHeaders := map[string]string{}
		for _, v := range rc.GetAccessLogInfo() {
			if v.Key == "headers" {
				vals := v.Value.Group()
				for _, val := range vals {
					loggedHeaders[val.Key] = val.Value.String()
				}
			}
		}
		if diff := cmp.Diff(tt.expectedLoggedHeaders, loggedHeaders); diff != "" {
			t.Fatalf("maps differ (-expected +actual):\n%s", diff)
		}

	}
}
