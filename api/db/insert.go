package handler

import (
	"io"
	"net/http"

	"github.com/jwx233s/a-service/pkg/db"
	"github.com/jwx233s/a-service/pkg/response"
)

// POST /api/db/insert?table=user
// Body: {"user_id": "123", "json": {"name": "Tom"}}
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
