package handler

import (
	"io"
	"net/http"

	"github.com/jwx233s/a-service/pkg/db"
	"github.com/jwx233s/a-service/pkg/response"
)

// POST /api/db/update?table=user&id=1
// POST /api/db/update?table=user&user_id=123
// Body: {"json": {"name": "Jerry"}}
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

	filter := db.BuildFilter(r)
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
