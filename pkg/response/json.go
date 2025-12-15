package response

import (
	"encoding/json"
	"net/http"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`    // 状态码: 200 成功, 其他失败
	Data    interface{} `json:"data"`    // 数据
	Message string      `json:"message"` // 消息
}

func SetHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// JSON 返回成功响应，将 id 和 json 字段内容合并到同一对象
func JSON(w http.ResponseWriter, data []byte) {
	SetHeaders(w)

	// 解析原始数据
	var rawData []map[string]interface{}
	json.Unmarshal(data, &rawData)

	// 合并 id 和 json 字段内容
	var result []map[string]interface{}
	for _, item := range rawData {
		record := make(map[string]interface{})
		
		// 先添加 json 字段的所有内容
		if jsonField, ok := item["json"].(map[string]interface{}); ok {
			for key, value := range jsonField {
				// 如果 json 中有 id 字段，重命名为 json.id
				if key == "id" {
					record["json.id"] = value
				} else {
					record[key] = value
				}
			}
		}
		
		// 再添加数据库的 id 字段
		if id, ok := item["id"]; ok {
			record["id"] = id
		}
		
		result = append(result, record)
	}

	resp := Response{
		Code:    200,
		Data:    result,
		Message: "success",
	}
	json.NewEncoder(w).Encode(resp)
}

// Success 返回成功响应
func Success(w http.ResponseWriter, data interface{}) {
	SetHeaders(w)
	resp := Response{
		Code:    200,
		Data:    data,
		Message: "success",
	}
	json.NewEncoder(w).Encode(resp)
}

// Error 返回错误响应
func Error(w http.ResponseWriter, msg string, code int) {
	SetHeaders(w)
	w.WriteHeader(code)
	resp := Response{
		Code:    code,
		Data:    nil,
		Message: msg,
	}
	json.NewEncoder(w).Encode(resp)
}
