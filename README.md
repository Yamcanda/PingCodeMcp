# PingCode MCP 服务器

[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)](https://github.com/your-repo/PingCodeMcp)

PingCode MCP (Model Context Protocol) 服务器是一个基于 Go 语言开发的高性能 MCP 服务器，为 AI 模型提供与 PingCode 平台的集成能力。通过丰富的工具集，AI 可以直接操作 PingCode 的项目管理、工作项、工时、用户管理等功能。

## 🎯 项目概述

PingCode MCP 服务器实现了 Model Context Protocol 规范，为 AI 助手提供了与 PingCode 研发管理平台的无缝集成。支持项目管理、工作项管理、工时记录、用户管理、测试管理等核心功能，让 AI 能够智能化地协助团队进行研发管理。

## ✨ 功能特性

### 🔧 核心功能
- **MCP 协议支持**: 完全兼容 Model Context Protocol 规范
- **HTTP 服务器**: 提供 RESTful API 接口
- **认证机制**: 支持 Bearer Token 认证
- **日志系统**: 结构化日志记录，支持每日轮转
- **配置管理**: 灵活的 YAML 配置文件支持
- **Docker 支持**: 容器化部署，支持健康检查

### 🛠️ 工具集合

#### 项目管理工具
- **get_project_list**: 获取项目列表
- **create_project**: 创建新项目
- **add_project_member**: 添加项目成员
- **remove_project_member**: 移除项目成员
- **get_project_members**: 获取项目成员列表

#### 工作项管理工具
- **create_work_item**: 创建工作项
- **update_work_item**: 更新工作项
- **list_work_items**: 获取工作项列表
- **delete_work_item**: 删除工作项
- **list_work_item_states**: 获取工作项状态列表

#### 工时管理工具
- **create_workload**: 创建工时记录
- **update_workload**: 更新工时记录
- **list_workloads**: 获取工时列表
- **delete_workload**: 删除工时记录
- **get_workload_types**: 获取工时类型列表

#### 用户管理工具
- **get_user_info**: 获取用户基本信息
- **list_enterprise_users**: 获取企业用户列表
- **create_enterprise_user**: 创建企业用户

#### 实用工具
- **ip_search**: IP 地址地理位置查询
- **time_search**: 时间查询工具

## 🚀 快速开始

### 环境要求

- Go 1.23 或更高版本
- PingCode 平台访问权限
- 有效的 API Token

### 安装方式

#### 方式一：源码编译

```bash
# 克隆项目
git clone https://github.com/peach-zhang/PingCodeMcp.git
cd PingCodeMcp

# 下载依赖
go mod download

# 编译项目
make build

# 运行服务器
./bin/server
```

#### 方式二：Docker 部署

```bash
# 构建镜像
make docker-build

# 运行容器
make docker-run
```

#### 方式三：直接运行

```bash
# 开发模式运行
make run
```

### 配置文件

创建 `config.yaml` 配置文件：

```yaml
# PingCode MCP 服务器配置文件
version: "1.0.0"

# 服务器配置
server:
  name: "ping-code-mcp-server"
  version: "1.0.0"
  port: 8080
  host: "0.0.0.0"

# 日志配置
logging:
  level: "info"          # debug, info, warn, error, fatal
  format: "json"         # json, console
  output: "both"         # console, file, both
  
  # 文件日志配置
  file:
    path: "logs"
    filename: "PingCodeMcp.log"
    max_size: 100        # MB
    max_age: 7           # days
    max_backups: 10
    compress: true
  
  # 日志轮转配置
  rotation:
    enabled: true
    daily: true          # 启用每日轮转

# 环境变量覆盖配置
env_override:
  enabled: true
  prefix: "PingCodeMcp"
```

### 环境变量

支持通过环境变量覆盖配置：

```bash
# 服务器配置
export PingCodeMcp_SERVER_PORT=9090
export PingCodeMcp_SERVER_HOST=localhost

# 日志配置
export PingCodeMcp_LOGGING_LEVEL=debug
export PingCodeMcp_LOGGING_FORMAT=console
```

## 📖 使用指南

### API 认证

所有 PingCode API 调用都需要提供有效的 Bearer Token：

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -H "Content-Type: application/json" \
     http://localhost:8080/mcp/tools
```

### 工具调用示例

#### 获取项目列表

```json
{
  "method": "tools/call",
  "params": {
    "name": "get_project_list",
    "arguments": {
      "page": 1,
      "per_page": 20
    }
  }
}
```

#### 创建工作项

```json
{
  "method": "tools/call",
  "params": {
    "name": "create_work_item",
    "arguments": {
      "title": "新功能开发",
      "description": "实现用户登录功能",
      "project_id": "project_123",
      "work_item_type_id": "story",
      "assignee_id": "user_456"
    }
  }
}
```

#### 记录工时

```json
{
  "method": "tools/call",
  "params": {
    "name": "create_workload",
    "arguments": {
      "work_item_id": "work_item_789",
      "hours": 4.5,
      "description": "完成登录页面开发",
      "date": "2024-01-15"
    }
  }
}
```

## 🏗️ 项目结构

```
PingCodeMcp/
├── cmd/
│   └── server/
│       └── main.go              # 服务器入口文件
├── internal/
│   ├── auth/                    # 认证相关
│   ├── config/                  # 配置管理
│   ├── models/                  # 数据模型
│   ├── server/                  # MCP服务器实现
│   ├── tools/                   # 工具实现
│   │   ├── tool_interface.go    # 工具接口定义
│   │   ├── project_manage.go    # 项目管理工具
│   │   ├── work_items_manage.go # 工作项管理工具
│   │   ├── workloads_manage.go  # 工时管理工具
│   │   ├── enterprise_users.go  # 用户管理工具
│   │   ├── user_info.go         # 用户信息工具
│   │   ├── ip_search.go         # IP查询工具
│   │   └── time_search.go       # 时间查询工具
│   └── utils/                   # 工具函数
├── pkg/
│   └── logger/                  # 日志包
│       ├── logger.go            # 日志器实现
│       └── logger_test.go       # 日志器测试
├── examples/                    # 示例代码
├── docs/                        # 项目文档
├── logs/                        # 日志文件目录
├── config.yaml                  # 配置文件
├── Dockerfile                   # Docker构建文件
├── Makefile                     # 构建脚本
├── go.mod                       # Go模块文件
└── README.md                    # 项目说明
```

## 🔧 开发指南

### 添加新工具

1. 在 `internal/tools/` 目录创建新工具文件
2. 实现 `MCPTool` 接口：

```go
type MCPTool interface {
    GetToolDefinition() mcp.Tool
    Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
    GetName() string
    GetDescription() string
}
```

3. 在工具注册函数中注册新工具

### 代码规范

- 遵循 Go 官方代码规范
- 使用结构化日志记录
- 所有日志记录必须单行格式
- 错误处理要完整
- 添加适当的单元测试

### 日志记录规范

```go
// ✅ 正确的日志记录 - 单行格式
log.With("user_id", 123, "action", "login", "ip", "192.168.1.1").Info("用户登录成功")

// ❌ 错误的日志记录 - 多行格式
log.With(
    "user_id", 123,
    "action", "login",
).Info("用户登录成功")
```

## 🧪 测试

### 运行测试

```bash
# 运行所有测试
make test

# 运行测试并生成覆盖率报告
make test-coverage

# 代码格式化
make fmt

# 代码检查
make lint
```

### 测试覆盖率

项目要求测试覆盖率不低于 80%。

## 📊 监控和日志

### 日志功能

- **结构化日志**: 使用 zap 库实现高性能结构化日志
- **每日轮转**: 自动按日期生成日志文件
- **文件压缩**: 自动压缩旧日志文件
- **自动清理**: 自动删除过期日志文件

### 日志文件命名

- 当前日志: `PingCodeMcp-2024-01-15.log`
- 压缩日志: `PingCodeMcp-2024-01-14.log.gz`

### 健康检查

Docker 容器支持健康检查：

```bash
# 检查服务状态
curl http://localhost:8080/health
```

## 🐳 Docker 部署

### 构建镜像

```bash
docker build -t pingcode-mcp-server .
```

### 运行容器

```bash
docker run -d \
  --name pingcode-mcp \
  -p 8080:8080 \
  -v $(pwd)/logs:/root/logs \
  -v $(pwd)/config.yaml:/root/config.yaml \
  pingcode-mcp-server
```

### Docker Compose

```yaml
version: '3.8'
services:
  pingcode-mcp:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./logs:/root/logs
      - ./config.yaml:/root/config.yaml
    environment:
      - PingCodeMcp_LOGGING_LEVEL=info
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 3s
      retries: 3
```

## 🔒 安全考虑

- 使用 HTTPS 进行生产部署
- 定期轮换 API Token
- 限制网络访问权限
- 监控异常访问行为
- 不在日志中记录敏感信息

## 📈 性能优化

- 使用连接池管理 HTTP 连接
- 实现请求缓存机制
- 优化日志写入性能
- 合理设置超时时间
- 监控内存使用情况

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

### 提交规范

```
类型: 简短描述

详细描述（可选）

类型包括：
- feat: 新功能
- fix: 修复
- docs: 文档
- style: 格式
- refactor: 重构
- test: 测试
```

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🆘 支持

- 📧 邮箱: support@pingcode.com
- 📖 文档: [PingCode API 文档](https://open.pingcode.com/docs)
- 🐛 问题反馈: [GitHub Issues](https://github.com/your-repo/PingCodeMcp/issues)

## 🔗 相关链接

- [PingCode 官网](https://pingcode.com)
- [PingCode API 文档](https://open.pingcode.com)
- [Model Context Protocol](https://modelcontextprotocol.io)
- [Go 官方文档](https://golang.org/doc)

## 📝 更新日志

### v1.0.0 (2024-01-15)
- 🎉 初始版本发布
- ✨ 支持完整的 MCP 协议
- 🔧 实现项目管理工具集
- 📊 添加结构化日志系统
- 🐳 支持 Docker 部署

---
