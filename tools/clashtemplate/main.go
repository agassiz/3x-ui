// Command clashtemplate reports where deploy/clash/acl4ssr-routing.yaml has
// drifted from the group layout sub.datapipe.top currently generates.
package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// probeLink is a placeholder node with an all-zero UUID; no real server data
// is ever sent to the converter.
const (
	probeName = "probe-node"
	probeLink = "vless://00000000-0000-4000-8000-000000000000@example.com:443?type=tcp&security=reality&pbk=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA&fp=chrome&sni=www.example.com&sid=0123abcd&flow=xtls-rprx-vision#" + probeName
)

func main() {
	templatePath := flag.String("template", "deploy/clash/acl4ssr-routing.yaml", "template to check")
	converterURL := flag.String("converter", "https://sub.datapipe.top/sub", "subconverter endpoint")
	savePath := flag.String("save", "", "also write the converter output to this file")
	flag.Parse()

	if err := run(*templatePath, *converterURL, *savePath); err != nil {
		fmt.Fprintln(os.Stderr, "clashtemplate:", err)
		os.Exit(2)
	}
}

func run(templatePath, converterURL, savePath string) error {
	template, err := os.ReadFile(templatePath)
	if err != nil {
		return err
	}
	converted, err := fetchConverted(converterURL)
	if err != nil {
		return err
	}
	if savePath != "" {
		if err := os.WriteFile(savePath, converted, 0o644); err != nil {
			return err
		}
	}
	want, err := converterLayout(converted, probeName)
	if err != nil {
		return fmt.Errorf("converter output: %w", err)
	}
	got, err := templateLayout(template)
	if err != nil {
		return fmt.Errorf("template: %w", err)
	}
	diffs := diffLayouts(want, got)
	if len(diffs) == 0 {
		fmt.Println("模板与 datapipe 当前的分组和规则顺序一致。")
		return nil
	}
	fmt.Println("模板与 datapipe 不一致：")
	for _, d := range diffs {
		fmt.Println("- " + d)
	}
	os.Exit(1)
	return nil
}

func fetchConverted(converterURL string) ([]byte, error) {
	query := url.Values{}
	query.Set("target", "clash")
	query.Set("url", probeLink)
	query.Set("insert", "false")
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(converterURL + "?" + query.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("converter returned status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
