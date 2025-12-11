package db

import (
	"io"
	"net/http"
	"strings"
)

// ============ 配置（修改这里） ============

const (
	SUPABASE_URL = "https://arwnlqnzofqqxjnlvgqm.supabase.co"
	SUPABASE_KEY = "sb_publishable_eGii7SaCtXirDh2O5suItQ_L_eZtxuJ"
)

// 允许操作的表
var AllowedTables = map[string]bool{
	"user":      true,
	"community": true,
}

// ============ CRUD 方法 ============

func Select(table, query string) ([]byte, error) {
	endpoint := SUPABASE_URL + "/rest/v1/" + table + "?" + query
	return request("GET", endpoint, "")
}

func Insert(table, body string) ([]byte, error) {
	endpoint := SUPABASE_URL + "/rest/v1/" + table
	return request("POST", endpoint, body)
}

func Update(table, filter, body string) ([]byte, error) {
	endpoint := SUPABASE_URL + "/rest/v1/" + table + "?" + filter
	return request("PATCH", endpoint, body)
}

func Delete(table, filter string) ([]byte, error) {
	endpoint := SUPABASE_URL + "/rest/v1/" + table + "?" + filter
	return request("DELETE", endpoint, "")
}

// ============ 内部方法 ============

func request(method, endpoint, body string) ([]byte, error) {
	var reqBody io.Reader
	if body != "" {
		reqBody = strings.NewReader(body)
	}

	req, _ := http.NewRequest(method, endpoint, reqBody)
	req.Header.Set("apikey", SUPABASE_KEY)
	req.Header.Set("Authorization", "Bearer "+SUPABASE_KEY)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
