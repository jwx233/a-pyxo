package handler

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/jwx233s/a-service/pkg/db"
	"github.com/jwx233s/a-service/pkg/response"
)

// parsePath 从路径中提取 table 和 action
func parsePath(path string) (table, action string) {
	parts := strings.Split(path, "/")
	if len(parts) >= 5 {
		return parts[3], parts[4]
	}
	return "", ""
}

// readBody 读取请求体
func readBody(r *http.Request) string {
	body, _ := io.ReadAll(r.Body)
	return string(body)
}

// reqContext 请求上下文
type reqContext struct {
	table  string
	filter string
	body   string
}

// Handler 数据库 CRUD 统一入口
func Handler(w http.ResponseWriter, r *http.Request) {
	response.SetHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	table, action := parsePath(r.URL.Path)
	if action == "" || table == "" {
		response.Error(w, "Invalid path. Use: /api/db/{table}/{action}", 400)
		return
	}

	ctx := &reqContext{
		table:  table,
		filter: db.BuildFilter(r),
		body:   readBody(r),
	}

	handlers := map[string]func(*reqContext) ([]byte, error){
		"get":    doGet,
		"insert": doInsert,
		"update": doUpdate,
		"delete": doDelete,
	}

	handler, ok := handlers[action]
	if !ok {
		response.Error(w, "Invalid action", 400)
		return
	}

	data, err := handler(ctx)
	if err != nil {
		response.Error(w, err.Error(), 400)
		return
	}
	response.JSON(w, data)
}

func doGet(ctx *reqContext) ([]byte, error) {
	query := "select=*"
	if ctx.filter != "" {
		query += "&" + ctx.filter
	}
	fmt.Printf("[DEBUG doGet] table=%s, query=%s\n", ctx.table, query)
	return db.Select(ctx.table, query)
}

func wrapJsonBody(body string) string {
	return `{"json":` + body + `}`
}

func doInsert(ctx *reqContext) ([]byte, error) {
	if ctx.body == "" {
		return nil, fmt.Errorf("Missing body")
	}
	return db.Insert(ctx.table, wrapJsonBody(ctx.body))
}

func doUpdate(ctx *reqContext) ([]byte, error) {
	if ctx.filter == "" || ctx.body == "" {
		return nil, fmt.Errorf("Missing filter or body")
	}
	// 合并更新：只更新传递的字段，保留原有字段
	mergedBody, err := db.MergeUpdate(ctx.table, ctx.filter, ctx.body)
	if err != nil {
		return nil, err
	}
	return db.Update(ctx.table, ctx.filter, mergedBody)
}

func doDelete(ctx *reqContext) ([]byte, error) {
	if ctx.filter == "" {
		return nil, fmt.Errorf("Missing filter")
	}
	return db.Delete(ctx.table, ctx.filter)
}
