package handler

import (
	"encoding/json"
	"net/http"

	"github.com/jwx233s/a-service/pkg/response"
	"github.com/jwx233s/a-service/pkg/storage"
)

// UploadHandler 文件上传统一入口
// 路由格式: POST /api/upload?bucket=xxx
//
// 参数:
//   - bucket: 存储桶名称（必填，通过 query 参数传递）
//   - file: 文件内容（必填，通过 multipart/form-data 上传）
//
// 示例:
//   POST /api/upload?bucket=avatars
//   Content-Type: multipart/form-data
//   Body: file=@image.jpg
//
// 响应:
//   {
//     "code": 200,
//     "data": {
//       "url": "https://xxx.supabase.co/storage/v1/object/public/avatars/123456.jpg"
//     },
//     "message": "success"
//   }
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	response.SetHeaders(w)
	
	// 处理 OPTIONS 预检请求
	if r.Method == "OPTIONS" {
		return
	}
	
	// 只允许 POST 请求
	if r.Method != "POST" {
		response.Error(w, "Only POST method is allowed", 405)
		return
	}
	
	// 获取存储桶名称
	bucket := r.URL.Query().Get("bucket")
	if bucket == "" {
		response.Error(w, "Missing bucket parameter", 400)
		return
	}
	
	// 解析 multipart form（最大 32MB）
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		response.Error(w, "Failed to parse form: "+err.Error(), 400)
		return
	}
	
	// 获取上传的文件
	file, header, err := r.FormFile("file")
	if err != nil {
		response.Error(w, "Missing file in request: "+err.Error(), 400)
		return
	}
	defer file.Close()
	
	// 上传文件到 Supabase Storage
	fileURL, err := storage.UploadFile(bucket, file, header)
	if err != nil {
		response.Error(w, "Upload failed: "+err.Error(), 500)
		return
	}
	
	// 返回成功响应
	result := map[string]interface{}{
		"url": fileURL,
	}
	jsonData, _ := json.Marshal(result)
	response.Success(w, json.RawMessage(jsonData))
}
