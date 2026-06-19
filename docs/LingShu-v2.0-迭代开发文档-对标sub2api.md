# 灵枢（LingShu）v1.6 → v2.0 迭代开发文档（对标 sub2api 全维度版）

> **本文档供执行 agent 直接落地。** 基于对参考项目 **sub2api**（Wei-Shaw/sub2api v0.1.137，已克隆至 `D:\code\LingShu\Sub`）的全维度源码研究，逐子系统对照 **LingShu 现状**（`D:\code\LingShu\backend` + `frontend/admin` + `frontend/user`），给出可执行的差距清单与里程碑路线。
>
> **三条铁律（任何里程碑都不得违反）：**
> 1. **计费语义不可变**：`charge = base_cost × rate_multiplier`。所有计费改动只能往 `base_cost` 里加 token 分量，**绝不改公式、不改扣费时机（预扣）**。
> 2. **用户端零敏感字段**：用户侧 API/页面绝不返回 `base_cost / rate_multiplier / gross_profit / 上游成本 / upstream_model_name`。
> 3. **存量计费不回退**：任何迁移（新增列、分组、缓存价默认 0）必须保证存量请求计费结果逐分不变。
>
> **裁剪原则**：sub2api 是商业化 SaaS，含支付/订单/套餐/联盟返利/促销码/风控等中台。LingShu 是私有网关，**这些默认不做**，只补「计费正确性 + 调度健壮性 + 运营可观测性 + UI 现代化」。
>
> 关联文档：`docs/LingShu-v1.6-UI重构方案-Anthropic学术风.md`（纯视觉层，与本文档互补）。

---

## 第一部分：全维度差距矩阵

### 1.1 技术栈对比

| 维度 | sub2api（参考） | LingShu（现状） | 启示 |
|---|---|---|---|
| 后端语言 | Go | Go | 同源，机制可借鉴 |
| ORM | **ent**（生成式强类型，35 schema） | 手写 repository + golang-migrate SQL（12 表） | 不迁 ent，沿用手写迁移 |
| 路由 | gin | chi | — |
| 前端 | Vue3 Composition + Vite5 + pnpm | React19 + Vite7 + npm workspace | — |
| 前端 app | 单 app（按角色显隐） | **双 app**：admin（AntD6）+ user（shadcn/Tailwind4） | admin 是 UI 短板 |
| UI 库 | 纯 Tailwind 自建组件 | admin=Ant Design 6；user=Radix+cva+lucide+recharts | admin 需去「裸表格」 |
| 计费模式 | 余额 standard + 订阅 subscription 双模式 | 仅余额预扣 | 订阅按需，不做 |
| 缓存计费 | **完整**（cache_creation/cache_read，5m/1h 分档） | **完全缺失** | ← **P0 漏收** |
| 入站协议 | OpenAI/Anthropic/Gemini/Responses(Codex)/Antigravity | OpenAI chat + Anthropic messages | 够用 |
| 调度 | 账号状态机 + 粘性 + excludedIDs + 混合 + WS 池 | 渠道级粘性30min + 加权随机 + 简单重试 | ← 健壮性缺口 |
| 限流 | 4 层（IP/账号RPM/模型冷却/USD窗口）+ 并发ZSet 等待队列 | key 级 RPM + 并发（基础） | ← 缺口 |
| 管理端 | 17+ 模块 / 30+ API 域 | 12 页面 | ← 功能缺口 |

### 1.2 数据模型差距（事实表是计费命脉）

**LingShu 现有 12 表**（`backend/migrations/0001_init.up.sql`）：users、api_keys、models、upstream_channels、channel_models、gateway_requests、balance_ledger、redeem_codes、redeem_records、announcements、audit_logs、system_settings。

**关键事实表 `gateway_requests` 现状字段**：`prompt_tokens / completion_tokens / total_tokens`（仅 2 个 token 口径）、`base_cost / rate_multiplier / charge`（单一成本口径）、`endpoint / status / http_status / is_stream / is_estimated / latency_ms / error_code / client_ip`。
**缺失**（对比 sub2api usage_log）：缓存 token（creation/read，5m/1h 分档）、image token、**双模型名**（requested vs upstream + mapping_chain）、**TTFT（首字延迟 first_token_ms）**、**三成本口径**（standard/account/actual）、billing_mode/tier 快照。

**`models` 表现状**：`input_price_per_1k / output_price_per_1k / price_per_call / rate_multiplier`（billing_mode ∈ token/per_call）。**缺失**：`cache_creation_price_per_1k / cache_read_price_per_1k`、image 单价、价格区间（intervals）。

**`upstream_channels` 现状**：`provider_type ∈ (openai, anthropic)`、`weight / rpm_limit / concurrency_limit / fail_threshold / fail_count / health(healthy|unhealthy) / last_success_at / last_error_at`。**缺失**：账号级状态机字段（rate_limited_at/reset_at、overload_until、temp_unschedulable、session_window、priority、load_factor、proxy 绑定、expires_at）。

> **概念错位提醒**：sub2api 把「上游凭证」建模为 **account**（账号，富状态机），把「可见性 + 倍率 + 限额」建模为 **group**（分组）；LingShu 把上游凭证建模为 **upstream_channel**，倍率挂在 **model** 上，无分组层。本文档统一用 LingShu 术语（channel），但调度状态机的思路直接借鉴 sub2api 的 account。

### 1.3 差距优先级总表

| # | 差距 | sub2api 实现位置 | LingShu 现状 | 优先级 |
|---|---|---|---|---|
| G1 | **缓存计费**（creation/read） | `service/billing_service.go:819 computeTokenBreakdown` / `:919 computeCacheCreationCost` | 仅 input/output，缓存 token 漏收 | **P0** |
| G2 | **usage_log 富化**（缓存/双模型/TTFT/三口径） | usage_log schema 多字段 | `gateway_requests` 仅基础字段 | **P0** |
| G3 | **渠道故障转移健壮性**（excludedIDs 累积排除 + 健康熔断冷却） | `gateway_service.go:4999` 重试循环 + `:1571` 账号选择 | 粘性30min + 加权随机 + maxRetry 简单循环 | **P1** |
| G4 | **运维监控 Dashboard**（吞吐/延迟/错误/切换率） | OpsDashboard.vue + ops API | 无 | **P1** |
| G5 | **用量明细页**（筛选+导出） | UsageView.vue + xlsx 导出 + 虚拟滚动 | 仅 reports 聚合，无明细查询 | **P1** |
| G6 | **分组（Group）体系**（倍率/可见模型/RPM 下放分组） | group schema + group_service | 倍率挂 model，无分组 | **P2** |
| G7 | **多层限流**（账号RPM/模型冷却/USD窗口） | 4 层限流（见 2.4） | key 级 RPM+并发 | P2 |
| G8 | **管理仪表盘升级**（StatCard 网格+趋势+成本三层） | OpsDashboard / Dashboard.vue | admin-dashboard 基础 | P2 |
| G9 | **管理端去裸表格**（AntD → 学术风/卡片化） | 纯 Tailwind 自建组件 | AntD6 默认蓝裸表 | P2 |
| G10 | 价格区间（intervals）阶梯计费 | `account_stats_pricing.go` intervals | 单一单价 | P3 |
| G11 | service tier（priority/flex）差异计费 | `billing_service.go` serviceTier | 无 | P3 |
| G12 | 订阅计费 / 支付 / 联盟 / 风控 | 商业化中台 | 无 | **不做** |

---

## 第二部分：sub2api 核心机制详解（落地依据）

### 2.1 计费机制

**客户计费公式**（`backend/internal/service/billing_service.go:819 computeTokenBreakdown`）：

```
TotalCost  = InputCost + OutputCost + ImageOutputCost + CacheCreationCost + CacheReadCost
ActualCost = TotalCost × rateMultiplier
```

与 LingShu `charge = base_cost × rate_multiplier` **完全同构**，差别仅在 `base_cost(=TotalCost)` 由 5 项构成，LingShu 只算 2 项 → **缓存 token 漏收**。

**五档单价**：input / output / cache_creation / cache_read / image_output，各自独立 `*PricePerToken`。
**缓存创建分档**（`:919 computeCacheCreationCost`）：支持 `CacheCreation5mPrice` / `CacheCreation1hPrice`；上游未返回 ephemeral 明细时全部回退按 5m 价计。
**定价优先级链**（`account_stats_pricing.go:19`）：自定义规则 > 渠道 ApplyPricingToAccountStats 开关 > LiteLLM 模型定价文件 > 默认公式。
**扣费时机**：sub2api 是**后结算**（`checkBalanceEligibility` 余额>0 放行 → 异步 `QueueDeductBalance`）。**LingShu 是预扣**（`backend/internal/billing/reserve.go`）。**预扣更安全，本文档保留预扣，只补缓存分量。**

### 2.2 网关调度与故障转移

**账号选择**（`gateway_service.go:1571 SelectAccountForModelWithExclusions`）：粘性会话（hash→account，TTL 1h）+ 优先级 + 负载感知 + `excludedIDs` 排除集 + 混合调度。
**重试循环**（`gateway_service.go:4999`）：失败账号入 `excludedIDs`，下轮自动排除；400 签名错误专项重试；`maxRetryElapsed` 时间预算上限；上游错误透传规则。
**账号状态机**（`ent/schema/account.go`）：`schedulable / rate_limited_at / rate_limit_reset_at / overload_until(529) / temp_unschedulable_until+reason / session_window_* / priority(50) / concurrency(3) / load_factor / expires_at + auto_pause_on_expired`。429→记 reset_at 暂避；529→overload_until；过期→自动暂停。

> **LingShu 现状**（`backend/internal/service/gateway_service.go`）：`orderChannels`（粘性30min via frozen.GetStickyChannel + weightedRandomOrder）→ `forwardWithRetry`/`openStreamWithRetry` 循环 + `shouldRetryStatus` + `rememberSticky`(30min)。**缺 excludedIDs 累积排除、缺健康熔断冷却窗口、缺 429/529 状态记忆**。失败渠道下一轮可能被再次选中。

### 2.3 usage_log 富化（事实表设计）

sub2api usage_log 字段：6 token（input/output/cache_creation/cache_read/cache_creation_5m/cache_creation_1h）、7 cost（input/output/cache_creation/cache_read_cost/total_cost/actual_cost + rate_multiplier/account_rate_multiplier 快照）、双模型名（requested + upstream + mapping_chain）、billing_mode/tier、duration_ms、**first_token_ms(TTFT)**、image 尺寸。三成本口径：standard（标准价）/ account（账号成本）/ actual（实际计费）。

### 2.4 四层限流 + 并发控制

1. **IP 级**（Redis Lua，仅 auth 端点）。
2. **账号 RPM**（Redis INCR，服务端时间分钟 key）。
3. **模型冷却**（上游 429 reset 持久化到 Account.Extra）。
4. **USD 滑窗**（订阅 daily/weekly/monthly 限额）。
**并发**：Redis ZSet（账号级+用户级，含等待队列、进程前缀过期清理）。

> LingShu 现状：`api_keys.rpm_limit / concurrency_limit` + 渠道级 rpm/concurrency，无账号 RPM、无模型冷却、无 USD 窗口。本文档只把 **G7 模型冷却** 列入 P2（直接关系调度健壮性），其余按需。

### 2.5 管理端能力面（17 模块，裁剪后看）

sub2api 17+ 模块中，**LingShu 应补**的：运维监控 Ops、用量明细、分组管理、渠道监控探活、仪表盘成本三层透明。**应裁掉**的：支付/订单/套餐、联盟返利、促销码、风控、订阅、TLS 指纹、用户属性。

---

## 第三部分：里程碑路线图（agent 可执行）

> **本次审计已核对代码真实状态（2026-06-19）**。已完成且验证通过的里程碑只留「完成基线」一行备查，不再展开实现细节（避免 agent 重做）；**未做 / 部分完成 / 新增的里程碑保留完整可执行描述**。

### ✅ 完成基线（已验证，agent 勿重做）

| 里程碑 | 状态 | 证据（核心文件） |
|---|---|---|
| **M1 缓存计费补全** | ✅ 约 95%（留 1 个补丁，见下方 M1-补丁） | `billing/calculator.go:8-9,16-17,31-33`；`migrations/0008`；`anthropic_adapter.go` 流式/非流式均解析缓存 token；`openai_client.go:275` 映射 cached_tokens；`repository/gateway.go:185-245` 写两列；`model-form.tsx:12-13` 有单价输入；`calculator_test.go` 3 个缓存用例（含缓存价=0 回归） |
| **M2 usage_log 富化** | ✅ 100% | `migrations/0009` 加 first_token_ms/upstream_model_name/image_output_tokens；`gateway_service.go:328,373-374` 落库；`handler/gateway/handler.go:86-87`、`messages.go:95-101` 打首字时间；`dto/user_log_dto.go` 已脱敏 base_cost/rate_multiplier/upstream_model_name + 回归测试 |
| **M4 后端聚合** | ✅ 后端完成（前端有小增量，见下方 M4-剩余） | `service/ops_service.go:88-201` 输出 RPM/TPM/延迟分位/错误率/渠道健康/切换次数/24h 趋势；`handler/admin/ops.go` + 路由 `server.go:107`；`admin-dashboard.tsx:24-62` 已含成本三层透明 |
| **M7 admin 去 antd + 迁 @lingshu/ui** | ✅ 约 90%（剩余风险继承自 M6 弱组件） | admin 全树 + package.json `antd` grep = **0**；12 页全部 import `@lingshu/ui`；PageHeader/StatCard/DataTable 范式已落地 |

> **以下为待办**：M1-补丁、M3 剩余、M4-剩余、M5（可选）、**M6（核心短板）**，以及第六部分 M8–M13。

### M1-补丁（P1，小）缓存计费全链路闭环

**唯一缺口**：`backend/internal/upstream/anthropic_inbound.go`（Anthropic 入站 → OpenAI 上游 这条转换链路）内部的 `openAIChatResponse.Usage` 与 `openAIStreamChunk.Usage` 只声明了 `prompt_tokens`/`completion_tokens`（约 `anthropic_inbound.go:101-104,159-162`），**没解析 OpenAI 上游可能返回的 `prompt_tokens_details.cached_tokens`**，也没回填到下游 usage。结果：Anthropic 客户端打过来、底层走 OpenAI 通道时，`cache_read_tokens` 会丢，与 M1「全链路缓存计费」不完全闭环。

**改动**：给该文件的两个 Usage 结构体补 `PromptTokensDetails.CachedTokens` 解析，并把 cached 部分映射进 `cache_read`，回填 GatewayUsage。**不碰计费公式**，只补 token 透传。

**验收**：构造经 anthropic_inbound 链路、上游 OpenAI 返回 cached_tokens 的请求，查 `gateway_requests.cache_read_tokens > 0`。

### M3（P1）渠道故障转移健壮性 —— 部分完成，补剩余

**已完成（勿重做）**：① 累积排除已失败渠道——`forwardWithRetry`/`forwardEmbeddingsWithRetry`/`openStreamWithRetry` 三循环用 `excluded map` + `markChannelFailure` + `orderChannels(excluded)` 过滤（`gateway_service.go:445,494,543,595-629`）。② 基础冷却——固定 30s `channelFailureCooldown`（`:54`）经 redis `channel_cooldown:<id>` TTL + `IsChannelCooling` 过滤（`:602`）。③ DB 侧 `fail_count>=fail_threshold → health=unhealthy`（`repository/gateway.go:160`，存量机制）。

**剩余待办（本次要做的）**：
1. **可变冷却窗口**（当前是不分错误类型的固定 30s）：按错误类型/连续失败次数区分冷却时长——5xx 短冷却、429/529 长冷却。可在 `markChannelFailure` 依据 status 写不同 TTL。
2. **上游 429/529 专属记忆与降权**（当前 `shouldRetryStatus` 把 429/529/5xx 一律同等重试、同一 30s 冷却，`:718-720,628`）：429 记 `reset_at` 短期跳过该渠道、529 记 `overload_until`，冷却内 `orderChannels` 跳过（参考 sub2api `account.go` 的 rate_limited_at/overload_until）。

**验收 M3**：单测验证 429 命中后该渠道在 reset 窗口内被 `orderChannels` 跳过、5xx 与 429 冷却时长不同。`go test ./internal/service/...`。

### M4-剩余（P2，前端小增量）—— 后端已完成，仅补前端

**后端聚合已全部完成**（见完成基线）。**剩余仅前端两点**：
1. `frontend/admin/src/pages/ops.tsx`：后端已返回 **P95 延迟、HTTP 状态分布、charge 趋势** 字段但前端未渲染（当前只有 6 张 StatCard + 请求数柱状 + 渠道健康表）。补：延迟折线、错误率折线、状态分布。
2. `frontend/admin/src/pages/admin-dashboard.tsx`：模型分布当前是表格，补环形/条形图可视化（可选，低优先）。

**验收 M4-剩余**：ops 页渲染出 P95 与状态分布；与后端聚合数值一致。

### M5（P2，可选）分组（Group）体系

**目标**：补分组抽象，让倍率/可见模型/RPM 可按分组下放（多租户/多套餐前置能力）。

- 新表 `groups`（rate_multiplier、platform、可见 model 集合、rpm_limit 覆盖）+ `user_groups`/`key_groups` 关联。
- 倍率优先级：key 覆盖 > 分组 > 默认。**迁移策略**：存量 user/key 倍率落入「默认分组」，保证存量计费逐分不变。
- 风险高，仅在确有多租户/多套餐运营需求时启动。

**验收 M5**：分组迁移后存量计费逐分不变；倍率优先级链有单测覆盖。

### M6 ✅ 已完成（验收通过，留 1 个死代码尾巴）

**验收结果**：`packages/ui/src/components/` 已从 441 行单文件拆成 12 个文件；Dialog/Sheet 用 `@radix-ui/react-dialog`、DropdownMenu 用 `react-dropdown-menu`、Popover 用 `react-popover`、Tooltip 用 `react-tooltip`、Command 用 `cmdk`，**全部真实现，非手搓**；补齐 Form；design-tokens 扩展齐全；user 端全走 `@lingshu/ui`。**组件库主体做好，无需重做。**

> **尾巴**：`frontend/user/src/components/ui/*.tsx` 7 个本地 shadcn 文件已成死代码（无人 import），需清理 → 已并入 **M8-收尾**（见 6.1）统一处理。

### M7 ✅ 完成（antd 零残留、12 页迁 `@lingshu/ui`、范式落地）

M6 真 Radix 化已完成，M7 交互质量随之达标，**无遗留**。

### M-X（不做，仅备案）
订阅双模式、支付/订单/套餐、联盟返利、促销码、风控、TLS 指纹、价格区间 intervals、service tier。**有真实需求再单独立项。**

---

## 第四部分：执行铁律与参考索引

### 4.1 全局路线图（二轮审计后真实状态，2026-06-19）

**✅ 已完成（勿重做）**：M1（缓存计费，留 1 小补丁）、M2（usage 富化）、M4 后端、M6（组件库真 Radix 化）、M7（admin 去 antd）、M8（单 app + 角色门控，留清理尾巴）、M9（i18n）、M11（主题切换）、M12.1 仪表盘、M13（兑换码）。

**🔴 待办，建议执行顺序**：

**第一梯队 · 收尾（清理这轮尾巴，快）**
1. **M8-收尾**：删旧 `frontend/admin`/`user` workspace + `user/src/components/ui/*` 死代码，让 `app` 成唯一前端。
2. **M12 补完**：usage 时间过滤 + 真实时区块；pricing 详情抽屉；api-keys 端点白名单（依赖 M17 后端 `allowed_endpoints` 字段）。
3. **M10 落地页**：仅剩这一项前端新建（framer-motion + Hero）。

**第二梯队 · 对外开门（P0，不做不能上线）— 第七部分**
4. **M14 邮箱体系**（SMTP + 注册验证 + 找回密码 + 注册模式开关）——头号阻断项。
5. **M15 登录防爆破**（IP/账号限流锁定 + 可选 captcha + banned 拦截）。
6. **M16 法务合规**（ICP/公安备案位 + ToS/隐私 + 全局 footer + 品牌可配置）——国内上线硬门槛。

**第三梯队 · 运营增长（P1，上线即补）— 第七部分**
7. **M17 接入文档页**（base_url + curl/Python/Node 示例 + 能力矩阵）+ Key 端点白名单后端字段。
8. **M18 admin 用户运营**（调额/封禁/重置/下线五连）+ CSV 导出（user 端脱敏）。
9. **M19 告警 + 渠道自愈**（规则引擎 + 邮件/Webhook + 连续失败自动 disabled）。

**第四梯队 · 后端零散 + 增长期（随时可插 / P2）**
10. 后端零散：**M1-补丁**（inbound cache）、**M3 剩余**（可变冷却 + 429 降权）、**M4-剩余**（ops 前端闲置字段）——独立于前端主线，可并行。
11. **M20**（状态页 + SEO）、**M5 分组**、M-X：增长期按需。

> **依赖图**：前端主线已基本完成（M6→M8→M9/M11/M12.1）。剩余 = 收尾梯队（M8-收尾/M12 补完/M10）→ 运营 P0（M14→M15→M16）→ 运营 P1（M17→M18→M19）。后端零散（M1-补丁/M3/M4-剩余）无依赖可随时并行。**M16 是国内对外上线的硬门槛，M14 是用户可用的前提。**

### 4.2 不可违反的铁律
1. 计费语义不变：`charge = base_cost × rate_multiplier`，缓存只是往 base_cost 加分量。
2. 用户端零敏感字段：base_cost / rate_multiplier / 毛利 / upstream_model_name 永不下发用户端。
3. 存量计费不回退：新列默认 0、分组迁移落默认分组，存量请求逐分一致。
4. 不照搬商业化臃肿（支付/联盟/风控套件/订阅默认不做）；变现仅兑换码。运营层新增（邮箱/合规/告警/导出/状态页）**不触碰计费链路**。
5. 每里程碑独立可上线 + 构建验证：`go build ./... && go vet ./... && go test ./...` + 前端 `app` build（旧 admin/user 待 M8-收尾删除）。

### 4.3 sub2api 参考文件索引（`D:\code\LingShu\Sub`）

| 主题 | 文件 |
|---|---|
| 计费公式核心 | `backend/internal/service/billing_service.go:819` |
| 缓存分档计费 | `backend/internal/service/billing_service.go:919` |
| 定价优先级链 | `backend/internal/service/account_stats_pricing.go:19` |
| 余额 preflight | `backend/internal/service/billing_cache_service.go:837` |
| 账号选择（粘性+排除） | `backend/internal/service/gateway_service.go:1571` |
| 故障转移重试循环 | `backend/internal/service/gateway_service.go:4999` |
| 流式 usage 解析 | `backend/internal/service/gateway_service.go:8380` |
| 入站协议分流 | `backend/internal/server/routes/gateway.go:44` |
| 账号状态机 schema | `backend/ent/schema/account.go` |
| 分组 schema | `backend/ent/schema/group.go` |

### 4.4 LingShu 现状文件索引（`D:\code\LingShu`）

| 主题 | 文件 |
|---|---|
| 计费公式（待扩缓存） | `backend/internal/billing/calculator.go:24` |
| 预扣 | `backend/internal/billing/reserve.go` |
| 事实表 schema | `backend/migrations/0001_init.up.sql:79`（gateway_requests）|
| 网关调度 | `backend/internal/service/gateway_service.go`（orderChannels/forwardWithRetry/rememberSticky）|
| Anthropic 入站 | `backend/internal/upstream/anthropic_inbound.go` |
| Anthropic 出站适配 | `backend/internal/upstream/anthropic_adapter.go` |
| messages handler | `backend/internal/handler/gateway/messages.go` |
| 管理端 12 页 | `frontend/admin/src/pages/*.tsx` |
| 模型表单 | `frontend/admin/src/pages/model-form.tsx` |
| 用户端（shadcn 风） | `frontend/user/src/` |
| UI 视觉重构方案 | `docs/LingShu-v1.6-UI重构方案-Anthropic学术风.md` |

---

## 第五部分：UI 设计系统重构详规（M6 / M7 落地依据）

> **核心诉求（用户原话）**：「目前只是有了 claudecode 颜色的空壳，只换了颜色跟字体」「需要完美复刻 claudecode 风格而不是只是一个模版」。
> **现状诊断**：design-token（clay/暖白/衬线）已配（`frontend/packages/shared/src/design-tokens.ts` + `user/src/styles.css`），但**组件库薄**（user 仅 7 件：badge/button/card/input/skeleton/sonner/tabs）、**admin 组件目录为空靠 AntD 默认样式**、页面是卡片+div 平铺、缺阴影层次/动效/状态色系统/页面范式。
> **目标**：从「换皮」升级为完整设计系统，两端复用，完美复刻 Claude Code 质感。

### 5.1 Claude Code 风格还原度自检表（M6 验收逐项核对）

复刻不是「调对主色」，而是还原以下质感细节：

| 维度 | Claude Code 特征 | 落地要求 |
|---|---|---|
| 纸感背景 | 暖 ivory 非纯白，微噪点纸张感 | `--bg:#F0EEE6`；可叠 `radial-gradient` 极淡噪点；卡面 `--surface:#FAF9F5` 比底色更亮 |
| 暖阴影 | 阴影带暖调、极柔，非灰色硬投影 | shadow 用 `rgba(20,20,19,0.04~0.08)` 暖墨色，多层叠加，扩散大模糊 |
| 衬线节奏 | 标题衬线（Tiempos 系）+ 负字距 + 舒展行高 | h1/h2/display 用 `--font-serif`，`letter-spacing:-0.02em`，字号阶梯 5 档 |
| 克制圆角 | 6px 主圆角，不圆润不锐利 | 卡/按钮 6px，小元素 4px，大容器 10px |
| 细腻边框 | 低对比暖灰描边，1px | `--border-c:#D8D4CA`，hover 升 `--border-strong:#B0AEA5` |
| 软交互 | hover 暖色微变、focus ring 用 clay 而非系统蓝 | 过渡 150~200ms ease-out；focus-visible ring=clay |
| 图标语言 | 细描边、统一尺寸 | lucide `stroke-width:1.5`，主 16/20px |
| 内容密度 | 舒适非紧凑，大量留白 | 行高 1.5+，区块间距走 spacing 尺度 |
| 低饱和语义色 | success/warning/danger 都偏哑光、与暖底协调 | success#4F7A4D / warning#B5821F / danger#A84031，各配 soft 底 |
| 文档感排版 | 正文像文档，强调 typographic hierarchy | 标题衬线、正文 sans、数据/代码 mono 三体系分工 |

### 5.2 组件清单（`@lingshu/ui` 目标 ~22 件）

**已有迁入（7）**：Button、Card、Input、Badge、Skeleton、Tabs、Sonner(Toast)。
**新建（15）**：
- **数据展示**：`DataTable`（排序/分页/空态/加载骨架/行选中，列表页核心）、`Table`（原子表格）、`Pagination`、`Avatar`、`Separator`、`Progress`、`Tag`（语义色版，区别于 Badge）。
- **覆盖层**：`Dialog`（Radix）、`Sheet`/`Drawer`（详情侧滑，替代跳页）、`DropdownMenu`（行内操作收敛）、`Popover`、`Tooltip`、`Command`（⌘K 命令面板）。
- **表单**：`Select`、`Switch`、`Textarea`、`Label`、`Form`（封装校验/错误态，配 react-hook-form 或轻量自封装）、`Alert`（页内提示）。

> 全部基于 Radix primitives + cva + Tailwind4，与 user 端现有技术栈一致（不引入新 UI 库）。每个组件支持 `className` 透传、`asChild`、暗色变量。

### 5.3 设计令牌扩展（在现有 design-tokens.ts 基础上补）

现有：colors（含 success/warning/danger）、radius、spacing（仅 page/section）、font、shadow（sm/md）。**需补**：

```ts
// 阴影层次（暖墨色，多层柔和）
shadow: { xs, sm, md, lg, xl }  // 当前只有 sm/md，补 xs/lg/xl
// 状态色系统（每色 3 件套：底/边/字）
status: { success:{bg,border,text}, warning:{...}, danger:{...}, info:{...} }
// 间距尺度（补全 4/8/12/16/24/32/48）
spacing: { xs:4, sm:8, md:12, lg:16, xl:24, 2xl:32, 3xl:48, page, section }
// 排版尺度（衬线 display 阶梯 + 行高 + 字距）
typography: { display, h1, h2, h3, body, small, caption }  // 各含 size/lineHeight/letterSpacing/weight
// 动效
motion: { fast:120ms, base:160ms, slow:240ms, ease:"cubic-bezier(0.32,0.72,0,1)" }
// 层级
z: { dropdown, sticky, overlay, modal, toast }
```
令牌同步进 `user/src/styles.css` 与 `admin/src/styles.css` 的 `@theme` / `:root`，两端共用 CSS 变量名。

### 5.4 页面范式（可复用模板，M7 重做 admin 12 页依据）

定义 3 套模板，所有页面归入其一，杜绝「div 平铺」：

**A. 列表页模板**（users/channels/models/api-keys/redeem/announcements/audit/usage）：
```
PageHeader（eyebrow + 衬线标题 + 描述 + 右上主操作）
└ StatCard 网格（该资源的 3~5 个关键指标）
└ 筛选栏（搜索 + 状态/类型筛选 + 时间范围，沉浸在卡片内）
└ DataTable（排序表头 / 语义色 Tag 状态 / 行内操作收进 DropdownMenu / 分页 / 空态 / 加载骨架）
└ 详情与编辑走 Sheet 侧滑，不跳页
```

**B. 仪表盘模板**（admin-dashboard / ops / user dashboard）：
```
PageHeader
└ 主指标 StatCard 网格（带趋势小箭头/环比）
└ 趋势区（折线/面积图，分时段切换）
└ 分布区（环形/条形：模型分布、渠道健康）
└ 明细区（最近调用 / 告警）
admin 仪表盘额外：成本三层透明（base_cost/charge/毛利，仅 admin；铁律 2）
```

**C. 表单/设置模板**（settings / model-form）：
```
分区卡片（每组配置一张卡，衬线小标题 + 描述）
└ Form 组件统一标签/校验/错误态/保存反馈
└ 多 Tab 时用 Tabs（settings 的 general/security/gateway/...）
```

**三态规范**：每个数据区都要有 加载（Skeleton 形态贴合内容）/ 空（EmptyState 含微文案+引导操作）/ 错误（Alert + 重试）三态，不允许裸 loading 文字或空白。

### 5.5 产品级细节（让它「像产品」非「模板」）

1. **⌘K 命令面板**（`Command` 组件 + 全局快捷键）：跨页导航、资源搜索、快捷操作。两端共用。
2. **全局快捷键**：`g+d` 仪表盘、`g+u` 用户、`/` 聚焦搜索、`⌘K` 面板、`esc` 关 Sheet。
3. **品牌区**：侧栏顶部「灵枢 LingShu」衬线字标 + 极简 mark；登录页品牌化。
4. **空态插画/微文案**：克制线性插画或大图标 + 引导性文案（非「暂无数据」四字）。
5. **暗色模式（可选）**：CSS 变量已抽象，补一套暗色 `:root[data-theme=dark]` 暖墨底（非纯黑），`prefers-color-scheme` + 手动切换。
6. **微交互**：数字滚动（StatCard 进场 count-up）、骨架→内容淡入、Sheet 弹簧过渡、hover 暖色微亮、表格行 hover 高亮。
7. **可访问性**：focus-visible ring（clay）、aria 标签、键盘可达（Radix 自带）、对比度达 AA。

### 5.6 包结构与迁移步骤

**`frontend/packages/ui` 结构**：
```
packages/ui/
├ package.json            # name:"@lingshu/ui", peerDeps: react/radix/lucide/cva
├ src/
│  ├ index.ts             # 统一导出
│  ├ components/          # 22 组件
│  ├ hooks/               # useMediaQuery / useHotkeys / useCommandPalette
│  ├ lib/cn.ts            # clsx+tailwind-merge
│  └ styles/tokens.css    # 共享 @theme 令牌（被两端 import）
```

**迁移步骤（M6→M7）**：
1. 建 `@lingshu/ui` 包，加入 `frontend/package.json` workspaces。
2. design-tokens.ts 扩展（5.3）+ tokens.css 落地；两端 styles.css import。
3. user 现有 7 组件迁入包，补 15 新组件，逐个写最小 story/示例自测。
4. user 端 9 页改 import（`@/components/ui/*` → `@lingshu/ui`），按页面范式（5.4）薄重构。
5. 加 ⌘K/快捷键/品牌区/暗色（5.5）。
6. **M7**：admin 删 antd 依赖，12 页按范式重做，全部 import `@lingshu/ui`。
7. 验收：两端 build 通过；`grep -r "antd" frontend/admin/src` 为空；5.1 自检表逐项过。

> **铁律对 UI 的约束**：admin 仪表盘可显示成本三层（base_cost/charge/毛利），**user 端组件与页面绝不渲染 base_cost/rate_multiplier/毛利/upstream_model_name**（铁律 2）。共享组件若同时用于两端，敏感字段由调用方传入，组件本身不内置展示逻辑。

---

# 第六部分：v2.0 前端体验重构（对标 RelaxyCode）

> 本部分基于与用户确认的范围增量追加，**不改动 M1–M7 的计费/后端契约**，仅重塑前端架构与体验。涉及的架构转向（双 app → 单 app）会在 6.1 显式说明对 M6/M7 的影响与衔接。
>
> **铁律重申（贯穿本部分所有页面）**：user 端**绝不渲染** `base_cost` / `rate_multiplier` / 毛利 / `upstream_model_name`。RelaxyCode 截图中对普通用户展示"倍率 2x/1x~1.2x"的做法**不照搬**——LingShu user 端只展示「对客最终单价」（即 `价格 × 倍率` 预先算好的每百万 token 价），日志页只展示最终费用、不显示倍率列。admin 端不受此限。

## 6.0 范围与决策快照（用户已确认 + 审计后真实状态）

| 维度 | 决策 | 当前状态 |
|---|---|---|
| 前端架构 | 合并 user + admin 为**单 React app**，角色路由门控 | 🔴 M8 全新未做（仍双 app） |
| 国际化 | **react-i18next**，中英双语全量 | 🔴 M9 全新未做（仅自写简易词典） |
| 落地页 | 完整落地页 + **framer-motion** 动画 | 🔴 M10 全新未做（有 pricing 公开页可作基础） |
| 主题 | 亮/暗双主题，CSS 变量 + `[data-theme]` | 🟡 M11 暗色 token 已就绪，仅缺切换器接线 |
| 运营功能（4 项） | 仪表盘/日志/价格页/API Key 增强 | 🟡 M12 四项均有简单版，做增量增强 |
| 营销/变现 | **仅兑换码充值** | ✅ M13 已闭环可用 |

## 6.1 里程碑 M8：单 app 合并 + 角色路由门控　✅ 已落地（留清理尾巴）

**验收结果**：已新建 `frontend/app` 单 app，`app/src/main.tsx` 统一路由，`router/guards.tsx` 真正按 `user.role !== "admin"` 重定向，`/admin/*` 全部 `requiresAdmin` 包裹，`HomeRoute`/`LoginPage` 按角色跳转。**机制完整，已做好。**

> **⚠️ 清理尾巴（M8-收尾，必做，含 M6 遗留死代码）**：
> 1. `frontend/package.json` workspaces 仍同时挂着 `app`、`user`、`admin` 三个；旧 `frontend/admin`、`frontend/user` 已被新 app 取代但**未删除**，旧 app 还能独立 build，易混淆来源。→ 确认新 app 覆盖全部页面后，删除旧 `admin`/`user` workspace（或至少从 workspaces 摘除并标记废弃），让 `app` 成为唯一前端。
> 2. **M6 死代码**：`frontend/user/src/components/ui/*.tsx`（badge/button/card/input/skeleton/sonner/tabs 共 7 个）已无人 import（全树 grep `from ".../components/ui"` = 0，页面都走 `@lingshu/ui`），属死代码且被改过 → `git rm` 清理，避免双份组件样式漂移。
> 3. admin 旧 app 复用了 `../../app/src/i18n`、`../../app/src/providers/theme` 等跨目录相对引用；删除旧 app 时一并清理这些悬挂引用。

**原始设计参考（已实现，留档）**：
- M6/M7 原假设「user 端 9 页 + admin 端 12 页」两套独立 app，各自 build。**M8 起改为单 app**：21 页合并进同一路由表，admin 页挂在 `/admin/*` 前缀下，由路由 meta + 守卫门控。
- `@lingshu/ui` 组件包（M6 产出）**继续复用**，不受合并影响——它本就是跨端共享层。
- M7 的「admin 删 antd」目标**并入 M8**：合并时 admin 页一并迁到 `@lingshu/ui`，不再保留独立 antd 技术栈。

**目录结构（合并后）**：
```
frontend/
  packages/shared, packages/ui   # 不变，跨端共享
  app/                            # 新建：合并后的单 app（取代 user/ 与 admin/）
    src/
      router/index.tsx            # 单一路由表，含 user + admin 全部页面
      router/guards.ts            # requiresAuth / requiresAdmin 守卫
      router/meta.ts              # RouteMeta 类型（参照 sub2api meta.d.ts）
      layout/AppSidebar.tsx       # 角色过滤侧边栏（userNav / adminNav）
      pages/...                   # user 页 + admin/* 页
```

**路由 meta 与守卫**（参照 `Sub/frontend/src/router/meta.d.ts` + `index.ts` 的 `beforeEach`）：
```ts
// meta.ts
interface RouteMeta {
  requiresAuth?: boolean   // 默认 true
  requiresAdmin?: boolean  // 默认 false，true 则普通用户不可见/不可达
  titleKey?: string        // i18n 页面标题 key
  icon?: string
  hideInMenu?: boolean
}

// guards.ts —— 单 app 守卫核心逻辑
function authGuard(to, auth) {
  if (to.meta.requiresAuth !== false && !auth.isLoggedIn) return redirect('/login')
  if (to.meta.requiresAdmin && !auth.isAdmin) return redirect('/dashboard') // 越权回落
  return allow()
}
```

**侧边栏角色过滤**（参照 `AppSidebar.vue` 的 `isAdmin` + `adminNavItems`）：
- `userNavItems`：概览/API Keys/用量/兑换码/设置——所有登录用户可见。
- `adminNavItems`：运营仪表盘/用户/分组/渠道/模型定价/系统设置——`{isAdmin && <AdminSection/>}` 条件渲染。
- React 实现：`const isAdmin = useAuthStore(s => s.isAdmin)`，菜单数组按 `isAdmin` 过滤后渲染。

**验收**：单 app build 通过；普通用户登录后 DOM 中无 admin 菜单项、手动访问 `/admin/*` 被守卫重定向到 `/dashboard`；admin 登录后两类菜单齐全。

## 6.2 里程碑 M9：react-i18next 中英双语　✅ 已完成（验收通过）

**验收结果**：已接入 `react-i18next`/`i18next`/`i18next-browser-languagedetector`，`app/src/i18n/index.ts` 完整 init（detection→localStorage `lingshu_locale`），`locales/en` 与 `locales/zh` 各 12 命名空间，`AppShell` 有语言切换器，28 文件用 `useTranslation`。**无需重做。** 后续仅需随新页面补译文。

**实现要点**（React 等价于 `Sub/frontend/src/i18n/index.ts`）：
```
frontend/app/src/i18n/
  index.ts          # createI18n 等价：i18next.init + 懒加载
  locales/en/*.ts   # 按命名空间拆分（common/dashboard/keys/...）
  locales/zh/*.ts
```
- 库：`react-i18next` + `i18next` + `i18next-browser-languagedetector`。
- 持久化 key：`lingshu_locale`（对应 sub2api 的 `sub2api_locale`）；默认 `en`，`navigator.language` 以 `zh` 开头则用 `zh`。
- 懒加载命名空间：路由进入时按需 `i18n.loadNamespaces([...])`，避免首屏全量打包。
- 切换副作用：写 localStorage + `document.documentElement.lang = locale` + 更新 `document.title`（跟随语言）。
- topbar 语言切换器：`availableLocales = [{code:'en',name:'English',flag:'🇺🇸'},{code:'zh',name:'中文',flag:'🇨🇳'}]`。
- 是否加 URL 语言前缀（`/zh/dashboard`，RelaxyCode 有、sub2api 无）：**本期不做**，仅 localStorage 记忆，降低路由复杂度。

**验收**：切换语言后全站文案即时更新、刷新后保持；`<html lang>` 与页签标题同步；无文案残留硬编码（抽查 dashboard/keys 页）。

## 6.3 里程碑 M10：完整落地页 + framer-motion 动画　🔴 全新未做

**现状**：grep `framer-motion` = 0，无 Landing/Hero 文件。仅 `frontend/user/src/routes/pricing.tsx` 是未登录可访问的公开页（带 SiteNav），可作落地页的导航/公开访问基础，但无 Hero/动画/落地页结构。

**目标**：新增公开落地页（未登录可访问），RelaxyCode 设计风格 + Claude Code 学术质感，配合滚动入场/悬停/点击微交互。

**路由**：`/`（公开，`requiresAuth: false`），已登录用户访问可保留落地页或提供「进入控制台」CTA。

**结构**（参照 RelaxyCode 与 imge 截图）：
```
pages/landing/
  Landing.tsx
  sections/Hero.tsx        # 主标题 + 副文案 + 双 CTA（开始使用/查看文档）
  sections/Features.tsx    # 特性卡片网格（多模型/计费透明/高可用）
  sections/Pricing.tsx     # 模型价格预览（仅对客最终单价，复用 6.5 价格组件）
  sections/CTA.tsx         # 底部行动召唤 + 兑换码入口
  sections/Footer.tsx
```

**动画规范（framer-motion）**：
- 入场：section 进入视口 `whileInView`，`initial={{opacity:0,y:24}} animate={{opacity:1,y:0}}`，`viewport={{once:true}}`，子项 `staggerChildren` 错峰。
- 悬停：卡片 `whileHover={{y:-4}}` + 阴影过渡；CTA 按钮 `whileHover/whileTap` 缩放。
- 过渡：统一缓动 `transition={{duration:0.4, ease:[0.22,1,0.36,1]}}`（与 Claude Code 学术调性一致，避免弹跳）。
- 性能：仅用 transform/opacity（GPU 友好）；`prefers-reduced-motion` 时禁用位移动画。

**验收**：落地页未登录可访问；滚动各 section 依次入场；卡片/按钮悬停点击有反馈；移动端布局自适应；开启系统「减少动态效果」时动画降级。

## 6.4 里程碑 M11：亮/暗双主题　✅ 已完成（验收通过）

**验收结果**：`app/src/providers/theme.tsx` 提供 `ThemeProvider`/`useTheme`/`toggleTheme`/`setTheme`，`lingshu_theme` 持久化 + `prefers-color-scheme` 回退，经 `document.documentElement.dataset.theme` 应用，`AppShell` 顶栏有切换按钮。**无需重做。**

**实现要点**：
- `tokens.css`：亮色为 `:root` 默认，暗色用 `[data-theme="dark"]` 覆写同名变量（clay/ivory/surface/ink 等映射到暗色等价值）。
- 切换：`document.documentElement.dataset.theme = 'dark' | 'light'`，写 localStorage（key `lingshu_theme`）；首屏读取 + `prefers-color-scheme` 兜底。
- topbar 主题切换图标（参照截图右上角图标组）。
- 组件不写死颜色，一律走 CSS 变量，确保双主题自动适配。

**验收**：切换主题全站即时变色、刷新保持；首屏无闪白（在 `<head>` 内联脚本预设 `data-theme`）；暗色下文字对比度达 WCAG AA（抽查正文/次要文字/链接）。

## 6.5 里程碑 M12：四项运营功能增强　🟡 验收后：2 页完成、2 页待补

> 均为 user 端面向客户的呈现，严守铁律：不显示 base_cost/rate_multiplier/毛利/上游模型名。**合规已验收通过（user 全树 grep 敏感字段 = 0）。** 以下为审计后逐页真实状态。

### 6.5.1 仪表盘增强（`frontend/user/src/routes/dashboard.tsx`）　✅ 完成
配额进度条、Claude Code/Codex 快速配置脚本 Tab + 一键复制均已落地，原 5 StatCard + 趋势图保留。**无需再动。**

### 6.5.2 请求日志增强（`frontend/user/src/routes/usage.tsx`）　🟡 ~75%，补两点
**已有**：关键字+状态过滤、行点击详情抽屉、合规字段。
**待补**：① **时间范围过滤器**（当前缺）；② **"实时进行中请求"目前只是静态计数**（`logs.filter(status!=="success" && http_status===0)`），需改成独立区块 + 轮询（建议 5s）或 SSE，展示当前 in-progress 调用。

### 6.5.3 模型价格展示页（`frontend/user/src/routes/pricing.tsx`）　🟡 ~80%，补一点
**已有**：search/计费方式/分组过滤、网格⇄列表切换、合规对客单价。
**待补**：**点击卡片的详情抽屉未做**（Dialog/Sheet）——补模型详情（上下文长度、支持端点、对客单价明细）。筛选可顺手从顶部条改成左侧面板（可选）。

### 6.5.4 API Key 管理增强（`frontend/user/src/routes/api-keys.tsx`）　🔴 ~40%，主要待补
**已有**：状态点、创建时一次性明文 Dialog。
**待补**：① **端点白名单复选框**——需后端先加 `allowed_endpoints` 字段（见 M14 后端配套）；create 表单与列表增该维度；② **费用分组展示**（user 端不显示倍率徽章，RelaxyCode 的 1x/2x 不照搬）；③ "二次查看明文"按安全实践**不开放**（仅创建时可见即可，文档定调，不再列为缺口）。

**验收（M12 整体）**：四页 user 端 grep 不到 `rate_multiplier`/`base_cost`/`gross_profit`/`upstream_model_name` 渲染（✅ 已过）；usage 时间过滤 + 真实时区块到位；pricing 详情抽屉到位；api-keys 端点白名单到位。

## 6.6 里程碑 M13：兑换码充值　✅ 基本完成（可选小增强）

**现状**：`frontend/user/src/routes/redeem.tsx` 已实现完整闭环——输入兑换码 → 调 `api.redeem` → toast 提示 + 刷新余额 + 展示本次到账金额。**核心功能可用，无需重做。**

**可选小增强**（按需，非必做）：兑换记录列表、更细的失败原因提示（已使用/无效/过期分别文案）。后端核销幂等若未保证，需确认并发重复提交不重复入账。**不做在线支付/套餐/订单/优惠券/分销。**

---

> **第六部分铁律总结**：M8–M13 仅重塑前端体验与新增兑换码入账，**计费公式 `charge = base_cost × rate_multiplier` 全程不动**；user 端任何页面/组件不渲染 base_cost/rate_multiplier/毛利/upstream_model_name；新增展示用的「对客单价」字段由服务端预算后下发，不改既有扣费链路。

---

# 第七部分：对外运营开门清单（M14–M20）

> **定位**：前六部分把产品做成了「功能完整的内部系统」，但**离「能对外开门做生意」还差一层运营基建**。本部分对照 sub2api 商业 SaaS，补齐「一个真正对外运营的站点」的刚需，砍掉重商业化臃肿（仍不做在线支付/订阅/联盟/风控）。
>
> **三条铁律继续生效**：计费公式不动；user 端零敏感字段；存量不回退。本部分新增的都是**用户生命周期 / 信任合规 / 运营干预 / 可观测告警**层，与计费链路解耦。
>
> **优先级**：M14–M16 标 **P0（不做不能上线）**；M17–M18 标 **P1（上线即需）**；M19–M20 标 **P2（增长期补）**。

## 7.1 里程碑 M14（P0）邮箱体系 + 用户生命周期闭环

**现状**：`backend/internal/handler/auth/handler.go` 仅 Login/Register/ChangePassword 三件套；**无 SMTP、无邮箱验证、无找回密码**。用户忘密码=账号废了；注册无验证=垃圾号泛滥。这是对外运营的**头号阻断项**。

**后端改动**：
1. **SMTP 服务** `backend/internal/service/email_service.go`（参考 sub2api `service/email_service.go`）：配置走 `system_settings`（smtp_host/port/user/pass/from/tls），发送验证码与通知邮件，模板中英双语。
2. **验证码** `backend/internal/redis/`：注册/找回密码验证码存 Redis（TTL 10min，发送频率限制 60s/次），key 如 `email_code:<purpose>:<email>`。
3. **新表/字段**：`users` 增 `email_verified BOOL DEFAULT false`；新表 `email_verifications` 可省（用 Redis 即可）。迁移 `0010_user_lifecycle.up.sql`。
4. **新端点**（`handler/auth/`）：`POST /auth/email/send-code`（注册/找回各一 purpose）、`POST /auth/register`（校验码后建号）、`POST /auth/forgot`（发重置码）、`POST /auth/reset`（校验码改密）。
5. **注册模式开关**：`system_settings.registration_mode ∈ open|invite|closed`。`invite` 复用兑换码模型加一类「邀请码」，`closed` 关闭注册入口。

**前端改动**（`frontend/app/src/`）：
- `routes/register.tsx`：邮箱 + 发码按钮（倒计时）+ 验证码 + 密码；`registration_mode` 控制可见性。
- `routes/forgot.tsx` + `routes/reset.tsx`：找回密码两步流程。
- 登录页底部「忘记密码？」入口。

**验收 M14**：未验证邮箱无法注册成功；找回密码全流程可走通；`registration_mode=closed` 时注册入口隐藏且接口拒绝；发码有频率限制。SMTP 未配置时给 admin 明确提示而非静默失败。

## 7.2 里程碑 M15（P0）登录防爆破 + 基础风控

**现状**：登录接口无 IP 限流、无失败锁定、无验证码。对外当天就会被撞库/爆破。

**后端改动**：
1. **登录限流**（`backend/internal/redis/` + auth 中间件）：IP 级 + 账号级失败计数，连续失败 N 次（如 5 次）锁定 M 分钟（如 15min），key `login_fail:<ip>` / `login_fail:<email>`。参考 sub2api IP 限流 Lua。
2. **人机验证（可选开关）**：`system_settings.captcha_enabled` + provider（Turnstile/hCaptcha），注册/登录前端挂载、后端校验 token。默认关，运营按需开。
3. **admin 用户封禁联动**：`users.status` 已有，确保 banned 用户登录直接拒绝并返回明确文案。

**前端改动**：登录失败计数提示（剩余尝试次数）；锁定时显示倒计时；captcha 开启时渲染验证组件。

**验收 M15**：暴力尝试触发锁定并返回明确错误；锁定窗口内拒绝登录；captcha 开启后无 token 登录被拒；banned 用户无法登录。

## 7.3 里程碑 M16（P0）法务合规 + 全局品牌可配置

**现状**：全仓 grep 无 ToS/隐私/ICP/备案；登录页与 landing 底部无法务链接；landing 站名/主题色硬编码。**国内服务器没有 ICP 备案位 + 法务页 = 无法合规上线。**

**后端改动**：
1. `system_settings` 增运营字段：`site_name`、`site_logo_url`、`site_icp`（ICP 备案号）、`site_police_beian`（公安备案号）、`tos_url`/`privacy_url`（或内置 Markdown 页 slug）、`contact_email`、`brand_primary_color`（可选主题色覆盖）。迁移 `0011_site_settings.up.sql`，给合理默认值。
2. （可选）`GET /legal/:slug` 渲染存于 settings 的 Markdown 法务文本，免单独建表。

**前端改动**（`frontend/app/src/`）：
1. **全局 Footer 组件**：全站底部展示 站名/ICP 号（链接到 beian.miit.gov.cn）/公安备案/联系邮箱/ToS/隐私 链接；从 `/site-config` 公开接口拉取（无需登录）。
2. **法务页** `routes/legal/[slug].tsx`：渲染 ToS / 隐私政策（Markdown）。
3. **注册/登录页**：底部「注册即代表同意《服务条款》《隐私政策》」勾选（必勾才能提交）。
4. **品牌可配置**：landing 与 AppShell 的站名/logo/主题色从 `/site-config` 读，移除硬编码。

**验收 M16**：footer 全站可见且 ICP/公安号正确链接；法务页可访问；注册必须勾选同意条款；admin 改 `site_name` 后前台即时生效（含 `<title>`）。

## 7.4 里程碑 M17（P1）用户自助：接入文档页 + Key 易用性

**现状**：用户拿到 API Key 后**不知道怎么用**——无 base_url 展示、无 curl/SDK 示例、无能力矩阵。这是「半成品感」最强的地方。

**前端改动**（`frontend/app/src/routes/`）：
1. **接入指引页** `docs.tsx`（或 `quickstart.tsx`）：
   - 展示完整 `base_url`（从 `/site-config` 读，不硬编码）。
   - `/v1/chat/completions`（OpenAI 兼容）与 `/v1/messages`（Anthropic 原生）两套示例，各给 **curl / Python / Node** 三段可复制代码，key 用占位符。
   - Claude Code / Codex 接入片段（与 dashboard 的快速配置复用同一份 ConfigSnippet）。
   - 能力矩阵表：模型 × 是否支持 流式/工具/视觉（数据来自模型表的能力标记，**只读对客信息**）。
2. **api-keys 页增强**（接 M12-6.5.4 待补项）：每个 key 加「复制 base_url + 示例」按钮；端点白名单复选框（依赖下方后端字段）。

**后端配套**（M17 后端，亦是 M12 api-keys 的前置）：
- `api_keys` 增 `allowed_endpoints TEXT[]`（或 JSON）字段，迁移 `0012_key_endpoints.up.sql`，默认空=全部放行（存量不回退）。网关鉴权时校验请求端点是否在白名单内。
- 模型能力标记：`models` 增 `supports_stream/supports_tools/supports_vision BOOL`（默认按 provider 推断），供能力矩阵展示。

**验收 M17**：接入页三语言示例可复制即用；能力矩阵与模型配置一致；设置了端点白名单的 key 调用非白名单端点被拒；存量无白名单的 key 行为不变。

## 7.5 里程碑 M18（P1）admin 用户运营 + 数据导出

**现状**：`admin_user_service.go` 有 `AdjustBalance` 和改 status，但缺集中的用户运营面；report 只返 JSON，用户/财务无法对账。

**后端改动**：
1. **用户运营动作**（`handler/admin/` + service）：在用户详情聚合「调余额 / 调用户级 RPM/并发上限 / 封禁解封 / 重置密码（发邮件）/ 强制下线（吊销 token）」。用户级 RPM/并发若无字段则 `users` 增 `rpm_limit/concurrency_limit`（默认 0=不限，覆盖优先级：key > 用户 > 全局默认）。迁移 `0013_user_ops.up.sql`。
2. **CSV 导出**（流式，避免大内存）：`GET /admin/usage/export.csv`、`/admin/ledger/export.csv`、用户侧 `GET /user/usage/export.csv`（**仅导出对客字段，铁律 2**）。直接 `csv.Writer` 流式写 response，无需 xlsx 依赖。

**前端改动**：
1. admin 用户详情页 `Sheet`：上述运营动作五连按钮（破坏性操作走确认弹窗）。
2. usage/ledger/logs 列表加「导出 CSV」按钮（admin 全量；user 仅自己且脱敏）。

**验收 M18**：admin 可调额/封禁/重置/下线并留审计日志；CSV 导出字段正确、user 端导出无敏感字段；大数据量导出不 OOM（流式）。

## 7.6 里程碑 M19（P1）可观测告警 + 渠道自愈

**现状**：`ops_service.go` 只查询聚合，**无告警规则、无通知出口**；渠道 `fail_count/health` 已有但无自动动作。渠道挂了 admin 最后一个才知道。

**后端改动**：
1. **告警规则引擎**（`backend/internal/service/ops_alert_service.go`，参考 sub2api `ops_alert_evaluator_service.go`）：定时（如每分钟）评估规则——渠道连续失败 N 次、网关 5xx 率超阈值、上游 401/429 比例超阈值、用户余额低于阈值。规则可存 `system_settings` 或新表 `alert_rules`。
2. **通知出口**（`notification_service.go`）：邮件（复用 M14 SMTP）+ Webhook（企业微信/飞书/Discord 通用 JSON）。`system_settings` 配收件人/webhook url。
3. **渠道自愈**：连续失败达阈值自动置 `disabled` 并告警（与 M3 冷却互补——M3 是请求内排除，M19 是跨请求持久熔断 + 通知）。

**前端改动**：admin settings 增「告警」tab（规则开关 + 阈值 + 通知渠道配置）；ops 页顶部展示当前活跃告警。

**验收 M19**：构造渠道连续失败触发自动 disabled + 通知；余额低阈值触发用户邮件预警；Webhook 收到结构化告警。

## 7.7 里程碑 M20（P2）公开状态页 + SEO/品牌收尾

**后端**：`GET /status`（公开，无需登录）输出渠道健康概览 + 近 24h 成功率（**仅聚合健康度，不泄漏渠道密钥/上游名**），复用 ops 聚合数据脱敏后下发。

**前端**：
1. **状态页** `routes/status.tsx`：各服务/渠道分组的健康指示 + 24h 可用率 + 最近事件。对外可分享，降低客服压力。
2. **SEO/品牌**：`<title>`/favicon/OG 图/`robots.txt`/`sitemap.xml`；这些与 M16 品牌可配置打通（站名改了 title 跟随）。
3. **健康/版本接口**：`GET /healthz`、`GET /version` 公开，便于监控与用户排障。

**验收 M20**：状态页未登录可访问且无敏感泄漏；改站名后 `<title>`/OG 同步；`/healthz` 返回 200。

---

## 第七部分执行顺序与铁律

**建议顺序**：M14（邮箱，最阻断）→ M15（防爆破）→ M16（合规，上线硬门槛）为 **P0 三连，必须先于对外开放**；M17（接入文档）→ M18（运营+导出）→ M19（告警）为 **P1 上线即补**；M20（状态页/SEO）为 **P2 增长期**。后端字段迁移（0010–0013）均 `DEFAULT` 安全、存量不回退。

> **第七部分铁律**：所有新增（邮箱/风控/合规/告警/导出/状态页）均在**用户生命周期与运营层**，**不触碰计费公式**；任何对外/对用户的导出与状态页**继续遵守铁律 2**（不泄漏 base_cost/rate_multiplier/毛利/upstream_model_name/渠道密钥）。变现仍**仅兑换码**，不做在线支付/订阅/联盟/风控套件。



