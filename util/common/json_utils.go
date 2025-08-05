package common

import (
	"strings"
)

// SearchKey 在嵌套的map/slice结构中搜索指定的键
func SearchKey(data any, key string) (any, bool) {
	switch val := data.(type) {
	case map[string]any:
		for k, v := range val {
			if k == key {
				return v, true
			}
			if result, ok := SearchKey(v, key); ok {
				return result, true
			}
		}
	case []any:
		for _, v := range val {
			if result, ok := SearchKey(v, key); ok {
				return result, true
			}
		}
	}
	return nil, false
}

// SearchHost 在headers中搜索Host头信息
func SearchHost(headers any) string {
	data, _ := headers.(map[string]any)
	for k, v := range data {
		if strings.EqualFold(k, "host") {
			switch v.(type) {
			case []any:
				hosts, _ := v.([]any)
				if len(hosts) > 0 {
					return hosts[0].(string)
				} else {
					return ""
				}
			case any:
				return v.(string)
			}
		}
	}
	return ""
}
