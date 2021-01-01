package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lobsterdore/release-dash/web/handler"

	"github.com/stretchr/testify/assert"
)

func TestHealthcheckOk(t *testing.T) {
	// ctrl := gomock.NewController(t)

	req, err := http.NewRequest("GET", "/healthcheck", nil)
	if err != nil {
		t.Fatal(err)
	}

	mockCtx := context.Background()
	req = req.WithContext(mockCtx)

	rr := httptest.NewRecorder()

	healthcheckHandler := handler.HealthcheckHandler{}
	handler := http.HandlerFunc(healthcheckHandler.Http)

	handler.ServeHTTP(rr, req)
	resBody := rr.Body.String()

	assert.Equal(t, rr.Code, http.StatusOK)
	assert.Contains(t, resBody, `{"status":"OK","errors":[]}`)
}
