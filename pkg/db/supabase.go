package db

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// ============ 配置（修改这里） ============

const (
	SUPABASE_URL = "https://fyhpixiizagsprclogct.supabase.co"
	SUPABASE_KEY = "sb_secret_RqEBxd7vHS1Tkwk4rJu67A_x5GiAhro" // 服务端密钥
)


// Debug 开关，生产环境设为 false
var Debug = true

// ============ CRUD 方法 ============

func Select(table, query string) ([]byte, error) {
	endpoint := SUPABASE_URL + "/rest/v1/" + table + "?" + query
	debugLog("SELECT", endpoint, "")
	return request("GET", endpoint, "")
}

func Insert(table, body string) ([]byte, error) {
	endpoint := SUPABASE_URL + "/rest/v1/" + table
	debugLog("INSERT", endpoint, body)
	return request("POST", endpoint, body)
}

func Update(table, filter, body string) ([]byte, error) {
	endpoint := SUPABASE_URL + "/rest/v1/" + table + "?" + filter
	debugLog("UPDATE", endpoint, body)
	return request("PATCH", endpoint, body)
}

func Delete(table, filter string) ([]byte, error) {
	endpoint := SUPABASE_URL + "/rest/v1/" + table + "?" + filter
	debugLog("DELETE", endpoint, "")
	return request("DELETE", endpoint, "")
}

// 系统保留参数，不作为查询条件
var reservedParams = map[string]bool{
	"action": true,
	"table":  true,
}

// BuildFilter 构建过滤条件
// 规则:
//   - id 参数查主键: ?id=1 -> id=eq.1
//   - 其他参数查 jsonb 字段: ?user_id=123 -> json->>user_id=eq.123
/*
	操作符	含义	REST API 查询示例
	gt	大于	age=gt.20
	gte	大于等于	age=gte.18
	lt	小于	age=lt.30
	lte	小于等于	age=lte.60
	eq	等于	id=eq.10
	neq	不等于	status=neq.inactive
	like	模糊匹配	name=like.%John%
	ilike	不区分大小写的模糊匹配	name=ilike.%john%
*/
func BuildFilter(r *http.Request) string {
	var filters []string

	debugLog("BuildFilter", "URL RawQuery", r.URL.RawQuery)
	// 遍历所有 Query 参数 （包括/:action/:table 这个是vercel的特殊处理）
	for key, values := range r.URL.Query() {
		debugLog("BuildFilter", "Param", fmt.Sprintf("key=%s, values=%v", key, values))
		if len(values) == 0 {
			continue
		}
		// 跳过系统保留参数
		if reservedParams[key] {
			continue
		}
		value := values[0]

		// 构建过滤条件（id 直接查询，其他查 json 字段）
		filter := buildFilter(key, value)
		debugLog("BuildFilter", key+" filter", filter)
		filters = append(filters, filter)
	}

	result := strings.Join(filters, "&")
	debugLog("BuildFilter", "Final filter", result)
	return result
}

// 操作符列表（按顺序排列）
var operators = []struct {
	symbol string
	op     string
}{
	{"like:", "like"},  // 模糊匹配（必须在最前面）
	{">=", "gte"},      // 大于等于
	{"<=", "lte"},      // 小于等于
	{"!=", "neq"},      // 不等于
	{">", "gt"},        // 大于
	{"<", "lt"},        // 小于
}

// 数据库原生字段（不在 json 中）
var nativeFields = map[string]bool{
	"id": true,
}

// buildFilter 构建字段过滤条件
// id 字段直接查询，其他字段查询 json 内部
func buildFilter(key, value string) string {
	// 去掉可选的 json. 前缀
	key = strings.TrimPrefix(key, "json.")

	// 判断是否为原生字段
	isNative := nativeFields[key]

	// 按顺序检查操作符
	for _, item := range operators {
		if strings.HasPrefix(value, item.symbol) {
			actualValue := strings.TrimPrefix(value, item.symbol)
			if isNative {
				// 原生字段直接查询
				return key + "=" + item.op + "." + url.QueryEscape(actualValue)
			}
			// json 字段
			return "json->>" + key + "=" + item.op + "." + url.QueryEscape(actualValue)
		}
	}

	// 默认 eq 操作符
	if isNative {
		return key + "=eq." + url.QueryEscape(value)
	}
	return "json->>" + key + "=eq." + url.QueryEscape(value)
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
		debugLog("Request", "Error", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	debugLog("Request", "Response Status", fmt.Sprintf("%d", resp.StatusCode))
	debugLog("Request", "Response Body", string(data))
	return data, nil
}

// debugLog 调试日志
func debugLog(action, key, value string) {
	if Debug {
		fmt.Printf("[DEBUG] %s | %s: %s\n", action, key, value)
	}
}


// MergeUpdate 合并更新：先查询原数据，再合并新数据
// 只更新传递的字段，保留原有字段
func MergeUpdate(table, filter, newBody string) (string, error) {
	// 1. 先查询原数据
	query := "select=json&" + filter
	data, err := Select(table, query)
	if err != nil {
		return "", fmt.Errorf("Failed to fetch original data: %v", err)
	}

	// 2. 解析原数据
	var results []map[string]interface{}
	if err := json.Unmarshal(data, &results); err != nil {
		return "", fmt.Errorf("Failed to parse original data: %v", err)
	}

	if len(results) == 0 {
		return "", fmt.Errorf("Record not found")
	}

	// 获取原 json 字段
	originalJson, ok := results[0]["json"].(map[string]interface{})
	if !ok {
		originalJson = make(map[string]interface{})
	}

	// 3. 解析新数据
	var newData map[string]interface{}
	if err := json.Unmarshal([]byte(newBody), &newData); err != nil {
		return "", fmt.Errorf("Failed to parse new data: %v", err)
	}

	// 4. 合并：新数据覆盖原数据中的对应字段
	for key, value := range newData {
		originalJson[key] = value
	}

	// 5. 构建更新 body
	mergedJson, err := json.Marshal(originalJson)
	if err != nil {
		return "", fmt.Errorf("Failed to marshal merged data: %v", err)
	}

	return `{"json":` + string(mergedJson) + `}`, nil
}
