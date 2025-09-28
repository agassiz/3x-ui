# 3x-ui 线上环境升级指南

## 🎯 升级概述

本指南适用于通过官方安装脚本部署的3x-ui线上环境：
```bash
bash <(curl -Ls https://raw.githubusercontent.com/agassiz/3x-ui/master/install.sh)
```

## ⚠️ 升级前准备

### 1. 数据备份（必须执行）

```bash
# 备份配置文件
sudo cp -r /etc/x-ui /etc/x-ui.backup.$(date +%Y%m%d_%H%M%S)

# 备份数据库
sudo cp -r /usr/local/x-ui/db /usr/local/x-ui/db.backup.$(date +%Y%m%d_%H%M%S)

# 备份证书文件（如果有）
sudo cp -r /root/cert /root/cert.backup.$(date +%Y%m%d_%H%M%S) 2>/dev/null || echo "No certificates to backup"

# 查看当前版本
x-ui status
```

### 2. 检查系统状态

```bash
# 检查服务状态
systemctl status x-ui

# 检查端口占用
netstat -tlnp | grep :2053

# 检查磁盘空间
df -h

# 检查内存使用
free -h
```

## 🚀 升级方法

### 方法一：使用x-ui管理脚本升级（推荐）

```bash
# 进入x-ui管理界面
x-ui

# 选择选项 2 (Update)
# 或者直接执行升级命令
x-ui update
```

### 方法二：直接执行升级命令

```bash
# 一键升级到最新版本
bash <(curl -Ls https://raw.githubusercontent.com/agassiz/3x-ui/main/install.sh)
```

### 方法三：手动升级（高级用户）

```bash
# 停止服务
systemctl stop x-ui

# 下载最新版本
cd /tmp
wget https://github.com/agassiz/3x-ui/releases/latest/download/x-ui-linux-amd64.tar.gz

# 解压并安装
tar -xzf x-ui-linux-amd64.tar.gz
sudo cp x-ui /usr/local/x-ui/
sudo chmod +x /usr/local/x-ui/x-ui

# 重启服务
systemctl start x-ui
```

## 📊 升级验证

### 1. 检查服务状态

```bash
# 检查服务是否正常运行
systemctl status x-ui

# 检查版本信息
x-ui status

# 检查日志
journalctl -u x-ui -f --no-pager -n 50
```

### 2. 功能验证

```bash
# 检查Web面板访问
curl -I http://localhost:2053

# 检查配置是否保留
x-ui settings

# 检查用户数据
# 登录Web面板验证用户配置是否完整
```

### 3. 性能验证

```bash
# 检查内存使用
ps aux | grep x-ui

# 检查网络连接
netstat -tlnp | grep x-ui

# 运行速度测试
x-ui
# 选择选项 25 (Speedtest)
```

## 🔧 升级后优化

### 1. 依赖更新收益验证

升级后您将获得以下改进：

**安全性提升**：
- ✅ golang.org/x/crypto v0.40.0 - 最新加密算法
- ✅ golang.org/x/net v0.42.0 - 网络安全修复
- ✅ golang.org/x/sys v0.34.0 - 系统调用安全

**性能优化**：
- ✅ fasthttp v1.64.0 - HTTP处理性能提升
- ✅ cpuid v2.3.0 - CPU检测优化
- ✅ sagernet/sing v0.6.11 - 协议处理优化

### 2. 配置优化建议

```bash
# 启用BBR加速（可选）
x-ui
# 选择选项 23 (Enable BBR)

# 更新地理位置数据
x-ui
# 选择选项 24 (Update Geo Files)

# 配置防火墙（推荐）
x-ui
# 选择选项 21 (Firewall Management)
```

## 🚨 故障排除

### 升级失败处理

```bash
# 如果升级失败，恢复备份
sudo systemctl stop x-ui
sudo rm -rf /etc/x-ui
sudo mv /etc/x-ui.backup.* /etc/x-ui
sudo rm -rf /usr/local/x-ui/db
sudo mv /usr/local/x-ui/db.backup.* /usr/local/x-ui/db
sudo systemctl start x-ui
```

### 常见问题解决

**问题1：服务启动失败**
```bash
# 检查日志
journalctl -u x-ui -n 50

# 重置配置
x-ui
# 选择选项 8 (Reset Settings)
```

**问题2：Web面板无法访问**
```bash
# 检查端口
netstat -tlnp | grep :2053

# 重置端口
x-ui
# 选择选项 9 (Change Port)
```

**问题3：用户数据丢失**
```bash
# 恢复数据库备份
sudo systemctl stop x-ui
sudo cp -r /usr/local/x-ui/db.backup.* /usr/local/x-ui/db
sudo systemctl start x-ui
```

## 📋 升级检查清单

### 升级前检查
- [ ] 已备份配置文件 (/etc/x-ui)
- [ ] 已备份数据库 (/usr/local/x-ui/db)
- [ ] 已备份证书文件 (/root/cert)
- [ ] 已记录当前版本信息
- [ ] 已检查磁盘空间充足
- [ ] 已通知用户维护时间

### 升级后验证
- [ ] 服务正常运行 (systemctl status x-ui)
- [ ] Web面板可正常访问
- [ ] 用户配置完整保留
- [ ] 代理功能正常工作
- [ ] 证书配置正确
- [ ] 日志无错误信息

### 性能验证
- [ ] 内存使用正常
- [ ] CPU使用率稳定
- [ ] 网络连接正常
- [ ] 响应速度提升

## 🔄 回滚方案

如果升级后出现问题，可以快速回滚：

```bash
# 方法1：使用备份恢复
sudo systemctl stop x-ui
sudo rm -rf /etc/x-ui /usr/local/x-ui/db
sudo mv /etc/x-ui.backup.* /etc/x-ui
sudo mv /usr/local/x-ui/db.backup.* /usr/local/x-ui/db
sudo systemctl start x-ui

# 方法2：重新安装旧版本
x-ui
# 选择选项 4 (Legacy Version)
# 输入之前的版本号，如：2.6.2
```

## 📞 技术支持

### 官方资源
- **GitHub**: https://github.com/agassiz/3x-ui
- **Telegram**: https://t.me/XrayUI
- **文档**: https://github.com/agassiz/3x-ui/wiki

### 升级后监控

建议升级后持续监控1-2天：

```bash
# 实时监控日志
journalctl -u x-ui -f

# 监控系统资源
htop

# 检查网络连接
watch -n 5 'netstat -tlnp | grep x-ui'
```

## 🎉 升级完成

升级完成后，您的3x-ui将包含：
- ✅ **最新安全更新** - 修复已知安全漏洞
- ✅ **性能优化** - HTTP处理和系统调用性能提升
- ✅ **协议支持** - 最新的代理协议支持
- ✅ **稳定性改进** - 更好的错误处理和恢复机制

---

**重要提醒**：
1. 升级过程中服务会短暂中断（通常1-2分钟）
2. 建议在用户使用较少的时间段进行升级
3. 升级前务必完成数据备份
4. 如有问题，可随时使用备份进行回滚
