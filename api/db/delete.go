package handler

import (
	"net/http"

	"github.com/jwx233s/a-service/pkg/db"
	"github.com/jwx233s/a-service/pkg/response"
)

// POST /api/db/delete?table=user&id=1
// POST /api/db/delete?table=user&user_id=123
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

	data, err := db.Delete(table, filter)
	if err != nil {
		response.Error(w, "Delete failed", 500)
		return
	}
	response.JSON(w, data)
}
