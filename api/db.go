package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/jwx233s/a-service/pkg/db"
	"github.com/jwx233s/a-service/pkg/response"
)

// GET  /api/db?action=get&table=user
// GET  /api/db?action=get&table=user&id=1
// GET  /api/db?action=get&table=user&user_id=123
// POST /api/db?action=insert&table=user         Body: {"user_id":"123","json":{}}
// POST /api/db?action=update&table=user&id=1    Body: {"json":{}}
// POST /api/db?action=delete&table=user&id=1
func Handler(w http.ResponseWriter, r *http.Request) {
	response.SetHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	action := r.URL.Query().Get("action")
	table := r.URL.Query().Get("table")

	if action == "" {
		response.Error(w, "Missing 'action'", 400)
		return
	}
	if table == "" {
		response.Error(w, "Missing 'table'", 400)
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
		response.Error(w, "Invalid action", 400)
	}
}

func handleGet(w http.ResponseWriter, r *http.Request, table string) {
	query := "select=*"
	if id := r.URL.Query().Get("id"); id != "" {
		query += "&id=eq." + id
	}
	if userID := r.URL.Query().Get("user_id"); userID != "" {
		query += "&user_id=eq." + userID
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
	filter := buildFilter(r)
	if filter == "" {
		response.Error(w, "Missing 'id' or 'user_id'", 400)
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
	filter := buildFilter(r)
	if filter == "" {
		response.Error(w, "Missing 'id' or 'user_id'", 400)
		return
	}

	data, err := db.Delete(table, filter)
	if err != nil {
		response.Error(w, "Delete failed", 500)
		return
	}
	response.JSON(w, data)
}

func buildFilter(r *http.Request) string {
	var filters []string
	if id := r.URL.Query().Get("id"); id != "" {
		filters = append(filters, "id=eq."+id)
	}
	if userID := r.URL.Query().Get("user_id"); userID != "" {
		filters = append(filters, "user_id=eq."+userID)
	}
	return strings.Join(filters, "&")
}
