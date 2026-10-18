<p align="center">
  <a href="https://docs.sanaei.dev">
    <picture>
      <source media="(prefers-color-scheme: dark)" srcset="public/logo-dark.png" />
      <img src="public/logo-light.png" alt="3x-ui" width="180" />
    </picture>
  </a>
</p>

<h1 align="center">3x-ui Documentation</h1>

<p align="center">
  The official documentation and product site for
  <a href="https://github.com/agassiz/3x-ui"><b>3x-ui</b></a> —
  an advanced web panel for managing Xray-core servers.
</p>

<p align="center">
  <a href="https://docs.sanaei.dev"><img src="https://img.shields.io/badge/docs-docs.sanaei.dev-22d3ee?style=flat-square" alt="Live site" /></a>
  <a href="https://github.com/agassiz/3x-ui/actions/workflows/docs-ci.yml"><img src="https://github.com/agassiz/3x-ui/actions/workflows/docs-ci.yml/badge.svg" alt="CI" /></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-GPL--3.0-blue?style=flat-square" alt="License: GPL-3.0" /></a>
  <img src="https://img.shields.io/badge/Next.js-16-black?style=flat-square&logo=next.js" alt="Next.js 16" />
  <img src="https://img.shields.io/badge/Fumadocs-16-0ea5e9?style=flat-square" alt="Fumadocs 16" />
</p>

<p align="center">
  <a href="https://docs.sanaei.dev"><b>Read the docs →</b></a>
</p>

---

## Overview

This directory (`docs/` in the [3x-ui](https://github.com/agassiz/3x-ui) monorepo) contains
the source for [docs.sanaei.dev](https://docs.sanaei.dev) — a static-first documentation and
marketing site built with [Fumadocs](https://fumadocs.dev) on Next.js. It has **no backend,
no database, and no auth**: every page is prerendered and every tool runs entirely in the
browser.

## What's inside

The documentation walks you through 3x-ui from first install to day-to-day operation:

- **Getting Started** — installation, first login, and updating or uninstalling the panel.
- **Configuration** — the panel, inbounds, REALITY, transports, clients, subscriptions, and share links.
- **Operations** — reverse proxy, multi-node setups, outbounds & routing, backup/restore, Telegram and Discord bots, and security.
- **Reference** — environment variables, the database, ports & firewall, and the HTTP API.
- **Help** — troubleshooting, FAQ, migration, and how to contribute.

## Interactive tools

The site ships with in-browser helpers that generate configuration for you — **no data
ever leaves your browser**:

| Tool                         | What it does                                            |
| ---------------------------- | ------------------------------------------------------- |
| **REALITY Config Generator** | Build a valid REALITY inbound configuration.            |
| **Share Link Inspector**     | Decode and inspect `vless://` / `vmess://` share links. |
| **Install Command Builder**  | Assemble the right install command for your setup.      |
| **Reverse Proxy Generator**  | Generate reverse-proxy configs (Nginx / Caddy).         |
| **Protocol Wizard**          | Pick and configure the right protocol for your needs.   |
| **Firewall Rules Generator** | Produce firewall rules for your ports.                  |

## Tech stack

| Layer     | Technology                                                  |
| --------- | ----------------------------------------------------------- |
| Framework | [Next.js 16](https://nextjs.org) (App Router) · React 19    |
| Docs      | [Fumadocs](https://fumadocs.dev) (`-ui` / `-core` / `-mdx`) |
| Styling   | [Tailwind CSS v4](https://tailwindcss.com)                  |
| Search    | [Orama](https://orama.com) static index                     |
| Language  | TypeScript (strict)                                         |
| Tests     | [Vitest](https://vitest.dev) for the pure `lib/xray` logic  |
| Tooling   | pnpm · oxlint · oxfmt                                       |

## Quick start

This project uses **[pnpm](https://pnpm.io)** (npm lockfiles are gitignored). Run everything
from the `docs/` directory:

```bash
cd docs
pnpm install
pnpm dev        # http://localhost:3000
```

Useful scripts:

| Script           | Description                                 |
| ---------------- | ------------------------------------------- |
| `pnpm dev`       | Start the dev server                        |
| `pnpm build`     | Production build (also typechecks)          |
| `pnpm typecheck` | Generate MDX/route types and `tsc --noEmit` |
| `pnpm lint`      | Run oxlint (`.oxlintrc.json`)               |
| `pnpm test`      | Run unit tests (Vitest)                     |

See [`CONTRIBUTING.md`](./CONTRIBUTING.md) for the full list and project conventions.

## Project structure

```
app/             # Next.js App Router — layouts, home, docs, OG images, search, llms.txt
components/      # React components — interactive tools, home sections, MDX bindings
content/docs/    # MDX documentation, one folder per locale (en · fa · ru · zh)
lib/             # source config, i18n, GitHub stats, and the unit-tested lib/xray logic
public/          # static assets — logos, favicon, openapi.json, CNAME
scripts/         # build-time scripts (API reference generation)
source.config.ts # Fumadocs MDX schema & collection config
next.config.mjs  # Next.js config (static-export gating)
proxy.ts         # i18n middleware
```

## Internationalization

Documentation is authored in **English**. Persian (`fa`, RTL), Russian (`ru`), and
Chinese (`zh`) locales are wired up; untranslated pages fall back to English so they
never 404. English URLs are unprefixed; other locales live under `/fa`, `/ru`, `/zh`.

## Deployment

The site builds for two targets:

- **Vercel / Node** — `pnpm build` (static search index + prerendered OG images).
- **GitHub Pages (static export)** — `DEPLOY_TARGET=static pnpm build` → `out/`.

## Contributing

Contributions are welcome! Setup, scripts, and project conventions live in
[`CONTRIBUTING.md`](./CONTRIBUTING.md).

## License

Licensed under [GPL-3.0](./LICENSE).

---

## 3x-ui 技术文档

欢迎来到 3x-ui 项目的技术文档中心。本文档集合提供了项目的完整技术说明和开发指南。

### 📚 文档目录

#### 🏗️ 架构设计
- [项目概览](./01-project-overview.md) - 项目介绍、功能特性、技术栈
- [系统架构](./02-architecture.md) - 整体架构设计和组件关系
- [数据库设计](./03-database-design.md) - 数据模型和表结构设计

#### 🔌 接口文档
- [API 接口规范](./04-api-documentation.md) - REST API 完整规范
- [前端架构](./05-frontend-architecture.md) - 前端技术栈和页面结构

#### ⚙️ 核心服务
- [核心服务文档](./06-core-services.md) - Xray集成和业务逻辑
- [流程图和时序图](./07-diagrams.md) - 关键业务流程可视化

#### 🚀 部署运维
- [部署运维指南](./08-deployment.md) - Docker部署和配置管理
- [开发指南](./09-development-guide.md) - 开发环境搭建和代码规范

### 🎯 快速导航

#### 新手入门
1. 先阅读 [项目概览](./01-project-overview.md) 了解项目基本情况
2. 查看 [系统架构](./02-architecture.md) 理解整体设计
3. 参考 [开发指南](./09-development-guide.md) 搭建开发环境

#### 开发者
- [API 接口规范](./04-api-documentation.md) - 接口开发必读
- [数据库设计](./03-database-design.md) - 数据操作参考
- [核心服务文档](./06-core-services.md) - 业务逻辑实现

#### 运维人员
- [部署运维指南](./08-deployment.md) - 生产环境部署
- [系统架构](./02-architecture.md) - 系统组件理解

### 📝 文档维护

本文档基于项目实际代码生成，定期更新以保持与代码同步。

- **生成时间**: 2024-07-14
- **项目版本**: v2.6.4
- **维护方式**: 自动化分析 + 人工审核

### 🤝 贡献指南

如发现文档错误或需要补充，请参考 [开发指南](./09-development-guide.md) 中的贡献流程。

---

*本文档由 AI 助手基于项目实际代码自动生成，确保内容的准确性和时效性。*
