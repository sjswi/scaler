package controller

import (
	"encoding/json"
	"io"
	"net/http"
)

func tableHandler(writer http.ResponseWriter, request *http.Request) {
	resp := new(NameToClusterResponse)
	req := new(NameToClusterRequest)
	bytes, err := io.ReadAll(request.Body)
	if err != nil {
		resp.Status = "error"

	}
	json.Unmarshal(bytes, &req)

}
