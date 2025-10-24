# 后端重构方案（Go + SQLite + Docker + Chi）（可勾选）

> 目标：彻底取消浏览器扩展与 Popup，重构为纯后端系统，提供 Dashboard 管理与“一键切换”（自动化登录）能力。

---

## 0. 决策清单（先选）

- [x] 取消浏览器扩展与 Popup（不再使用扩展形态）
- [x] 不考虑数据迁移（当前无存量数据）
- [x] Web 框架：`chi`
- [x] 数据库：`SQLite`（驱动：`modernc.org/sqlite`，免 CGO）
- [x] DB 访问：`sqlx`
- [x] 自动化引擎：`chromedp`（通过 Chrome DevTools Protocol）
- [x] Chrome 连接模式：本机 Chrome 开启远程调试端口（默认 9222）
- [x] Dashboard 技术栈：`html/template` + `htmx`/`Alpine.js`（无 Node）
- [x] 部署：多阶段 `Dockerfile` + `docker-compose`（卷持久化 `./data/mss.db`）
- [ ] 认证方式
  - [ ] Dashboard：管理员用户名/密码（Cookie 会话）
  - [ ] 切换 API：仅 Dashboard 内部调用（CSRF 防护 + Origin 校验）
- [ ] MVP 站点适配器范围
  - [ ] 仅 `example` 示例
  - [ ] 同时实现实际目标站点：`__________`

---

## 1. 目录结构与基础骨架

- [ ] 创建 `server/` 目录与模块初始化
  - [ ] `server/go.mod`（chi、sqlx、sqlite 依赖）
  - [ ] `server/cmd/mss-server/main.go`（启动、路由、中间件、健康检查）
- [ ] 路由与中间件（`chi`）
  - [ ] `RequestID/Logger/Recoverer/Timeout`
  - [ ] CORS（按需）
- [ ] 健康检查接口
  - [ ] `GET /healthz` 返回 `ok`

---

## 2. 数据库与迁移（SQLite）

- [ ] 迁移脚本目录：`server/internal/migrate/*.sql`
- [ ] 表结构
  - [ ] `sites(key PK, name, login_url, created_at, updated_at)`
  - [ ] `accounts(id PK, site_key FK, username, password, extra(JSON), created_at, updated_at)`（索引：`(site_key, username)`）
  - [ ] `active_accounts(site_key PK, account_id NULL, updated_at)`
- [ ] 启动自动迁移（`migrate.Apply()`）
- [ ] `server/internal/store/`
  - [ ] `Open()` 连接与 PRAGMA
  - [ ] CRUD 与事务封装

---

## 3. REST API（MVP）

- [ ] 站点
  - [ ] `GET /api/sites` → `[ { key,name,loginUrl } ]`
  - [ ] `GET /api/sites/{key}` → `{ key,name,loginUrl }`
- [ ] 账号
  - [ ] `GET /api/sites/{key}/accounts` → `{ accounts:[...], activeId }`
  - [ ] `POST /api/sites/{key}/accounts` → 新增
  - [ ] `PUT /api/sites/{key}/accounts/{id}` → 编辑
  - [ ] `DELETE /api/sites/{key}/accounts/{id}` → 删除
- [ ] 活跃账号
  - [ ] `GET /api/sites/{key}/active-account` → `{ accountId }`
  - [ ] `PUT /api/sites/{key}/active-account` → `{ accountId }`
- [ ] 切换执行
  - [ ] `POST /api/sites/{key}/switch` → `{ accountId, options? }`（触发自动化；返回执行结果/日志）
- [ ] 统一返回：`{ ok: boolean, data?: any, error?: string }`
- [ ] OpenAPI 草案（可选）

---

## 4. 自动化切换（chromedp）

- [ ] Chrome 连接发现
  - [ ] 读取 `MSS_CDP_DISCOVERY`（默认 `http://127.0.0.1:9222/json/version`）
  - [ ] 获取 `webSocketDebuggerUrl` 建立连接
- [ ] 会话与上下文
  - [ ] 独立上下文（可选独立临时 profile）
  - [ ] 超时/取消（默认 90s）
- [ ] 站点适配器（`server/internal/adapter/`）
  - [ ] 接口：`Key() Name() LoginURL() Logout(ctx, cdp) Login(ctx, cdp, cred)`
  - [ ] 示例：`example`（导航登录页→填充→提交）
- [ ] 任务调度
  - [ ] 按站点串行化
  - [ ] 日志与错误归档

---

## 5. Dashboard（无 Node）

- [ ] 模板：`server/web/templates/`，静态：`server/web/static/`
- [ ] 页面
  - [ ] 站点列表/详情（`key/name/loginUrl`）
  - [ ] 账号列表（新增/编辑/删除/设为活跃）
  - [ ] 触发“切换”与结果提示
- [ ] 交互
  - [ ] `htmx` 局部刷新（CRUD/设为活跃）
  - [ ] 简单表单校验（用户名必填）

---

## 6. 安全与认证

- [ ] 管理员登录
  - [ ] `MSS_ADMIN_USER/MSS_ADMIN_PASS` 初始凭据
  - [ ] 会话 Cookie（Secure、HttpOnly、SameSite）
- [ ] CSRF
  - [ ] 表单/接口保护（`Origin` 检查 + Token）
- [ ] 切换接口保护
  - [ ] 仅 Dashboard 内部可调用（会话 + CSRF）
- [ ] 速率限制（可选）

---

## 7. 配置与部署

- [ ] 环境变量
  - [ ] `MSS_LISTEN_ADDR=:8080`
  - [ ] `MSS_DB_PATH=./data/mss.db`
  - [ ] `MSS_CDP_DISCOVERY=http://127.0.0.1:9222/json/version`
  - [ ] `MSS_ADMIN_USER=admin` / `MSS_ADMIN_PASS=...`
  - [ ] `MSS_LOG_LEVEL=info`
- [ ] Docker 化
  - [ ] `server/Dockerfile`（多阶段 + distroless/alpine）
  - [ ] `docker-compose.yml`（`8080:8080`、卷 `./data:/data`、健康检查 `/healthz`）
- [ ] Windows Chrome 启动示例
  - [ ] `chrome.exe --remote-debugging-port=9222 --user-data-dir="%LOCALAPPDATA%\ChromeDevProfile"`

---

## 8. 日志与可观测

- [ ] 结构化日志（level、req_id、site_key、account_id）
- [ ] 切换任务日志持久化（可选）
- [ ] `pprof`（可选，开发模式）

---

## 9. 验收测试（手测）

- [ ] 站点：创建 `example`（`key/name/loginUrl`）
- [ ] 账号：新增/编辑/删除
- [ ] 活跃：设为活跃后能被读取
- [ ] 切换：触发后打开登录页并自动填充提交（示例逻辑）
- [ ] 错误：网络异常/登录失败时提示清晰
- [ ] 并发：连续触发多次串行执行

---

## 10. 里程碑（建议）

- [ ] M1（1.5 天）：服务骨架（chi）+ SQLite 迁移 + 基础 REST + Docker/Compose + `/healthz`
- [ ] M2（1 天）：`chromedp` 接入 + `example` 站点适配器 + `POST /switch` + 基础日志
- [ ] M3（1 天）：Dashboard（站点/账号/活跃/切换）+ 会话登录 + CSRF
- [ ] M4（0.5 天）：文档与部署指南、冒烟测试与打磨

---

## 11. 风险与备选

- [ ] Chrome 远程调试端口被占用 → 文档提供备用端口
- [ ] 目标站风控/验证码 → 预留人工确认或半自动策略
- [ ] SQLite 并发写压力 → 当前可接受；后续可升级 Postgres

---

## 12. 交付清单

- [ ] 代码：`server/`（chi 路由、store、migrate、adapter、web）
- [ ] Docker：`server/Dockerfile`、`docker-compose.yml`
- [ ] 配置：`.env.example` 或 README 环境变量说明
- [ ] 文档：运行指南、开发指南、站点适配器规范
