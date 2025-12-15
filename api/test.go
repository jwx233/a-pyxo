package handler

import (
	"net/http"

	"github.com/jwx233s/a-service/pkg/response"
)

// Handler 测试接口
func Handler(w http.ResponseWriter, r *http.Request) {
	response.SetHeaders(w)
	
	if r.Method == "OPTIONS" {
		return
	}

	response.Success(w, map[string]string{
		"message": "Hello World",
	})
}
