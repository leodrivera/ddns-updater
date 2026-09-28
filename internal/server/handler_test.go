package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/qdm12/ddns-updater/internal/records"
	"github.com/stretchr/testify/assert"
)

type emptyDatabase struct{}

func (emptyDatabase) SelectAll() []records.Record { return nil }

type countingUpdater struct{ calls atomic.Int32 }

func (u *countingUpdater) ForceUpdate(context.Context) []error {
	u.calls.Add(1)
	return nil
}

func Test_update_postOnlyAndSameOrigin(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		method        string
		headers       map[string]string
		status        int
		updatesForced int32
	}{
		"same_origin_post": {
			method:        http.MethodPost,
			headers:       map[string]string{"Sec-Fetch-Site": "same-origin"},
			status:        http.StatusAccepted,
			updatesForced: 1,
		},
		"non_browser_post": { // no Sec-Fetch-Site nor Origin, e.g. curl
			method:        http.MethodPost,
			status:        http.StatusAccepted,
			updatesForced: 1,
		},
		"get_refused": {
			method: http.MethodGet,
			status: http.StatusMethodNotAllowed,
		},
		"cross_site_post_refused": {
			method:  http.MethodPost,
			headers: map[string]string{"Sec-Fetch-Site": "cross-site"},
			status:  http.StatusForbidden,
		},
		"foreign_origin_post_refused": {
			method:  http.MethodPost,
			headers: map[string]string{"Origin": "https://attacker.example"},
			status:  http.StatusForbidden,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			updater := &countingUpdater{}
			handler := newHandler(t.Context(), "/", emptyDatabase{}, updater, nil)

			request := httptest.NewRequest(testCase.method, "http://ddns.example.com/update", nil)
			for key, value := range testCase.headers {
				request.Header.Set(key, value)
			}
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)

			assert.Equal(t, testCase.status, recorder.Code)
			assert.Equal(t, testCase.updatesForced, updater.calls.Load())
		})
	}
}
