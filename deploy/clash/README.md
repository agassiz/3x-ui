# Clash 分流模板（ACL4SSR 风格）

[`acl4ssr-routing.yaml`](acl4ssr-routing.yaml) 是给面板"Clash 路由规则"用的分流模板：15 个策略组，规则由客户端从 ACL4SSR 直接下载。

## 来源

- **分组结构**复刻自 `https://sub.datapipe.top/sub` 的输出：分组名称、顺序、类型、成员，以及规则的分段顺序都和它一致。最初这个 fork 用 datapipe 把节点链接转换成 Clash 配置；现在节点由面板本地生成，只借用它的分组和规则顺序。
- **获取方式**：2026-10-18 用一个占位节点（全零 UUID、`example.com`）请求 datapipe 得到，没有发送任何真实节点信息。
- **规则内容**：datapipe 把约 1 万条规则直接写在订阅里（约 530KB）。模板改为通过 `rule-providers` 引用 ACL4SSR 的 `.list` 规则集，订阅只有几 KB，规则由客户端每 24 小时自动更新。

与 datapipe 输出的差异：

| 项目 | datapipe | 本模板 |
| --- | --- | --- |
| 广告规则 | 约 8500 个域名的自定义组合 | ACL4SSR 的 BanAD + BanProgramAD，较少，误杀也更少 |
| 分组里的节点 | 直接写出节点名 | `include-all-proxies: true`，需要 Mihomo（Clash Meta）内核客户端 |
| `external-controller` | `:9090`（对局域网开放） | `127.0.0.1:9090` |
| `URL-REGEX` 规则 | 转换时丢弃 | Mihomo 加载时跳过 |

## 分组与规则集

规则按下表顺序匹配，最后两条是 `GEOIP,CN,🎯 全球直连` 和 `MATCH,🐟 漏网之鱼`：

| 规则集（ACL4SSR `Clash/` 下） | 分组 |
| --- | --- |
| `LocalAreaNetwork` | 🎯 全球直连 |
| `Ruleset/OpenAi` | 🤖GPT |
| `Ruleset/Gemini` | 🤖Gemini |
| `Ruleset/Microsoft` | Ⓜ️ 微软服务 |
| `UnBan`、`Ruleset/GoogleFCM`、`GoogleCN` | 🎯 全球直连 |
| `BanAD`、`BanProgramAD` | 🛑 全球拦截 |
| `Ruleset/Netflix` | 🎥 NETFLIX |
| `ProxyMedia` | 🌍 国外媒体 |
| `ChinaMedia` | 🌏 国内媒体 |
| `Telegram` | 📲 电报信息 |
| `ProxyGFWlist` | 🔰 节点选择 |
| `Apple` | 🍎 苹果服务 |
| `ChinaDomain`、`ChinaCompanyIp`、`Download` | 🎯 全球直连 |

⛔️ 广告拦截、🚫 运营劫持 两个分组在 datapipe 输出里就没有规则指向，这里同样保留为空。

## 使用

前提：面板设置 → 订阅中已开启订阅服务和 Clash。客户端列表的 Clash 按钮显示的是 `subClashURI` 加客户端的 `subId`。

这个 fork 的默认设置已经开启"Clash 路由"，并且"Clash 路由规则"默认填的就是下面的远程 URL。新安装的面板直接生效；已经在设置页保存过的面板，数据库里存的是当时的值，需要手动改一次。

手动设置时，在面板设置 → 订阅中开启"Clash 路由"，然后在"Clash 路由规则"里二选一：

- **远程 URL（推荐）**：填 `https://raw.githubusercontent.com/agassiz/3x-ui/main/deploy/clash/acl4ssr-routing.yaml`。订阅服务会缓存 10 分钟，模板推送到 `main` 后 10 分钟内生效。远程方式只合并 `proxy-groups`、`rule-providers`、`rules`，`port`、`allow-lan` 等顶层设置由客户端自己决定。
- **粘贴内容**：顶层设置也会生效，但模板更新后需要重新粘贴。

## 同步更新

1. **规则内容**：不用处理，客户端会按 `interval: 86400` 自动从 ACL4SSR 更新。
2. **检查分组结构是否变化**：在仓库根目录运行

   ```bash
   go run ./tools/clashtemplate -save /tmp/datapipe.yaml
   ```

   一致时输出"一致"并以 0 退出；不一致时列出差异并以 1 退出，例如新增的分组、成员变化、规则分段顺序变化。工具只发送占位节点；`-save` 会保存 datapipe 的原始输出，方便查看新增规则段具体包含哪些域名。
3. **按差异修改模板**：datapipe 里列出节点名的分组，在模板里写成 `include-all-proxies: true`。新增的规则段要从 ACL4SSR 中挑选对应的规则集，并加到 `rule-providers` 和 `rules` 里。
4. **校验**：运行

   ```bash
   go test ./internal/sub -run TestACL4SSR
   ```

   它会检查分组、规则集名称有没有写错（写错时 Clash 客户端会拒绝整份配置），也会检查分组列表是否等于 `internal/sub/acl4ssr_template_test.go` 里的期望值；如果分组有增删，要同步改这份期望值。
5. **提交并推送**：远程 URL 方式会自动生效；粘贴方式需要重新粘贴。
