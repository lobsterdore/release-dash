package handler

import (
	"encoding/json"
	"net/http"
)

type HealthcheckHandler struct{}

type healthcheckData struct {
	Status string    `json:"status"`
	Errors [0]string `json:"errors"`
}

func NewHealthcheckHandler() *HealthcheckHandler {
	return &HealthcheckHandler{}
}

func (h *HealthcheckHandler) Http(respWriter http.ResponseWriter, request *http.Request) {
	hcData := healthcheckData{
		Status: "OK",
	}

	responseBytes, err := json.Marshal(hcData)
	if err != nil {
		respWriter.WriteHeader(http.StatusInternalServerError)
		return
	}

	respWriter.Header().Set("Content-Type", "application/json")
	_, _ = respWriter.Write(responseBytes)
}
