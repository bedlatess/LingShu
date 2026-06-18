# LingShu API

## 核心概念

LingShu 的统一入口是“平台 API Key”。一个 Key 即可调用所有已上线模型，只需在 OpenAI SDK 的 `model` 字段填写模型 `public_name`，例如 `gpt-5.5`、`claude-opus-4-7`。不需要为每个上游创建独立 Key。

模型可见性 = 该模型 `status='enabled'` 且至少绑定了一个 `enabled` 渠道。管理员新增渠道不会自动新增模型，模型和渠道是 N:N 关系，需要在管理端绑定。v1.3 起支持“渠道 -> 批量拉取上游 `/models` -> 一键导入为模型”功能。

All API responses are JSON unless explicitly noted. Admin and user APIs use
`Authorization: Bearer <jwt>`. Gateway APIs use a platform API key in the same
header.

## Auth

- `POST /api/auth/login` - username/email + password, returns JWT and user.
- `POST /api/auth/logout` - client-side token discard helper.
- `POST /api/auth/register` - available only when registration is enabled.
- `GET /api/auth/me` - current user profile.
- `POST /api/auth/change-password` - change current user's password.

## User Console

- `GET /api/user/dashboard` - balance, usage summary, frozen amount, model count.
- `GET /api/user/models` - enabled models and user-facing unit prices only.
- `GET /api/user/api-keys` - user's API keys.
- `POST /api/user/api-keys` - create key; plaintext is returned once.
- `PATCH /api/user/api-keys/{id}` - rename or disable key.
- `DELETE /api/user/api-keys/{id}` - delete/disable key.
- `GET /api/user/announcements` - visible announcements.
- `POST /api/user/redeem` - redeem a code.
- `GET /api/user/usage/logs` - user request logs without base cost, multiplier, or gross profit.
- `GET /api/user/usage/ledger` - user ledger without internal cost fields.
- `GET /api/user/usage/stats/daily?days=7` - daily usage stats.
- `GET /api/user/usage/stats/models` - usage by model.

User-facing DTOs must not expose `base_cost`, `rate_multiplier`, or
`gross_profit`.

## Admin

List endpoints return:

```json
{ "items": [], "total": 0, "page": 1, "limit": 20 }
```

Core resources:

- `GET/POST /api/admin/users`
- `GET/PATCH /api/admin/users/{id}`
- `POST /api/admin/users/{id}/reset-password`
- `POST /api/admin/users/{id}/ban`
- `POST /api/admin/users/{id}/balance` - manual grant/deduct; remark required.
- `GET /api/admin/users/{id}/logs`
- `GET /api/admin/users/{id}/ledger`
- `GET /api/admin/users/{id}/api-keys`
- `GET/POST /api/admin/api-keys`
- `PATCH/DELETE /api/admin/api-keys/{id}`
- `GET/POST /api/admin/models`
- `GET/PATCH/DELETE /api/admin/models/{id}`
- `POST /api/admin/models/{id}/disable`
- `GET/POST /api/admin/channels`
- `GET/PATCH/DELETE /api/admin/channels/{id}`
- `POST /api/admin/channels/{id}/test`
- `POST /api/admin/channels/{id}/disable`
- `POST /api/admin/channel-models`
- `DELETE /api/admin/channels/{channelID}/models/{modelID}`
- `GET/POST /api/admin/announcements`
- `PATCH/DELETE /api/admin/announcements/{id}`
- `GET/POST /api/admin/redeem-codes`
- `POST /api/admin/redeem-codes/batch`
- `POST /api/admin/redeem-codes/{id}/disable`

Operations and reporting:

- `GET /api/admin/dashboard`
- `GET /api/admin/gateway-requests`
- `GET /api/admin/balance-ledger`
- `GET /api/admin/reports/daily?from=&to=`
- `GET /api/admin/reports/by-user?from=&to=`
- `GET /api/admin/reports/by-model?from=&to=`
- `GET /api/admin/reports/by-channel?from=&to=`
- `GET/PATCH /api/admin/settings`
- `GET /api/admin/audit-logs?actor_id=&action=&target_type=&from=&to=`
- `GET /api/admin/audit-count`

## Gateway

- `GET /v1/models` - OpenAI-compatible model list.
- `POST /v1/chat/completions` - OpenAI-compatible chat completions.

Clients use:

```text
base_url = https://your-domain/v1
api_key  = lsk_...
```

Billing invariant:

```text
charge = base_cost * rate_multiplier
```

`gateway_requests` and `balance_ledger` both store `base_cost`,
`rate_multiplier`, and `charge`. Streaming requests request upstream usage and
finalize billing after the response body is captured. If upstream usage is not
available, the gateway marks the request as estimated.
