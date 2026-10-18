package sub

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	yaml "github.com/goccy/go-yaml"

	"github.com/agassiz/3x-ui/v3/internal/web/service"
)

var acl4ssrGroupNames = []string{
	"🔰 节点选择", "♻️ 自动选择", "🤖GPT", "🤖Gemini", "🎥 NETFLIX", "⛔️ 广告拦截", "🚫 运营劫持", "🌍 国外媒体",
	"🌏 国内媒体", "Ⓜ️ 微软服务", "📲 电报信息", "🍎 苹果服务", "🎯 全球直连", "🛑 全球拦截", "🐟 漏网之鱼",
}

func readACL4SSRTemplate(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "deploy", "clash", "acl4ssr-routing.yaml"))
	if err != nil {
		t.Fatalf("read template: %v", err)
	}
	return string(raw)
}

// A misspelled group or rule-provider name makes Clash reject the whole profile,
// so the shipped template must survive the generator and its route-graph check.
func TestACL4SSRRoutingTemplateProducesValidRouteGraph(t *testing.T) {
	seedSubDB(t)
	seedSubInbound(t, "s-acl", "acl", 4431, 1, `{"network":"tcp","security":"none"}`)

	out, _, err := NewSubClashService(true, readACL4SSRTemplate(t), NewSubService(service.DefaultRemarkTemplate)).GetClash("s-acl", "req.example.com")
	if err != nil {
		t.Fatalf("GetClash: %v", err)
	}
	var config map[string]any
	if err := yaml.Unmarshal([]byte(out), &config); err != nil {
		t.Fatalf("generated yaml does not parse: %v\n%s", err, out)
	}
	if err := validateClashRouteGraph(config); err != nil {
		t.Fatalf("route graph invalid: %v", err)
	}

	groups, _ := asAnySlice(config["proxy-groups"])
	names := make([]string, 0, len(groups))
	for _, g := range groups {
		names = append(names, clashProxyGroupName(g))
	}
	if !reflect.DeepEqual(names, acl4ssrGroupNames) {
		t.Fatalf("proxy-groups = %v, want %v", names, acl4ssrGroupNames)
	}
	rules, _ := asAnySlice(config["rules"])
	if len(rules) == 0 || rules[len(rules)-1] != "MATCH,🐟 漏网之鱼" {
		t.Fatalf("last rule = %v, want MATCH,🐟 漏网之鱼", rules[len(rules)-1:])
	}
	if config["port"] != uint64(7890) {
		t.Fatalf("port = %#v, want the template's 7890", config["port"])
	}
}

// Served from a raw URL instead of pasted, the template goes through the remote
// merge, which keeps only groups, providers and rules and re-validates them.
func TestACL4SSRRoutingTemplateMergesAsRemoteSource(t *testing.T) {
	var remote map[string]any
	if err := yaml.Unmarshal([]byte(readACL4SSRTemplate(t)), &remote); err != nil {
		t.Fatalf("template does not parse: %v", err)
	}
	base := map[string]any{
		"proxies":      []any{map[string]any{"name": "node-a", "type": "vless", "server": "203.0.113.5", "port": 443}},
		"proxy-groups": []any{map[string]any{"name": "PROXY", "type": "select", "proxies": []any{"node-a", "DIRECT"}}},
		"rules":        []any{"MATCH,PROXY"},
	}
	if err := mergeRemoteClashRules(base, remote); err != nil {
		t.Fatalf("remote merge rejected the template: %v", err)
	}
}
