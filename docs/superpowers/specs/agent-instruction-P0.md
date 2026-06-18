# 灵枢 v1.1 P0 返工任务 — Agent 执行指令

## 任务目标

按照 `灵枢-v1.1-上线返工方案.md` 的 **P0 范围**完成以下工作，使线上版本从"能跑"变成"可运营、可理解、可操作"。

---

## 执行优先级（必须按顺序）

### 阶段 1：后端安全收口（最高优先级，先做）

**目标**：剥离 user-scoped 接口响应中的敏感字段，确保用户即便直接 curl 也看不到进价和毛利。

#### 任务清单

1. **新建 DTO 层**：
   - 在 `backend/internal/dto/` 目录下新建以下文件：
     - `user_dashboard_dto.go` — 定义 `UserDashboardDTO`，去掉 `GrossProfit`
     - `user_ledger_dto.go` — 定义 `UserLedgerRecordDTO`，去掉 `BaseCost`、`RateMultiplier`
     - `user_log_dto.go` — 定义 `UserGatewayLogDTO`，去掉 `BaseCost`、`RateMultiplier`
     - `user_model_dto.go` — 定义 `UserModelConfigDTO`，去掉 `RateMultiplier`、`InputPricePer1K`、`OutputPricePer1K`、`PricePerCall`（只保留用户需要看的：`PublicName`、`Type`、`Group`、`BillingMode`、`Status`、`SortOrder`）

2. **修改 user handler**：
   - 编辑 `backend/internal/handler/user_handler.go`：
     - `/api/user/dashboard` 接口：将 `repository.AdminDashboard` 转换为 `dto.UserDashboardDTO` 后返回
     - `/api/user/usage/ledger` 接口：将 `[]repository.LedgerRecord` 转换为 `[]dto.UserLedgerRecordDTO` 后返回
     - `/api/user/usage/logs` 接口：将 `[]repository.GatewayLog` 转换为 `[]dto.UserGatewayLogDTO` 后返回
     - `/api/user/models` 接口：将 `[]repository.ModelConfig` 转换为 `[]dto.UserModelConfigDTO` 后返回

3. **验证**：
   - `go build ./...` 通过
   - 启动服务，用 curl 直接请求 `http://localhost:8080/api/user/dashboard`（带 Bearer token），确认响应体里不再有 `gross_profit`、`base_cost`、`rate_multiplier`

#### 涉及文件
- `backend/internal/dto/` （新建目录和 4 个文件）
- `backend/internal/handler/user_handler.go` （修改 4 个接口）

---

### 阶段 2：前端类型拆分

**目标**：将 shared types 拆分为 AdminTypes 和 UserTypes，与后端 DTO 对齐。

#### 任务清单

1. **拆分类型文件**：
   - 在 `frontend/packages/shared/src/` 下新建：
     - `admin-types.ts` — 导出完整类型（现有 `types.ts` 里的所有类型）
     - `user-types.ts` — 导出精简类型：
       - `UserDashboard` — 去掉 `gross_profit`
       - `UserLedgerRecord` — 去掉 `base_cost`、`rate_multiplier`
       - `UserGatewayLog` — 去掉 `base_cost`、`rate_multiplier`
       - `UserModelConfig` — 只保留 `id`、`public_name`、`type`、`group`、`billing_mode`、`status`、`sort_order`

2. **修改 api-client.ts**：
   - user-scoped 接口（`userDashboard`、`userLedger`、`userLogs`、`userModels`）的返回类型改为 `user-types.ts` 里的类型
   - admin-scoped 接口保持用 `admin-types.ts` 里的类型

3. **更新前端引用**：
   - `frontend/user/src/` 下所有文件：从 `@lingshu/shared` 导入改为从 `@lingshu/shared/user-types` 导入
   - `frontend/admin/src/main.tsx`：从 `@lingshu/shared` 导入改为从 `@lingshu/shared/admin-types` 导入

4. **验证**：
   - `npm --workspace @lingshu/user run build` 通过
   - `npm --workspace @lingshu/admin run build` 通过

#### 涉及文件
- `frontend/packages/shared/src/admin-types.ts` （新建）
- `frontend/packages/shared/src/user-types.ts` （新建）
- `frontend/packages/shared/src/api-client.ts` （修改返回类型）
- `frontend/packages/shared/src/index.ts` （导出拆分后的类型）
- `frontend/user/src/**/*.tsx` （批量修改 import）
- `frontend/admin/src/main.tsx` （修改 import）

---

### 阶段 3：用户端中文化 + 脱敏

**目标**：清理所有中英混用、枚举裸显、内部概念泄露。

#### 任务清单

1. **新建枚举映射文件**：
   - `frontend/user/src/lib/i18n.ts`：
     ```typescript
     export const statusMap: Record<string, string> = {
       active: "启用",
       disabled: "禁用",
       enabled: "启用",
       success: "成功",
       failed: "失败",
     };
     
     export const billingModeMap: Record<string, string> = {
       token: "按量计费",
       per_call: "按次计费",
     };
     
     export const typeMap: Record<string, string> = {
       chat: "对话",
       embedding: "向量",
       image: "图像",
       video: "视频",
     };
     
     export const ledgerTypeMap: Record<string, string> = {
       charge: "扣费",
       redeem: "兑换",
       refund: "退款",
       adjust: "调整",
     };
     ```

2. **修改 app-layout.tsx**：
   - 导航栏全部中文：`Dashboard → 概览`，`API Keys → API 密钥`，`Usage → 用量`，`Models → 模型`，`Redeem → 充值`，`News → 公告`，`Settings → 设置`
   - 删除英文标语 `"Private AI API gateway"`

3. **修改 login.tsx**：
   - 删除英文标语 `"A calm cockpit for private AI traffic."`，改为 `"私有 AI 网关，安全可控"`
   - 删除 `"追踪扣费和倍率毛利链路"` 改为 `"统一接入，余额透明"`
   - 删除 `"不再粘贴 JWT"` 改为 `"账号密码登录"`
   - 删除 `"余额事务行锁"` 改为 `"余额实时扣费"`

4. **修改 dashboard.tsx**：
   - 删除 `"Postgres 持久余额"` 改为 `"账户余额"`
   - 删除 `"Redis frozen reservation"` 改为 `"预扣金额"`
   - 删除 `"按本月 gateway_requests 聚合"` 改为 `"本月请求数"`
   - 表格中 `"X tokens"` 改为 `"X 个 token"`

5. **修改 api-keys.tsx**：
   - 表头 `"API Keys"` 改为 `"API 密钥"`
   - 删除 `"OpenAI SDK 的 base_url 指向 LingShu"` 改为 `"在 OpenAI SDK 中配置 base_url 指向本平台地址"`
   - 状态用 `statusMap` 映射显示

6. **修改 usage.tsx**：
   - 表头 `"Usage"` 改为 `"用量统计"`
   - `` `${item.total_tokens} tokens` `` 改为 `` `${item.total_tokens} tokens` ``（保留 tokens 因为是专业术语）
   - 状态用 `statusMap`、`ledgerTypeMap` 映射显示

7. **修改 models.tsx**：
   - 表头 `"Models"` 改为 `"可用模型"`
   - **删除倍率显示**：移除 `<Price label="倍率" value={...} />` 整行
   - 删除 `"展示基准成本乘倍率后的实际扣费口径..."` 改为 `"以下为各模型的计费标准"`
   - 状态、类型、计费模式用映射表显示

8. **修改 redeem.tsx**：
   - 表头 `"Redeem"` 改为 `"充值兑换"`
   - 删除 `"私有运营不接在线支付。余额来源只有管理员手动充值和兑换码。"` 改为 `"请输入管理员提供的兑换码进行充值"`
   - 删除 `"兑换成功后，账本会同时记录 redeem 类型，余额变动可追溯。"` 改为 `"兑换成功后余额即时到账"`

9. **修改 settings.tsx**：
   - 表头 `"Settings"` 改为 `"账户设置"`
   - **删除充值方式行**：移除 `<Row label="充值方式" value="..." />`
   - **删除扣费公式行**：移除 `<Row label="扣费公式" value="base_cost × rate_multiplier" />`
   - 删除 `"密码修改仍使用后端 /api/auth/change-password"` 改为 `"如需修改密码请联系管理员"`

10. **添加全局 Toast**：
    - 在 `main.tsx` 添加 `<Toaster />` 组件（shadcn）
    - 所有异步操作（login、createKey、deleteKey、redeem 等）添加 `toast.success()` / `toast.error()` 反馈

11. **添加 Error Boundary**：
    - 新建 `components/error-boundary.tsx`，捕获组件错误显示友好提示
    - 在 `main.tsx` 根组件外包裹

#### 涉及文件
- `frontend/user/src/lib/i18n.ts` （新建）
- `frontend/user/src/components/app-layout.tsx`
- `frontend/user/src/components/error-boundary.tsx` （新建）
- `frontend/user/src/routes/login.tsx`
- `frontend/user/src/routes/dashboard.tsx`
- `frontend/user/src/routes/api-keys.tsx`
- `frontend/user/src/routes/usage.tsx`
- `frontend/user/src/routes/models.tsx`
- `frontend/user/src/routes/redeem.tsx`
- `frontend/user/src/routes/settings.tsx`
- `frontend/user/src/main.tsx`

---

### 阶段 4：管理端路由重构 + 中文化

**目标**：将静态死的侧栏换成真正的 react-router + AntD Menu，拆分单文件为多页面。

#### 任务清单

1. **安装依赖**：
   - `cd frontend/admin && npm install react-router-dom antd-dayjs-webpack-plugin`

2. **新建路由配置**：
   - `frontend/admin/src/routes.tsx` — 定义路由表

3. **新建布局组件**：
   - `frontend/admin/src/layouts/AdminLayout.tsx` — 左侧 Menu + Header + Content
   - Menu 配置：
     ```typescript
     [
       { key: "/dashboard", icon: <DashboardOutlined />, label: "概览" },
       { key: "/users", icon: <TeamOutlined />, label: "用户管理" },
       { key: "/api-keys", icon: <KeyOutlined />, label: "API 密钥" },
       { key: "/models", icon: <AppstoreOutlined />, label: "模型管理" },
       { key: "/channels", icon: <CloudOutlined />, label: "渠道管理" },
       { key: "/announcements", icon: <NotificationOutlined />, label: "公告管理" },
       { key: "/redeem", icon: <GiftOutlined />, label: "兑换码" },
       { key: "/reports", icon: <BarChartOutlined />, label: "数据报表" },
       { key: "/settings", icon: <SettingOutlined />, label: "系统设置" },
       { key: "/audit", icon: <AuditOutlined />, label: "审计日志" },
     ]
     ```
   - Menu 的 `onClick` 调用 `navigate(key)`

4. **拆分页面组件**：
   - 从 `main.tsx` 的 Tabs 里提取出来，每个 Tab 内容变成独立页面：
     - `pages/Dashboard.tsx` — 统计卡 + Summary
     - `pages/Users.tsx` — UsersPane
     - `pages/APIKeys.tsx` — KeysPane
     - `pages/Models.tsx` — ModelsPane
     - `pages/Channels.tsx` — ChannelsPane
     - `pages/Announcements.tsx` — AnnouncementsPane
     - `pages/RedeemCodes.tsx` — RedeemPane
     - `pages/Reports.tsx` — ReportsPane
     - `pages/Settings.tsx` — SettingsPane
     - `pages/AuditLogs.tsx` — 审计日志表格

5. **重构 main.tsx**：
   - 删除 `<Tabs>`、`MenuPlaceholder`、所有 Pane 组件
   - 改为 `<RouterProvider router={router} />`，router 用 `createBrowserRouter` 定义
   - 保留 `refresh()` 逻辑，但改为在各页面按需调用（通过 Context 或 zustand）

6. **配置 ConfigProvider**：
   - 在 `Theme` 组件里添加 `locale={zhCN}`（从 `antd/locale/zh_CN` 导入）

7. **统一 Toast/Message**：
   - 删除所有 `<Alert>` 一次性提示
   - 改用 `message.success()` / `message.error()`

8. **表格空状态中文化**：
   - 所有 `<Table>` 组件添加 `locale={{ emptyText: "暂无数据，点击上方按钮创建" }}`

9. **去掉与侧栏冲突的统计卡**：
   - Dashboard 页面保留统计卡，但侧栏 Menu 不再重复展示（删除原 `main.tsx` 里顶部 Summary 组件重复渲染）

#### 涉及文件
- `frontend/admin/src/routes.tsx` （新建）
- `frontend/admin/src/layouts/AdminLayout.tsx` （新建）
- `frontend/admin/src/pages/` （新建目录 + 10 个页面文件）
- `frontend/admin/src/main.tsx` （大幅重构，保留 App 入口和 Theme，删除所有业务逻辑）

---

### 阶段 5：补齐基础 CRUD

**目标**：补全编辑/删除/禁用/重启用/解绑等操作。

#### 任务清单

1. **用户详情页**（新建 `pages/UserDetail.tsx`）：
   - 路由：`/users/:id`
   - 显示：余额、今日消费、API Key 列表（带禁用/删除按钮）、最近 20 条账本、最近 20 条请求日志
   - 用户列表表格的用户名列改为可点击，跳转到详情页

2. **模型删除**：
   - 在 `ModelColumns` 的操作列添加"删除"按钮
   - 调用 `api.deleteModel(id)`（需后端补接口：`DELETE /api/admin/models/:id`）
   - 删除前弹 `Modal.confirm` 确认

3. **渠道编辑/删除/解绑**：
   - 渠道表格添加"编辑"、"删除"按钮
   - 编辑弹窗修改 base_url/api_key/weight/status
   - 新增"解绑模型"表单：选择渠道 + 模型，调用 `api.unbindChannelModel(channel_id, model_id)`（需后端补接口：`DELETE /api/admin/channels/:channel_id/models/:model_id`）

4. **公告编辑/删除**：
   - 公告表格添加"编辑"、"删除"按钮
   - 编辑弹窗修改 title/content/status/pinned
   - 调用 `api.updateAnnouncement(id, ...)` 和 `api.deleteAnnouncement(id)`（需后端补接口）

5. **兑换码禁用**：
   - 兑换码表格添加"禁用"按钮（仅对 status=active 的显示）
   - 调用 `api.disableRedeemCode(id)`（需后端补接口：`PATCH /api/admin/redeem-codes/:id`）

6. **API Key 重命名/重启用**：
   - 用户端 `api-keys.tsx` 添加"重命名"按钮（弹窗输入新 name，调用 `api.updateUserAPIKey(id, { name })`）
   - 添加"重新启用"按钮（仅对 status=disabled 的显示，调用 `api.updateUserAPIKey(id, { status: "active" })`）

#### 涉及文件
- `frontend/admin/src/pages/UserDetail.tsx` （新建）
- `frontend/admin/src/pages/Models.tsx` （添加删除按钮）
- `frontend/admin/src/pages/Channels.tsx` （添加编辑/删除/解绑）
- `frontend/admin/src/pages/Announcements.tsx` （添加编辑/删除）
- `frontend/admin/src/pages/RedeemCodes.tsx` （添加禁用）
- `frontend/user/src/routes/api-keys.tsx` （添加重命名/重启用）
- `frontend/packages/shared/src/api-client.ts` （补充缺失的接口函数）
- **后端需补接口**（标注清单，agent 如果能做就做，不能做就在报告里列出来）：
  - `DELETE /api/admin/models/:id`
  - `PATCH /api/admin/channels/:id`
  - `DELETE /api/admin/channels/:id`
  - `DELETE /api/admin/channels/:channel_id/models/:model_id`
  - `PATCH /api/admin/announcements/:id`
  - `DELETE /api/admin/announcements/:id`
  - `PATCH /api/admin/redeem-codes/:id`

---

### 阶段 6：分页支持

**目标**：管理端所有列表支持分页。

#### 任务清单

1. **后端补分页参数**（标注清单，agent 如果能做就做，不能做就在报告里列出来）：
   - 所有 List 接口接受 `?page=1&limit=20` 参数
   - 返回格式：`{ items: [], total: 100, page: 1, limit: 20 }`
   - 涉及接口：
     - `GET /api/admin/users`
     - `GET /api/admin/api-keys`
     - `GET /api/admin/models`
     - `GET /api/admin/channels`
     - `GET /api/admin/announcements`
     - `GET /api/admin/redeem-codes`
     - `GET /api/admin/logs`
     - `GET /api/admin/ledger`
     - `GET /api/admin/audit-logs`

2. **前端分页**：
   - 所有 `<Table>` 组件添加 `pagination` 配置：
     ```typescript
     pagination={{
       current: page,
       pageSize: limit,
       total: total,
       onChange: (p, l) => { setPage(p); setLimit(l); },
       showSizeChanger: true,
       showTotal: (t) => `共 ${t} 条`,
     }}
     ```
   - 页面内维护 `page`、`limit`、`total` 状态
   - `useEffect` 监听 page/limit 变化重新请求

#### 涉及文件
- `frontend/admin/src/pages/*.tsx` （所有列表页面）
- `backend/internal/handler/admin_handler.go` （补分页逻辑）
- `backend/internal/repository/*.go` （List 方法补 page/limit 参数）

---

## 验收标准（agent 完成后必须自测）

### 后端验收
- [ ] `go build ./...` 通过
- [ ] 启动服务，用 curl 请求 `/api/user/dashboard`（带 token），确认响应不含 `gross_profit`、`base_cost`、`rate_multiplier`
- [ ] 用 curl 请求 `/api/user/usage/ledger`、`/api/user/usage/logs`、`/api/user/models`，确认响应都已剥敏

### 前端验收
- [ ] `npm --workspace @lingshu/user run build` 通过
- [ ] `npm --workspace @lingshu/admin run build` 通过
- [ ] 用户端登录页无英文标语，无"倍率"、"毛利"、"JWT"、"Postgres"等词
- [ ] 用户端模型列表不再显示倍率
- [ ] 用户端设置页不再显示扣费公式和充值方式
- [ ] 用户端所有状态枚举显示中文（启用/禁用/成功/失败/按量计费/按次计费等）
- [ ] 用户端所有异步操作有 toast 反馈
- [ ] 管理端左侧菜单可点击，点击后路由切换正常
- [ ] 管理端所有表格显示中文列名和空状态提示
- [ ] 管理端所有列表支持分页（如果后端接口已补）
- [ ] 用户详情页可正常访问，显示余额、API Key、账本、日志
- [ ] 模型/渠道/公告/兑换码的编辑/删除按钮可见且可操作（如果后端接口已补）

---

## 注意事项

1. **优先级顺序不可打乱**：必须先做阶段 1（后端剥敏），再做阶段 2（类型拆分），最后做前端中文化和重构。否则会出现类型不匹配导致的编译错误。

2. **后端接口缺失时的处理**：如果阶段 5、阶段 6 需要的后端接口当前不存在，agent 应该：
   - 在报告里列出缺失的接口清单（方法、路径、请求体、响应体）
   - 前端先写 UI 和调用代码，接口函数先返回 `Promise.reject(new Error("接口未实现"))`
   - 标注为"待后端补充"

3. **Git 提交策略**：
   - 阶段 1 完成后提交一次：`fix(backend): 剥离 user 接口敏感字段`
   - 阶段 2 完成后提交一次：`refactor(frontend): 拆分 shared types 为 admin/user`
   - 阶段 3 完成后提交一次：`feat(user): 中文化 + 脱敏 + toast`
   - 阶段 4 完成后提交一次：`refactor(admin): 路由重构 + 左菜单 + 中文化`
   - 阶段 5/6 完成后提交一次：`feat(admin): 补齐 CRUD + 分页`

4. **测试环境**：
   - 后端：`cd backend && go run cmd/server/main.go`
   - 前端：`cd frontend && npm --workspace @lingshu/user run dev`（端口 5173）和 `npm --workspace @lingshu/admin run dev`（端口 5174）
   - 全栈：`docker compose up --build`

5. **遇到问题时**：
   - 先检查是否按优先级顺序执行
   - 类型错误优先检查 `admin-types.ts` / `user-types.ts` 是否正确拆分
   - 编译错误优先检查 import 路径是否正确
   - 运行时错误检查后端响应结构是否和前端类型匹配

---

## 完成报告模板

agent 完成后请按以下格式输出报告：

```markdown
# 灵枢 v1.1 P0 完成报告

## 已完成项
- [x] 阶段 1：后端剥敏（涉及文件：...）
- [x] 阶段 2：前端类型拆分（涉及文件：...）
- [x] 阶段 3：用户端中文化 + 脱敏（涉及文件：...）
- [x] 阶段 4：管理端路由重构（涉及文件：...）
- [ ] 阶段 5：补齐 CRUD（部分完成，缺后端接口）
- [ ] 阶段 6：分页支持（前端已完成，后端待补）

## 验收结果
- [ ] 后端编译通过
- [ ] 前端编译通过
- [ ] curl 测试已剥敏
- [ ] 用户端无敏感词
- [ ] 管理端路由可用

## 待补充后端接口清单
1. `DELETE /api/admin/models/:id` — 删除模型
2. `PATCH /api/admin/channels/:id` — 编辑渠道
3. ...

## 遇到的问题
1. 问题描述
2. 解决方案或建议

## 下一步建议
...
```

---

## 开始执行

agent 请按照上述 6 个阶段顺序执行，每完成一个阶段自测通过后再进入下一阶段。遇到无法解决的问题时，记录在报告里并继续后续阶段（如果不依赖当前阶段）。
