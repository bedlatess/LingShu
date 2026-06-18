# LingShu v1.5.1 上线热修复

> 本文档为 **agent 可直接执行** 的精确修复说明。每一项都给出文件路径、行号定位、改前/改后代码。请严格按文件逐项执行，不要做文档外的额外改动（不新增营销类或其他无关功能）。

修复范围（4 项，按优先级）：

- **P0-A 管理端 `列"deleted_at"不存在 (SQLSTATE 42703)`**：后端 dashboard SQL 对 `users` 表误加了软删过滤，但 `users` 表无 `deleted_at` 列。
- **P0-B 用户端删除 API Key 报错文案丑陋**：删除链逻辑正确，但删 0 行时把 `pgx.ErrNoRows` 原样以 `no rows in result set` 当 400 抛出。
- **P1-A 用户端设置页缺少"修改密码"功能**：后端接口已就绪，仅需前端 UI + api-client 方法。
- **P1-B 用户端模型列表按分组展示**：`group` 字段已贯穿全链路，纯前端按 group 聚合渲染即可。

---

## P0-A：修复 dashboard 的 deleted_at 42703

**根因**：`backend/migrations/0003_soft_delete.up.sql` 只给 `upstream_channels`/`models`/`api_keys` 加了 `deleted_at`，`users` 表从未有此列，且全代码库无用户软删逻辑。v1.5 dashboard SQL 对 `users` 误加 `deleted_at IS NULL`。

**修复方式**：去掉 `users` 三处的 `deleted_at IS NULL` 过滤（保留 channels/models/api_keys 的，因为这三张表确实有该列）。

**文件**：`backend/internal/repository/reports.go`，函数 `AdminDashboard`（约 258 行起）。

将第 266-268 行三处 `users` 子查询中的 `AND deleted_at IS NULL` / `WHERE deleted_at IS NULL` 删除。

改前：
```go
		  COALESCE((SELECT count(*) FROM users WHERE status='active' AND deleted_at IS NULL),0)::int,
		  COALESCE((SELECT sum(balance) FROM users WHERE deleted_at IS NULL),0)::text,
		  COALESCE((SELECT count(*) FROM users WHERE deleted_at IS NULL),0)::int,
```

改后：
```go
		  COALESCE((SELECT count(*) FROM users WHERE status='active'),0)::int,
		  COALESCE((SELECT sum(balance) FROM users),0)::text,
		  COALESCE((SELECT count(*) FROM users),0)::int,
```

> 注意：仅改这 3 行。第 269-275 行对 `upstream_channels`/`models`/`api_keys` 的 `deleted_at IS NULL` **保持不变**。

**验证**：`cd backend && go build ./... && go vet ./...`。上线后访问管理端首页 dashboard 不再报 42703。

---

## P0-B：用户端删除 API Key 返回友好错误

**根因**：删 0 行（key 已不存在/已被删）时 repo 返回 `pgx.ErrNoRows`，handler 直接 `httpx.Error(w, 400, err.Error())`，前端 toast 显示 `删除失败：no rows in result set`。逻辑本身正确，只是文案差。

**文件**：`backend/internal/handler/user/user.go`

**第一步**：在 import 块加入 `errors` 与 `pgx`。

改前（第 3-14 行）：
```go
import (
	"net"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"lingshu/backend/internal/dto"
	"lingshu/backend/internal/middleware"
	"lingshu/backend/internal/pkg/httpx"
	"lingshu/backend/internal/service"
)
```

改后：
```go
import (
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"lingshu/backend/internal/dto"
	"lingshu/backend/internal/middleware"
	"lingshu/backend/internal/pkg/httpx"
	"lingshu/backend/internal/service"
)
```

> 若该文件 import 中 pgx 路径与项目其它文件不同，请以仓库现有 `github.com/jackc/pgx/v5` 版本为准（与 `backend/internal/repository/api_keys.go` 顶部一致）。

**第二步**：`DeleteAPIKey`（约 137-148 行）对 `ErrNoRows` 返回 404 友好文案。

改前：
```go
	if err := h.keys.DeleteForUser(r.Context(), current.ID, chi.URLParam(r, "id")); err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
```

改后：
```go
	if err := h.keys.DeleteForUser(r.Context(), current.ID, chi.URLParam(r, "id")); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.Error(w, http.StatusNotFound, "密钥不存在或已被删除")
			return
		}
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
```

**验证**：`cd backend && go build ./...`。用户端重复删除同一 key 时提示"密钥不存在或已被删除"，而非 `no rows in result set`。

---

## P1-A：用户端设置页新增"修改密码"

**后端契约（已存在，无需改后端）**：
- `POST /api/auth/change-password`（需 JWT）
- 请求体：`{ "old_password": "...", "new_password": "..." }`
- 成功：`200 {"status":"ok"}`；失败：`4xx {"error":"<msg>"}`（旧密码错误时为 `invalid credentials`）

### 步骤 1：api-client 增加方法

**文件**：`frontend/packages/shared/src/api-client.ts`

在 `login` 方法之后（约第 86 行附近，紧邻其它 auth 方法处）新增 `changePassword`。找到现有 `login:` 定义，在其后插入：

```ts
    changePassword: (payload: { old_password: string; new_password: string }) =>
      request<{ status: string }>("/api/auth/change-password", {
        method: "POST",
        body: JSON.stringify(payload)
      }),
```

> 该方法须放在 `createAPI` 返回的对象内，与 `login`/`redeem` 等同级。

### 步骤 2：设置页加入修改密码表单

**文件**：`frontend/user/src/routes/settings.tsx`（整文件替换为以下内容）

```tsx
import React from "react";
import { KeyRound, ShieldCheck, UserRound } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { PageHeader } from "@/components/page-header";
import { useAuth } from "@/providers/auth";
import { formatMoney } from "@/lib/utils";
import { zhStatus } from "@/lib/i18n";

export function SettingsPage() {
  const { user, api } = useAuth();
  const [oldPassword, setOldPassword] = React.useState("");
  const [newPassword, setNewPassword] = React.useState("");
  const [confirmPassword, setConfirmPassword] = React.useState("");
  const [submitting, setSubmitting] = React.useState(false);

  async function changePassword(event: React.FormEvent) {
    event.preventDefault();
    if (newPassword.length < 6) {
      toast.error("新密码至少 6 位");
      return;
    }
    if (newPassword !== confirmPassword) {
      toast.error("两次输入的新密码不一致");
      return;
    }
    setSubmitting(true);
    try {
      await api.changePassword({ old_password: oldPassword, new_password: newPassword });
      toast.success("密码已修改");
      setOldPassword("");
      setNewPassword("");
      setConfirmPassword("");
    } catch (err) {
      const message = err instanceof Error ? err.message : "修改失败";
      toast.error(`修改失败：${message === "invalid credentials" ? "原密码错误" : message}`);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="page-grid">
      <PageHeader eyebrow="账户设置" title="账户设置" description="查看当前账户信息并管理登录密码。" />
      <div className="grid gap-5 lg:grid-cols-2">
        <Card className="glass">
          <CardHeader><CardTitle className="flex items-center gap-2"><UserRound className="h-4 w-4 text-primary" />账户</CardTitle></CardHeader>
          <CardContent className="grid gap-3 text-sm">
            <Row label="用户名" value={user?.username ?? "-"} />
            <Row label="邮箱" value={user?.email || "-"} />
            <Row label="角色" value={user?.role ?? "-"} />
            <Row label="状态" value={zhStatus(user?.status)} />
          </CardContent>
        </Card>
        <Card className="glass">
          <CardHeader><CardTitle className="flex items-center gap-2"><ShieldCheck className="h-4 w-4 text-primary" />余额安全</CardTitle></CardHeader>
          <CardContent className="grid gap-3 text-sm">
            <Row label="当前余额" value={formatMoney(user?.balance)} />
            <Row label="安全提示" value="请妥善保管 API 密钥" />
          </CardContent>
        </Card>
      </div>
      <Card className="glass">
        <CardHeader><CardTitle className="flex items-center gap-2"><KeyRound className="h-4 w-4 text-primary" />修改密码</CardTitle></CardHeader>
        <CardContent>
          <form className="grid max-w-md gap-4" onSubmit={changePassword}>
            <div className="grid gap-1.5">
              <label className="text-sm text-muted-foreground">原密码</label>
              <Input type="password" value={oldPassword} onChange={(event) => setOldPassword(event.target.value)} placeholder="请输入原密码" required />
            </div>
            <div className="grid gap-1.5">
              <label className="text-sm text-muted-foreground">新密码</label>
              <Input type="password" value={newPassword} onChange={(event) => setNewPassword(event.target.value)} placeholder="至少 6 位" required />
            </div>
            <div className="grid gap-1.5">
              <label className="text-sm text-muted-foreground">确认新密码</label>
              <Input type="password" value={confirmPassword} onChange={(event) => setConfirmPassword(event.target.value)} placeholder="再次输入新密码" required />
            </div>
            <Button type="submit" disabled={submitting}>{submitting ? "提交中…" : "确认修改"}</Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between rounded-lg border border-white/10 bg-white/[0.035] px-3 py-2">
      <span className="text-muted-foreground">{label}</span>
      <strong>{value}</strong>
    </div>
  );
}
```

**验证**：`cd frontend/user && npm run build`（workspace 单独构建，勿并行）。设置页出现"修改密码"卡片；原密码错误时提示"原密码错误"，成功提示"密码已修改"。

---

## P1-B：用户端模型列表按分组展示

**说明**：`group` 字段已从 DB(`model_group`)→Go(`Group json:"group"`)→DTO→前端类型(`UserModelConfig.group`) 全链路就绪，**仅改前端**，按 `group` 聚合分段渲染，并加一个搜索框便于 30+ 模型检索。无分页（分组+搜索已足够；不要引入额外依赖）。

**文件**：`frontend/user/src/routes/models.tsx`（整文件替换为以下内容）

```tsx
import React from "react";
import { Boxes, Image, MessageSquareText, Search, Zap } from "lucide-react";
import type { UserModelConfig } from "@lingshu/shared/user-types";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { PageHeader } from "@/components/page-header";
import { EmptyState } from "@/components/empty-state";
import { useAuth } from "@/providers/auth";
import { zhBillingMode, zhStatus, zhType } from "@/lib/i18n";

export function ModelsPage() {
  const { api } = useAuth();
  const [models, setModels] = React.useState<UserModelConfig[]>([]);
  const [keyword, setKeyword] = React.useState("");

  React.useEffect(() => {
    api.userModels().then((result) => setModels(result.items));
  }, [api]);

  const filtered = React.useMemo(() => {
    const kw = keyword.trim().toLowerCase();
    if (!kw) return models;
    return models.filter((model) =>
      [model.public_name, model.group, zhType(model.type)]
        .filter(Boolean)
        .some((field) => field.toLowerCase().includes(kw))
    );
  }, [models, keyword]);

  const groups = React.useMemo(() => {
    const map = new Map<string, UserModelConfig[]>();
    for (const model of filtered) {
      const name = model.group?.trim() || "默认分组";
      const list = map.get(name) ?? [];
      list.push(model);
      map.set(name, list);
    }
    return Array.from(map.entries()).sort((a, b) => a[0].localeCompare(b[0], "zh"));
  }, [filtered]);

  return (
    <div className="page-grid">
      <PageHeader eyebrow="可用模型" title="模型列表" description="以下模型已为你的账户开放，可直接通过平台密钥调用。" />
      <div className="relative max-w-sm">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input className="pl-9" value={keyword} onChange={(event) => setKeyword(event.target.value)} placeholder="搜索模型名称或分组" />
      </div>
      {filtered.length === 0 ? (
        <EmptyState title="暂无可用模型" description={keyword ? "没有匹配的模型，试试其他关键字。" : "管理员启用模型并绑定渠道后，这里会展示可用模型。"} />
      ) : (
        <div className="grid gap-8">
          {groups.map(([groupName, items]) => (
            <section key={groupName} className="grid gap-4">
              <div className="flex items-center gap-3">
                <h2 className="text-base font-semibold">{groupName}</h2>
                <Badge variant="secondary">{items.length}</Badge>
              </div>
              <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
                {items.map((model) => (
                  <Card key={model.id} className="glass overflow-hidden transition-all hover:-translate-y-1 hover:border-primary/35">
                    <CardContent className="p-5">
                      <div className="mb-5 flex items-start justify-between gap-3">
                        <div className="grid h-11 w-11 place-items-center rounded-lg bg-primary/10 text-primary">{iconFor(model.type)}</div>
                        <Badge>{zhBillingMode(model.billing_mode)}</Badge>
                      </div>
                      <h3 className="text-lg font-semibold">{model.public_name}</h3>
                      <p className="mt-1 text-sm text-muted-foreground">{zhType(model.type)}</p>
                      <div className="mt-5 grid gap-2 rounded-lg border border-white/10 bg-white/[0.035] p-4 text-sm">
                        <Meta label="计费方式" value={zhBillingMode(model.billing_mode)} />
                        <Meta label="状态" value={zhStatus(model.status)} />
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            </section>
          ))}
        </div>
      )}
    </div>
  );
}

function iconFor(type: string) {
  if (type === "image") return <Image className="h-5 w-5" />;
  if (type === "embedding") return <Boxes className="h-5 w-5" />;
  if (type === "video") return <Zap className="h-5 w-5" />;
  return <MessageSquareText className="h-5 w-5" />;
}

function Meta({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-muted-foreground">{label}</span>
      <strong>{value}</strong>
    </div>
  );
}
```

> 若 `Badge` 组件不支持 `variant="secondary"`，请改为去掉该 prop 用默认 `<Badge>`。请先看 `frontend/user/src/components/ui/badge.tsx` 的 variant 定义再决定。

**验证**：`cd frontend/user && npm run build`。模型按分组分段展示，每段标题带数量徽标；搜索框可按名称/分组过滤。

---

## 总验证清单（agent 执行完后逐项跑）

1. 后端：`cd backend && go build ./... && go vet ./... && go test ./...`（全绿）
2. 用户端：`cd frontend/user && npm run build`（单独构建，勿与 admin 并行）
3. 共享包改动后用户端会重新拾取 `api-client.ts`；若类型未更新，先 `cd frontend/packages/shared && npm run build` 再构建 user。
4. 冒烟：
   - 管理端首页 dashboard 正常加载，无 42703。
   - 用户端删除已不存在的 key，提示"密钥不存在或已被删除"。
   - 用户端设置页修改密码：原密码错→"原密码错误"；正确→"密码已修改"。
   - 用户端模型页按分组分段 + 搜索可用。

## 注意事项

- 不要改 `backend/migrations/`（本次无需新建迁移；users 表本就不需要 deleted_at）。
- 不要触碰计费链路、SSE usage 解析、上游错误透传等money path。
- 不要新增营销类或其它本文档之外的功能。
