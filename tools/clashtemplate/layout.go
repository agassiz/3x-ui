package main

import (
	"fmt"
	"slices"
	"strings"

	yaml "github.com/goccy/go-yaml"
)

// groupShape is a proxy group with its node list reduced to a flag, the one
// part a template written for unknown node names can express.
type groupShape struct {
	Name          string
	Type          string
	Members       []string
	IncludesNodes bool
}

type layout struct {
	Groups      []groupShape
	RuleTargets []string
}

type rawGroup struct {
	Name              string   `yaml:"name"`
	Type              string   `yaml:"type"`
	Proxies           []string `yaml:"proxies"`
	IncludeAllProxies bool     `yaml:"include-all-proxies"`
}

type rawConfig struct {
	ProxyGroups []rawGroup `yaml:"proxy-groups"`
	Rules       []string   `yaml:"rules"`
}

func parseConfig(raw []byte) (rawConfig, error) {
	var cfg rawConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return rawConfig{}, err
	}
	if len(cfg.ProxyGroups) == 0 || len(cfg.Rules) == 0 {
		return rawConfig{}, fmt.Errorf("no proxy-groups or rules found")
	}
	return cfg, nil
}

// converterLayout reads the converter output, where nodes appear by name.
func converterLayout(raw []byte, probeName string) (layout, error) {
	cfg, err := parseConfig(raw)
	if err != nil {
		return layout{}, err
	}
	out := layout{RuleTargets: ruleTargets(cfg.Rules)}
	for _, g := range cfg.ProxyGroups {
		shape := groupShape{Name: g.Name, Type: g.Type, Members: []string{}}
		for _, p := range g.Proxies {
			if p == probeName {
				shape.IncludesNodes = true
				continue
			}
			shape.Members = append(shape.Members, p)
		}
		out.Groups = append(out.Groups, shape)
	}
	return out, nil
}

// templateLayout reads the template, where include-all-proxies stands for the nodes.
func templateLayout(raw []byte) (layout, error) {
	cfg, err := parseConfig(raw)
	if err != nil {
		return layout{}, err
	}
	out := layout{RuleTargets: ruleTargets(cfg.Rules)}
	for _, g := range cfg.ProxyGroups {
		members := g.Proxies
		if members == nil {
			members = []string{}
		}
		out.Groups = append(out.Groups, groupShape{Name: g.Name, Type: g.Type, Members: members, IncludesNodes: g.IncludeAllProxies})
	}
	return out, nil
}

// ruleTargets lists the group each rule routes to with consecutive repeats
// collapsed, so inline rules and the RULE-SETs replacing them compare equal.
func ruleTargets(rules []string) []string {
	var out []string
	for _, rule := range rules {
		parts := strings.Split(rule, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		end := len(parts) - 1
		for end > 1 && (strings.EqualFold(parts[end], "no-resolve") || strings.EqualFold(parts[end], "src")) {
			end--
		}
		target := parts[end]
		if strings.EqualFold(parts[0], "MATCH") && len(parts) > 1 {
			target = parts[1]
		}
		if len(out) == 0 || out[len(out)-1] != target {
			out = append(out, target)
		}
	}
	return out
}

// diffLayouts describes every way the template (got) differs from the converter (want).
func diffLayouts(want, got layout) []string {
	var diffs []string
	wantNames, gotNames := groupNames(want.Groups), groupNames(got.Groups)
	if !slices.Equal(wantNames, gotNames) {
		diffs = append(diffs, fmt.Sprintf("分组名称或顺序不同\n  datapipe: %s\n  模板:     %s", strings.Join(wantNames, " | "), strings.Join(gotNames, " | ")))
	}
	gotByName := make(map[string]groupShape, len(got.Groups))
	for _, g := range got.Groups {
		gotByName[g.Name] = g
	}
	for _, w := range want.Groups {
		g, ok := gotByName[w.Name]
		if !ok {
			diffs = append(diffs, fmt.Sprintf("模板缺少分组 %s（类型 %s，成员 %v，收录节点 %v）", w.Name, w.Type, w.Members, w.IncludesNodes))
			continue
		}
		if w.Type != g.Type {
			diffs = append(diffs, fmt.Sprintf("分组 %s 类型不同：datapipe=%s 模板=%s", w.Name, w.Type, g.Type))
		}
		if !slices.Equal(w.Members, g.Members) {
			diffs = append(diffs, fmt.Sprintf("分组 %s 成员不同：datapipe=%v 模板=%v", w.Name, w.Members, g.Members))
		}
		if w.IncludesNodes != g.IncludesNodes {
			diffs = append(diffs, fmt.Sprintf("分组 %s 是否收录节点不同：datapipe=%v 模板=%v", w.Name, w.IncludesNodes, g.IncludesNodes))
		}
	}
	if !slices.Equal(want.RuleTargets, got.RuleTargets) {
		diffs = append(diffs, fmt.Sprintf("规则分段顺序不同\n  datapipe: %s\n  模板:     %s", strings.Join(want.RuleTargets, " → "), strings.Join(got.RuleTargets, " → ")))
	}
	return diffs
}

func groupNames(groups []groupShape) []string {
	names := make([]string, 0, len(groups))
	for _, g := range groups {
		names = append(names, g.Name)
	}
	return names
}
