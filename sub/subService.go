package sub

import (
	"fmt"

	"x-ui/database"
	"x-ui/database/model"
	"x-ui/logger"
	"x-ui/util/common"
	"x-ui/util/proxy"
	"x-ui/web/service"
	"x-ui/xray"

	"github.com/goccy/go-json"
)

type SubService struct {
	address        string
	showInfo       bool
	remarkModel    string
	datepicker     string
	inboundService service.InboundService
	settingService service.SettingService
	linkGenerator  *proxy.LinkGenerator
}

func NewSubService(showInfo bool, remarkModel string) *SubService {
	s := &SubService{
		showInfo:    showInfo,
		remarkModel: remarkModel,
	}

	// 初始化链接生成器
	s.initLinkGenerator()

	return s
}

func (s *SubService) initLinkGenerator() {
	config := &proxy.LinkGeneratorConfig{
		Address:     s.address,
		Port:        0, // 使用inbound的端口
		RemarkModel: s.remarkModel,
		ShowInfo:    s.showInfo,
	}
	s.linkGenerator = proxy.NewLinkGenerator(config)
}

func (s *SubService) updateLinkGeneratorConfig() {
	if s.linkGenerator != nil {
		s.linkGenerator = proxy.NewLinkGenerator(&proxy.LinkGeneratorConfig{
			Address:     s.address,
			Port:        0,
			RemarkModel: s.remarkModel,
			ShowInfo:    s.showInfo,
		})
	}
}

func (s *SubService) GetSubs(subId string, host string) ([]string, string, error) {
	s.address = host

	// 更新链接生成器配置
	s.updateLinkGeneratorConfig()

	var result []string
	var header string
	var traffic xray.ClientTraffic
	var clientTraffics []xray.ClientTraffic
	inbounds, err := s.getInboundsBySubId(subId)
	if err != nil {
		return nil, "", err
	}

	if len(inbounds) == 0 {
		return nil, "", common.NewError("No inbounds found with ", subId)
	}

	s.datepicker, err = s.settingService.GetDatepicker()
	if err != nil {
		s.datepicker = "gregorian"
	}
	for _, inbound := range inbounds {
		clients, err := s.inboundService.GetClients(inbound)
		if err != nil {
			logger.Error("SubService - GetClients: Unable to get clients from inbound")
		}
		if clients == nil {
			continue
		}
		if len(inbound.Listen) > 0 && inbound.Listen[0] == '@' {
			listen, port, streamSettings, err := s.getFallbackMaster(inbound.Listen, inbound.StreamSettings)
			if err == nil {
				inbound.Listen = listen
				inbound.Port = port
				inbound.StreamSettings = streamSettings
			}
		}
		for _, client := range clients {
			if client.Enable && client.SubID == subId {
				link := s.getLink(inbound, client.Email)
				result = append(result, link)
				clientTraffics = append(clientTraffics, s.getClientTraffics(inbound.ClientStats, client.Email))
			}
		}
	}

	// Prepare statistics
	for index, clientTraffic := range clientTraffics {
		if index == 0 {
			traffic.Up = clientTraffic.Up
			traffic.Down = clientTraffic.Down
			traffic.Total = clientTraffic.Total
			if clientTraffic.ExpiryTime > 0 {
				traffic.ExpiryTime = clientTraffic.ExpiryTime
			}
		} else {
			traffic.Up += clientTraffic.Up
			traffic.Down += clientTraffic.Down
			if traffic.Total == 0 || clientTraffic.Total == 0 {
				traffic.Total = 0
			} else {
				traffic.Total += clientTraffic.Total
			}
			if clientTraffic.ExpiryTime != traffic.ExpiryTime {
				traffic.ExpiryTime = 0
			}
		}
	}
	header = fmt.Sprintf("upload=%d; download=%d; total=%d; expire=%d", traffic.Up, traffic.Down, traffic.Total, traffic.ExpiryTime/1000)
	return result, header, nil
}

func (s *SubService) getInboundsBySubId(subId string) ([]*model.Inbound, error) {
	db := database.GetDB()
	var inbounds []*model.Inbound
	err := db.Model(model.Inbound{}).Preload("ClientStats").Where(`id in (
		SELECT DISTINCT inbounds.id
		FROM inbounds,
			JSON_EACH(JSON_EXTRACT(inbounds.settings, '$.clients')) AS client
		WHERE
			protocol in ('vmess','vless','trojan','shadowsocks')
			AND JSON_EXTRACT(client.value, '$.subId') = ? AND enable = ?
	)`, subId, true).Find(&inbounds).Error
	if err != nil {
		return nil, err
	}
	return inbounds, nil
}

func (s *SubService) getLink(inbound *model.Inbound, email string) string {
	// 更新链接生成器的外部代理配置
	s.updateExternalProxies(inbound)

	// 获取客户端信息
	clients, err := s.inboundService.GetClients(inbound)
	if err != nil {
		return ""
	}

	switch inbound.Protocol {
	case "vmess":
		return s.linkGenerator.GenerateVmessLink(inbound, email, clients)
	case "vless":
		return s.linkGenerator.GenerateVlessLink(inbound, email, clients)
	case "trojan":
		return s.linkGenerator.GenerateTrojanLink(inbound, email, clients)
	case "shadowsocks":
		return s.linkGenerator.GenerateShadowsocksLink(inbound, email, clients)
	}
	return ""
}

func (s *SubService) updateExternalProxies(inbound *model.Inbound) {
	var stream map[string]any
	if err := json.Unmarshal([]byte(inbound.StreamSettings), &stream); err != nil {
		return
	}
	externalProxies, _ := stream["externalProxy"].([]any)

	var proxies []map[string]any
	for _, proxy := range externalProxies {
		if p, ok := proxy.(map[string]any); ok {
			proxies = append(proxies, p)
		}
	}

	// 重新创建链接生成器以更新外部代理配置
	s.linkGenerator = proxy.NewLinkGenerator(&proxy.LinkGeneratorConfig{
		Address:         s.address,
		Port:            0,
		RemarkModel:     s.remarkModel,
		ShowInfo:        s.showInfo,
		ExternalProxies: proxies,
	})
}

// 保留genRemark方法，但现在主要使用proxy包的GenerateRemark
func (s *SubService) genRemark(inbound *model.Inbound, email string, extra string) string {
	return s.linkGenerator.GenerateRemark(inbound, email, extra)
}

func (s *SubService) getFallbackMaster(dest string, streamSettings string) (string, int, string, error) {
	db := database.GetDB()
	var inbound *model.Inbound
	err := db.Model(model.Inbound{}).
		Where("JSON_TYPE(settings, '$.fallbacks') = 'array'").
		Where("EXISTS (SELECT * FROM json_each(settings, '$.fallbacks') WHERE json_extract(value, '$.dest') = ?)", dest).
		Find(&inbound).Error
	if err != nil {
		return "", 0, "", err
	}

	var stream map[string]any
	if err := json.Unmarshal([]byte(streamSettings), &stream); err != nil {
		return "", 0, "", err
	}
	var masterStream map[string]any
	if err := json.Unmarshal([]byte(inbound.StreamSettings), &masterStream); err != nil {
		return "", 0, "", err
	}
	stream["security"] = masterStream["security"]
	stream["tlsSettings"] = masterStream["tlsSettings"]
	stream["externalProxy"] = masterStream["externalProxy"]
	modifiedStream, _ := json.MarshalIndent(stream, "", "  ")

	return inbound.Listen, inbound.Port, string(modifiedStream), nil
}

func (s *SubService) getClientTraffics(traffics []xray.ClientTraffic, email string) xray.ClientTraffic {
	for _, traffic := range traffics {
		if traffic.Email == email {
			return traffic
		}
	}
	return xray.ClientTraffic{}
}
