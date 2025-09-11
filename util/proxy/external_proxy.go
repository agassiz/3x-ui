package proxy

import (
	"encoding/base64"
	"fmt"
	"net/url"

	"x-ui/database/model"

	"github.com/goccy/go-json"
)

// processExternalProxiesVmess 处理VMESS的外部代理
func (g *LinkGenerator) processExternalProxiesVmess(obj map[string]any, inbound *model.Inbound, email string) string {
	links := ""
	for index, externalProxy := range g.config.ExternalProxies {
		ep := externalProxy
		newSecurity, _ := ep["forceTls"].(string)
		newObj := map[string]any{}
		for key, value := range obj {
			if !(newSecurity == "none" && (key == "alpn" || key == "sni" || key == "fp" || key == "allowInsecure")) {
				newObj[key] = value
			}
		}
		newObj["ps"] = g.GenerateRemark(inbound, email, ep["remark"].(string))
		newObj["add"] = ep["dest"].(string)
		newObj["port"] = int(ep["port"].(float64))

		if newSecurity != "same" {
			newObj["tls"] = newSecurity
		}
		if index > 0 {
			links += "\n"
		}
		jsonStr, _ := json.MarshalIndent(newObj, "", "  ")
		links += "vmess://" + base64.StdEncoding.EncodeToString(jsonStr)
	}
	return links
}

// processExternalProxiesVless 处理VLESS的外部代理
func (g *LinkGenerator) processExternalProxiesVless(uuid string, params map[string]string, security string, inbound *model.Inbound, email string) string {
	links := ""
	for index, externalProxy := range g.config.ExternalProxies {
		ep := externalProxy
		newSecurity, _ := ep["forceTls"].(string)
		dest, _ := ep["dest"].(string)
		port := int(ep["port"].(float64))
		link := fmt.Sprintf("vless://%s@%s:%d", uuid, dest, port)

		if newSecurity != "same" {
			params["security"] = newSecurity
		} else {
			params["security"] = security
		}
		url, _ := url.Parse(link)
		q := url.Query()

		for k, v := range params {
			if !(newSecurity == "none" && (k == "alpn" || k == "sni" || k == "fp" || k == "allowInsecure")) {
				q.Add(k, v)
			}
		}

		url.RawQuery = q.Encode()
		url.Fragment = g.GenerateRemark(inbound, email, ep["remark"].(string))

		if index > 0 {
			links += "\n"
		}
		links += url.String()
	}
	return links
}

// processExternalProxiesTrojan 处理Trojan的外部代理
func (g *LinkGenerator) processExternalProxiesTrojan(password string, params map[string]string, security string, inbound *model.Inbound, email string) string {
	links := ""
	for index, externalProxy := range g.config.ExternalProxies {
		ep := externalProxy
		newSecurity, _ := ep["forceTls"].(string)
		dest, _ := ep["dest"].(string)
		port := int(ep["port"].(float64))
		link := fmt.Sprintf("trojan://%s@%s:%d", password, dest, port)

		if newSecurity != "same" {
			params["security"] = newSecurity
		} else {
			params["security"] = security
		}
		url, _ := url.Parse(link)
		q := url.Query()

		for k, v := range params {
			if !(newSecurity == "none" && (k == "alpn" || k == "sni" || k == "fp" || k == "allowInsecure")) {
				q.Add(k, v)
			}
		}

		url.RawQuery = q.Encode()
		url.Fragment = g.GenerateRemark(inbound, email, ep["remark"].(string))

		if index > 0 {
			links += "\n"
		}
		links += url.String()
	}
	return links
}

// processExternalProxiesSS 处理Shadowsocks的外部代理
func (g *LinkGenerator) processExternalProxiesSS(encPart string, params map[string]string, security string, inbound *model.Inbound, email string) string {
	links := ""
	for index, externalProxy := range g.config.ExternalProxies {
		ep := externalProxy
		newSecurity, _ := ep["forceTls"].(string)
		dest, _ := ep["dest"].(string)
		port := int(ep["port"].(float64))
		link := fmt.Sprintf("ss://%s@%s:%d", base64.StdEncoding.EncodeToString([]byte(encPart)), dest, port)

		if newSecurity != "same" {
			params["security"] = newSecurity
		} else {
			params["security"] = security
		}
		url, _ := url.Parse(link)
		q := url.Query()

		for k, v := range params {
			if !(newSecurity == "none" && (k == "alpn" || k == "sni" || k == "fp" || k == "allowInsecure")) {
				q.Add(k, v)
			}
		}

		url.RawQuery = q.Encode()
		url.Fragment = g.GenerateRemark(inbound, email, ep["remark"].(string))

		if index > 0 {
			links += "\n"
		}
		links += url.String()
	}
	return links
}
