package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/jwx233s/a-service/pkg/db"
	"github.com/jwx233s/a-service/pkg/response"
)
// parsePath 从路径中提取 table 和 action
// /api/db/user/get -> table="user", action="get"
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

// reqContext 请求上下文，封装每次请求的公共参数
type reqContext struct {
	table  string // 表名
	filter string // 过滤条件 (id=eq.1 或 json->>name=eq.Tom)
	body   string // 请求体 JSON
}

// Handler 数据库 CRUD 统一入口
// 路由格式: /api/db/{action}/{table}
//
// 示例:
//   GET  /api/db/get/user              - 查询 user 表全部
//   GET  /api/db/get/user?id=1         - 按 id 查询
//   GET  /api/db/get/user?name=Tom - 按 json 字段查询
//   POST /api/db/insert/user           - 新增记录
//   POST /api/db/update/user?id=1      - 更新记录
//   POST /api/db/delete/user?id=1      - 删除记录
func Handler(w http.ResponseWriter, r *http.Request) {
	response.SetHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	// 解析路径，提取 action 和 table
	table, action := parsePath(r.URL.Path)
	if action == "" || table == "" {
		response.Error(w, "Invalid path. Use: /api/db/{table}/{action}", 400)
		return
	}

	// 封装请求上下文
	ctx := &reqContext{
		table:  table,
		filter: db.BuildFilter(r), // 从 URL 参数构建过滤条件
		body:   readBody(r),
	}

	// action -> handler 映射
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

	// 执行对应操作
	data, err := handler(ctx)
	if err != nil {
		response.Error(w, err.Error(), 400)
		return
	}
	response.JSON(w, data)
}



// doGet 查询操作
func doGet(ctx *reqContext) ([]byte, error) {
	query := "select=*"
	if ctx.filter != "" {
		query += "&" + ctx.filter
	}
	// 调试: 打印最终查询
	fmt.Printf("[DEBUG doGet] table=%s, query=%s\n", ctx.table, query)
	return db.Select(ctx.table, query)
}

// wrapJsonBody 将 body 包装成 {"json": body} 格式
func wrapJsonBody(body string) string {
	return `{"json":` + body + `}`
}

// doInsert 新增操作
func doInsert(ctx *reqContext) ([]byte, error) {
	if ctx.body == "" {
		return nil, fmt.Errorf("Missing body")
	}
	// 自动包装成 {"json": ...} 格式
	return db.Insert(ctx.table, wrapJsonBody(ctx.body))
}

// doUpdate 更新操作，需要 filter 和 body
func doUpdate(ctx *reqContext) ([]byte, error) {
	if ctx.filter == "" || ctx.body == "" {
		return nil, fmt.Errorf("Missing filter or body")
	}
	// 自动包装成 {"json": ...} 格式
	return db.Update(ctx.table, ctx.filter, wrapJsonBody(ctx.body))
}

// doDelete 删除操作，需要 filter
func doDelete(ctx *reqContext) ([]byte, error) {
	if ctx.filter == "" {
		return nil, fmt.Errorf("Missing filter")
	}
	return db.Delete(ctx.table, ctx.filter)
}

