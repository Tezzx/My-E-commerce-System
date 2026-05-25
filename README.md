# Order Payment System (订单支付系统)

基于 Go 语言构建的高并发订单与支付系统，采用了典型的微服务/模块化分层设计，内置了完整的基础设施集成，包括关系型数据库、缓存中间件和消息队列。

## 🚀 技术栈

- **Web 框架**: [Gin](https://github.com/gin-gonic/gin)
- **数据库 ORM**: [GORM](https://gorm.io/) + MySQL
- **缓存**: [Go-Redis](https://github.com/redis/go-redis) 
- **消息队列**: RabbitMQ (`streadway/amqp`)
- **配置管理**: [Viper](https://github.com/spf13/viper)
- **日志**: [Zap](https://github.com/uber-go/zap)
- **支付集成**: [Gopay](https://github.com/go-pay/gopay)
- **其他**: JWT (鉴权), Bcrypt (密码加密), Zap (日志收集)

## 🏗 项目结构

```text
├── cmd/                # 程序入口
│   └── main.go         # 启动文件
├── config/             # 配置管理 (Viper & YAML config)
├── internal/           # 内部业务逻辑
│   ├── app/            # 应用启停、路由与模块初始化
│   ├── errs/           # 自定义错误处理
│   ├── handler/        # 控制器层 (HTTP handlers: 订单, 支付, 用户, 商品)
│   ├── model/          # 数据库模型层 (GORM Models)
│   ├── repository/     # 数据访问层 (DAO)
│   ├── service/        # 核心业务逻辑层
│   └── types/          # 请求响应等结构体定义
├── job/                # 异步任务/定时任务 (如缓存预热、订单超时取消、异步创单)
├── pkg/                # 公共基础组件
│   ├── database/       # DB、Redis、MQ初始化
│   ├── jwt/            # JWT工具
│   ├── logger/         # Zap日志封装
│   ├── middleware/     # Gin中间件 (CORS、限流、日志、链路追踪、鉴权)
│   ├── response/       # 统一响应封装
│   └── util/           # 实用工具包 (Bcrypt等)
├── templates/          # 前端 HTML 模板 (登录, 首页, 支付页)
├── docker-compose.yml  # 本地环境编排 (MySQL, Redis, RabbitMQ等)
├── Dockerfile          # 服务容器化构建
└── go.mod              # 依赖管理
```

## ✨ 核心功能模块

1. **用户模块**: 用户注册、登录、基于 JWT 的认证访问，密码 bcrypt 加密。
2. **商品模块**: 商品列表展示与秒杀/抢购概念。
3. **订单模块**: 
   - 包含同步或通过 MQ 异步创建订单处理。
   - 订单超时未支付自动取消处理（Job 处理）。
   - 库存扣减逻辑。
4. **支付模块**: 集成了三方支付（如支付宝/微信支付等），结合订单状态完成支付回调与更新。
5. **任务调度 (Jobs)**:
   - `cache_preheat.go`: 缓存预热，防止缓存击穿。
   - `order_create.go`: 通过 MQ 消费实现订单异步创建，削峰填谷。
   - `order_timeout.go`: 处理超时未支付订单，自动变更状态并回滚库存。

## 📦 快速启动

### 1. 环境准备

确保本地已安装 [Docker](https://www.docker.com/) 和 Docker Compose，以及 Go 1.25+ 环境。

```bash
# 启动依赖服务 (MySQL, Redis, RabbitMQ)
docker-compose up -d
```

### 2. 配置文件

根据本地环境修改 `config/config.yaml`（如果是首次启动，注意初始化相关数据库实例和表）。

### 3. 运行服务

```bash
go mod tidy
go run cmd/main.go
```
服务默认会在配置的端口（如 `8080`）启动。

## 📌 TODO & 优化计划 (Future Enhancements)

开发过程中的待办与优化项记录：

- [ ] **高并发库存扣减**: 使用 Redis + Lua 脚本来保证超并发环境下的库存扣减原子性，避免超卖问题。
- [ ] **登录注册优化**: 用户登录注册功能引入 Redis 缓存与校验，减少对数据库的查询压力。
- [ ] **创单流程原子化**: 异步创建订单时，将 Redis 缓存更新和发送 MQ 消息两个操作保证原子化或最终一致性。
- [ ] **用户模型完善**: `user_model` 补充 User ID (用户编号) 等基础字段规范与索引。
- [ ] **限流升级**: 现有的基于令牌桶的限流器改进，支持更精准的分布式限流（如基于 Redis 的 Token Bucket 算法）。
