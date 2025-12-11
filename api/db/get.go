package handler

import (
	"net/http"

	"github.com/jwx233s/a-service/pkg/db"
	"github.com/jwx233s/a-service/pkg/response"
)

// GET /api/db/get?table=user
// GET /api/db/get?table=user&id=1
// GET /api/db/get?table=user&user_id=123
func Handler(w http.ResponseWriter, r *http.Request) {
	response.SetHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	table := r.URL.Query().Get("table")
	if table == "" {
		response.Error(w, "Missing 'table'", 400)
		return
	}
	if !db.AllowedTables[table] {
		response.Error(w, "Table not allowed", 403)
		return
	}

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
