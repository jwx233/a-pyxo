package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jwx233s/a-service/pkg/response"
	"github.com/jwx233s/a-service/pkg/storage"
)

// parsePath 从路径中提取 action
// /api/file/upload -> action="upload"
func parsePath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) >= 4 {
		return parts[3]
	}
	return ""
}

// Handler 文件操作统一入口
// 路由格式: /api/file/:action
//
// 示例:
//   POST /api/file/upload - 上传文件
//   POST /api/file/delete?filename=xxx.jpg - 删除文件
func Handler(w http.ResponseWriter, r *http.Request) {
	response.SetHeaders(w)

	if r.Method == "OPTIONS" {
		return
	}

	action := parsePath(r.URL.Path)
	if action == "" {
		response.Error(w, "Invalid path. Use: /api/file/{action}", 400)
		return
	}

	switch action {
	case "upload":
		handleUpload(w, r)
	case "delete":
		handleDelete(w, r)
	default:
		response.Error(w, "Invalid action. Use: upload or delete", 400)
	}
}

// handleUpload 处理文件上传
func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		response.Error(w, "Only POST method is allowed", 405)
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

	fileURL, err := storage.UploadFile("file", file, header)
	if err != nil {
		response.Error(w, "Upload failed: "+err.Error(), 500)
		return
	}

	result := map[string]interface{}{"url": fileURL}
	jsonData, _ := json.Marshal(result)
	response.Success(w, json.RawMessage(jsonData))
}

// handleDelete 处理文件删除
func handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		response.Error(w, "Only POST method is allowed", 405)
		return
	}

	filename := r.URL.Query().Get("filename")
	if filename == "" {
		response.Error(w, "Missing filename parameter", 400)
		return
	}

	err := storage.DeleteFile("file", filename)
	if err != nil {
		response.Error(w, "Delete failed: "+err.Error(), 500)
		return
	}

	response.Success(w, map[string]string{"message": "File deleted successfully"})
}
