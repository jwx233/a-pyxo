package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/jwx233s/a-service/pkg/db"
	"github.com/jwx233s/a-service/pkg/response"
)

// /api/db/get/user           - 查询全部
// /api/db/get/user?id=1      - 按 id 查询
// /api/db/get/user?json.name=Tom - 按 json 字段查询
// /api/db/insert/user        - 新增
// /api/db/update/user?id=1   - 更新
// /api/db/delete/user?id=1   - 删除
func Handler(w http.ResponseWriter, r *http.Request) {
	response.SetHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	// 解析路径: /api/db/get/user -> action=get, table=user
	action, table := parsePath(r.URL.Path)
	if action == "" || table == "" {
		response.Error(w, "Invalid path. Use: /api/db/{action}/{table}", 400)
		return
	}
	if !db.AllowedTables[table] {
		response.Error(w, "Table not allowed", 403)
		return
	}

	switch action {
	case "get":
		handleGet(w, r, table)
	case "insert":
		handleInsert(w, r, table)
	case "update":
		handleUpdate(w, r, table)
	case "delete":
		handleDelete(w, r, table)
	default:
		response.Error(w, "Invalid action. Use: get, insert, update, delete", 400)
	}
}

func parsePath(path string) (action, table string) {
	// /api/db/get/user -> ["", "api", "db", "get", "user"]
	parts := strings.Split(path, "/")
	if len(parts) >= 5 {
		return parts[3], parts[4]
	}
	return "", ""
}

func handleGet(w http.ResponseWriter, r *http.Request, table string) {
	query := "select=*"
	filter := db.BuildFilter(r)
	if filter != "" {
		query += "&" + filter
	}

	data, err := db.Select(table, query)
	if err != nil {
		response.Error(w, "Query failed", 500)
		return
	}
	response.JSON(w, data)
}

func handleInsert(w http.ResponseWriter, r *http.Request, table string) {
	body, _ := io.ReadAll(r.Body)
	if len(body) == 0 {
		response.Error(w, "Missing body", 400)
		return
	}

	data, err := db.Insert(table, string(body))
	if err != nil {
		response.Error(w, "Insert failed", 500)
		return
	}
	response.JSON(w, data)
}

func handleUpdate(w http.ResponseWriter, r *http.Request, table string) {
	filter := db.BuildFilter(r)
	if filter == "" {
		response.Error(w, "Missing filter (id or json.xxx)", 400)
		return
	}

	body, _ := io.ReadAll(r.Body)
	if len(body) == 0 {
		response.Error(w, "Missing body", 400)
		return
	}

	data, err := db.Update(table, filter, string(body))
	if err != nil {
		response.Error(w, "Update failed", 500)
		return
	}
	response.JSON(w, data)
}

func handleDelete(w http.ResponseWriter, r *http.Request, table string) {
	filter := db.BuildFilter(r)
	if filter == "" {
		response.Error(w, "Missing filter (id or json.xxx)", 400)
		return
	}

	data, err := db.Delete(table, filter)
	if err != nil {
		response.Error(w, "Delete failed", 500)
		return
	}
	response.JSON(w, data)
}
