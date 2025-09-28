package controller

import (
	"net"
	"net/http"
	"strings"

	"github.com/agassiz/3x-ui/v2/logger"
	"github.com/agassiz/3x-ui/v2/web/service"

	"github.com/gin-gonic/gin"
)

type ClashController struct {
	clashService *service.ClashService
}

func NewClashController(g *gin.RouterGroup) *ClashController {
	controller := &ClashController{
		clashService: service.NewClashService(),
	}
	controller.initRouter(g)
	return controller
}

func (c *ClashController) initRouter(g *gin.RouterGroup) {
	// 创建公开的API路由组，不需要登录验证
	clashGroup := g.Group("/clash")

	// 注册路由 - 只需要订阅获取接口
	clashGroup.GET("/subscription/:email", c.getClashSubscription)
}

// getClashSubscription 获取Clash订阅配置
func (c *ClashController) getClashSubscription(ctx *gin.Context) {
	email := ctx.Param("email")

	// 验证email参数
	if email == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Email parameter is required",
		})
		return
	}

	// 清理email参数（去除空格等）
	email = strings.TrimSpace(email)

	// 获取服务器地址（正确处理端口分离）
	var host string
	if rawForwarded := ctx.GetHeader("X-Forwarded-Host"); rawForwarded != "" {
		parts := strings.Split(rawForwarded, ",")
		candidate := strings.TrimSpace(parts[0])
		if h, err := getHostFromXFH(candidate); err == nil {
			host = h
		}
	}
	if host == "" {
		var err error
		host, _, err = net.SplitHostPort(ctx.Request.Host)
		if err != nil {
			host = ctx.Request.Host
		}
	}

	logger.Info("Generating Clash subscription for email:", email, "host:", host)

	// 调用服务获取Clash配置
	yamlContent, err := c.clashService.GetClashSubscription(email, host)
	if err != nil {
		logger.Error("Failed to get Clash subscription for email", email, ":", err)

		// 根据错误类型返回不同的HTTP状态码
		if strings.Contains(err.Error(), "not found") {
			ctx.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Client not found for the specified email",
			})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to generate Clash subscription",
				"error":   err.Error(),
			})
		}
		return
	}

	// 设置响应头，指示这是YAML文件
	ctx.Header("Content-Type", "application/x-yaml; charset=utf-8")
	ctx.Header("Content-Disposition", "attachment; filename=clash-config.yaml")

	// 添加缓存控制头
	ctx.Header("Cache-Control", "public, max-age=300") // 5分钟缓存

	// 返回YAML内容
	ctx.String(http.StatusOK, yamlContent)
}

// getHostFromXFH 从X-Forwarded-Host头中提取主机地址（去除端口）
func getHostFromXFH(s string) (string, error) {
	if strings.Contains(s, ":") {
		realHost, _, err := net.SplitHostPort(s)
		if err != nil {
			return "", err
		}
		return realHost, nil
	}
	return s, nil
}
