package storage

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"

	"github.com/jwx233s/a-service/pkg/db"
)

// ============ 配置 ============

const (
	STORAGE_BASE_URL = "https://arwnlqnzofqqxjnlvgqm.supabase.co/storage/v1"
)

// UploadFile 上传文件到 Supabase Storage
// bucket: 存储桶名称
// file: 文件内容
// header: 文件头信息（包含文件名、类型等）
// 返回: 文件 URL 和错误信息
func UploadFile(bucket string, file multipart.File, header *multipart.FileHeader) (string, error) {
	// 生成唯一文件名：时间戳_原文件名
	timestamp := time.Now().UnixMilli()
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%d%s", timestamp, ext)
	
	debugLog("UploadFile", "Original filename", header.Filename)
	debugLog("UploadFile", "Generated filename", filename)
	debugLog("UploadFile", "Content-Type", header.Header.Get("Content-Type"))
	
	// 读取文件内容
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("Failed to read file: %v", err)
	}
	
	// 构建上传 URL（使用 Supabase Storage API）
	uploadURL := fmt.Sprintf("%s/object/%s/%s", STORAGE_BASE_URL, bucket, filename)
	debugLog("UploadFile", "Upload URL", uploadURL)
	
	// 创建请求
	req, err := http.NewRequest("POST", uploadURL, bytes.NewReader(fileBytes))
	if err != nil {
		return "", fmt.Errorf("Failed to create request: %v", err)
	}
	
	// 设置请求头
	req.Header.Set("Authorization", "Bearer "+db.SUPABASE_KEY)
	req.Header.Set("apikey", db.SUPABASE_KEY)
	req.Header.Set("Content-Type", header.Header.Get("Content-Type"))
	
	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Failed to upload file: %v", err)
	}
	defer resp.Body.Close()
	
	// 读取响应
	respBody, _ := io.ReadAll(resp.Body)
	debugLog("UploadFile", "Response Status", fmt.Sprintf("%d", resp.StatusCode))
	debugLog("UploadFile", "Response Body", string(respBody))
	
	// 检查响应状态
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return "", fmt.Errorf("Upload failed with status %d: %s", resp.StatusCode, string(respBody))
	}
	
	// 构建公开访问 URL
	publicURL := fmt.Sprintf("%s/object/public/%s/%s", STORAGE_BASE_URL, bucket, filename)
	debugLog("UploadFile", "Public URL", publicURL)
	
	return publicURL, nil
}

// debugLog 调试日志
func debugLog(action, key, value string) {
	if db.Debug {
		fmt.Printf("[DEBUG Storage] %s | %s: %s\n", action, key, value)
	}
}
