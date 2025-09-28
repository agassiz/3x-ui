package proxy

import (
	"strings"

	"github.com/agassiz/3x-ui/v2/database/model"
	"github.com/agassiz/3x-ui/v2/util/common"
	"github.com/agassiz/3x-ui/v2/util/random"
)

// processNetworkSettings 处理VMESS的网络设置
func (g *LinkGenerator) processNetworkSettings(obj map[string]any, stream map[string]any, network string) {
	switch network {
	case "tcp":
		tcp, ok := stream["tcpSettings"].(map[string]any)
		if !ok || tcp == nil {
			return
		}
		header, ok := tcp["header"].(map[string]any)
		if !ok || header == nil {
			return
		}
		typeStr, _ := header["type"].(string)
		obj["type"] = typeStr
		if typeStr == "http" {
			request, ok := header["request"].(map[string]any)
			if !ok || request == nil {
				return
			}
			requestPath, ok := request["path"].([]any)
			if !ok || len(requestPath) == 0 {
				return
			}
			if pathStr, ok := requestPath[0].(string); ok {
				obj["path"] = pathStr
			}
			headers, _ := request["headers"].(map[string]any)
			obj["host"] = common.SearchHost(headers)
		}
	case "kcp":
		kcp, ok := stream["kcpSettings"].(map[string]any)
		if !ok || kcp == nil {
			return
		}
		header, ok := kcp["header"].(map[string]any)
		if !ok || header == nil {
			return
		}
		obj["type"], _ = header["type"].(string)
		obj["path"], _ = kcp["seed"].(string)
	case "ws":
		ws, ok := stream["wsSettings"].(map[string]any)
		if !ok || ws == nil {
			return
		}
		if path, ok := ws["path"].(string); ok {
			obj["path"] = path
		}
		if host, ok := ws["host"].(string); ok && len(host) > 0 {
			obj["host"] = host
		} else {
			headers, _ := ws["headers"].(map[string]any)
			obj["host"] = common.SearchHost(headers)
		}
	case "grpc":
		grpc, ok := stream["grpcSettings"].(map[string]any)
		if !ok || grpc == nil {
			return
		}
		if serviceName, ok := grpc["serviceName"].(string); ok {
			obj["path"] = serviceName
		}
		if authority, ok := grpc["authority"].(string); ok {
			obj["authority"] = authority
		}
		if multiMode, ok := grpc["multiMode"].(bool); ok && multiMode {
			obj["type"] = "multi"
		}
	case "httpupgrade":
		httpupgrade, ok := stream["httpupgradeSettings"].(map[string]any)
		if !ok || httpupgrade == nil {
			return
		}
		if path, ok := httpupgrade["path"].(string); ok {
			obj["path"] = path
		}
		if host, ok := httpupgrade["host"].(string); ok && len(host) > 0 {
			obj["host"] = host
		}
	case "xhttp":
		xhttp, ok := stream["xhttpSettings"].(map[string]any)
		if !ok || xhttp == nil {
			return
		}
		if path, ok := xhttp["path"].(string); ok {
			obj["path"] = path
		}
		if host, ok := xhttp["host"].(string); ok && len(host) > 0 {
			obj["host"] = host
		}
	}
}

// processSecuritySettings 处理VMESS的安全设置
func (g *LinkGenerator) processSecuritySettings(obj map[string]any, stream map[string]any) {
	security, _ := stream["security"].(string)
	if security == "tls" {
		obj["tls"] = "tls"
		tlsSetting, ok := stream["tlsSettings"].(map[string]any)
		if !ok || tlsSetting == nil {
			return
		}
		alpns, ok := tlsSetting["alpn"].([]any)
		if ok && alpns != nil {
			var alpn []string
			for _, a := range alpns {
				if alpnStr, ok := a.(string); ok {
					alpn = append(alpn, alpnStr)
				}
			}
			if len(alpn) > 0 {
				obj["alpn"] = strings.Join(alpn, ",")
			}
		}
		if sniValue, ok := common.SearchKey(tlsSetting, "serverName"); ok {
			obj["sni"], _ = sniValue.(string)
		}

		tlsSettings, _ := common.SearchKey(tlsSetting, "settings")
		if tlsSettings != nil {
			if tlsSettingsMap, ok := tlsSettings.(map[string]any); ok {
				if fpValue, ok := common.SearchKey(tlsSettingsMap, "fingerprint"); ok {
					obj["fp"], _ = fpValue.(string)
				}
				if insecure, ok := common.SearchKey(tlsSettingsMap, "allowInsecure"); ok {
					if insecureBool, ok := insecure.(bool); ok && insecureBool {
						obj["allowInsecure"] = "1"
					}
				}
			}
		}
	}

	if security == "reality" {
		obj["tls"] = "reality"
		realitySetting, ok := stream["realitySettings"].(map[string]any)
		if !ok || realitySetting == nil {
			return
		}
		realitySettings, _ := common.SearchKey(realitySetting, "settings")

		if sniValue, ok := common.SearchKey(realitySetting, "serverNames"); ok {
			if sNames, ok := sniValue.([]any); ok && len(sNames) > 0 {
				// 固定选择第一个serverName，与前端JavaScript保持一致
				// 前端逻辑：this.stream.reality.serverNames.split(",")[0]
				if sniStr, ok := sNames[0].(string); ok {
					obj["sni"] = sniStr
				}
			}
		}

		if realitySettings != nil {
			if realitySettingsMap, ok := realitySettings.(map[string]any); ok {
				if pbkValue, ok := common.SearchKey(realitySettingsMap, "publicKey"); ok {
					obj["pbk"], _ = pbkValue.(string)
				}
				if fpValue, ok := common.SearchKey(realitySettingsMap, "fingerprint"); ok {
					if fp, ok := fpValue.(string); ok && len(fp) > 0 {
						obj["fp"] = fp
					}
				}
			}
		}

		if sidValue, ok := common.SearchKey(realitySetting, "shortIds"); ok {
			if shortIds, ok := sidValue.([]any); ok {
				if sidStr, ok := common.FirstRealityShortIDFromAny(shortIds); ok {
					obj["sid"] = sidStr
				}
			}
		}

		obj["spx"] = "/" + random.Seq(15)
	}
}

// processNetworkParams 处理网络参数（用于VLESS/Trojan/SS）
func (g *LinkGenerator) processNetworkParams(params map[string]string, stream map[string]any, streamNetwork string) {
	switch streamNetwork {
	case "tcp":
		tcp, ok := stream["tcpSettings"].(map[string]any)
		if !ok || tcp == nil {
			return
		}
		header, ok := tcp["header"].(map[string]any)
		if !ok || header == nil {
			return
		}
		typeStr, _ := header["type"].(string)
		if typeStr == "http" {
			request, ok := header["request"].(map[string]any)
			if !ok || request == nil {
				return
			}
			requestPath, ok := request["path"].([]any)
			if !ok || len(requestPath) == 0 {
				return
			}
			if pathStr, ok := requestPath[0].(string); ok {
				params["path"] = pathStr
			}
			headers, _ := request["headers"].(map[string]any)
			params["host"] = common.SearchHost(headers)
			params["headerType"] = "http"
		}
	case "kcp":
		kcp, ok := stream["kcpSettings"].(map[string]any)
		if !ok || kcp == nil {
			return
		}
		header, ok := kcp["header"].(map[string]any)
		if !ok || header == nil {
			return
		}
		params["headerType"], _ = header["type"].(string)
		params["seed"], _ = kcp["seed"].(string)
	case "ws":
		ws, ok := stream["wsSettings"].(map[string]any)
		if !ok || ws == nil {
			return
		}
		if path, ok := ws["path"].(string); ok {
			params["path"] = path
		}
		if host, ok := ws["host"].(string); ok && len(host) > 0 {
			params["host"] = host
		} else {
			headers, _ := ws["headers"].(map[string]any)
			params["host"] = common.SearchHost(headers)
		}
	case "grpc":
		grpc, ok := stream["grpcSettings"].(map[string]any)
		if !ok || grpc == nil {
			return
		}
		if serviceName, ok := grpc["serviceName"].(string); ok {
			params["serviceName"] = serviceName
		}
		if authority, ok := grpc["authority"].(string); ok {
			params["authority"] = authority
		}
		if multiMode, ok := grpc["multiMode"].(bool); ok && multiMode {
			params["mode"] = "multi"
		}
	case "httpupgrade":
		httpupgrade, ok := stream["httpupgradeSettings"].(map[string]any)
		if !ok || httpupgrade == nil {
			return
		}
		if path, ok := httpupgrade["path"].(string); ok {
			params["path"] = path
		}
		if host, ok := httpupgrade["host"].(string); ok && len(host) > 0 {
			params["host"] = host
		}
	case "xhttp":
		xhttp, ok := stream["xhttpSettings"].(map[string]any)
		if !ok || xhttp == nil {
			return
		}
		if path, ok := xhttp["path"].(string); ok {
			params["path"] = path
		}
		if host, ok := xhttp["host"].(string); ok && len(host) > 0 {
			params["host"] = host
		}
	}
}

// processSecurityParams 处理安全参数（用于VLESS/Trojan）
func (g *LinkGenerator) processSecurityParams(params map[string]string, stream map[string]any, client model.Client, streamNetwork string) string {
	security, _ := stream["security"].(string)
	if security == "tls" {
		params["security"] = "tls"
		tlsSetting, ok := stream["tlsSettings"].(map[string]any)
		if !ok || tlsSetting == nil {
			return security
		}
		alpns, ok := tlsSetting["alpn"].([]any)
		if ok && alpns != nil {
			var alpn []string
			for _, a := range alpns {
				if alpnStr, ok := a.(string); ok {
					alpn = append(alpn, alpnStr)
				}
			}
			if len(alpn) > 0 {
				params["alpn"] = strings.Join(alpn, ",")
			}
		}
		if sniValue, ok := common.SearchKey(tlsSetting, "serverName"); ok {
			params["sni"], _ = sniValue.(string)
		}

		tlsSettings, _ := common.SearchKey(tlsSetting, "settings")
		if tlsSettings != nil {
			if tlsSettingsMap, ok := tlsSettings.(map[string]any); ok {
				if fpValue, ok := common.SearchKey(tlsSettingsMap, "fingerprint"); ok {
					params["fp"], _ = fpValue.(string)
				}
				if insecure, ok := common.SearchKey(tlsSettingsMap, "allowInsecure"); ok {
					if insecureBool, ok := insecure.(bool); ok && insecureBool {
						params["allowInsecure"] = "1"
					}
				}
			}
		}
	}

	if security == "reality" {
		params["security"] = "reality"
		realitySetting, ok := stream["realitySettings"].(map[string]any)
		if !ok || realitySetting == nil {
			return security
		}
		realitySettings, _ := common.SearchKey(realitySetting, "settings")

		if sniValue, ok := common.SearchKey(realitySetting, "serverNames"); ok {
			if sNames, ok := sniValue.([]any); ok && len(sNames) > 0 {
				// 固定选择第一个serverName，与前端JavaScript保持一致
				// 前端逻辑：this.stream.reality.serverNames.split(",")[0]
				if sniStr, ok := sNames[0].(string); ok {
					params["sni"] = sniStr
				}
			}
		}

		if realitySettings != nil {
			if realitySettingsMap, ok := realitySettings.(map[string]any); ok {
				if pbkValue, ok := common.SearchKey(realitySettingsMap, "publicKey"); ok {
					params["pbk"], _ = pbkValue.(string)
				}
				if fpValue, ok := common.SearchKey(realitySettingsMap, "fingerprint"); ok {
					if fp, ok := fpValue.(string); ok && len(fp) > 0 {
						params["fp"] = fp
					}
				}
			}
		}

		if sidValue, ok := common.SearchKey(realitySetting, "shortIds"); ok {
			if shortIds, ok := sidValue.([]any); ok {
				if sidStr, ok := common.FirstRealityShortIDFromAny(shortIds); ok {
					params["sid"] = sidStr
				}
			}
		}

		params["spx"] = "/" + random.Seq(15)

		if streamNetwork == "tcp" && len(client.Flow) > 0 {
			params["flow"] = client.Flow
		}
	}

	if security != "tls" && security != "reality" {
		params["security"] = "none"
	}

	return security
}

// processSecurityParamsSS 处理Shadowsocks的安全参数
func (g *LinkGenerator) processSecurityParamsSS(params map[string]string, stream map[string]any) string {
	security, _ := stream["security"].(string)
	if security == "tls" {
		params["security"] = "tls"
		tlsSetting, ok := stream["tlsSettings"].(map[string]any)
		if !ok || tlsSetting == nil {
			return security
		}
		alpns, ok := tlsSetting["alpn"].([]any)
		if ok && alpns != nil {
			var alpn []string
			for _, a := range alpns {
				if alpnStr, ok := a.(string); ok {
					alpn = append(alpn, alpnStr)
				}
			}
			if len(alpn) > 0 {
				params["alpn"] = strings.Join(alpn, ",")
			}
		}
		if sniValue, ok := common.SearchKey(tlsSetting, "serverName"); ok {
			params["sni"], _ = sniValue.(string)
		}

		tlsSettings, _ := common.SearchKey(tlsSetting, "settings")
		if tlsSettings != nil {
			if tlsSettingsMap, ok := tlsSettings.(map[string]any); ok {
				if fpValue, ok := common.SearchKey(tlsSettingsMap, "fingerprint"); ok {
					params["fp"], _ = fpValue.(string)
				}
				if insecure, ok := common.SearchKey(tlsSettingsMap, "allowInsecure"); ok {
					if insecureBool, ok := insecure.(bool); ok && insecureBool {
						params["allowInsecure"] = "1"
					}
				}
			}
		}
	}

	if security != "tls" {
		params["security"] = "none"
	}

	return security
}
