package service

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"x-ui/database"
	"x-ui/database/model"
	"x-ui/logger"
	"x-ui/util/common"
	"x-ui/util/proxy"

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
	logger.Info("🔍 [Clash Debug] ========== 开始生成Clash订阅 ==========")
	logger.Info("🔍 [Clash Debug] 请求参数 - Email:", email, "Host:", host)

	// 设置服务器地址并更新链接生成器配置（参考SubService的实现）
	s.address = host
	s.updateLinkGeneratorConfig()

	// 1. 根据email查找客户端并生成链接
	clientUrl, realPort, err := s.getClientUrl(email, host)
	if err != nil {
		logger.Error("🔍 [Clash Debug] 获取客户端链接失败:", err)
		return "", err
	}

	if clientUrl == "" {
		logger.Error("🔍 [Clash Debug] 未找到客户端，Email:", email)
		return "", common.NewError("Client not found for email: " + email)
	}

	logger.Info("🔍 [Clash Debug] 成功获取客户端信息 - URL:", clientUrl, "端口:", realPort)

	// 2. 计算URL的MD5签名
	urlMd5 := s.calculateMD5(clientUrl)
	logger.Info("🔍 [Clash Debug] 客户端链接MD5:", urlMd5)

	// 3. 检查数据库中是否已有缓存
	db := database.GetDB()
	var subscription model.ClashSubscription
	queryErr := db.Where("email = ?", email).First(&subscription).Error

	var isNewRecord bool
	if queryErr == nil {
		// 找到记录，检查MD5是否一致
		logger.Info("🔍 [Clash Debug] 找到缓存记录 - 缓存MD5:", subscription.UrlMd5, "当前MD5:", urlMd5)
		if subscription.UrlMd5 == urlMd5 {
			// MD5一致，直接返回缓存的YAML
			logger.Info("🔍 [Clash Debug] 缓存命中，直接返回缓存内容，长度:", len(subscription.YamlContent))
			return subscription.YamlContent, nil
		}
		// MD5不一致，需要重新生成
		logger.Info("🔍 [Clash Debug] 缓存失效（URL已变更），需要重新生成")
		isNewRecord = false
	} else if queryErr != gorm.ErrRecordNotFound {
		// 数据库查询出错
		logger.Error("🔍 [Clash Debug] 数据库查询错误:", queryErr)
		return "", queryErr
	} else {
		logger.Info("🔍 [Clash Debug] 未找到缓存记录，需要首次生成")
		isNewRecord = true
	}

	// 4. 调用外部API生成新的Clash配置
	yamlContent, err := s.generateClashConfig(clientUrl, realPort)
	if err != nil {
		return "", err
	}

	// 5. 保存或更新缓存（使用安全的Upsert操作）
	now := time.Now().Unix()
	clashSubscription := model.ClashSubscription{
		Email:       email,
		UrlMd5:      urlMd5,
		YamlContent: yamlContent,
		UpdatedAt:   now,
	}

	if isNewRecord {
		// 创建新记录
		clashSubscription.CreatedAt = now
		err = db.Create(&clashSubscription).Error
	} else {
		// 更新现有记录，保留原有的CreatedAt
		clashSubscription.Id = subscription.Id
		clashSubscription.CreatedAt = subscription.CreatedAt
		err = db.Save(&clashSubscription).Error
	}

	if err != nil {
		logger.Error("🔍 [Clash Debug] 保存缓存到数据库失败:", err)
		// 即使保存失败，也返回生成的配置
	} else {
		logger.Info("🔍 [Clash Debug] 成功保存缓存到数据库")
	}

	logger.Info("🔍 [Clash Debug] ========== Clash订阅生成完成 ==========")
	logger.Info("🔍 [Clash Debug] 最终返回内容长度:", len(yamlContent), "字符")

	return yamlContent, nil
}

// getClientUrl 根据email查找客户端并生成链接
func (s *ClashService) getClientUrl(email string, host string) (string, int, error) {
	db := database.GetDB()
	var inbounds []model.Inbound

	// 使用SQL查询找到包含指定email的入站规则
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
	`, email).Scan(&inbounds).Error

	if err != nil {
		return "", 0, err
	}

	if len(inbounds) == 0 {
		return "", 0, nil
	}

	inbound := inbounds[0]

	// 打印找到的入站规则信息
	logger.Info("🔍 [Clash Debug] 找到入站规则 - ID:", inbound.Id, "协议:", inbound.Protocol, "端口:", inbound.Port)
	logger.Info("🔍 [Clash Debug] 入站规则备注:", inbound.Remark)

	// 直接调用简化的链接生成方法（基于SubService逻辑）
	clientLink := s.generateClientLink(&inbound, email, host)
	logger.Info("🔍 [Clash Debug] 生成的客户端链接:", clientLink)

	return clientLink, inbound.Port, nil
}

// generateClashConfig 调用外部API生成Clash配置，并替换真实端口
func (s *ClashService) generateClashConfig(clientUrl string, realPort int) (string, error) {
	// 打印原始客户端链接用于调试
	logger.Info("🔍 [Clash Debug] 原始客户端链接:", clientUrl)

	// URL编码客户端链接
	encodedUrl := url.QueryEscape(clientUrl)
	logger.Info("🔍 [Clash Debug] URL编码后:", encodedUrl)

	// 构建外部API URL
	apiUrl := fmt.Sprintf("https://sub.datapipe.top/sub?target=clash&url=%s&insert=false", encodedUrl)
	logger.Info("🔍 [Clash Debug] 完整API URL:", apiUrl)
	logger.Info("🔍 [Clash Debug] 真实端口:", realPort)

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
	logger.Info("🔍 [Clash Debug] 外部API返回内容长度:", len(yamlContent), "字符")

	// 打印前200个字符用于调试
	if len(yamlContent) > 200 {
		logger.Info("🔍 [Clash Debug] API返回内容预览:", yamlContent[:200], "...")
	} else {
		logger.Info("🔍 [Clash Debug] API返回完整内容:", yamlContent)
	}

	// 将固定端口45556替换为真实端口
	beforeReplace := strings.Count(yamlContent, "45556")
	yamlContent = strings.ReplaceAll(yamlContent, "45556", fmt.Sprintf("%d", realPort))
	afterReplace := strings.Count(yamlContent, fmt.Sprintf("%d", realPort))

	logger.Info("🔍 [Clash Debug] 端口替换统计 - 替换前45556出现次数:", beforeReplace, "替换后", realPort, "出现次数:", afterReplace)

	return yamlContent, nil
}

// calculateMD5 计算字符串的MD5值
func (s *ClashService) calculateMD5(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// generateClientLink 生成客户端链接（使用固定端口45556隐藏真实端口）
func (s *ClashService) generateClientLink(inbound *model.Inbound, email string, address string) string {
	logger.Info("🔍 [Clash Debug] 开始生成客户端链接 - Email:", email, "Address:", address)
	logger.Info("🔍 [Clash Debug] 当前配置 - RemarkModel:", s.remarkModel, "ShowInfo:", s.showInfo)

	// 获取客户端信息
	clients, err := s.inboundService.GetClients(inbound)
	if err != nil {
		return ""
	}

	// 根据协议类型生成完整的链接
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
