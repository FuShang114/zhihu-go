# zhihu-go 学习路线图（演进式：痛点驱动）

> 这是一份"边学边建"的路线图：从最简单的单文件 HTTP 服务起步，每一层架构都因为**上一版的真实痛点**才被引入。
> 最终目标：手写 DDD 四层架构的知乎后端微服务（用户/内容/互动/Feed 召回/通知/搜索）。
> **一节课 = 一次 git 提交**，本文件随进度更新。最终会把整个学习过程可视化为教学网页。

## 技术栈（终点形态）

Go 1.27 · net/http → Gin · gRPC + Protobuf · MySQL 8（按服务分库）· Redis（缓存/Stream 事件总线）· 手写 DDD 四层（interfaces / application / domain / infrastructure，domain 不依赖任何层——依赖倒置）

## 课程表

- [x] **L1 net/http 起步**（2026-10-08）
  单文件 `main.go`：`GET /ping` + `POST /register`（内存 map 存储）。
  学到：`http.Handler` 接口与 ServeHTTP、响应=往 `ResponseWriter` 写字节（return 无法跨网络）、struct/json tag、map 与 comma-ok 惯用法、`any`、error 处理、参数校验必须先于业务逻辑、业务码 0/1xxxx/2xxxx 规范。
  经典 bug 存档：校验后漏 `return` 导致一个响应写出两段 JSON、id 被烧、空用户名入库。
- [ ] **L2 登录态**（进行中）
  - [ ] L2.1 bcrypt 密码哈希 + `POST /login`（正在做）：`go get x/crypto/bcrypt`、[]byte 转换、盐与 cost、防用户枚举（不存在/密码错同一报错）
  - [ ] L2.2 JWT 登录态 + `sync.Mutex`（并发写 map 的坑）：中间件雏形（闭包包装 handler）
  - [ ] L2.3 标准库中间件的繁琐 → 引入 Gin（终于知道框架省了什么）：参数绑定/校验、路由分组
- [ ] **L3 MySQL 持久化**：重启数据不丢（痛点：内存 map 一重启全没）；database/sql → GORM；DSN 时区坑
- [ ] **L4 拆包**：main.go 膨胀成面条 → handler / service / repository 三层
- [ ] **L5 DDD 四层**：依赖倒置（接口在 domain、实现在 infrastructure）、PO/DO 分离、依赖注入手写、单测（stub/mock）
- [ ] **L6 内容模块**：Question/Answer 聚合根、游标分页、批量查作者防 N+1、Article 自助练习
- [ ] **L7 微服务拆分**：go.work 多 module、proto 契约、gRPC 打通 user↔content、防腐层
- [ ] **L8 网关 + 缓存**：统一入口 8080、JWT 认证上收、BFF 聚合、cache-aside + singleflight
- [ ] **L9 互动 + 事件**：投票状态机、两级评论、同步计数"埋雷" → Redis Stream 事件总线"拆雷"（消费者组/幂等/最终一致）
- [ ] **L10 Feed 召回 + 热榜 + 搜索**：推拉结合 ZSET 收件箱、多路召回合并、时间衰减热榜、MySQL ngram 中文全文
- [ ] **L11 收尾**：trace_id 全链路、令牌桶限流、pprof + 压测、README 收官
- [ ] **L12 学习过程可视化网页**：从 git 提交历史生成课程时间线页面

## 全局约定

- 响应体 `{"code":0,"msg":"ok","data":...}`；错误码 1xxxx 通用 / 2xxxx user / 3xxxx content / 4xxxx interaction
- 端口：user 8001/9001、content 8002/9002、gateway 8080、Redis 6379、MySQL 3306
- 服务重启丢内存数据（L1 观察到的痛点，L3 解决）
- 环境：Windows 11 + GoLand + goproxy.cn；MySQL/Memurai/protoc 待安装（L3/L7 前装）

## 测试脚本备忘（每课收官跑一遍）

```bash
B=http://localhost:8001
curl -s $B/ping
curl -s -X POST $B/register -d '{"username":"alice","password":"123"}'
curl -s -X POST $B/register -d '{"username":"alice","password":"456"}'  # 期望 20001
curl -s -X POST $B/register -d '{"username":"","password":""}'          # 期望 10001
curl -s -X POST $B/login    -d '{"username":"alice","password":"123"}'  # L2 起
curl -s $B/register                                                      # 期望 10001 方法检查
```
