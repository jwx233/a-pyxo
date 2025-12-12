package db

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
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
//   - 支持操作符: ?age=gt.18 -> json->>age=gt.18
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

		if key == "id" {
			// id 查主键
			filter := "id=eq." + value
			debugLog("BuildFilter", "id filter", filter)
			filters = append(filters, filter)
		} else {
			// 其他参数都查 jsonb 字段
			filter := buildJsonFilter(key, value)
			debugLog("BuildFilter", key+" filter", filter)
			filters = append(filters, filter)
		}
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

// buildJsonFilter 构建 jsonb 字段过滤条件
// 支持: ?age=>18  ?age=<60  ?age=>=18  ?age=<=60  ?name=!=Tom  ?name=like:%Tom%
func buildJsonFilter(key, value string) string {
	// 去掉可选的 json. 前缀
	key = strings.TrimPrefix(key, "json.")

	// 按顺序检查操作符
	for _, item := range operators {
		// 自定义标识
		symbol := item.symbol
		// 系统操作标识
		op := item.op
		
		// 匹配：输入的参数内容以自定义标识开头
		if strings.HasPrefix(value, symbol){
			// 去掉操作符符号
    		actualValue := strings.TrimPrefix(value, o.symbol)
			return "json->>" + key + "=" + op + "." + url.QueryEscape(actualValue)
		}
	}

	// 默认 eq 操作符
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
