package proxy

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"time"

	"x-ui/database/model"
	"x-ui/xray"

	"github.com/goccy/go-json"
)

// LinkGeneratorConfig 链接生成器配置
type LinkGeneratorConfig struct {
	Address         string           // 服务器地址
	Port            int              // 端口号（0表示使用inbound的端口）
	RemarkModel     string           // 备注模式
	ShowInfo        bool             // 是否显示流量信息
	ExternalProxies []map[string]any // 外部代理配置
}

// LinkGenerator 通用链接生成器
type LinkGenerator struct {
	config *LinkGeneratorConfig
}

// NewLinkGenerator 创建新的链接生成器
func NewLinkGenerator(config *LinkGeneratorConfig) *LinkGenerator {
	return &LinkGenerator{
		config: config,
	}
}

// GenerateVmessLink 生成VMESS链接
func (g *LinkGenerator) GenerateVmessLink(inbound *model.Inbound, email string, clients []model.Client) string {
	if inbound.Protocol != "vmess" {
		return ""
	}

	port := g.getPort(inbound)
	obj := map[string]any{
		"v":    "2",
		"add":  g.config.Address,
		"port": port,
		"type": "none",
	}

	var stream map[string]any
	if err := json.Unmarshal([]byte(inbound.StreamSettings), &stream); err != nil {
		return ""
	}
	network, ok := stream["network"].(string)
	if !ok {
		network = "tcp" // 默认值
	}
	obj["net"] = network

	// 处理网络类型配置
	g.processNetworkSettings(obj, stream, network)

	// 处理安全设置
	g.processSecuritySettings(obj, stream)

	// 查找客户端信息
	clientIndex := -1
	for i, client := range clients {
		if client.Email == email {
			clientIndex = i
			break
		}
	}
	if clientIndex == -1 {
		return ""
	}

	client := clients[clientIndex]
	obj["id"] = client.ID
	obj["aid"] = 0
	obj["scy"] = client.Security

	// 处理外部代理
	if len(g.config.ExternalProxies) > 0 {
		return g.processExternalProxiesVmess(obj, inbound, email)
	}

	obj["ps"] = g.GenerateRemark(inbound, email, "")
	jsonStr, _ := json.MarshalIndent(obj, "", "  ")
	return "vmess://" + base64.StdEncoding.EncodeToString(jsonStr)
}

// GenerateVlessLink 生成VLESS链接
func (g *LinkGenerator) GenerateVlessLink(inbound *model.Inbound, email string, clients []model.Client) string {
	if inbound.Protocol != "vless" {
		return ""
	}

	var stream map[string]any
	if err := json.Unmarshal([]byte(inbound.StreamSettings), &stream); err != nil {
		return ""
	}

	clientIndex := -1
	for i, client := range clients {
		if client.Email == email {
			clientIndex = i
			break
		}
	}
	if clientIndex == -1 {
		return ""
	}

	uuid := clients[clientIndex].ID
	port := g.getPort(inbound)
	streamNetwork, ok := stream["network"].(string)
	if !ok {
		streamNetwork = "tcp" // 默认值
	}
	params := make(map[string]string)
	params["type"] = streamNetwork

	// 处理网络类型参数
	g.processNetworkParams(params, stream, streamNetwork)

	// 处理安全参数
	security := g.processSecurityParams(params, stream, clients[clientIndex], streamNetwork)

	// 处理外部代理
	if len(g.config.ExternalProxies) > 0 {
		return g.processExternalProxiesVless(uuid, params, security, inbound, email)
	}

	link := fmt.Sprintf("vless://%s@%s:%d", uuid, g.config.Address, port)
	url, _ := url.Parse(link)
	q := url.Query()

	for k, v := range params {
		q.Add(k, v)
	}

	url.RawQuery = q.Encode()
	url.Fragment = g.GenerateRemark(inbound, email, "")
	return url.String()
}

// GenerateTrojanLink 生成Trojan链接
func (g *LinkGenerator) GenerateTrojanLink(inbound *model.Inbound, email string, clients []model.Client) string {
	if inbound.Protocol != "trojan" {
		return ""
	}

	var stream map[string]any
	if err := json.Unmarshal([]byte(inbound.StreamSettings), &stream); err != nil {
		return ""
	}

	clientIndex := -1
	for i, client := range clients {
		if client.Email == email {
			clientIndex = i
			break
		}
	}
	if clientIndex == -1 {
		return ""
	}

	password := clients[clientIndex].Password
	port := g.getPort(inbound)
	streamNetwork, ok := stream["network"].(string)
	if !ok {
		streamNetwork = "tcp" // 默认值
	}
	params := make(map[string]string)
	params["type"] = streamNetwork

	// 处理网络类型参数
	g.processNetworkParams(params, stream, streamNetwork)

	// 处理安全参数
	security := g.processSecurityParams(params, stream, clients[clientIndex], streamNetwork)

	// 处理外部代理
	if len(g.config.ExternalProxies) > 0 {
		return g.processExternalProxiesTrojan(password, params, security, inbound, email)
	}

	link := fmt.Sprintf("trojan://%s@%s:%d", password, g.config.Address, port)
	url, _ := url.Parse(link)
	q := url.Query()

	for k, v := range params {
		q.Add(k, v)
	}

	url.RawQuery = q.Encode()
	url.Fragment = g.GenerateRemark(inbound, email, "")
	return url.String()
}

// GenerateShadowsocksLink 生成Shadowsocks链接
func (g *LinkGenerator) GenerateShadowsocksLink(inbound *model.Inbound, email string, clients []model.Client) string {
	if inbound.Protocol != "shadowsocks" {
		return ""
	}

	var stream map[string]any
	if err := json.Unmarshal([]byte(inbound.StreamSettings), &stream); err != nil {
		return ""
	}

	var settings map[string]any
	if err := json.Unmarshal([]byte(inbound.Settings), &settings); err != nil {
		return ""
	}
	inboundPassword, ok := settings["password"].(string)
	if !ok {
		return ""
	}
	method, ok := settings["method"].(string)
	if !ok {
		return ""
	}

	clientIndex := -1
	for i, client := range clients {
		if client.Email == email {
			clientIndex = i
			break
		}
	}
	if clientIndex == -1 {
		return ""
	}

	port := g.getPort(inbound)
	streamNetwork := stream["network"].(string)
	params := make(map[string]string)
	params["type"] = streamNetwork

	// 处理网络类型参数
	g.processNetworkParams(params, stream, streamNetwork)

	// 处理安全参数
	security := g.processSecurityParamsSS(params, stream)

	// 构建加密部分
	var encPart string
	if strings.Contains(method, "2022") {
		encPart = fmt.Sprintf("%s:%s", method, inboundPassword)
	} else {
		clientPassword := clients[clientIndex].Password
		if clientPassword == "" {
			clientPassword = inboundPassword
		}
		encPart = fmt.Sprintf("%s:%s", method, clientPassword)
	}

	// 处理外部代理
	if len(g.config.ExternalProxies) > 0 {
		return g.processExternalProxiesSS(encPart, params, security, inbound, email)
	}

	link := fmt.Sprintf("ss://%s@%s:%d", base64.StdEncoding.EncodeToString([]byte(encPart)), g.config.Address, port)
	url, _ := url.Parse(link)
	q := url.Query()

	for k, v := range params {
		q.Add(k, v)
	}

	url.RawQuery = q.Encode()
	url.Fragment = g.GenerateRemark(inbound, email, "")
	return url.String()
}

// GenerateRemark 生成备注
func (g *LinkGenerator) GenerateRemark(inbound *model.Inbound, email string, extra string) string {
	remarkModel := g.config.RemarkModel
	if remarkModel == "" {
		remarkModel = "-ieo" // 默认备注模式
	}

	separationChar := string(remarkModel[0])
	orderChars := remarkModel[1:]
	orders := map[byte]string{
		'i': "",
		'e': "",
		'o': "",
	}

	if len(email) > 0 {
		orders['e'] = email
	}
	if len(inbound.Remark) > 0 {
		orders['i'] = inbound.Remark
	}
	if len(extra) > 0 {
		orders['o'] = extra
	}

	var remark []string
	for i := 0; i < len(orderChars); i++ {
		char := orderChars[i]
		order, exists := orders[char]
		if exists && order != "" {
			remark = append(remark, order)
		}
	}

	result := strings.Join(remark, separationChar)

	// 如果需要显示流量信息
	if g.config.ShowInfo {
		result = g.addTrafficInfo(result, inbound, email)
	}

	return result
}

// getPort 获取端口号
func (g *LinkGenerator) getPort(inbound *model.Inbound) int {
	if g.config.Port > 0 {
		return g.config.Port
	}
	return inbound.Port
}

// addTrafficInfo 添加流量信息到备注
func (g *LinkGenerator) addTrafficInfo(remark string, inbound *model.Inbound, email string) string {
	statsExist := false
	var stats xray.ClientTraffic
	for _, clientStat := range inbound.ClientStats {
		if clientStat.Email == email {
			stats = clientStat
			statsExist = true
			break
		}
	}

	if !statsExist {
		return remark
	}

	// 添加流量信息逻辑（简化版本）
	if stats.Total > 0 {
		remark += fmt.Sprintf(" | %s", formatBytes(stats.Total-stats.Up-stats.Down))
	}

	if stats.ExpiryTime > 0 {
		remark += fmt.Sprintf(" | %s", time.Unix(stats.ExpiryTime/1000, 0).Format("2006-01-02"))
	}

	return remark
}

// formatBytes 格式化字节数
func formatBytes(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%dB", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1fKB", float64(bytes)/1024)
	} else if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.1fMB", float64(bytes)/(1024*1024))
	} else {
		return fmt.Sprintf("%.1fGB", float64(bytes)/(1024*1024*1024))
	}
}
