# Order Payment System（订单支付系统）

基于 Go 语言构建的高并发电商订单与支付系统，采用分层架构（Handler → Service → Repository）与依赖注入设计。系统涵盖用户鉴权、商品管理、购物车、订单交易、支付宝支付、退款审批、评价审核等完整交易闭环，并引入 Redis 缓存策略、RabbitMQ 异步削峰、分布式链路追踪等生产级特性。

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Docker](https://img.shields.io/badge/Docker-✓-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

---

## 🚀 技术栈

| 类别 | 技术 |
|------|------|
| **Web 框架** | [Gin](https://github.com/gin-gonic/gin) |
| **ORM** | [GORM](https://gorm.io/) + MySQL 8.0 |
| **缓存** | [go-redis](https://github.com/redis/go-redis) + Lua 脚本 |
| **消息队列** | RabbitMQ（AMQP）— 异步创单、延迟队列、死信队列 |
| **配置管理** | [Viper](https://github.com/spf13/viper) — YAML 配置文件 |
| **日志** | [Zap](https://github.com/uber-go/zap) — 结构化高性能日志 |
| **支付** | [Gopay](https://github.com/go-pay/gopay) — 支付宝支付集成 |
| **鉴权** | JWT（golang-jwt） + Bcrypt 密码加密 |
| **API 文档** | Swagger（swaggo） |
| **容器化** | Docker + Docker Compose |

---

## 🏗 项目结构

```text
order-payment-system/
├── cmd/                    # 程序入口
│   └── main.go             # 启动文件（优雅启停）
├── config/                 # 配置管理（Viper & YAML）
│   ├── config.go
│   └── config.yaml
├── internal/               # 内部业务逻辑
│   ├── app/                # 应用初始化、路由注册、后台 Job 启停
│   ├── errs/               # 统一错误码与 Sentinel Error
│   ├── handler/            # HTTP 控制器层（8 个 Handler）
│   ├── model/              # GORM 数据模型（9 张表）
│   ├── repository/         # 数据访问层（DAO）
│   ├── service/            # 核心业务逻辑层
│   └── types/              # 请求/响应 DTO 定义
├── job/                    # 异步任务
│   ├── cache_preheat.go    # 缓存预热（热卖商品预加载）
│   ├── order_create.go     # MQ 消费者（异步订单落库）
│   └── order_timeout.go    # 延迟队列消费者（超时关单 + 库存回滚）
├── pkg/                    # 公共基础设施
│   ├── database/           # MySQL / Redis / RabbitMQ 连接初始化
│   ├── jwt/                # JWT 工具
│   ├── logger/             # Zap 日志封装
│   ├── middleware/         # 中间件（CORS、限流、TraceID、鉴权、角色控制）
│   ├── response/           # 统一 JSON 响应封装
│   └── util/               # 工具函数（Bcrypt 等）
├── templates/              # 前端 HTML 模板（首页、登录页、支付页）
├── Dockerfile              # 多阶段构建镜像
├── docker-compose.yml      # 本地开发环境（MySQL + Redis + RabbitMQ + App）
└── go.mod                  # Go 模块依赖
```

---

## ✨ 核心功能

### 用户与认证
- 用户注册/登录，密码 Bcrypt 加密存储
- 基于 JWT 的无状态鉴权，支持普通用户 / 商家 / 管理员三种角色
- 角色中间件（`MerchantOnly` / `AdminOnly`）控制接口访问权限

### 商品与分类
- 商品 CRUD（商家/管理员）、分页查询、多条件筛选
- 商品分类树形结构管理
- 商品详情缓存策略（防穿透、防击穿、防雪崩）

### 购物车
- 添加/删除/修改/选中商品
- 购物车结算 → 一键生成订单

### 订单管理
- **同步创单**：直接下单并扣减库存
- **异步创单**：通过 RabbitMQ 削峰填谷，消息持久化 + Confirm 机制保障可靠性
- **订单状态机**：待支付 → 已支付 → 已发货 → 已收货 → 已完成，带状态转换校验
- **超时关单**：基于 RabbitMQ 延迟队列，30 分钟未支付自动取消并回滚库存
- **操作日志**：记录订单全生命周期变更（`order_logs` 表）

### 支付模块
- 支付宝扫码支付（Gopay 集成）
- 支付异步通知回调验签与订单状态更新
- 退款申请 → 商家审批 → 自动回滚库存

### 评价审核
- 用户评价提交后默认待审核
- 管理员可审批通过/驳回，前端仅展示已通过评价
- 评价列表支持分页查询

### 缓存策略（高可用设计）
- **缓存预热**：启动时预加载热卖商品与库存
- **防缓存穿透**：不存在商品写入短 TTL 的 EMPTY 标记
- **防缓存击穿**：singleflight 合并同一 Key 的并发回源请求
- **防缓存雪崩**：随机 TTL（Hour + Random）分散过期时间

### 库存并发控制
- 高并发下单使用 Redis + Lua 实现库存原子扣减（`HINCRBY`）
- Redis 预扣 + MySQL 最终扣减的双层库存模型

### 可观测性
- TraceID 中间件（支持透传 `X-Trace-ID`），日志记录 trace_id、path、method、耗时
- Zap 结构化日志，支持请求级别链路追踪

### 服务治理
- 优雅启停（`context` + `sync.WaitGroup` + 信号监听）
- Redis + Lua 令牌桶限流中间件
- CORS 跨域支持

---

## 📦 快速开始

### 方式一：Docker Compose（推荐，一键启动全套环境）

```bash
# 1. 克隆项目
git clone https://github.com/Tezzx/order-payment-system.git
cd order-payment-system

# 2. 一键启动所有服务（MySQL + Redis + RabbitMQ + App）
docker-compose up -d

# 3. 查看运行状态
docker-compose ps

# 4. 查看应用日志
docker-compose logs -f app

# 5. 停止所有服务
docker-compose down
```

启动后：
- **应用 API**：http://localhost:8081
- **Swagger 文档**：http://localhost:8081/swagger/index.html
- **RabbitMQ 管理面板**：http://localhost:15672（guest/guest）

### 方式二：本地开发运行

**前置条件**：Go 1.25+、MySQL 8.0、Redis 7+、RabbitMQ 3.x

```bash
# 1. 启动基础设施（仅启动 MySQL/Redis/RabbitMQ）
docker-compose up -d mysql redis rabbitmq

# 2. 修改配置（根据实际环境调整）
cp config/config.yaml config/config.yaml.example
vim config/config.yaml

# 3. 安装依赖并运行
go mod tidy
go run cmd/main.go
```

### 配置文件说明

编辑 `config/config.yaml`：

```yaml
server:
  port: 8081               # 服务端口

database:
  host: 127.0.0.1          # 本地开发用 127.0.0.1，Docker 环境用 mysql
  port: 3306
  user: root
  password: root
  dbname: order_payment

redis:
  host: 127.0.0.1          # 本地开发用 127.0.0.1，Docker 环境用 redis
  port: "6379"
  password: ""
  DB: 0
  PoolSize: 10

rabbitmq:
  host: 127.0.0.1          # 本地开发用 127.0.0.1，Docker 环境用 rabbitmq
  port: 5672
  user: guest
  password: guest
  vhost: "/"

jwt:
  key: "your-secret-key"   # JWT 签名密钥，生产环境请更换

AliPay:
  AppID: "your-app-id"
  PrivateKey: "your-private-key"
  AliPayKey: "your-alipay-public-key"
  notify_url: "https://your-domain/api/notify/alipay"
```

---

## 📖 API 文档

项目集成了 Swagger 自动生成 API 文档。启动服务后访问：

```
http://localhost:8081/swagger/index.html
```

主要 API 模块：

| 模块 | 前缀 | 说明 |
|------|------|------|
| 认证 | `/login`, `/register` | 用户登录注册 |
| 用户 | `/home/me` | 当前用户信息 |
| 商品 | `/home/`, `/home/search/goods`, `/home/goods/*` | 商品浏览与管理 |
| 订单 | `/home/order/*` | 订单创建、支付、发货、退款 |
| 支付 | `/home/pay/*` | 支付宝支付创建 |
| 购物车 | `/home/cart/*` | 购物车增删改查 |
| 地址 | `/home/address/*` | 收货地址管理 |
| 分类 | `/home/category/*` | 商品分类（管理员） |
| 评价 | `/home/review/*` | 评价提交与审核 |

---

## 🛠 开发计划

### ✅ 已完成
- [x] 用户注册登录 + JWT 鉴权 + 角色控制
- [x] 商品 CRUD（分页、筛选、分类）
- [x] 购物车管理 + 结算下单
- [x] 订单状态机（7 种状态，带转换校验）
- [x] 支付宝扫码支付 + 回调通知
- [x] 退款申请 → 审批 → 库存回滚
- [x] 评价提交 → 审核 → 展示
- [x] 异步创单（RabbitMQ）
- [x] 超时关单（延迟队列 + 库存回滚）
- [x] 缓存预热 + 防穿透/击穿/雪崩
- [x] Redis + Lua 库存原子扣减
- [x] TraceID 链路追踪
- [x] Redis + Lua 令牌桶限流
- [x] 优雅启停
- [x] 统一错误码
- [x] Docker 容器化部署
- [x] Swagger API 文档
- [x] CORS 跨域中间件

### 🔲 待完成
- [ ] 分布式事务（Seata/Saga 模式）
- [ ] 分布式限流升级（Redis Cluster 支持）
- [ ] 读写分离（MySQL 主从）
- [ ] 压测报告与性能调优文档
- [ ] Kubernetes 部署配置（Helm Chart）
- [ ] 前端 SPA 重构（Vue/React）

---

## 📄 License

MIT License
