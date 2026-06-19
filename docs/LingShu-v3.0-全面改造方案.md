# 灵枢 v3.0 全面改造方案

> **本轮不按旧文档执行。直接按本文件所列 7 大项逐一完成，顺序不可乱，每项自测通过再进下一项。**
>
> 代码位置：`D:\code\LingShu\backend`（Go chi）+ `D:\code\LingShu\frontend`（React 单 app + packages/shared + packages/ui）
>
> 参考项目（只读，不改）：
> - `D:\code\LingShu\Sub` —— sub2api
> - `D:\code\LingShu\new` —— new-api
>
> **三条铁律**：① 计费公式 `charge = base_cost × rate_multiplier` 绝对不改；② user 端绝不返回 base_cost/rate_multiplier/毛利/upstream_model_name/client_ip；③ 新增 DB 字段 DEFAULT 安全、存量不回退。

---

# 第 1 项：侧栏导航分组（一级 → 二级）

## 当前问题

`frontend/app/src/layout/AppShell.tsx:21-45` 两个平铺数组 `userNavItems` 和 `adminNavItems`，所有菜单混在一级展示，太长/无层次。

## 改动

### 1.1 类型定义改造（同文件顶部）

把 `NavItemConfig` 拆成支持分组嵌套：

```ts
type NavItemConfig = {
  to: string;
  labelKey: string;
  fallbackLabel: string;
  icon: React.ComponentType<{ className?: string }>;
  hint?: string;
};

type NavGroup = {
  groupLabelKey: string;
  groupFallbackLabel: string;
  children: NavItemConfig[];
};
```

### 1.2 用户端分组（替换原 `userNavItems`）

```
概览 (userGroup.overview)
  Gauge         Dashboard          /dashboard       g+d

资源 (userGroup.resources)
  KeyRound      API Keys           /api-keys
  Activity      用量                /usage           g+u
  PanelTop      模型价格            /models

服务 (userGroup.services)
  BookOpen      接入文档            /docs
  Bell          公告                /announcements
  Ticket        兑换码              /redeem

系统 (userGroup.system)
  Settings      设置                /settings
```

### 1.3 管理端分组（替换原 `adminNavItems`）

```
📊 运营 (adminGroup.ops)
  Gauge         仪表盘             /admin/dashboard
  Activity      Ops                /admin/ops
  FileText      报表               /admin/reports

👤 用户 (adminGroup.users)
  Users         用户管理           /admin/users
  KeyRound      API Keys           /admin/api-keys
  ShieldAlert   黑名单             /admin/blacklist

🔧 系统 (adminGroup.system)
  RadioTower    渠道管理           /admin/channels
  Waypoints     模型定价           /admin/models
  Bell          公告               /admin/announcements
  Ticket        兑换码             /admin/redeem
  ScrollText    审计日志           /admin/audit
  Settings      系统设置           /admin/settings
```

### 1.4 渲染改造

- 每个 `NavGroup` 渲染一个**分组标题**：`text-[10px] uppercase tracking-wider text-muted-foreground/60 px-3 pt-4 pb-1`，上方留有 padding 与上组隔开。
- 子项缩进渲染（`pl-3`），保持现有 active/hover 样式（`bg-gradient-to-r from-accent/10 to-transparent` + 左侧 2px 暖色竖条）。
- 移动端底部 tab 导航不分级，保持原样（使用原 `userNavItems` 一级列表）。
- 分组 key 走 i18n，新增翻译 key 到 `navigation:userGroup.*` 和 `navigation:adminGroup.*`，在 `en/navigation.ts` / `zh/navigation.ts` 补充。

### 1.5 快捷键标签

保持 hint 标签（`g+d`、`g+u`）在子项右侧展示。

---

# 第 2 项：删除旧端口 8081

## 当前问题

`frontend/nginx.conf` 有两个 `server` 块——`listen 80`（统一前端）和 `listen 8081`（注释写着"兼容旧管理端入口"）。双端口已不需要，且 8081 端口无 `/v1/` 反代（漏配），不安全。

## 改动

**直接删除**整个 `server { listen 8081; ... }` 块。从 `server {`（当前第 56 行附近）到对应的 `}`（文件末尾倒数第 2 行）全部删除。保留文件中的 `listen 80` 块不变。

---

# 第 3 项：弹窗风格重构（Claude Code 风格化）

## 当前问题

`packages/ui/src/components/dialog.tsx` 使用 Radix Dialog 默认样式，视觉简陋。Overlay 颜色太暗，弹窗缺乏 Claude Code 的暖质感，关闭按钮是原生样式。

## 改动（只改 `dialog.tsx` 一个文件，`confirm-dialog.tsx` 复用它的样式）

### 3.1 Overlay

```tsx
<DialogPrimitive.Overlay asChild forceMount>
  <motion.div
    className="fixed inset-0 z-50 bg-[rgba(20,20,19,0.25)] backdrop-blur-[1px]"
    initial={{ opacity: 0 }}
    animate={{ opacity: 1 }}
    exit={{ opacity: 0 }}
    transition={{ duration: 0.15 }}
  />
</DialogPrimitive.Overlay>
```

### 3.2 弹窗容器

```tsx
<DialogPrimitive.Content asChild forceMount>
  <motion.div
    className="fixed left-1/2 top-1/2 z-50 w-full max-w-lg -translate-x-1/2 -translate-y-1/2 rounded-xl bg-[#FAF9F5] p-0 shadow-[0_4px_24px_rgba(20,20,19,0.08),0_1px_4px_rgba(20,20,19,0.04)]"
    initial={{ opacity: 0, scale: 0.95, y: -4 }}
    animate={{ opacity: 1, scale: 1, y: 0 }}
    exit={{ opacity: 0, scale: 0.95, y: -4 }}
    transition={{ duration: 0.15, ease: [0.22, 1, 0.36, 1] }}
  >
    {/* 头部：标题 + 关闭按钮 */}
    <div className="flex items-center justify-between border-b border-border px-5 py-4">
      <DialogPrimitive.Title className="font-serif text-lg font-semibold text-foreground">
        {title}
      </DialogPrimitive.Title>
      <DialogPrimitive.Close asChild>
        <button className="flex h-7 w-7 items-center justify-center rounded-full text-muted-foreground hover:bg-accent/10 hover:text-foreground transition-colors">
          <X className="h-4 w-4" />
        </button>
      </DialogPrimitive.Close>
    </div>

    {/* 内容区 */}
    <div className="px-5 py-4 text-sm text-foreground">
      {children}
    </div>

    {/* Footer（可选） */}
    {footer && (
      <div className="flex flex-row-reverse items-center gap-2 border-t border-border px-5 py-3">
        {footer}
      </div>
    )}
  </motion.div>
</DialogPrimitive.Content>
```

### 3.3 import 补充

```ts
import { X } from "lucide-react";
import { motion } from "framer-motion";
```

**不需要用 `AnimatePresence`**——Radix Portal 的 `forceMount` + 组件的挂载/卸载由 Radix 控制，`initial/animate/exit` 配合 `AnimatePresence` 需要在调用方包裹。但 Radix Dialog 内部不自动提供 exit 动画，如果 exit 动画不起作用则使用 `onAnimationComplete` 做降级，或者确认 Dialog 的 open/close 确实触发 exit。

### 3.4 ConfirmDialog

`frontend/app/src/components/confirm-dialog.tsx` 确认视觉一致——不做样式覆盖，直接使用重构后的 Dialog。

---

# 第 4 项：Codex / OpenAI 快速配置补 /v1

## 当前问题

上一轮改端点路径时，`dashboard.tsx:124-129` 的 `codexSnippet()` 中的 `baseURL` 去掉了 `/v1`，但 Codex/OpenAI SDK 的 `OPENAI_BASE_URL` 需要带 `/v1`。

## 改动

改 `frontend/app/src/routes/dashboard.tsx`，`codexSnippet()`（约 124 行）：

当前：
```ts
const codexSnippet = (baseURL: string) => ({
  label: t("quickConfig.codex"),
  value: `OPENAI_BASE_URL=${baseURL}\nOPENAI_API_KEY=${sampleKey}`
});
```

改为：
```ts
const codexSnippet = (baseURL: string) => ({
  label: t("quickConfig.codex"),
  value: `OPENAI_BASE_URL=${baseURL}/v1\nOPENAI_API_KEY=${sampleKey}`
});
```

**不要改 claudeSnippet**——Anthropic SDK 自动处理 `/v1`，保持不带。

---

# 第 5 项：多语言切换器改图标式下拉菜单

## 当前问题

`AppShell.tsx:89-100` 是行内按钮组 `🇺🇸 EN | 🇨🇳 ZH`，文字感太强，视觉上像标签而不是切换器。

## 改动

### 5.1 替换当前语言切换区

删除：
```tsx
<div className="hidden items-center gap-1 rounded-md border border-border bg-card p-1 sm:flex" aria-label={t("common:language")}>
  {availableLocales.map((locale) => (
    <Button key={locale.code} size="sm" variant={i18n.resolvedLanguage === locale.code ? "secondary" : "ghost"} className="px-2" onClick={() => void i18n.changeLanguage(locale.code)}>
      <span aria-hidden>{locale.flag}</span>
      <span className="text-xs">{locale.code.toUpperCase()}</span>
    </Button>
  ))}
</div>
```

替换为：
```tsx
<Popover>
  <PopoverTrigger asChild>
    <Button variant="ghost" size="icon" aria-label={t("common:language")}>
      <Globe className="h-4 w-4" />
    </Button>
  </PopoverTrigger>
  <PopoverContent align="end" className="w-36 p-1">
    {availableLocales.map((locale) => (
      <button
        key={locale.code}
        onClick={() => void i18n.changeLanguage(locale.code)}
        className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm hover:bg-accent/10 transition-colors"
      >
        <span>{locale.flag}</span>
        <span className="flex-1 text-left">{locale.name}</span>
        {i18n.resolvedLanguage === locale.code && (
          <Check className="h-3.5 w-3.5 text-accent" />
        )}
      </button>
    ))}
  </PopoverContent>
</Popover>
```

### 5.2 新增 imports

```ts
import { Globe, Check } from "lucide-react";
import { Popover, PopoverContent, PopoverTrigger } from "@lingshu/ui";
```

### 5.3 确认 Popover 组件存在

如果 `packages/ui/src/components/popover.tsx` 不存在，基于 `@radix-ui/react-popover` 建一个：

```tsx
import React from "react";
import * as PopoverPrimitive from "@radix-ui/react-popover";
import { motion } from "framer-motion";
import { cn } from "../lib/cn";

export function Popover({ children, open, onOpenChange }: { children: React.ReactNode; open?: boolean; onOpenChange?: (open: boolean) => void }) {
  return <PopoverPrimitive.Root open={open} onOpenChange={onOpenChange}>{children}</PopoverPrimitive.Root>;
}

export const PopoverTrigger = PopoverPrimitive.Trigger;

export function PopoverContent({ children, align = "center", className }: { children: React.ReactNode; align?: "start" | "center" | "end"; className?: string }) {
  return (
    <PopoverPrimitive.Portal>
      <PopoverPrimitive.Content asChild align={align} sideOffset={4}>
        <motion.div
          className={cn("z-50 rounded-lg border border-border bg-card p-1 shadow-md", className)}
          initial={{ opacity: 0, scale: 0.96, y: -2 }}
          animate={{ opacity: 1, scale: 1, y: 0 }}
          exit={{ opacity: 0, scale: 0.96, y: -2 }}
          transition={{ duration: 0.12, ease: [0.22, 1, 0.36, 1] }}
        >
          {children}
        </motion.div>
      </PopoverPrimitive.Content>
    </PopoverPrimitive.Portal>
  );
}
```

### 5.4 清理 `availableLocales`

当前 `frontend/app/src/i18n/index.ts` 中 `availableLocales` 格式确认包含 `name` 字段（如 `{code:'en', name:'English', flag:'🇺🇸'}`）。如果只有 `code` 和 `flag`，加 `name` 字段。

---

# 第 6 项：对比 new-api / sub2api 补充功能

> 参考了 `D:\code\LingShu\new`（new-api）和 `D:\code\LingShu\Sub`（sub2api）。只做以下两项：(A) 渠道一键测试 + (B) 模型自动同步。分组体系另排一期。

## 6.1 (A) 渠道一键测试

### 后端：新增 Test endpoint

**文件**：`backend/internal/handler/admin/channels.go` 新增 handler

```go
func (h ChannelHandler) Test(w http.ResponseWriter, r *http.Request) {
    channelID := chi.URLParam(r, "id")
    
    var req struct {
        TestModel string `json:"test_model"`
    }
    if err := httpx.Decode(r, &req); err != nil {
        httpx.ErrorJSON(w, http.StatusBadRequest, "invalid_request", "invalid json", "invalid_json")
        return
    }
    
    result, err := h.channels.Test(r.Context(), channelID, req.TestModel)
    if err != nil {
        httpx.ErrorJSON(w, http.StatusBadGateway, "test_failed", err.Error(), "test_failed")
        return
    }
    
    httpx.JSON(w, http.StatusOK, result)
}
```

**文件**：`backend/internal/service/admin_channel_service.go` 新增 `Test` 方法

```go
type ChannelTestResult struct {
    Success   bool   `json:"success"`
    LatencyMS int64  `json:"latency_ms"`
    Message   string `json:"message,omitempty"`
    Model     string `json:"model,omitempty"`
}

func (s *ChannelService) Test(ctx context.Context, channelID, testModel string) (*ChannelTestResult, error) {
    // 1. 从 repository 获取渠道信息
    // 2. 根据渠道 provider_type 构造一个简单的 chat 请求
    // 3. 发起真实请求（如 OpenAI 兼容的 "Hello" 请求）
    // 4. 记录成功/失败 + 延迟
    // 5. 调用 MarkTest 写回 upstream_channels 的 last_success_at/health
    // 6. 返回结果
}
```

**路由注册**（`backend/internal/server/server.go`，在 channels 路由组内）：

```go
r.Post("/channels/{id}/test", adminChannels.Test)
```

### 前端：渠道详情加"测试"按钮

**文件**：`frontend/app/src/pages/admin/channels.tsx`

- 渠道列表每行操作区加一个 "测试" 按钮（`Button variant="outline" size="sm"`），点击后打开一个简单 Dialog 或 Sheet：
  - 选择测试模型（可选，如果没选则用渠道默认模型）
  - 点"开始测试"调 `api.testChannel(channelID, testModel)`
  - 结果展示：成功（绿色）+ 延迟 ms / 失败（红色）+ 错误信息
- 测试结果同时更新渠道行内的 `health` 状态（调后端 test 接口后返回的 health 已更新）。

**api-client**（`packages/shared/src/api-client.ts`）新增方法：

```ts
async testChannel(id: string, testModel?: string): Promise<{ success: boolean; latency_ms: number; message?: string }>
```

## 6.2 (B) 模型自动同步

### 后端：新增 SyncModels endpoint

**文件**：`backend/internal/handler/admin/channels.go` 新增 handler

```go
func (h ChannelHandler) SyncModels(w http.ResponseWriter, r *http.Request) {
    channelID := chi.URLParam(r, "id")
    
    result, err := h.channels.SyncModels(r.Context(), channelID)
    if err != nil {
        httpx.ErrorJSON(w, http.StatusBadGateway, "sync_failed", err.Error(), "sync_failed")
        return
    }
    
    httpx.JSON(w, http.StatusOK, result)
}

func (h ChannelHandler) ConfirmSyncModels(w http.ResponseWriter, r *http.Request) {
    var req struct {
        ChannelID  string   `json:"channel_id"`
        ModelIDs   []string `json:"model_ids"`
    }
    if err := httpx.Decode(r, &req); err != nil {
        httpx.ErrorJSON(w, http.StatusBadRequest, "invalid_request", "invalid json", "invalid_json")
        return
    }
    
    err := h.channels.ConfirmSyncModels(r.Context(), req.ChannelID, req.ModelIDs)
    if err != nil {
        httpx.ErrorJSON(w, http.StatusInternalServerError, "sync_confirm_failed", err.Error(), "sync_confirm_failed")
        return
    }
    
    httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
```

**文件**：`backend/internal/service/admin_channel_service.go` 新增方法

```go
type SyncModelsResult struct {
    Existing  int                     `json:"existing"`
    NewModels []SyncModelCandidate    `json:"new_models"`
    Removed   []string               `json:"removed"`
}

type SyncModelCandidate struct {
    ID       string `json:"id,omitempty"`
    Name     string `json:"name"`
    Type     string `json:"type"`
    Action   string `json:"action"` // "add" / "remove"
}

func (s *ChannelService) SyncModels(ctx context.Context, channelID string) (*SyncModelsResult, error) {
    // 1. 获取渠道信息（provider_type、api_key、base_url）
    // 2. 调该渠道的 /v1/models 或对应 endpoint 获取模型列表
    // 3. 对比已绑定的 channel_models，找出"新增"和"可能已停用"的
    // 4. 返回差异清单（不自动写入）
    // 5. 注意不同 provider 的模型列表接口不同：
    //    - OpenAI: GET /v1/models
    //    - Anthropic: GET /v1/models（Anthropic API 也支持）
    //    - 其他：若有就调，没有就返回"该渠道类型不支持自动同步"
    return result, nil
}

func (s *ChannelService) ConfirmSyncModels(ctx context.Context, channelID string, modelIDs []string) error {
    // 用户确认后，批量创建/绑定 model + channel_models 记录
}
```

**路由注册**：

```go
r.Post("/channels/{id}/sync-models", adminChannels.SyncModels)
r.Post("/channels/{id}/sync-models/confirm", adminChannels.ConfirmSyncModels)
```

### 前端：渠道详情加"同步模型"按钮

**文件**：`frontend/app/src/pages/admin/channels.tsx`

- 渠道详情（Sheet 或 Dialog）里，操作区加"同步模型"按钮。
- 点击后调 `api.syncChannelModels(channelID)`，返回差异清单：
  - 显示对比列表：已有的模型（灰色）/ 新增的模型（绿色，带复选框）/ 可能已停用的模型（红色）
  - 底部"确认同步"按钮，勾选需要新增的模型后调 `api.confirmSyncChannelModels(channelID, modelIDs)`
- 确认后刷新渠道绑定模型列表。

**api-client 新增方法**：

```ts
async syncChannelModels(id: string): Promise<{ existing: number; new_models: { id?: string; name: string; type: string; action: string }[]; removed: string[] }>
async confirmSyncChannelModels(id: string, modelIds: string[]): Promise<void>
```

---

# 第 7 项：全站 UI/风格重构（去廉价感）

> 目标：不做大重写，做**一次性润色**，让整个站点统一到 Claude Code 质感。

## 7.1 侧栏重构（AppShell.tsx）

| 项目 | 当前 | 改为 |
|---|---|---|
| 宽度 | `w-64` | `w-56` |
| 背景 | `bg-card` | `bg-surface`（暖米白），右边框 `border-r border-border` |
| 导航项 active | 浅背景色 | `bg-gradient-to-r from-accent/10 to-transparent` + 左侧 2px 竖条 `border-l-2 border-accent` |
| 导航项 hover | 无特效 | `bg-accent/5 transition-colors duration-100` |
| 导航文字 | 统一颜色 | active：`font-medium text-foreground`，非活跃：`text-muted-foreground` |
| 分组标题 | 无 | `text-[10px] uppercase tracking-wider text-muted-foreground/60 px-3 pt-4 pb-1` |
| 头像区 padding | 当前值 | 减少为 `px-3 py-3`，更紧凑 |
| 底部操作区 | 等大按钮 | 缩小图标尺寸（`h-4 w-4`），文字用 `text-xs` |

## 7.2 页面布局统一

- 内容区统一 `max-w-7xl mx-auto px-6 py-8`（后端所有 admin 页的容器类）。
- 每个页面顶部 `PageHeader`：eyebrow（`text-xs text-accent font-medium uppercase tracking-wider`）+ 衬线标题（`font-serif text-2xl font-semibold tracking-tight`）+ 描述（`text-sm text-muted-foreground mt-1`）。
- 功能区之间：`space-y-6`。
- 特别检查：`audit.tsx`、`ops.tsx`、`blacklist.tsx` 与其他 admin 页布局是否一致，不一致就改。

## 7.3 Card 组件增强（packages/ui/src/components/card.tsx）

当前 Card 平淡。改为：

```tsx
// Card 容器
<div className="rounded-lg border border-border bg-card shadow-sm hover:shadow-md transition-shadow duration-150">
  {children}
</div>

// CardHeader
<div className="border-b border-border px-5 py-3">
  {title && <h3 className="font-serif text-base font-semibold text-foreground">{title}</h3>}
  {description && <p className="mt-0.5 text-sm text-muted-foreground">{description}</p>}
</div>

// CardContent
<div className="p-5">
  {children}
</div>
```

**注意**：改后可能导致之前引用的 Card 组件间距错乱。检查所有 `packages/ui/src/index.ts` 导出 + 检查所有 `frontend/app/src/pages/admin/` 中使用 Card 的地方 padding 是否正常。

## 7.4 DataTable 增强（packages/ui/src/components/table.tsx）

| 项目 | 当前 | 改为 |
|---|---|---|
| 表头背景 | 无 | `bg-muted/30` |
| 表头文字 | 普通 | `text-xs font-medium text-muted-foreground uppercase tracking-wider` |
| 行高 | 默认 | `h-11` |
| 行 hover | 无 | `hover:bg-accent/3 transition-colors duration-100` |
| 分页 | 默认样式 | 居右，`text-sm text-muted-foreground` |
| 空态 | 纯文字 | `EmptyState` 组件（lucide `Inbox` 或 `FileSearch` 图标 + 标题 + 描述） |

### EmptyState 组件

如果 `packages/ui/src/components/empty-state.tsx` 不存在，新建一个简单组件：

```tsx
import { cn } from "../lib/cn";

export function EmptyState({ icon: Icon, title, description, className }: {
  icon?: React.ComponentType<{ className?: string }>;
  title: string;
  description?: string;
  className?: string;
}) {
  return (
    <div className={cn("flex flex-col items-center justify-center py-12 text-center", className)}>
      {Icon && <Icon className="mb-3 h-8 w-8 text-muted-foreground/40" />}
      <p className="text-sm font-medium text-foreground">{title}</p>
      {description && <p className="mt-1 text-xs text-muted-foreground">{description}</p>}
    </div>
  );
}
```

### Skeleton → 内容 exit fadeIn

之前确认了 DataTable 的三个条件分支（skeleton/table/empty）缺外层 `AnimatePresence`，导致切换时无 exit 动画。修复：

```tsx
import { AnimatePresence, motion } from "framer-motion";

// 在 return 处：
<AnimatePresence mode="wait">
  {loading ? (
    <motion.div key="loading" {...fadeProps}>
      {/* skeleton */}
    </motion.div>
  ) : items.length === 0 ? (
    <motion.div key="empty" {...fadeProps}>
      <EmptyState icon={Inbox} title="暂无数据" />
    </motion.div>
  ) : (
    <motion.div key="table" {...fadeProps}>
      <table>...</table>
    </motion.div>
  )}
</AnimatePresence>
```

## 7.5 StatCard 增强（packages/ui/src/components/feedback.tsx）

| 项目 | 当前 | 改为 |
|---|---|---|
| 网格 | 因页面而异 | 统一 `grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4` |
| 数字 | 默认 | `font-serif text-2xl font-semibold tracking-tight` |
| 标签 | 默认 | `text-xs text-muted-foreground` |
| 趋势指示 | 无 | 如果有趋势字段（百分比），显示 `text-xs text-green-600/destructive` + `TrendingUp`/`TrendingDown` 图标 |

## 7.6 表单控件统一

在 `packages/ui/src/components/` 中检查 Input、Select、Switch 的样式：

**Input**（`input.tsx`）：
```tsx
<input className="h-9 w-full rounded-md border border-border bg-surface px-3 text-sm text-foreground placeholder:text-muted-foreground/50 focus:outline-none focus:ring-1 focus:ring-accent focus:border-accent transition-colors" />
```

**Button**（`button.tsx`）确认 variant 配色：
- `primary`：`bg-accent text-white hover:bg-accent/90`
- `secondary`：`bg-surface text-foreground border border-border hover:bg-accent/5`
- `ghost`：`text-muted-foreground hover:text-foreground hover:bg-accent/5`
- `danger`：`bg-destructive text-white hover:bg-destructive/90`

**Switch**（`switch.tsx`，如果已用 Radix Switch）：活跃色用 `bg-accent`，非活跃 `bg-muted`。

## 7.7 品牌色点缀

在以下位置加入少量 accent 色（clay 暖色）：
- 侧栏活跃项左侧竖条 `border-l-2 border-accent`
- 页面标题的 eyebrow：`text-accent`
- Tag/Badge 的小圆点指示器
- 链接 hover 状态

**不要过度**。Claude Code 风格是克制的——大多数字体暖灰/墨色，accent 只出现在交互态。

## 7.8 动画一致性检查

确认所有覆盖层（Dialog/Sheet/Popover/Tooltip）都有入场/出场动画，参数一致：
- Dialog：opacity + scale（0.15s，`cubic-bezier(0.22, 1, 0.36, 1)`）
- Sheet：opacity + translateX（0.2s，同缓动）
- Popover：opacity + scale（0.12s，同缓动）
- Tooltip：opacity only（0.1s）

检查 **管理端用户详情 Sheet**（`pages/admin/users.tsx`）、**渠道编辑 Sheet**（`channels.tsx`）、**模型编辑 Dialog/Sheet**（`model-form.tsx`）是否都有动画。如果某个 Sheet 还没动画，给它加上。

---

# 验证清单

全部做完后逐条跑：

- [ ] 1. `go build ./... && go vet ./... && go test ./...` 全通过
- [ ] 2. `npm run build` 在 `frontend/` 下通过
- [ ] 3. 侧栏分组正确展示，子项缩进，分组标题可见，展开收起正常
- [ ] 4. `nginx.conf` 不再有 `listen 8081` 块
- [ ] 5. Dialog 弹窗暖白+暖阴影，关闭按钮是 lucide X，入场动画有 scale
- [ ] 6. Dashboard 快速配置 Codex Tab 的 `OPENAI_BASE_URL` 带 `/v1`
- [ ] 7. 语言切换器是 Globe 图标 + 下拉菜单，不再是行内 EN/ZH 按钮
- [ ] 8. 渠道列表有"测试"按钮，点击后显示测试结果（成功/失败+延迟）
- [ ] 9. 渠道详情有"同步模型"按钮，显示差异清单，确认后写入
- [ ] 10. 所有 admin 页风格一致（PageHeader + Card + DataTable 范式）
- [ ] 11. 所有覆盖层（Dialog/Sheet/Popover/Tooltip）有入场动画
- [ ] 12. DataTable skeleton → table 切换有 exit fadeOut
- [ ] 13. 整体观感不再是"廉价感"——暖色调，一致间距，舒适留白

---

做完后列出清单：改了哪些文件（按项分类）、每项验证结果、无法通过的项及原因。
