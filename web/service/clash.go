package service

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/agassiz/3x-ui/v2/database"
	"github.com/agassiz/3x-ui/v2/database/model"
	"github.com/agassiz/3x-ui/v2/logger"
	"github.com/agassiz/3x-ui/v2/util/common"
	"github.com/agassiz/3x-ui/v2/util/proxy"

	"gorm.io/gorm"
)

type ClashService struct {
	inboundService InboundService
	settingService SettingService
	linkGenerator  *proxy.LinkGenerator
	address        string
	remarkModel    string
	showInfo       bool
}

func NewClashService() *ClashService {
	service := &ClashService{}

	// 获取用户配置
	remarkModel, err := service.settingService.GetRemarkModel()
	if err != nil {
		remarkModel = "-ieo" // 默认备注模式
	}

	showInfo, err := service.settingService.GetSubShowInfo()
	if err != nil {
		showInfo = false // 默认不显示流量信息
	}

	service.remarkModel = remarkModel
	service.showInfo = showInfo

	// 初始化链接生成器
	service.initLinkGenerator()

	return service
}

// initLinkGenerator 初始化链接生成器（参考SubService的实现）
func (s *ClashService) initLinkGenerator() {
	config := &proxy.LinkGeneratorConfig{
		Address:     s.address,
		Port:        45556, // 使用隐藏端口（Clash特有）
		RemarkModel: s.remarkModel,
		ShowInfo:    s.showInfo,
	}
	s.linkGenerator = proxy.NewLinkGenerator(config)
}

// updateLinkGeneratorConfig 更新链接生成器配置（参考SubService的实现）
func (s *ClashService) updateLinkGeneratorConfig() {
	if s.linkGenerator != nil {
		s.linkGenerator = proxy.NewLinkGenerator(&proxy.LinkGeneratorConfig{
			Address:     s.address,
			Port:        45556, // 使用隐藏端口（Clash特有）
			RemarkModel: s.remarkModel,
			ShowInfo:    s.showInfo,
		})
	}
}

// GetClashSubscription 获取Clash订阅配置
func (s *ClashService) GetClashSubscription(email string, host string) (string, error) {
	logger.Infof("[clash] generating subscription email=%s host=%s", email, host)

	s.address = host
	s.updateLinkGeneratorConfig()

	clientUrl, realPort, err := s.getClientUrl(email, host)
	if err != nil {
		logger.Error("[clash] failed to resolve client link:", err)
		return "", err
	}
	if clientUrl == "" {
		return "", common.NewError("Client not found for email: " + email)
	}

	urlMd5 := s.calculateMD5(clientUrl)

	db := database.GetDB()
	var subscription model.ClashSubscription
	queryErr := db.Where("email = ?", email).First(&subscription).Error
	isNewRecord := false

	if queryErr == nil {
		if subscription.UrlMd5 == urlMd5 && subscription.YamlContent != "" {
			logger.Debug("[clash] cache hit", email)
			return subscription.YamlContent, nil
		}
	} else if queryErr != gorm.ErrRecordNotFound {
		logger.Error("[clash] cache lookup failed:", queryErr)
		return "", queryErr
	} else {
		isNewRecord = true
	}

	yamlContent, err := s.generateClashConfig(clientUrl, realPort)
	if err != nil {
		return "", err
	}

	now := time.Now().Unix()
	clashSubscription := model.ClashSubscription{
		Email:       email,
		UrlMd5:      urlMd5,
		YamlContent: yamlContent,
		UpdatedAt:   now,
	}

	if isNewRecord {
		clashSubscription.CreatedAt = now
		err = db.Create(&clashSubscription).Error
	} else {
		clashSubscription.Id = subscription.Id
		clashSubscription.CreatedAt = subscription.CreatedAt
		err = db.Save(&clashSubscription).Error
	}

	if err != nil {
		logger.Warning("[clash] failed to persist cache:", err)
	}

	logger.Debugf("[clash] generated subscription size=%d", len(yamlContent))
	return yamlContent, nil
}

// getClientUrl 根据email查找客户端并生成链接
func (s *ClashService) getClientUrl(email string, host string) (string, int, error) {
	db := database.GetDB()
	var inbound model.Inbound

	err := db.Raw(`
		SELECT * FROM inbounds
		WHERE enable = 1
		AND protocol IN ('vmess', 'vless', 'trojan', 'shadowsocks')
		AND JSON_EXTRACT(settings, '$.clients') IS NOT NULL
		AND EXISTS (
			SELECT 1 FROM JSON_EACH(JSON_EXTRACT(settings, '$.clients')) AS client
			WHERE JSON_EXTRACT(client.value, '$.email') = ?
		)
		LIMIT 1
	`, email).Take(&inbound).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", 0, nil
		}
		return "", 0, err
	}

	if len(inbound.Listen) > 0 && inbound.Listen[0] == '@' {
		listen, port, streamSettings, fbErr := s.getFallbackMaster(inbound.Listen, inbound.StreamSettings)
		if fbErr == nil {
			inbound.Listen = listen
			inbound.Port = port
			inbound.StreamSettings = streamSettings
		} else {
			logger.Warning("[clash] fallback lookup failed:", fbErr)
		}
	}

	clientLink := s.generateClientLink(&inbound, email)
	return clientLink, inbound.Port, nil
}

// generateClashConfig 调用外部API生成Clash配置，并替换真实端口
func (s *ClashService) generateClashConfig(clientUrl string, realPort int) (string, error) {
	// URL编码客户端链接
	encodedUrl := url.QueryEscape(clientUrl)
	apiUrl := fmt.Sprintf("https://sub.datapipe.top/sub?target=clash&url=%s&insert=false", encodedUrl)

	// 发起HTTP请求
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(apiUrl)
	if err != nil {
		return "", fmt.Errorf("failed to call external API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("external API returned status: %d", resp.StatusCode)
	}

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	yamlContent := string(body)
	logger.Debugf("[clash] external config received size=%d", len(yamlContent))

	// 将固定端口45556替换为真实端口
	beforeReplace := strings.Count(yamlContent, "45556")
	yamlContent = strings.ReplaceAll(yamlContent, "45556", fmt.Sprintf("%d", realPort))
	logger.Debugf("[clash] replaced hidden port occurrences=%d", beforeReplace)

	return yamlContent, nil
}

// calculateMD5 计算字符串的MD5值
func (s *ClashService) calculateMD5(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// generateClientLink 生成客户端链接（使用固定端口45556隐藏真实端口）
func (s *ClashService) generateClientLink(inbound *model.Inbound, email string) string {
	clients, err := s.inboundService.GetClients(inbound)
	if err != nil || len(clients) == 0 {
		return ""
	}

	activeClients := make([]model.Client, 0, len(clients))
	for _, client := range clients {
		if client.Enable {
			activeClients = append(activeClients, client)
		}
	}
	if len(activeClients) == 0 {
		return ""
	}

	switch inbound.Protocol {
	case "vmess":
		return s.linkGenerator.GenerateVmessLink(inbound, email, activeClients)
	case "vless":
		return s.linkGenerator.GenerateVlessLink(inbound, email, activeClients)
	case "trojan":
		return s.linkGenerator.GenerateTrojanLink(inbound, email, activeClients)
	case "shadowsocks":
		return s.linkGenerator.GenerateShadowsocksLink(inbound, email, activeClients)
	}
	return ""
}

func (s *ClashService) getFallbackMaster(dest string, streamSettings string) (string, int, string, error) {
	db := database.GetDB()
	var inbound model.Inbound

	err := db.Model(model.Inbound{}).
		Where("JSON_TYPE(settings, '$.fallbacks') = 'array'").
		Where("EXISTS (SELECT 1 FROM json_each(settings, '$.fallbacks') WHERE json_extract(value, '$.dest') = ?)", dest).
		Take(&inbound).Error
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

	modifiedStream, err := json.Marshal(stream)
	if err != nil {
		return "", 0, "", err
	}

	return inbound.Listen, inbound.Port, string(modifiedStream), nil
}
