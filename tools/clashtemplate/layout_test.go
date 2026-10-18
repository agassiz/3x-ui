package main

import (
	"strings"
	"testing"
)

const converterYAML = `
proxies:
  - {name: probe-node, server: example.com, port: 443, type: vless}
proxy-groups:
  - name: 🔰 节点选择
    type: select
    proxies:
      - ♻️ 自动选择
      - 🎯 全球直连
      - probe-node
  - name: ♻️ 自动选择
    type: url-test
    url: http://www.gstatic.com/generate_204
    interval: 300
    proxies:
      - probe-node
  - name: 🎯 全球直连
    type: select
    proxies:
      - DIRECT
rules:
  - IP-CIDR,10.0.0.0/8,🎯 全球直连,no-resolve
  - DOMAIN-SUFFIX,local,🎯 全球直连
  - DOMAIN-SUFFIX,google.com,🔰 节点选择
  - GEOIP,CN,🎯 全球直连
  - MATCH,🔰 节点选择
`

const templateYAML = `
port: 7890
proxy-groups:
  - name: 🔰 节点选择
    type: select
    proxies:
      - ♻️ 自动选择
      - 🎯 全球直连
    include-all-proxies: true
  - name: ♻️ 自动选择
    type: url-test
    url: http://www.gstatic.com/generate_204
    interval: 300
    include-all-proxies: true
  - name: 🎯 全球直连
    type: select
    proxies:
      - DIRECT
rule-providers:
  LocalAreaNetwork: {type: http, behavior: classical}
rules:
  - RULE-SET,LocalAreaNetwork,🎯 全球直连
  - RULE-SET,ProxyGFWlist,🔰 节点选择
  - GEOIP,CN,🎯 全球直连
  - MATCH,🔰 节点选择
`

func layoutsFor(t *testing.T, converter, template string) []string {
	t.Helper()
	want, err := converterLayout([]byte(converter), "probe-node")
	if err != nil {
		t.Fatalf("converterLayout: %v", err)
	}
	got, err := templateLayout([]byte(template))
	if err != nil {
		t.Fatalf("templateLayout: %v", err)
	}
	return diffLayouts(want, got)
}

// The converter lists the probe node by name where the template says
// include-all-proxies; treating those as different would report drift forever.
func TestDiffLayoutsMatchesNodeListingWithIncludeAll(t *testing.T) {
	if diffs := layoutsFor(t, converterYAML, templateYAML); len(diffs) != 0 {
		t.Fatalf("diffs = %q, want none", diffs)
	}
}

func TestDiffLayoutsReportsDrift(t *testing.T) {
	cases := []struct {
		name      string
		converter string
		want      string
	}{
		{
			"new group upstream",
			strings.Replace(converterYAML, "  - name: 🎯 全球直连\n", "  - name: 🎮 游戏平台\n    type: select\n    proxies:\n      - 🔰 节点选择\n  - name: 🎯 全球直连\n", 1),
			"🎮 游戏平台",
		},
		{
			"member removed",
			strings.Replace(converterYAML, "      - ♻️ 自动选择\n      - 🎯 全球直连\n      - probe-node\n", "      - ♻️ 自动选择\n      - probe-node\n", 1),
			"🔰 节点选择",
		},
		{
			"group stops listing nodes",
			strings.Replace(converterYAML, "    interval: 300\n    proxies:\n      - probe-node\n", "    interval: 300\n    proxies:\n      - DIRECT\n", 1),
			"♻️ 自动选择",
		},
		{
			"group type changed",
			strings.Replace(converterYAML, "type: url-test", "type: fallback", 1),
			"fallback",
		},
		{
			"rule order changed",
			strings.Replace(converterYAML, "  - GEOIP,CN,🎯 全球直连\n", "  - DOMAIN,ads.example.com,🛑 全球拦截\n  - GEOIP,CN,🎯 全球直连\n", 1),
			"🛑 全球拦截",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			diffs := layoutsFor(t, c.converter, templateYAML)
			if !strings.Contains(strings.Join(diffs, "\n"), c.want) {
				t.Fatalf("diffs = %q, want one mentioning %q", diffs, c.want)
			}
		})
	}
}
