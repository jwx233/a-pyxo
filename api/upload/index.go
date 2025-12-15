package handler

import (
	"encoding/json"
	"net/http"

	"github.com/jwx233s/a-service/pkg/response"
	"github.com/jwx233s/a-service/pkg/storage"
)

// Handler 文件上传统一入口
func Handler(w http.ResponseWriter, r *http.Request) {
	response.SetHeaders(w)

	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		response.Error(w, "Only POST method is allowed", 405)
		return
	}

	bucket := r.URL.Query().Get("bucket")
	if bucket == "" {
		response.Error(w, "Missing bucket parameter", 400)
		return
	}

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		response.Error(w, "Failed to parse form: "+err.Error(), 400)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.Error(w, "Missing file in request: "+err.Error(), 400)
		return
	}
	defer file.Close()

	fileURL, err := storage.UploadFile(bucket, file, header)
	if err != nil {
		response.Error(w, "Upload failed: "+err.Error(), 500)
		return
	}

	result := map[string]interface{}{
		"url": fileURL,
	}
	jsonData, _ := json.Marshal(result)
	response.Success(w, json.RawMessage(jsonData))
}
