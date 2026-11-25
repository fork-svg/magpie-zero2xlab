# Magpie Zero2XLab Gateway

一个 Go 编写的 API 网关服务，将 OpenAI 兼容的请求转发到 zero2xlab (之江实验室 AI 平台)。

## 功能特性

- **用户认证系统** - 注册/登录，JWT token 管理
- **API Key 管理** - 为每个用户生成/管理 API Key
- **OpenAI 兼容接口** - 支持 `/v1/chat/completions` 等标准端点
- **请求转发** - 代理请求到 zero2xlab API
- **使用统计** - 记录每个用户的请求量和 token 使用量
- **流式响应** - 支持 SSE 流式输出

## 快速开始

### 1. 构建

```bash
# 安装依赖
go mod tidy

# 构建
go build -o gateway ./cmd/main.go
```

### 2. 配置环境变量

```bash
# 必需：zero2xlab 的 API Key
export ZERO2XLAB_API_KEY="your-zero2xlab-api-key"

# 可选：JWT 密钥（生产环境必须修改）
export JWT_SECRET="your-jwt-secret"

# 可选：服务端口
export PORT="8080"
```

### 3. 运行

```bash
./gateway
```

## API 文档

### 用户注册

```bash
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'
```

响应：
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": 1,
    "username": "testuser",
    "email": "test@example.com"
  }
}
```

### 用户登录

```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'
```

### 创建 API Key

```bash
curl -X POST http://localhost:8080/api/api-keys \
  -H "Authorization: Bearer <jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{"name": "My API Key"}'
```

响应：
```json
{
  "key": "mg-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "name": "My API Key",
  "created_at": "2024-01-01T00:00:00Z"
}
```

### 列出 API Keys

```bash
curl http://localhost:8080/api/api-keys \
  -H "Authorization: Bearer <jwt-token>"
```

### 使用 API Key 调用 Chat

```bash
curl -X POST http://localhost:8080/api/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: mg-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" \
  -d '{
    "model": "oneastronomy-r1",
    "messages": [
      {"role": "user", "content": "你好"}
    ],
    "stream": false
  }'
```

### 查看请求日志

```bash
curl http://localhost:8080/api/logs \
  -H "Authorization: Bearer <jwt-token>"
```

## 项目结构

```
magpie-zero2xlab/
├── cmd/
│   └── main.go          # 主入口
├── config/
│   └── config.go        # 配置管理
├── internal/
│   ├── api/
│   │   └── handlers.go  # HTTP 处理器
│   ├── auth/
│   │   ├── jwt.go       # JWT 认证
│   │   └── password.go  # 密码哈希
│   ├── db/
│   │   └── database.go  # 数据库操作
│   ├── models/
│   │   └── user.go      # 数据模型
│   └── proxy/
│       └── proxy.go     # 代理转发
├── go.mod
└── README.md
```

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `PORT` | 服务端口 | `8080` |
| `HOST` | 服务地址 | `0.0.0.0` |
| `JWT_SECRET` | JWT 密钥 | `change-this-secret-in-production` |
| `ZERO2XLAB_BASE_URL` | zero2xlab API 地址 | `https://oneastronomy.zero2x.org/api/platform/v1` |
| `ZERO2XLAB_API_KEY` | zero2xlab API Key | - |
| `DB_PATH` | SQLite 数据库路径 | `./data/gateway.db` |

## License

MIT
