# 灵枢 LingShu v1.4 - 安全硬化 + 营销门面

> 整理人：Kiro
> 日期：2026-06-18
> 上一版：v1.3 核心已上线，6 个 BUG 收口（BUG1/2/3 + N1-N5 + H1-H4 + F1-F5）
> 本版主题：把"能跑"升级为"能投产"——堵住安全裸奔点 + 加营销门面承接新用户

---

## 一、决策与范围

**做什么**：

| 阶段 | 内容 | 预计 |
|---|---|---|
| **S1 安全硬化** | CORS 白名单、Body 大小限制、渠道自愈、错误体统一 | 1-2 天 |
| **S2 营销门面** | 公开价格表页 + 公共 API + 顶部导航 | 1 天 |

**不做什么**（明确推后，避免 agent 自由扩展）：

- 用户增长（签到/邀请/注册送余额）→ v1.5
- 监控大盘 / Prometheus → v1.5（先有流量再做）
- 管理员 2FA → v1.5
- 模型分组 / 令牌分组 → v1.5（先看用户实际需要）

**铁律保持不变**：
- 计费公式 `charge = base_cost × rate_multiplier`
- 用户端禁露 `base_cost / rate_multiplier / gross_profit`（v1.1 已收口）
- 软删除模式（v1.2 H1 已建立）

---

## 二、S1 安全硬化（按 S1.1 → S1.4 顺序）

### S1.1【🔴 P0】CORS 白名单

**现状**：`backend/internal/server/server.go:154` 的 `cors()` 中间件直接回显请求头的 Origin —— 任何站点都能跨域访问 admin/user/gateway API。

**步骤**：
1. `backend/internal/config/config.go` 加字段：
   ```go
   AllowedOrigins []string  // 解析逗号分隔的 ALLOWED_ORIGINS env
   ```
   `envStringSlice("ALLOWED_ORIGINS", []string{"*"})` 默认 `*`（保持向后兼容，部署时不破坏）。
2. `backend/internal/server/server.go` 把 `cors` 改成接受 `[]string` 参数的 factory：
   ```go
   func corsWith(allowed []string) func(http.Handler) http.Handler {
       allowAll := len(allowed) == 1 && allowed[0] == "*"
       set := make(map[string]bool, len(allowed))
       for _, o := range allowed { set[strings.TrimSpace(o)] = true }
       return func(next http.Handler) http.Handler {
           return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
               origin := r.Header.Get("Origin")
               if origin != "" && (allowAll || set[origin]) {
                   w.Header().Set("Access-Control-Allow-Origin", origin)
                   w.Header().Set("Vary", "Origin")
                   w.Header().Set("Access-Control-Allow-Credentials", "true")
                   w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
                   w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
               }
               if r.Method == http.MethodOptions {
                   w.WriteHeader(http.StatusNoContent)
                   return
               }
               next.ServeHTTP(w, r)
           })
       }
   }
   ```
   `New()` 改 `r.Use(corsWith(cfg.AllowedOrigins))`。
3. README 加配置说明：`ALLOWED_ORIGINS=https://lingshu.example.com,https://admin.lingshu.example.com`，**生产强制配具体值**。

**验收**：
- 不配 `ALLOWED_ORIGINS`：旧行为，任意 Origin 都能访问（兼容）
- 配 `ALLOWED_ORIGINS=https://lingshu.example.com`：从别的域名 fetch 浏览器拒绝（无 Allow-Origin 头）

---

### S1.2【🔴 P0】Gateway Body 大小限制

**现状**：`/v1/chat/completions` 没限 body，恶意客户端可以塞 100MB 消息撑爆 Go 内存。

**步骤**：
1. `config.go` 加 `GatewayMaxBodyBytes int64`，默认 `envInt64("GATEWAY_MAX_BODY_BYTES", 2*1024*1024)`（2MB）。
2. `backend/internal/middleware/` 新建 `max_body.go`：
   ```go
   func MaxBody(limit int64) func(http.Handler) http.Handler {
       return func(next http.Handler) http.Handler {
           return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
               r.Body = http.MaxBytesReader(w, r.Body, limit)
               next.ServeHTTP(w, r)
           })
       }
   }
   ```
3. `server.go` 的 `/v1` 路由组加 `r.Use(middleware.MaxBody(cfg.GatewayMaxBodyBytes))`。
4. `/v1/chat/completions` handler 读 body 失败时，如果是 `MaxBytesError`，返 HTTP 413 + `{"error":{"message":"request body exceeds 2 MiB","type":"request_too_large"}}`。

**验收**：发 3MB body 应该收到 413（用 `head -c 3000000 /dev/urandom | base64` 或类似构造）。

---

### S1.3【🟡 P1】渠道熔断自愈

**现状**：渠道 `fail_count >= fail_threshold` 后 `health='unhealthy'`，但**没看到自动恢复逻辑**，需要管理员手动点测试或重启服务。

**步骤**：
1. `backend/internal/job/` 新建 `channel_healer.go`：
   - 每 `CHANNEL_HEALER_INTERVAL_SECONDS`（默认 300 秒）扫一次所有 `health='unhealthy' AND deleted_at IS NULL` 的渠道
   - 对每个 unhealthy 渠道调 `channelService.Test(ctx, id, baseURL)`（复用已有方法）
   - 连续 **3 次** 成功后改回 `health='healthy', fail_count=0`
   - "连续 3 次"用 Redis 计数：`channel_heal:{channelID}` INCR + EXPIRE 30 分钟（防止次数无限累积）
2. `bootstrap` 启动时如果 `CHANNEL_HEALER_ENABLED=true`（默认 true）就 `NewChannelHealer(...).Start(ctx)`。
3. `config.go` 加：
   - `ChannelHealerEnabled bool`（默认 true）
   - `ChannelHealerIntervalSeconds int`（默认 300）
   - `ChannelHealerSuccessThreshold int`（默认 3）

**验收**：
- 让一个渠道连续失败到 unhealthy
- 修好上游
- 等 5-15 分钟（取决于阈值），观察 health 自动变回 healthy 且 fail_count=0

---

### S1.4【🟡 P1】统一错误体格式

**现状**：v1.3 已经把上游错误透传成 OpenAI 格式 `{"error":{"message","type","code"}}`，但 LingShu 自身的错误（401 invalid api key、402 insufficient balance、429 rate limited、413 request too large）还是简陋的 `{"error":"text"}`，前端不好统一处理。

**步骤**：
1. `backend/internal/pkg/httpx/` 加 `ErrorJSON(w, status, errorType, message, code)`：
   ```go
   func ErrorJSON(w http.ResponseWriter, status int, errType, message, code string) {
       w.Header().Set("Content-Type", "application/json")
       w.WriteHeader(status)
       _ = json.NewEncoder(w).Encode(map[string]any{
           "error": map[string]any{
               "message": message,
               "type":    errType,
               "code":    code,
           },
       })
   }
   ```
2. `/v1/*` 路径（gateway 网关接口）的所有错误用 `ErrorJSON`：
   - 401: `("invalid_api_key", "invalid api key", "invalid_api_key")`
   - 402: `("insufficient_balance", "insufficient account balance", "insufficient_balance")`
   - 413: `("request_too_large", "request body exceeds N bytes", "request_too_large")`
   - 429: `("rate_limit_exceeded", "rate limit exceeded", "rate_limit_exceeded")`
   - 502: `("upstream_unavailable", "no healthy upstream channel", "no_healthy_channel")`
3. `/api/admin/*`、`/api/user/*` 内部 API 保持现有 `{"error":"text"}` 不动（管理端/用户端自己消费，统一改成本高）。
4. 修改 `middleware/auth.go`、`middleware/apikey_auth.go`、`handler/gateway/handler.go` 改用 `ErrorJSON`（gateway 路径）。

**验收**：用错误的 key 调 `/v1/chat/completions` 应该返：
```json
{"error":{"message":"invalid api key","type":"invalid_api_key","code":"invalid_api_key"}}
```

---

## 三、S2 营销门面（按 S2.1 → S2.3 顺序）

### S2.1【后端】公共模型清单端点

**目的**：公开价格表页要的数据源，无需登录。

**步骤**：
1. `backend/internal/handler/public/` 新建 `pricing.go`：
   ```go
   type PublicModelDTO struct {
       ID                string `json:"id"`
       PublicName        string `json:"public_name"`
       Type              string `json:"type"`
       Group             string `json:"group,omitempty"`
       BillingMode       string `json:"billing_mode"`
       // 公示价 = 上游单价 × 倍率（脱敏：永远不暴露 base_cost / rate_multiplier 本身）
       InputPricePer1M   string `json:"input_price_per_1m"`   // 千进制 → 百万
       OutputPricePer1M  string `json:"output_price_per_1m"`
       PricePerCall      string `json:"price_per_call,omitempty"`
       Currency          string `json:"currency"`             // 固定 "USD" 或读 setting
   }
   ```
2. handler：查所有 `status='enabled' AND deleted_at IS NULL` 且至少绑了一个 enabled 渠道的模型（这是已有逻辑），按 sort_order 排序返回。
   - **公示价计算公式**：`input_price_per_1m = (input_price_per_1k * rate_multiplier).round(6) * 1000`
   - **绝对不要暴露 base_cost、rate_multiplier 这两个字段**
3. `backend/internal/server/server.go` 加路由（**不挂 JWTAuth**）：
   ```go
   r.Get("/api/public/models", publicHandler.ListModels)
   r.Get("/api/public/site-info", publicHandler.SiteInfo)  // 站点名、登录链接、注册开关
   ```
4. `SiteInfo` 从 `system_settings` 读 `site_name`、`registration_enabled`、`contact_info` 返给前端。

**验收**：`curl http://155.248.195.94:19080/api/public/models`（无任何鉴权）应返回模型清单，**响应体里不能 grep 到 `base_cost / rate_multiplier / gross_profit`**。

---

### S2.2【前端】价格表页 /pricing

**目的**：未登录用户也能看到模型清单和价格，作为营销门面。

**步骤**：
1. `frontend/user/src/routes/` 新建 `pricing.tsx`：
   - 调 `/api/public/models` 和 `/api/public/site-info`
   - 按 `group` 分组展示（没有 group 的归到"通用"）
   - 表格列：模型名 / 类型 / 计费方式 / 输入价(/1M) / 输出价(/1M) / 操作（"立即接入"按钮跳 `/login`）
   - 顶部 hero：`<h1>{siteName} - AI API 价格表</h1>` + 简介 + "免费注册"按钮（若 registration_enabled）
   - 卡片设计参考你给的 sub-api 截图（绿色高亮、模型卡片网格）
2. `frontend/user/src/App.tsx`（或 router 配置）加 `<Route path="/pricing" element={<PricingPage />} />`，**不要包在 RequireAuth 内**。
3. SEO：`pricing.tsx` 顶部用 `document.title = ...` 设置标题；index.html 加 `<meta name="description" content="LingShu - 私人 AI API 聚合网关，支持 OpenAI / Anthropic 模型转发">`。

**验收**：
- 浏览器无痕模式打开 `/pricing` 能看到价格表（不被强制登录跳转）
- 价格数字应该是上游价 × 倍率后的值（per 1M）
- 浏览器 Network 看响应**没有** `base_cost / rate_multiplier / gross_profit` 字段

---

### S2.3【前端】顶部导航统一

**目的**：让 `/pricing`、`/login`、`/dashboard` 之间能互相跳转。

**步骤**：
1. `frontend/user/src/components/` 加 `site-nav.tsx`：
   - 左侧：站点 logo + name（点跳 `/`）
   - 中间：链接 `/pricing` / `/docs`（docs 暂时跳 `/api.md` 静态 README）
   - 右侧：未登录显示 "登录"，已登录显示 "控制台" + 用户名
2. 把这个 nav 加到 `pricing.tsx`、`login.tsx` 顶部。已登录后的页面（dashboard 等）保持原有 sidebar 不变。
3. 首页路由 `/`：未登录 → 跳 `/pricing`；已登录 → 跳 `/dashboard`。

**验收**：从 `/pricing` 能一键跳 `/login`，从 `/login` 能跳 `/pricing` 不丢登录状态。

---

## 四、执行顺序与验收

### Agent 执行顺序

**严格按 S1.1 → S1.2 → S1.3 → S1.4 → S2.1 → S2.2 → S2.3 顺序**，每个小节内部步骤一气呵成，做完跳下一节。

### 收尾必做

1. `cd backend && go build ./... && go vet ./... && go test ./...` 全绿
2. `cd frontend && npm --workspace @lingshu/admin run build && npm --workspace @lingshu/user run build` 全绿
3. 给 S1.3 写一个 `channel_healer_test.go`，至少覆盖"unhealthy 连续 3 次成功 → healthy"这条路径
4. 给 S2.1 写一个 handler test，验证响应体 **不包含** `base_cost / rate_multiplier / gross_profit`
5. `git commit` 但 **不要 push**，等 Kiro 验收后由用户决定推

### 给 agent 的执行指令（直接复制）

```
按 docs/LingShu-v1.4-安全硬化与营销门面.md 顺序执行
S1.1 CORS 白名单 → S1.2 Body 限制 → S1.3 渠道自愈 → S1.4 错误体统一 → S2.1 公共 API → S2.2 价格表页 → S2.3 顶部导航
每节按文档精确步骤做不要自由扩展
不要做文档第一节里明确推后的项（增长/大盘/2FA/分组）
做完跑全套构建测试 commit 但不要 push 等我推
S2.1 的脱敏测试必须真跑过 响应体里不能出现 base_cost / rate_multiplier / gross_profit
```

---

## 五、v1.5 前瞻（不属于本轮，仅提示 agent 不要做）

- 用户增长：注册送余额、邀请码（双方奖励）、每日签到、余额低站内通知
- 管理员 2FA（TOTP）+ IP 白名单
- 路由策略明文化（同 public_name 多渠道时按权重 + 健康度选）+ 文档
- Prometheus `/metrics` + Grafana 大盘
- 模型分组 / 令牌分组 / 用户分组
- 公告富文本编辑器（v1.2 只做了 markdown 预览，没编辑器）
- API SDK 文档（Swagger UI）

这些都不要做。
