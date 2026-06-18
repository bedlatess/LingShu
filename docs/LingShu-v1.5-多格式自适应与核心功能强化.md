# 灵枢 LingShu v1.5 - 多格式自适应 + 核心功能强化 + 精致化

> 整理人：Kiro
> 日期：2026-06-18
> 上一版：v1.4 安全硬化已上线并验收通过（CORS 白名单 / Body 限制 / 渠道自愈 / 错误体统一），go build+vet+test 全绿
> 本版主题：**让接入"零配置自适应"** + 补齐核心接口 + 把管理端/用户端从"能用"做到"精致"

---

## 零、本版铁律与边界（必须先读）

**永不改动的铁律**：
- 计费公式 `charge = base_cost × rate_multiplier`，平台赚倍率差价。
- 用户端禁露 `base_cost / rate_multiplier / gross_profit`。
- 软删除模式（`deleted_at IS NULL`）。
- 流式计费优先用上游 SSE 末帧回灌的真实 usage（`ExtractStreamUsage`），取不到才回退 tiktoken 估算。**本版任何改动都不得破坏这条money路径。**

**明确不做**（用户已明令否决"营销/杂项"，agent 不要自由扩展）：
- ❌ 营销门面 / 公开价格表页 / 站点导航（v1.4 S2 已被否，本版不再强化，但已合入的 `/api/public/*` 代码保持不动，不删不扩）
- ❌ 注册送余额 / 签到 / 邀请码等增长功能
- ❌ Prometheus 大盘 / 管理员 2FA / 模型分组令牌分组
- ❌ 图片/TTS/音频转发（本版只做 embeddings，其余推后）

**本版范围总览**：

| 模块 | 内容 | 优先级 |
|---|---|---|
| **A 多格式自适应** | 供应商预设一键填充 + 协议自动探测 + 修复渠道连通性测试 + base_url 容错 | 🔴 P0（用户最痛点） |
| **B 核心接口扩展** | `/v1/embeddings` 透传与计费 | 🟡 P1 |
| **C 七处 UI/UX 修复** | 模型保存按钮 / 兑换码详情 / 审计清理 / 概览精致化 / 报表精致化 / key 复制 / 兑换页排版 | 🔴🟡 混合 |
| **D 附带修复** | TS 类型补全、剪贴板兜底封装 | 🟢 随手做 |

**执行顺序严格按 A → B → C → D**，每节内部步骤一气呵成。

---

## A. 多格式自适应（P0｜用户最痛点：「最主要能自适应格式而不是每次都要配置」）

### A0 现状与设计说明（agent 先理解再动手）

当前协议判定只有一处字符串 switch（`backend/internal/upstream/provider.go`）：
```go
func ProviderForType(providerType string) Provider {
    switch strings.ToLower(strings.TrimSpace(providerType)) {
    case "anthropic", "claude":
        return AnthropicAdapter{}
    default:
        return OpenAIAdapter{}   // openai/deepseek/zhipu/未知 全部走这里
    }
}
```

**关键事实（不要误解）**：
- DeepSeek、智谱 GLM、Kimi、通义、SiliconFlow、xAI、OpenRouter 等**本来就是 OpenAI 格式**，用 `provider_type=openai` + 正确 base_url 就能直接转发，**无需新增协议**。
- 真正只有两种"协议"：`openai` 与 `anthropic`。区别在代码里：

| 维度 | OpenAI 适配器 | Anthropic 适配器 |
|---|---|---|
| 聊天路径 | `<base>/chat/completions` | `<base>/v1/messages`（`anthropicURL` 缺 /v1 时自动补） |
| 模型列表 | `<base>/models` | `<base>/v1/models` |
| 鉴权头 | `Authorization: Bearer <key>` | `x-api-key: <key>` + `anthropic-version: 2023-06-01` |
| base_url 约定 | **必须含 /v1**（OpenAI）或 /v4（智谱），适配器不补 | 不含 /v1 也行（自动补） |

- **用户的"麻烦"本质**：每加一个渠道都要手填 provider_type、记住要不要加 /v1、记住每家 base_url。
- **本节解决方案三件套**：
  1. **A1 供应商预设**：下拉选「DeepSeek/智谱/Kimi/...」→ 自动填好 base_url + 协议，粘贴 key 即可。
  2. **A2 协议自动探测**：点「检测并填充」或「测试」时，用 key 真实探测上游，自动判出 openai/anthropic 并修正 base_url。覆盖"自定义未知上游"。
  3. **A3 base_url 容错**：缺 /v1 自动补一刀重试，降低手填错误。

> 决策：**不新增 provider_type 枚举值**（不动 DB CHECK 约束、不动 `validateChannel`）。新增供应商只是"预设"层面的便利，底层仍归一到 openai/anthropic 两种协议。这样既满足"支持更多家"，又不增加协议维护成本。

---

### A1【后端+前端】供应商预设一键填充

**目的**：把"每次手填 base_url + 选协议"变成"选一个供应商，自动填好"。

**A1.1 后端预设注册表** — 新建 `backend/internal/upstream/presets.go`：
```go
package upstream

// ChannelPreset 描述一个已知供应商的接入预设。
// Format 只能是 "openai" 或 "anthropic"（底层协议），其余家都是 openai 兼容。
type ChannelPreset struct {
    Key      string `json:"key"`
    Label    string `json:"label"`
    BaseURL  string `json:"base_url"`
    Format   string `json:"format"`   // openai | anthropic
    Note     string `json:"note"`
}

func ChannelPresets() []ChannelPreset {
    return []ChannelPreset{
        {Key: "openai", Label: "OpenAI 官方", BaseURL: "https://api.openai.com/v1", Format: "openai", Note: "GPT 系列"},
        {Key: "deepseek", Label: "DeepSeek 深度求索", BaseURL: "https://api.deepseek.com/v1", Format: "openai", Note: "deepseek-chat / deepseek-reasoner"},
        {Key: "zhipu", Label: "智谱 GLM", BaseURL: "https://open.bigmodel.cn/api/paas/v4", Format: "openai", Note: "glm-4 系列，注意是 /v4"},
        {Key: "moonshot", Label: "月之暗面 Kimi", BaseURL: "https://api.moonshot.cn/v1", Format: "openai", Note: "moonshot-v1 系列"},
        {Key: "qwen", Label: "阿里通义千问", BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1", Format: "openai", Note: "qwen 系列（兼容模式）"},
        {Key: "siliconflow", Label: "硅基流动 SiliconFlow", BaseURL: "https://api.siliconflow.cn/v1", Format: "openai", Note: "多模型聚合"},
        {Key: "xai", Label: "xAI Grok", BaseURL: "https://api.x.ai/v1", Format: "openai", Note: "grok 系列"},
        {Key: "openrouter", Label: "OpenRouter", BaseURL: "https://openrouter.ai/api/v1", Format: "openai", Note: "多供应商路由"},
        {Key: "anthropic", Label: "Anthropic Claude", BaseURL: "https://api.anthropic.com", Format: "anthropic", Note: "claude 系列，不要加 /v1"},
        {Key: "custom", Label: "自定义上游", BaseURL: "", Format: "openai", Note: "中转站/自建，填好地址后点检测自适应"},
    }
}
```

**A1.2 后端端点** — 在 `backend/internal/handler/admin/channels.go` 的 `ChannelHandler` 加方法：
```go
func (h ChannelHandler) Presets(w http.ResponseWriter, r *http.Request) {
    httpx.JSON(w, http.StatusOK, map[string]any{"items": upstream.ChannelPresets()})
}
```
（文件顶部 import 加 `"lingshu/backend/internal/upstream"`。）

在 `backend/internal/server/server.go` 的 `/api/admin` 路由组内加一行（紧挨现有 channels 路由）：
```go
r.Get("/channel-presets", adminChannels.Presets)
```

**A1.3 前端 API 客户端** — `frontend/packages/shared/src/api-client.ts` 加方法（admin 区）：
```ts
listChannelPresets: () =>
  request<{ items: ChannelPreset[] }>("/api/admin/channel-presets"),
```
并在 `frontend/packages/shared/src/types.ts` 加类型并从 `admin-types.ts` 导出：
```ts
export interface ChannelPreset {
  key: string;
  label: string;
  base_url: string;
  format: string;
  note: string;
}
```

**A1.4 前端创建渠道表单接入预设** — `frontend/admin/src/pages/channels.tsx` 的 `ChannelsPage` 创建表单（约 66-101 行）：
- 组件内 `useState` 拉取预设：`const [presets, setPresets] = useState<ChannelPreset[]>([]);`，`useEffect` 调 `api.listChannelPresets().then((r) => setPresets(r.items))`。
- 在 `<Form.Item name="name">` 之前加一个"供应商预设"下拉：
  ```tsx
  <Form.Item label="供应商预设">
    <Select
      style={{ width: 220 }}
      placeholder="选择后自动填充"
      options={presets.map((p) => ({ value: p.key, label: p.label }))}
      onChange={(key) => {
        const p = presets.find((item) => item.key === key);
        if (!p) return;
        form.setFieldsValue({ base_url: p.base_url, provider_type: p.format });
        if (p.note) message.info(p.note);
      }}
    />
  </Form.Item>
  ```
- 这样选「DeepSeek」就自动把 base_url 填成 `https://api.deepseek.com/v1`、provider_type 设为 `openai`，用户只剩粘贴 key。

**验收 A1**：进入渠道管理 → 选「智谱 GLM」→ base_url 自动变 `https://open.bigmodel.cn/api/paas/v4`、供应商变 OpenAI 兼容 → 填 key + 名称即可创建。

---

### A2【后端+前端】协议自动探测 + 修复渠道连通性测试

**现状两个问题**：
1. 没有任何"探测上游真实协议"的能力，自定义中转站只能用户自己猜 openai/anthropic。
2. `backend/internal/service/channel_service.go` 的 `Test()` 是**写死的** unauth `GET <base>/models`，**完全忽略 provider_type、不带鉴权**——Anthropic 渠道测试必假阳/假阴，OpenAI 需要鉴权的上游也测不准。

> ⚠️ 关键坑（必须避开）：`AnthropicAdapter.ListModels` 在任何失败时会**回退到硬编码的 `anthropicPresetModels()`**（恒返回 8 个 claude 模型）。所以**探测/测试绝不能直接复用 `AnthropicAdapter.ListModels`**，否则 anthropic 永远"成功"。探测必须用下面的**严格探针**（非 2xx 直接返回错误，不回退预设）。

**A2.1 严格探针** — 新建 `backend/internal/upstream/probe.go`：
```go
package upstream

import (
    "context"
    "encoding/json"
    "errors"
    "io"
    "net/http"
    "strings"
    "time"
)

// ProbeOpenAI 严格探测 OpenAI 兼容上游：GET <base>/models + Bearer。
// 返回模型 id 样本、HTTP 状态码、错误。非 2xx 或非 JSON 视为失败（不回退预设）。
func ProbeOpenAI(ctx context.Context, baseURL, apiKey string) ([]string, int, error) {
    url := strings.TrimRight(strings.TrimSpace(baseURL), "/") + "/models"
    return probeModels(ctx, http.MethodGet, url, func(req *http.Request) {
        req.Header.Set("Authorization", "Bearer "+apiKey)
    })
}

// ProbeAnthropic 严格探测 Anthropic 上游：GET <base>/v1/models + x-api-key（不回退预设）。
func ProbeAnthropic(ctx context.Context, baseURL, apiKey string) ([]string, int, error) {
    return probeModels(ctx, http.MethodGet, anthropicURL(baseURL, "/models"), func(req *http.Request) {
        req.Header.Set("x-api-key", apiKey)
        req.Header.Set("anthropic-version", "2023-06-01")
    })
}

func probeModels(ctx context.Context, method, url string, auth func(*http.Request)) ([]string, int, error) {
    client := &http.Client{Timeout: 15 * time.Second}
    req, err := http.NewRequestWithContext(ctx, method, url, nil)
    if err != nil {
        return nil, 0, err
    }
    auth(req)
    resp, err := client.Do(req)
    if err != nil {
        return nil, 0, err
    }
    defer resp.Body.Close()
    body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return nil, resp.StatusCode, errors.New(strings.TrimSpace(string(body)))
    }
    var parsed struct {
        Data []struct {
            ID string `json:"id"`
        } `json:"data"`
    }
    if err := json.Unmarshal(body, &parsed); err != nil {
        return nil, resp.StatusCode, errors.New("上游返回非 JSON，疑似 base_url 错误")
    }
    ids := make([]string, 0, len(parsed.Data))
    for _, item := range parsed.Data {
        if item.ID != "" {
            ids = append(ids, item.ID)
        }
    }
    return ids, resp.StatusCode, nil
}
```

**A2.2 协议探测编排** — 同文件追加：
```go
type DetectResult struct {
    Format         string   `json:"format"`              // openai | anthropic
    NormalizedBase string   `json:"normalized_base_url"` // 探测成功用的 base_url（可能补了 /v1）
    SampleModels   []string `json:"sample_models"`
}

// DetectProtocol 用 key 真实探测上游格式。顺序：
// 1) host 含 anthropic/claude → 先严格探 anthropic
// 2) 严格探 openai（GET <base>/models）
// 3) base_url 不含 /v 时补 /v1 再探 openai
// 4) 严格探 anthropic
func DetectProtocol(ctx context.Context, baseURL, apiKey string) (DetectResult, error) {
    trimmed := strings.TrimRight(strings.TrimSpace(baseURL), "/")
    lower := strings.ToLower(trimmed)

    if strings.Contains(lower, "anthropic") || strings.Contains(lower, "claude") {
        if models, _, err := ProbeAnthropic(ctx, trimmed, apiKey); err == nil {
            return DetectResult{Format: "anthropic", NormalizedBase: trimmed, SampleModels: models}, nil
        }
    }
    if models, _, err := ProbeOpenAI(ctx, trimmed, apiKey); err == nil {
        return DetectResult{Format: "openai", NormalizedBase: trimmed, SampleModels: models}, nil
    }
    if !strings.Contains(lower, "/v") {
        withV1 := trimmed + "/v1"
        if models, _, err := ProbeOpenAI(ctx, withV1, apiKey); err == nil {
            return DetectResult{Format: "openai", NormalizedBase: withV1, SampleModels: models}, nil
        }
    }
    if models, _, err := ProbeAnthropic(ctx, trimmed, apiKey); err == nil {
        return DetectResult{Format: "anthropic", NormalizedBase: trimmed, SampleModels: models}, nil
    }
    return DetectResult{}, errors.New("无法识别上游格式：openai /models 与 anthropic /v1/models 均探测失败，请检查 base_url 与密钥")
}
```

**A2.3 探测端点** — `handler/admin/channels.go` 加方法 + `server.go` 加路由：
```go
// handler/admin/channels.go
func (h ChannelHandler) Detect(w http.ResponseWriter, r *http.Request) {
    var input struct {
        BaseURL string `json:"base_url"`
        APIKey  string `json:"api_key"`
    }
    if err := httpx.Decode(r, &input); err != nil {
        httpx.Error(w, http.StatusBadRequest, "invalid json")
        return
    }
    result, err := upstream.DetectProtocol(r.Context(), input.BaseURL, input.APIKey)
    if err != nil {
        httpx.Error(w, http.StatusBadGateway, err.Error())
        return
    }
    httpx.JSON(w, http.StatusOK, result)
}
```
```go
// server.go /api/admin 组内
r.Post("/channels/detect", adminChannels.Detect)
```

**A2.4 修复 `ChannelService.Test` 走协议感知探针** — `backend/internal/service/channel_service.go` 的 `Test()`：
- 改为先 `FindSecretByID` 拿到该渠道的 `ProviderType` + 解密后的 key + 默认 base_url。
- 入参 baseURL 为空则用渠道自身 base_url。
- 按协议选探针：
  ```go
  var statusCode int
  var err error
  if strings.EqualFold(channelSecret.ProviderType, "anthropic") || strings.EqualFold(channelSecret.ProviderType, "claude") {
      _, statusCode, err = upstream.ProbeAnthropic(ctx, baseURL, key)
  } else {
      var models []string
      models, statusCode, err = upstream.ProbeOpenAI(ctx, baseURL, key)
      // openai 缺 /v1 容错重试一次
      if err != nil && statusCode == 0 && !strings.Contains(strings.ToLower(baseURL), "/v") {
          models, statusCode, err = upstream.ProbeOpenAI(ctx, strings.TrimRight(baseURL, "/")+"/v1", key)
      }
      _ = models
  }
  ```
- 用 `statusCode` 走现有 `categorizeChannelTest()`，`err==nil && statusCode 2xx` 才算 ok，再调 `MarkTest`（保留 v1.2 F2 的 `<300` 判定与延迟记录逻辑）。
- **保留**现有返回结构（ok/category/message/latency_ms），前端无需大改。

> 注意：`Test` 现在需要解密 key，确认 `channel_service.go` 里已能拿到 `secret.Unprotect(...)`（`SyncModels` 已这么用，照搬即可）。

**A2.5 前端"检测并填充"按钮** — `frontend/admin/src/pages/channels.tsx` 创建表单内，base_url + api_key 之后加一个按钮：
```tsx
<Form.Item label=" ">
  <Button onClick={async () => {
    const base = form.getFieldValue("base_url");
    const key = form.getFieldValue("api_key");
    if (!base || !key) { message.warning("请先填 base_url 和上游 key"); return; }
    try {
      const r = await api.detectChannel(base, key);
      form.setFieldsValue({ provider_type: r.format, base_url: r.normalized_base_url });
      message.success(`已识别为 ${r.format}，样例模型 ${r.sample_models.slice(0, 3).join(", ") || "无"}`);
    } catch (err) { message.error(`检测失败: ${errText(err)}`); }
  }}>检测并填充协议</Button>
</Form.Item>
```
`api-client.ts` 加：
```ts
detectChannel: (base_url: string, api_key: string) =>
  request<{ format: string; normalized_base_url: string; sample_models: string[] }>(
    "/api/admin/channels/detect",
    { method: "POST", body: JSON.stringify({ base_url, api_key }) }
  ),
```

**验收 A2**：
- 填一个自定义中转站 base_url + key，点「检测并填充」→ 自动判出 openai/anthropic 并修正 base_url。
- 对 Anthropic 渠道点「测试」→ 不再假阳；key 错误时如实报 auth 类错误。
- 对 base_url 漏了 /v1 的 OpenAI 渠道，测试能自动补 /v1 通过。

---

## B. 核心接口扩展：`/v1/embeddings`（P1）

**目的**：很多客户端（RAG/向量库）需要 embeddings，这是当前最缺的核心接口。只做 OpenAI 格式（Anthropic 无 embeddings）。

> 铁律守护：embeddings 只有输入 token，`charge = base_cost × rate_multiplier`，其中 `base_cost = 输入token × input_price`。直接复用现有 `actualChargeForModel`（令 `OutputTokens=0`），不新写计费公式。

**步骤**：

1. **Provider 接口** `backend/internal/upstream/provider.go` 加方法：
   ```go
   ForwardEmbeddings(ctx context.Context, baseURL, apiKey string, timeoutSeconds int, rawBody []byte, upstreamModelName string) (ChatResponse, error)
   ```
2. **OpenAI 适配器** `openai_client.go` 实现：POST `<base>/embeddings`，复用现有 HTTP 转发与 `ensureJSONResponse`，body 用一个**不注入 stream_options** 的 `PrepareEmbeddingsBody`（只覆盖 `model` 字段为 upstreamModelName）。响应解析 `usage{prompt_tokens,total_tokens}` 进 `ChatResponse.Usage`（`completion_tokens` 恒 0）。
3. **Anthropic 适配器** `anthropic_adapter.go` 实现：直接返回 `errors.New("anthropic 渠道不支持 embeddings")`。
4. **GatewayService** `gateway_service.go` 加 `Embeddings(ctx, principal, rawBody, clientIP) (int, []byte, error)`：
   - 仿 `Chat()`，但：解析 `{model, input}`；`req.MaxTokens` 恒 0（embeddings 无输出预算）；预扣用 `estimateChargeForModel`（输出 0）。
   - 新增 `forwardEmbeddingsWithRetry` + `forwardEmbeddingsOnce`，**克隆** `forwardWithRetry`/`forwardOnce` 的渠道选择/熔断/sticky 逻辑，仅把 `provider.ForwardChat` 换成 `provider.ForwardEmbeddings`。
   - `RecordAndCharge` 的 `Endpoint` 填 `"/v1/embeddings"`，`actualChargeForModel`（usage.CompletionTokens=0）。
5. **Handler** `handler/gateway/handler.go` 加 `Embeddings` 方法（**仅非流式**），错误体复用 v1.4 的 `writeGatewayError`/`writeGatewayBody`。
6. **路由** `server.go` `/v1` 组加：
   ```go
   r.Post("/embeddings", gatewayHandler.Embeddings)
   ```

**验收 B**：给一个绑定了 embedding 模型（如 `text-embedding-3-small`）的 openai 渠道，`curl /v1/embeddings -d '{"model":"...","input":"hello"}'` 返回标准向量，且 `gateway_requests` 落一条 `/v1/embeddings` 记录、按输入 token 扣费、`charge=base_cost×倍率`。

> 若时间紧张，B 可在 A、C 完成后单独提交；但**不得**为了 B 改动 chat 既有 money 路径。

---

## C. 七处 UI/UX 修复

### C1【admin】模型保存按钮位置异常（用户问题 1）

**现状**：`frontend/admin/src/pages/model-form.tsx` 第 30 行 `<Button ...>保存</Button>` 紧贴 `<Space>` 无外距；且该表单同时用于：①模型管理"创建模型"卡片（`models.tsx` 的 `ModelsPage`）②`main.tsx` 第 509-511 行的"编辑模型"Modal——Modal 自带 OK 按钮调 `modelForm.submit()`，导致 Modal 内出现**两个保存按钮**。

**步骤**：
1. `model-form.tsx` 给 `ModelForm` 加可选 prop `hideSubmit?: boolean`，并把按钮包进有上边距的 `Form.Item`：
   ```tsx
   export function ModelForm({ form, onFinish, hideSubmit }: { form: FormInstance<Omit<ModelConfig,"id">>; onFinish: (v: Omit<ModelConfig,"id">) => Promise<void>; hideSubmit?: boolean }) {
     // ...
     {!hideSubmit && (
       <Form.Item style={{ marginTop: 12, marginBottom: 0 }}>
         <Button type="primary" htmlType="submit">保存</Button>
       </Form.Item>
     )}
   ```
2. `main.tsx` 第 510 行编辑 Modal 内改为 `<ModelForm form={modelForm} onFinish={handleUpdateModel} hideSubmit />`（只保留 Modal 的 OK）。
3. `models.tsx` 的 `ModelsPage` 创建卡片保持不传 `hideSubmit`（显示带边距的保存按钮）。

**验收**：编辑模型 Modal 只剩一个底部"确定"按钮；创建卡片的"保存"按钮与表单有正常间距。

---

### C2【admin+后端】兑换码详情：完整卡号 + 谁用的 + 何时用（用户问题 2）

**现状盘点**：
- `redeem_codes` 表只存 `code_hash`（不可逆）+ `code_prefix`（前缀），**完整卡号生成后未持久化**，列表查不到。
- `redeem_records` 表**已有** `user_id / created_at / client_ip / amount / ledger_id`，但没有任何 repo/service/handler 把它 join 出来。
- TS `RedeemCode` 缺 `expires_at`。

**安全权衡说明（写给用户）**：兑换码是管理员发放的充值码、仅管理员可见，存明文卡号属业界常见取舍（sub-api 等同类站点也存）。本方案采用"存明文 + 仅管理员接口可见"。若你更在意安全，可改为只在生成时一次性展示、不落库——告知我即可切换。**默认实现存明文。**

**步骤**：

1. **迁移** 新建 `backend/migrations/0007_redeem_code_plain.up.sql`：
   ```sql
   ALTER TABLE redeem_codes ADD COLUMN IF NOT EXISTS code_plain TEXT NOT NULL DEFAULT '';
   ```
   （历史码该列为空，前端回退显示 `code_prefix + ***`。）
2. **repository** `backend/internal/repository/redeem.go`：
   - `CreateRedeemCodeInput` 加 `CodePlain string`；`Create` 的 INSERT 写入 `code_plain`。
   - `ListPaged` 的 SELECT 增 `code_plain`，扫到 `RedeemCode.Code` 字段（已存在 `Code string json:"code,omitempty"`）。同时 SELECT 出 `expires_at` 填充 `ExpiresAt`。
   - 新增方法 `Records(ctx, codeID string) ([]RedeemRecord, error)`：
     ```sql
     SELECT r.id, r.user_id, u.username, r.amount, r.client_ip::text, r.created_at
     FROM redeem_records r JOIN users u ON u.id = r.user_id
     WHERE r.redeem_code_id = $1 ORDER BY r.created_at DESC
     ```
     新增结构体 `RedeemRecord{ID, UserID, Username, Amount, ClientIP, CreatedAt}`。
3. **service** `redeem_service.go`：`Create` 里把生成的明文 `code` 同时塞进 `CreateRedeemCodeInput.CodePlain`；新增 `Records(ctx, codeID)` 透传 repo。
4. **handler** `handler/admin/redeem.go` 加：
   ```go
   func (h RedeemHandler) Records(w http.ResponseWriter, r *http.Request) {
       items, err := h.redeems.Records(r.Context(), chi.URLParam(r, "id"))
       if err != nil { httpx.Error(w, http.StatusInternalServerError, err.Error()); return }
       httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
   }
   ```
   `server.go` `/api/admin` 加 `r.Get("/redeem-codes/{id}/records", adminRedeems.Records)`。
5. **前端类型** `types.ts`：`RedeemCode` 补 `expires_at?: string`；新增 `RedeemRecord{ id; user_id; username; amount; client_ip; created_at }`。`api-client.ts` 加 `listRedeemRecords(id) => request<{items: RedeemRecord[]}>(\`/api/admin/redeem-codes/${id}/records\`)`。
6. **前端页面** `frontend/admin/src/pages/redeem.tsx`：
   - 列表新增列：「完整卡号」（`item.code || item.code_prefix + "****"`，带复制按钮，用 D2 的 `copyText`）、「有效期」（`expires_at || "永久"`）、「使用记录」按钮。
   - 点「使用记录」打开 Drawer/Modal，调 `api.listRedeemRecords(id)` 展示表格：用户名（链 `/users/{user_id}`）、入账金额、IP、兑换时间。
   - 生成成功后的展示从单行 Alert 升级为可复制的卡号表格（沿用 createdCodes）。

**验收**：兑换码列表能看到完整卡号并一键复制；点「使用记录」能看到谁、何时、从哪个 IP 兑换、入账多少。

---

### C3【admin+后端】审计日志清理（二次确认）；系统清理同理（用户问题 3）

**现状**：`settings.tsx` 的"系统清理"**已有** `Modal.confirm` 二次确认（满足"系统清理同理"）。审计日志页 `audit.tsx` 完全没有清理入口。

**步骤**：
1. **后端** `backend/internal/repository/` 的审计 repo（`audit*.go`）加：
   ```go
   func (r AuditRepository) DeleteOlderThan(ctx context.Context, days int) (int64, error) {
       tag, err := r.db.Exec(ctx, `DELETE FROM audit_logs WHERE created_at < now() - ($1 || ' days')::interval`, days)
       return tag.RowsAffected(), err
   }
   ```
2. **handler** 在已暴露 `AuditLogs` 的 settings handler（`adminSettings`）加：
   ```go
   func (h SettingsHandler) CleanupAuditLogs(w http.ResponseWriter, r *http.Request) {
       var input struct{ BeforeDays int `json:"before_days"` }
       _ = httpx.Decode(r, &input)
       if input.BeforeDays < 7 { input.BeforeDays = 90 } // 防误删，最少保留 7 天
       deleted, err := h.audits.DeleteOlderThan(r.Context(), input.BeforeDays)
       if err != nil { httpx.Error(w, http.StatusInternalServerError, err.Error()); return }
       httpx.JSON(w, http.StatusOK, map[string]any{"deleted": deleted})
   }
   ```
   （`SettingsHandler` 已持有 `auditRepo`，见 `server.go:71` 的 `NewSettingsHandler(settingsService, auditRepo)`。）
   `server.go` 加 `r.Post("/audit-logs/cleanup", adminSettings.CleanupAuditLogs)`。
3. **前端** `api-client.ts` 加 `cleanupAuditLogs: (before_days: number) => request<{deleted:number}>("/api/admin/audit-logs/cleanup", { method:"POST", body: JSON.stringify({ before_days }) })`。
4. **前端页面** `audit.tsx` 顶部筛选 Card 加一个 `InputNumber`（默认 90，min 7）+「清理审计日志」按钮，点击走 `Modal.confirm` 二次确认：
   ```tsx
   Modal.confirm({
     title: "确认清理审计日志？",
     content: `将永久删除 ${beforeDays} 天前的审计日志，不可恢复。`,
     okText: "确认清理", cancelText: "取消", okButtonProps: { danger: true },
     onOk: async () => { const r = await api.cleanupAuditLogs(beforeDays); message.success(`已清理 ${r.deleted} 条`); await loadAuditLogs(); }
   })
   ```

**验收**：审计日志页能按"保留天数"清理旧日志，点击有二次确认弹窗；系统清理保持已有的二次确认。

---

### C4【admin+后端】概览精致化（用户问题 5 主诉 + 问题 4）

**现状**：`admin-dashboard.tsx` 只渲染 5 张 `metricCards`（管理员/今日请求/今日扣费/毛利/审计日志），大片空白。后端 `AdminDashboard` 只有 6 个字段。

**步骤**：

1. **后端扩展 `AdminDashboard`** — `backend/internal/repository/reports.go` 的 `AdminDashboard` struct 增字段，并在其聚合 SQL（约 249-261 行）补子查询：
   ```go
   TotalUsers      int    `json:"total_users"`
   TotalChannels   int    `json:"total_channels"`
   HealthyChannels int    `json:"healthy_channels"`
   TotalModels     int    `json:"total_models"`
   EnabledModels   int    `json:"enabled_models"`
   TodaySuccesses  int    `json:"today_successes"`
   TodayFailures   int    `json:"today_failures"`
   ActiveAPIKeys   int    `json:"active_api_keys"`
   TotalRequests   int    `json:"total_requests"`
   ```
   SQL 追加（全部 `deleted_at IS NULL` 过滤，照搬现有风格）：
   ```sql
   COALESCE((SELECT count(*) FROM users WHERE deleted_at IS NULL),0)::int,
   COALESCE((SELECT count(*) FROM upstream_channels WHERE deleted_at IS NULL),0)::int,
   COALESCE((SELECT count(*) FROM upstream_channels WHERE deleted_at IS NULL AND health='healthy' AND status='enabled'),0)::int,
   COALESCE((SELECT count(*) FROM models WHERE deleted_at IS NULL),0)::int,
   COALESCE((SELECT count(*) FROM models WHERE deleted_at IS NULL AND status='enabled'),0)::int,
   COALESCE((SELECT count(*) FROM gateway_requests WHERE created_at::date=now()::date AND status='success'),0)::int,
   COALESCE((SELECT count(*) FROM gateway_requests WHERE created_at::date=now()::date AND status='failed'),0)::int,
   COALESCE((SELECT count(*) FROM api_keys WHERE deleted_at IS NULL AND status='active'),0)::int,
   COALESCE((SELECT count(*) FROM gateway_requests),0)::int
   ```
   > agent 务必核对各表真实列名（`users.deleted_at`、`api_keys.status`、`models` 表名等），以现有迁移为准；若某表无 `deleted_at` 则去掉该过滤。
2. **TS 类型** `types.ts` 的 `AdminDashboard` 同步加上述字段（可选 `?`）。
3. **dashboard 传 api** — `main.tsx:479` 路由改为传入 `api`：
   `<AdminDashboardPage dashboard={dashboard} auditCount={auditCount} me={me} api={api} />`。
4. **重写** `frontend/admin/src/pages/admin-dashboard.tsx` 为分组网格 + 近 7 日趋势 + Top 榜：
   - 顶部分三组卡片（用 antd `Row/Col gutter` 做响应式网格，别再用 `Space wrap`）：
     - 「今日」：今日请求 / 成功 / 失败 / 成功率 / 今日扣费 / 今日成本
     - 「全站资产」：总用户 / 活跃用户 / 余额池 / 累计毛利 / 累计请求
     - 「资源健康」：渠道(healthy/total) / 模型(enabled/total) / 活跃密钥 / 审计日志数
   - 近 7 日趋势：`useEffect` 调 `api.adminReportDaily("","")` 取最近数据，用**零依赖**的内联条形图（见下）展示每日请求量/扣费。
   - Top 模型 / Top 渠道：调 `api.adminReportByModel`、`api.adminReportByChannel`，各取前 5 用小表格展示（列：维度/请求数/扣费）。
   - **零依赖迷你条形图**（加到 `admin-page-utils.tsx`，避免引入图表库）：
     ```tsx
     export function MiniBars({ data }: { data: { label: string; value: number }[] }) {
       const max = Math.max(1, ...data.map((d) => d.value));
       return (
         <div style={{ display: "flex", alignItems: "flex-end", gap: 8, height: 140 }}>
           {data.map((d) => (
             <div key={d.label} style={{ flex: 1, textAlign: "center" }}>
               <div style={{ height: `${(d.value / max) * 110}px`, background: "#4f46e5", borderRadius: 4 }} title={`${d.value}`} />
               <div style={{ fontSize: 12, color: "#888", marginTop: 4 }}>{d.label.slice(5)}</div>
             </div>
           ))}
         </div>
       );
     }
     ```

**验收**：概览不再是孤零零 5 张卡 + 空白，而是三组指标 + 7 日趋势图 + Top5 模型/渠道，信息密度饱满。

---

### C5【admin】数据报表精致化（用户问题 4）

**现状**：`reports.tsx` 表格朴素、金额未格式化、概览卡只 3 个。

**步骤**（轻量、不引图表库）：
1. 顶部概览卡用 `Row/Col` 网格替代 `Space wrap`，金额统一过一个 `fmtMoney`（千分位，保留 2-4 位）。在 `admin-page-utils.tsx` 加：
   ```tsx
   export function fmtMoney(v?: string | number) {
     const n = Number(v ?? 0);
     return Number.isFinite(n) ? n.toLocaleString("zh-CN", { minimumFractionDigits: 2, maximumFractionDigits: 6 }) : "0";
   }
   ```
2. `reportColumns` 的 `成本/扣费/毛利` 列 `render: (v) => fmtMoney(v)`；为各报表表格加 `summary` 合计行（请求数求和、扣费求和），并设 `size="small"`、`pagination={{ pageSize: 10 }}`。
3. 「按日」Tab 顶部加 C4 的 `MiniBars`（按日扣费），让报表页也有趋势可视化。

**验收**：报表页金额带千分位、有合计行、按日有趋势条形图，观感明显提升。

---

### C6【user】生成的 Key 没有"复制功能"（用户问题 6）

**根因（重要）**：`api-keys.tsx` 其实**有**复制按钮，但 `copyKey()` 用 `navigator.clipboard.writeText`，而生产是 `http://155.248.195.94:19080`（**非安全上下文**），浏览器下 `navigator.clipboard` 为 `undefined` → 复制静默失败 → 用户感觉"有按钮没功能"。

**步骤**：
1. 新建 `frontend/user/src/lib/clipboard.ts`（剪贴板兜底）：
   ```ts
   export async function copyText(text: string): Promise<boolean> {
     try {
       if (navigator.clipboard && window.isSecureContext) {
         await navigator.clipboard.writeText(text);
         return true;
       }
     } catch { /* 落到兜底 */ }
     try {
       const ta = document.createElement("textarea");
       ta.value = text;
       ta.style.position = "fixed";
       ta.style.opacity = "0";
       document.body.appendChild(ta);
       ta.focus();
       ta.select();
       const ok = document.execCommand("copy");
       document.body.removeChild(ta);
       return ok;
     } catch {
       return false;
     }
   }
   ```
2. `frontend/user/src/routes/api-keys.tsx` 的 `copyKey` 改用 `copyText`：
   ```ts
   async function copyKey() {
     const ok = await copyText(plaintext);
     if (ok) { setCopied(true); toast.success("已复制到剪贴板"); setTimeout(() => setCopied(false), 1500); }
     else { toast.error("复制失败，请手动选择复制"); }
   }
   ```

**验收**：在 http（非 https）环境下点复制也能复制成功；失败时给明确提示而非静默。

---

### C7【user】兑换码页大片空白（用户问题 7）

**现状**：`frontend/user/src/routes/redeem.tsx` 右列（`lg:grid-cols-[1fr_0.75fr]`）只有一行「兑换成功后余额即时到账。」，右半屏空旷。

**步骤**：把右列那个 `<div className="rounded-lg border ...">` 替换为内容丰富的"充值说明 + 当前余额"卡片（`useAuth()` 暴露 `user`，可读 `user.balance`）：
```tsx
<div className="grid content-start gap-4 rounded-lg border border-white/10 bg-white/[0.035] p-5">
  <div>
    <p className="text-sm text-muted-foreground">当前余额</p>
    <p className="text-2xl font-semibold text-primary">¥ {formatMoney(user?.balance)}</p>
  </div>
  <div className="grid gap-2 text-sm text-muted-foreground">
    <p className="font-medium text-foreground">充值说明</p>
    <p>1. 向管理员获取兑换码（形如 LS-XXXX-XXXX）。</p>
    <p>2. 在左侧输入框粘贴兑换码并点击「兑换」。</p>
    <p>3. 兑换成功后余额即时到账，可在下方账本查看。</p>
    <p>4. 每个兑换码有使用次数与有效期限制，过期或用尽将失效。</p>
  </div>
  <p className="text-xs text-muted-foreground">如兑换异常请联系管理员核对卡号状态。</p>
</div>
```
（`user` 从 `const { api, refreshMe, user } = useAuth();` 取；`formatMoney` 已在 `lib/utils` 导出。）

**验收**：兑换页右列不再空旷，展示当前余额 + 4 步充值说明 + 提示。

---

## D. 附带修复（随手做）

- **D1** TS `RedeemCode` 补 `expires_at`；`createRedeemCodes` 入参补 `expires_at?` 与 `max_uses?`（后端 `CreateRedeemInput` 已支持 `ExpiresAt`，前端"生成兑换码"表单可加一个可选有效期 `Input type="date"`）。
- **D2** C2 的完整卡号复制、C6 的 key 复制都走统一剪贴板兜底；admin 端若需复制（如兑换码），可在 admin 侧同样加一个 `copyText`（admin 无 `lib/clipboard`，可在 `admin-page-utils.tsx` 内放一份同款实现）。

---

## 五、执行顺序与验收

### Agent 执行顺序（严格）
**A1 → A2 → B → C1 → C2 → C3 → C4 → C5 → C6 → C7 → D**，每节按文档精确步骤做，不要自由扩展，不要碰"明确不做"清单。

### 收尾必做
1. `cd backend && go build ./... && go vet ./... && go test ./...` 全绿。
2. `cd frontend && npm --workspace @lingshu/admin run build` 通过；**再单独**跑 `npm --workspace @lingshu/user run build` 通过（**不要并行跑两个 workspace build**，会触发 npm workspace 误报 ENOENT）。
3. 新增测试：
   - `backend/internal/upstream/probe_test.go`：用 `httptest` 起假上游，覆盖
     - OpenAI `/models` 返回 `{"data":[{"id":"gpt-x"}]}` → `DetectProtocol` 判 `openai`；
     - base_url 不含 /v1 时自动补 /v1 命中 → `NormalizedBase` 带 /v1；
     - Anthropic `/v1/models` 200 → 判 `anthropic`；
     - **严格性**：Anthropic 上游 401 时 `ProbeAnthropic` 必须返回错误（**不得**回退 `anthropicPresetModels`）。
   - 若做了 B：`gateway`/`upstream` 加一条 embeddings 计费用例，断言 `charge = base_cost × rate_multiplier` 且 `completion_tokens=0`。
4. 迁移自检：确认 `0007_redeem_code_plain.up.sql` 能在现有 migrate 流程跑过（`cmd/migrate`）。
5. `git commit` 但 **不要 push**，等 Kiro 验收后由用户决定推。

### 关键回归红线（改完务必自测）
- 流式 chat 仍逐块推送、usage 来自上游末帧回灌（v1.3 money 修复不能回退）。
- 上游错误透传仍保留原始 HTTP 状态码 + JSON body（v1.4 BUG3 不能回退）。
- 软删除过滤 `deleted_at IS NULL` 不能因为 C4 的新 SQL 而遗漏。

---

## 六、给 agent 的执行指令（直接复制）

```
按 docs/LingShu-v1.5-多格式自适应与核心功能强化.md 顺序执行
A1 供应商预设 → A2 协议自动探测+修复渠道测试 → B /v1/embeddings → C1 模型保存按钮 → C2 兑换码详情(完整卡号/谁用/何时) → C3 审计日志清理(二次确认) → C4 概览精致化 → C5 报表精致化 → C6 用户key复制(非安全上下文兜底) → C7 兑换页排版 → D 类型与剪贴板补全
每节按文档精确步骤做，不要自由扩展，不要做"明确不做"清单里的营销/增长/2FA/分组/图片TTS
铁律不变：charge=base_cost×rate_multiplier；用户端禁露 base_cost/rate_multiplier/gross_profit；软删除 deleted_at IS NULL
不得破坏 v1.3 流式真实usage回灌 与 v1.4 上游错误透传
关键坑：协议探测/渠道测试必须用严格探针 ProbeOpenAI/ProbeAnthropic，绝不能复用会回退预设的 AnthropicAdapter.ListModels
做完跑全套 go build/vet/test + 两个前端 workspace 分别 build，新增 probe_test.go，commit 但不要 push 等我推
```

---

## 七、给用户的话（你和 subapi 的差距 & 本版定位）

你问"不知道跟 subapi 差距在哪"。差距主要在三类，**本版精准补前两类**：
1. **接入体验**（本版 A）：subapi 选个供应商就自动配好；我们之前每次手填。A1 预设 + A2 自适应探测把这条补平——**以后加 DeepSeek/智谱/Kimi 就是"选一下、贴个 key"**。
2. **接口覆盖**（本版 B）：补 `/v1/embeddings`，向量类客户端能直接用；图片/TTS 等按你的"先不做杂项"原则推后。
3. **精致度**（本版 C）：概览从 5 个数字补到三组指标+趋势+榜单，报表加合计与可视化，修掉模型保存按钮、key 复制、兑换页空白等手感问题。

**仍坚持不做**：营销门面、增长玩法、2FA、分组——按你的要求全部排除，等核心打磨扎实再说。
