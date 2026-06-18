# 灵枢 LingShu v1.3 精致中转站 - 完整规划方案

> 整理人：Kiro
> 日期：2026-06-18
> 输入：v1.2 上线后用户反馈 7 项 + Kiro 自查发现
> 定位：从"手工绑定模式"升级为"渠道驱动 + 模型自动同步"的现代中转架构（参考 one-api / new-api / sub-api）

---

## 一、决策摘要

**不直接推 v1.2 上线**。当前 v1.2 已有 2 个 commit 待推，但存在 **2 个 P0 bug**（渠道解绑残留、封禁会话不失效），必须先收口。

**三阶段执行**：

| 阶段 | 范围 | 是否阻塞上线 |
|---|---|---|
| **v1.2 收尾**（1-2 天） | H1 渠道解绑修复、H2 封禁失效修复、H3 admin 清理 UI、H4 文档澄清 | ✅ 阻塞 |
| **v1.3 核心**（1 周） | C1 批量模型同步、C2 模型映射、C3 计费 M 单位、C4 渠道-模型 UX 重做 | 独立版本 |
| **v1.3 精致**（视进度） | P1 令牌/模型/用户分组、P2 价格表页、P3 监控大盘、P4 用户增长（签到/邀请） | 独立版本 |

---

## 二、v1.2 收尾（必修，阻塞上线）

### H1【🔴】渠道解绑后还显示 + 主列表看不到绑定

**根因 1**：`backend/internal/repository/channels.go:147-153` ChannelDetail SQL 没过滤 `cm.status='enabled'`，加上 `UnbindModel` 是软删（`UPDATE ... SET status='disabled'`），解绑后记录还在。

**步骤**：
1. `channels.go:147-153` 的 SQL 改为：
   ```sql
   SELECT cm.id, cm.model_id, m.public_name, cm.upstream_model_name, cm.status, cm.created_at
   FROM channel_models cm
   JOIN models m ON m.id = cm.model_id AND m.deleted_at IS NULL
   WHERE cm.channel_id=$1 AND cm.status='enabled'
   ORDER BY cm.created_at DESC
   ```
2. `models.go` 同样有反查 channel_models 的地方（搜 `channel_models`），统一加 `AND cm.status='enabled'`。
3. `gateway.go:108` 的路由 SQL 已经有 `cm.status='enabled'`，不动。

**根因 2**：admin 主列表 `ChannelsPage` 只显示 name/status/health 等字段，没"已绑模型数"或"绑定列表"入口，用户绑了不知道在哪看，会怀疑没绑上。

**步骤**：
1. 后端 `ChannelRepository.ListPaged` 多带一个聚合字段 `bound_count`：在原 SQL 加 `LEFT JOIN (SELECT channel_id, COUNT(*) FROM channel_models WHERE status='enabled' GROUP BY channel_id) AS b ON b.channel_id=c.id`。
2. `Channel` struct 增 `BoundCount int json:"bound_count"`。
3. 前端 `channels.tsx` 渠道列表加列「已绑模型」显示 `bound_count`，点击跳详情。

**验收**：建 2 个渠道 → 各绑 2 个模型 → 主列表分别显示 2 / 2 → 解绑 1 个 → 主列表变 1，详情列表里那条也消失。

---

### H2【🔴 安全】用户封禁后会话不失效

**双重根因**：
- `middleware/auth.go` `JWTAuth` 中间件**只验签不查 DB**，被封禁用户的 JWT 在到期前依然有效（无状态 token 的固有问题）。
- `APIKeyAuth` 中间件可能没 enforce `principal.UserStatus`（需要 agent 复查）。

**步骤**：

**方案 A（推荐，简单）**：每次请求查一次 DB 兜底
1. `JWTAuth` 中间件解析 JWT 后追加一次 `userRepo.FindByID(ctx, claims.UserID)` 查询，检查 `status='active'`，否则 401。读取开销很小（单条 PK 查询 +cache）。
2. `APIKeyAuth` 中间件确认 `principal.UserStatus == "active" && principal.KeyStatus == "active"`，否则 401。`FindPrincipalByHash` 已经 SELECT 了 status，handler 必须用。
3. `FindPrincipalByHash` 加 Redis 缓存（key=`apikey:{hash}`，TTL 60s），减少 DB 压力。**封禁用户时主动 DEL 这个 key**（在 `Ban` service 里调一次 `redisClient.Del`）。

**方案 B（更严，复杂）**：token version
1. `users` 表加 `token_version INT DEFAULT 1`。
2. JWT claims 加 `tv` 字段，签发时写当前 token_version。
3. 中间件解 JWT 后查 DB 比 tv，不一致就 401。
4. 封禁时 `UPDATE users SET token_version=token_version+1`，所有旧 JWT 立刻失效。

**建议先 A，后续再 B**。封禁时务必：
- DB 改 `users.status='banned'`
- DEL Redis 的 apikey 缓存（如果实现了 A 的缓存）
- 把该用户所有 API Key 一并 disable（已有 `BanWithCleanup`？没有就加）

**验收**：用户登录拿 JWT → admin 封禁 → 用户立刻拿 401（不能等 token 自然过期）；用户的 API Key 调用也立刻 401。

---

### H3【🟡】admin 加"立即清理"UI

后端 `POST /api/admin/cleanup/run` 和 `GET /history` 已经通，只缺前端按钮。

**步骤**：`frontend/admin/src/pages/settings.tsx` 文件末尾加一个新 Card：
```tsx
<Card title="系统清理">
  <Space direction="vertical" style={{ width: '100%' }}>
    <Alert message="清理 30 天前的请求日志、90 天前的审计日志、过期公告与失效兑换码。账本数据永不清理。" type="info" />
    <Space>
      <Button type="primary" loading={cleaning}
        onClick={() => Modal.confirm({
          title: '确认立即清理？', okText: '执行',
          onOk: () => runWrite(async () => {
            setCleaning(true);
            const result = await api.runCleanup();
            message.success(`清理完成：${result.map(r => `${r.table} -${r.deleted}`).join(', ')}`);
            await refreshHistory();
            setCleaning(false);
          }, '清理失败')
        })}>立即清理</Button>
      <Button onClick={refreshHistory}>刷新历史</Button>
    </Space>
    <Table size="small" rowKey="id" dataSource={history}
      columns={[
        { title: '开始时间', dataIndex: 'started_at', render: v => new Date(v).toLocaleString() },
        { title: '结束时间', dataIndex: 'ended_at', render: v => new Date(v).toLocaleString() },
        { title: '清理结果', dataIndex: 'result_json', render: r => r.map(x => `${x.table}: -${x.deleted}`).join('  ') }
      ]} pagination={false} />
  </Space>
</Card>
```

`api-client.ts` 加 `runCleanup: () => request<CleanupResult[]>('/api/admin/cleanup/run', { method: 'POST' })` 和 `cleanupHistory: (limit?: number) => request<...>('/api/admin/cleanup/history', ...)`。

---

### H4【📘】文档澄清：消除 #2 #4 误解

写在 `docs/api.md` 顶部加一段「核心概念」：

```
LingShu 的统一入口是"平台 API Key"。一个 Key 即可调用所有已上线模型，
只需在 OpenAI SDK 的 model 字段填模型 public_name（如 gpt-5.5、claude-opus-4-7）。
不需要为每个上游创建独立 Key。

模型可见性 = 该模型 status='enabled' 且至少绑定了一个 enabled 渠道。
管理员新增渠道不会自动新增模型——模型和渠道是 N:N 关系，需要在管理端绑定。
v1.3 起支持"渠道→批量拉取上游 /models→一键导入为模型"功能（见 C1）。
```

---

## 三、v1.3 核心功能（解决用户的真实痛点）

### C1【最重要】渠道批量同步模型（解决 #5）

**目标**：管理员添加渠道后，点一个按钮自动从上游拉所有可用模型，勾选要的模型一键创建+绑定。

**步骤**：

**后端**：
1. `backend/internal/upstream/provider.go` Provider 接口新增方法：
   ```go
   type ProviderModel struct {
       ID          string  // 上游模型名，如 gpt-4o-mini
       Type        string  // chat/embedding/image (从上游推断或默认 chat)
       Owned       string  // 上游 owned_by
   }
   ListModels(ctx context.Context, baseURL, apiKey string) ([]ProviderModel, error)
   ```
2. `OpenAIAdapter.ListModels`：GET `{baseURL}/models` 解析 OpenAI 格式。
3. `AnthropicAdapter.ListModels`：Anthropic 有 `/v1/models`（2024 后开放），解析它的格式；若上游不支持就返回静态预设列表（claude-3-5-sonnet-20241022 等）。
4. 新 handler `POST /api/admin/channels/{id}/sync-models`：
   - 查渠道，解密 key
   - 调 adapter.ListModels
   - 返回 `{upstream_models: ProviderModel[], existing_bindings: ChannelBinding[]}`，前端展示差集
5. 新 handler `POST /api/admin/channels/{id}/import-models`：
   - body: `{models: [{upstream_name, public_name, billing_mode, input_price_per_m, output_price_per_m, rate_multiplier, type}], strategy: "create_or_bind"|"bind_existing"}`
   - 事务里：每条若 `public_name` 在 models 表不存在则 INSERT，然后 INSERT channel_models 绑定
   - 支持 `ON CONFLICT (channel_id, model_id) DO UPDATE` 复用已有绑定

**前端**：
- `ChannelDetailPage` 加按钮「同步上游模型」→ 弹 Modal 展示拉取结果
- Modal 内是表格：上游模型名 / 建议 public_name（默认 = 上游名） / 计费方式 / 输入价 / 输出价 / 倍率 / 状态（新建 / 已存在仅绑定 / 已绑定）
- 顶部「批量设置」：一键给所有勾选项填同一倍率、同一价格
- 「确认导入」一次性建 + 绑

**验收**：新加一个 OpenAI 渠道 → 点同步 → 看到上游 80+ 模型列表 → 勾选 5 个 → 一键导入，模型表多 5 条且全部绑定到该渠道。

---

### C2【关键】模型映射 / 模型别名（让用户端友好）

**目标**：用户调 `gpt-5.5`，网关自动路由到不同渠道的真实模型（A 渠道是 `gpt-4o-mini-2024-07-18`，B 渠道是 `gpt-4o-mini`），用户感受到的是"统一抽象模型名"。

**现状**：当前 `channel_models.upstream_model_name` 已经做了渠道级映射（绑定时填）。**缺失的是统一的 public_name 抽象**——用户没意识到可以这么用。

**步骤**：
1. C1 导入时默认 `public_name = 上游名`，但允许编辑——这是关键！管理员可以让 `public_name=gpt-mini`，三个渠道分别绑 `gpt-4o-mini` / `gpt-4o-mini-2024-07-18` / `gpt-4o-mini-2024-12-17`。用户用 `gpt-mini` 调用，网关按权重/健康度选渠道。
2. 模型详情页要清楚展示「绑定渠道列表 + 各自的上游模型名」（已有，确认 UX 清晰）。
3. 文档示例必须画一张图：用户 → `model:gpt-mini` → 网关 → 选渠道 → 实际调用上游 `gpt-4o-mini-2024-07-18`。

**额外**：模型分组 = 把多个 public_name 归类（如"经济型"/"旗舰"），用户端按组展示，UX 更清爽（见 P1）。

---

### C3【UX】计费单位改 M

**步骤**：
1. **DB 字段不动**（`input_price_per_1k` / `output_price_per_1k`），避免破坏性迁移。
2. 前端**输入与展示统一显示 per 1M**，提交时除 1000 存 DB，显示时乘 1000。
3. 新建辅助函数 `frontend/packages/shared/src/pricing.ts`：
   ```ts
   export const perKToM = (perK: string | number) => (Number(perK) * 1000).toFixed(6);
   export const perMToK = (perM: string | number) => (Number(perM) / 1000).toFixed(8);
   ```
4. 所有模型表单（admin model-form.tsx、C1 导入 Modal）：label 改"输入价 / 1M token"，提交时 `perMToK`，回显 `perKToM`。
5. 模型详情、价格表(P2)、用户端价格展示统一 per 1M。

---

### C4【UX】渠道-模型管理重做

**当前痛点**：
- 创建渠道在一行 Form.inline，渠道字段塞不下
- 绑定渠道-模型用第三张表单，操作分散
- 渠道列表看不到健康度趋势 / 已绑模型
- 列表没有搜索

**步骤**：
1. `ChannelsPage` 创建表单改 vertical Form，加 Drawer 编辑模式。
2. 列表加列：`bound_count`、`last_latency_ms`（C4.1 增字段）、`last_success_at`。
3. 列表加搜索框（按 name 模糊）+ 过滤（provider_type / status / health）。
4. 详情页加：
   - 顶部 metric cards（已有）
   - 「同步上游模型」按钮（C1）
   - 「绑定列表」表格 + 批量解绑
   - 「最近测试」面板（最近 10 次测试结果，latency 趋势小图）
5. 模型详情页同步加搜索/筛选。

---

## 四、v1.3 精致 UX（参考 one-api / new-api / sub-api）

### P1 模型分组 / 令牌分组 / 用户分组

**模型分组**：`models` 表加 `group_name TEXT`，admin 创建模型时填或导入时填。用户端 `/api/user/models` 返回时按 group 聚合：`{groups: [{name:"经济型", models:[...]}]}`。

**令牌分组（API Key 分组）**：让用户能为不同业务创建带标签的 Key，列表里按标签筛选。`api_keys` 加 `group_name`。

**用户分组（VIP/普通）**：`users` 加 `group_name`，不同组享受不同倍率折扣或模型可见性。admin 设置分组规则。

### P2 公开价格表页

新增 `/pricing` 用户端页面（无需登录可见）：
- 按模型分组展示
- 输入/输出价（per 1M）
- 倍率、计费方式
- 数据源 `/api/public/pricing` 端点
- 营销价值高，外部用户能直接看到价格表

### P3 监控大盘

admin dashboard 增强：
- 实时 QPS 折线
- Top 10 模型用量
- Top 10 用户消费
- 渠道健康度热力图（绿/黄/红）
- 余额预警用户列表（< 阈值）
- 数据基于 v1.1 已有的 reports.go

### P4 用户增长

- 邀请码：用户邀请新人，双方获奖励余额
- 签到：每日签到送少量余额
- 充值套餐：充 100 送 10 这种营销
- 余额预警：低于阈值站内通知（v1.2 方案里 B1）

---

## 五、Kiro 额外建议（你"还有没想到的"）

| 项 | 价值 | 紧急度 |
|---|---|---|
| **限流分级**：当前 rpm/concurrency 是单 key 维度，缺用户维度（一个用户多 key 总和限制） | 防滥用 | 中 |
| **WebSocket / 长连接计费**：当前仅支持 chat completions，Anthropic computer use、OpenAI realtime 是长连接计费完全不同 | 未来 | 低 |
| **嵌入模型 / TTS / 图像**：endpoint 只暴露 `/chat/completions` 和 `/models`，没 `/embeddings`、`/audio/speech`、`/images/generations`。要做综合中转必须补 | 功能 | 高 |
| **请求级日志检索**：现在 logs 按用户/模型/日期筛，缺按 request_id / content 关键字检索 | 排错 | 中 |
| **流式响应中断处理**：客户端中途断开时，当前应该是 Reserve 已扣，正常 Finalize 会按实际 usage 结算——需要验证 |  风险 | 高 |
| **多租户隔离**：当前是单租户，未来要给企业卖账号需要 organization 表 | 商业化 | 低 |
| **Docker 镜像瘦身**：当前镜像应该几百 MB，alpine + multi-stage 可压到 50MB 以内 | 部署 | 低 |
| **管理员操作日志暴露**：已有 audit_logs，但管理端没好用的检索面板（filter 是有了但没大屏） | 合规 | 中 |
| **Prometheus metrics**：`/metrics` 端点暴露 QPS / latency p50/p95/p99 / 错误率，对接 Grafana | 可观测 | 中 |
| **失败请求自动重试 + 渠道自动切换**：当前重试只在同渠道，应该在主渠道连续失败 N 次后自动切到备用渠道 | 可靠性 | 高 |

---

## 六、执行顺序与给 agent 的指令

### 立即执行（v1.2 收尾，发 agent 做）

> 按本文档第二节 H1→H2→H3→H4 顺序执行。这四项做完跑全套构建测试，commit + push。完成后等用户验收，**不要**继续做 v1.3。

### v1.2 验收通过后

我（Kiro）会基于这份方案的第三节为 C1→C2→C3→C4 各自准备独立的执行指令文档，每个 C 单独发给 agent，避免一次塞太多。

### v1.3 第四节（P1-P4）

视 v1.3 核心稳定运行 1-2 周后再启动，期间收集真实用户反馈调整优先级。

---

## 七、回到你的决策点

**我的明确建议**：

1. **现在让 agent 做 v1.2 收尾的 H1+H2+H3+H4**（1-2 天完成），完成后推上线。这是阻塞项。
2. **跳过原方案的"直接推 A"**——封禁不失效是安全 P0，不能带病上线。
3. **C1 批量同步是 v1.3 第一个动手的功能**，因为它直接解决你最痛的 #5 + 顺带解释 #2。
4. **计费 M 单位（C3）** 跟 C1 一起做，导入界面默认就用 M。
5. **公告 markdown 已完成**，问题 1/2/6 的诊断我已写明，无需你做额外判断。

你要不要现在就开 v1.2 收尾的执行指令发给 agent？
