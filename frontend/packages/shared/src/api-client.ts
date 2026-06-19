import type { APIKey, AdminDashboard, Announcement, AuditLog, Channel, ChannelDetectResult, ChannelDetail, ChannelPreset, CreatedAPIKey, DailyStat, GatewayLog, HealthResponse, LedgerRecord, LoginResponse, ModelConfig, ModelDetail, ModelStat, RedeemCode, RedeemRecord, ReportRow, SystemSetting, User } from "./admin-types";
import type { ChannelModelImportInput, ChannelModelImportResult, ChannelModelSyncResult, CleanupHistoryEntry, CleanupResult, OpsDashboard, PaginatedResponse, PublicModel, PublicSiteInfo } from "./types";
import type { UserDashboard, UserGatewayLog, UserLedgerRecord, UserModelConfig } from "./user-types";

const env = (import.meta as unknown as { env?: Record<string, string | undefined> }).env;
const apiBaseURL = env?.VITE_API_BASE_URL ?? "http://localhost:8080";
const unauthorizedEvent = "lingshu:unauthorized";
const requestTimeoutMs = Number(env?.VITE_API_TIMEOUT_MS ?? 30000);

function withQuery(path: string, query?: Record<string, string | number | undefined>) {
  if (!query) return path;
  const params = new URLSearchParams();
  Object.entries(query).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== "") {
      params.set(key, String(value));
    }
  });
  const search = params.toString();
  return search ? `${path}?${search}` : path;
}

export async function getHealth(): Promise<HealthResponse> {
  const response = await fetch(`${apiBaseURL}/healthz`);
  if (!response.ok) {
    throw new Error(`health check failed: ${response.status}`);
  }
  return response.json();
}

export function createAPI(token?: string) {
  async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
    return requestWithRetry<T>(path, init);
  }

  async function requestWithRetry<T>(path: string, init: RequestInit = {}): Promise<T> {
    const method = String(init.method ?? "GET").toUpperCase();
    try {
      return await requestOnce<T>(path, init);
    } catch (err) {
      if (method === "GET" && isRetryableError(err)) {
        return requestOnce<T>(path, init);
      }
      throw err;
    }
  }

  async function requestOnce<T>(path: string, init: RequestInit = {}): Promise<T> {
    const headers = new Headers(init.headers);
    headers.set("Content-Type", "application/json");
    if (token) {
      headers.set("Authorization", `Bearer ${token}`);
    }
    const controller = new AbortController();
    const timeout = window.setTimeout(() => controller.abort(), requestTimeoutMs);
    let response: Response;
    try {
      response = await fetch(`${apiBaseURL}${path}`, { ...init, headers, signal: controller.signal });
    } catch (err) {
      if (err instanceof DOMException && err.name === "AbortError") {
        throw new Error("请求超时，请稍后重试");
      }
      throw err;
    } finally {
      window.clearTimeout(timeout);
    }
    if (!response.ok) {
      if (response.status === 401) {
        if (typeof window !== "undefined") {
          window.dispatchEvent(new CustomEvent(unauthorizedEvent));
        }
        throw new Error("登录已过期，请重新登录");
      }
      if (response.status === 403) {
        throw new Error("没有权限执行此操作");
      }
      if (response.status === 429) {
        throw new Error("请求过于频繁，请稍后再试");
      }
      const body = await response.json().catch(() => ({ error: response.statusText }));
      throw new Error(errorMessageFromBody(body, response.statusText));
    }
    return response.json();
  }

  return {
    login: (login: string, password: string, captcha_token?: string) =>
      request<LoginResponse>("/api/auth/login", {
        method: "POST",
        body: JSON.stringify({ login, password, captcha_token })
      }),
    sendEmailCode: (payload: { purpose: "register" | "reset"; email: string; captcha_token?: string }) =>
      request<{ status: string }>("/api/auth/email/send-code", {
        method: "POST",
        body: JSON.stringify(payload)
      }),
    register: (payload: { username: string; email: string; password: string; code: string; captcha_token?: string }) =>
      request<User>("/api/auth/register", {
        method: "POST",
        body: JSON.stringify(payload)
      }),
    forgotPassword: (email: string, captcha_token?: string) =>
      request<{ status: string }>("/api/auth/forgot", {
        method: "POST",
        body: JSON.stringify({ email, captcha_token })
      }),
    resetPassword: (payload: { email: string; code: string; new_password: string }) =>
      request<{ status: string }>("/api/auth/reset", {
        method: "POST",
        body: JSON.stringify(payload)
      }),
    changePassword: (payload: { old_password: string; new_password: string }) =>
      request<{ status: string }>("/api/auth/change-password", {
        method: "POST",
        body: JSON.stringify(payload)
      }),
    publicModels: () => request<{ items: PublicModel[] }>("/api/public/models"),
    siteInfo: () => request<PublicSiteInfo>("/api/public/site-info"),
    legal: (slug: "tos" | "privacy") => request<{ slug: string; markdown: string }>(`/api/public/legal/${slug}`),
    me: () => request<User>("/api/auth/me"),
    listUsers: (page?: number, limit?: number) => request<PaginatedResponse<User>>(withQuery("/api/admin/users", { page, limit })),
    getUser: (id: string) => request<User>(`/api/admin/users/${id}`),
    createUser: (payload: { username: string; email?: string; password: string; role: "admin" | "user" }) =>
      request<User>("/api/admin/users", {
        method: "POST",
        body: JSON.stringify(payload)
      }),
    updateUser: (id: string, payload: { username?: string; email?: string; status?: string }) =>
      request<User>(`/api/admin/users/${id}`, {
        method: "PATCH",
        body: JSON.stringify(payload)
      }),
    resetUserPassword: (id: string, password: string) =>
      request<{ status: string }>(`/api/admin/users/${id}/reset-password`, {
        method: "POST",
        body: JSON.stringify({ password })
      }),
    banUser: (id: string) =>
      request<User>(`/api/admin/users/${id}/ban`, {
        method: "POST"
      }),
    auditCount: () => request<{ count: number }>("/api/admin/audit-count"),
    listAPIKeys: (page?: number, limit?: number) => request<PaginatedResponse<APIKey>>(withQuery("/api/admin/api-keys", { page, limit })),
    createAPIKey: (payload: { user_id: string; name: string; allowed_endpoints?: string[] }) =>
      request<APIKey & { plaintext: string }>("/api/admin/api-keys", {
        method: "POST",
        body: JSON.stringify(payload)
      }),
    disableAPIKey: (id: string) =>
      request<{ status: string }>(`/api/admin/api-keys/${id}`, {
        method: "PATCH"
      }),
    deleteAPIKey: (id: string) =>
      request<{ status: string }>(`/api/admin/api-keys/${id}`, {
        method: "DELETE"
      }),
    adjustUserBalance: (userID: string, payload: { amount: string; remark: string }) =>
      request<User>(`/api/admin/users/${userID}/balance`, {
        method: "POST",
        body: JSON.stringify(payload)
      }),
    listModels: (page?: number, limit?: number) => request<PaginatedResponse<ModelConfig>>(withQuery("/api/admin/models", { page, limit })),
    getModelDetail: (id: string) => request<ModelDetail>(`/api/admin/models/${id}`),
    createModel: (payload: Omit<ModelConfig, "id">) =>
      request<ModelConfig>("/api/admin/models", {
        method: "POST",
        body: JSON.stringify(payload)
      }),
    updateModel: (id: string, payload: Omit<ModelConfig, "id">) =>
      request<ModelConfig>(`/api/admin/models/${id}`, {
        method: "PATCH",
        body: JSON.stringify(payload)
      }),
    disableModel: (id: string) =>
      request<{ status: string }>(`/api/admin/models/${id}/disable`, {
        method: "POST"
      }),
    deleteModel: (id: string) =>
      request<{ status: string }>(`/api/admin/models/${id}`, {
        method: "DELETE"
      }),
    listSettings: (page?: number, limit?: number) => request<PaginatedResponse<SystemSetting>>(withQuery("/api/admin/settings", { page, limit })),
    patchSettings: (items: { key: string; value: string }[]) =>
      request<{ items: SystemSetting[] }>("/api/admin/settings", {
        method: "PATCH",
        body: JSON.stringify({ items })
      }),
    runCleanup: () =>
      request<{ items: CleanupResult[] }>("/api/admin/cleanup/run", {
        method: "POST"
      }),
    cleanupHistory: (limit = 20) => request<{ items: CleanupHistoryEntry[] }>(withQuery("/api/admin/cleanup/history", { limit })),
    listAuditLogs: (page?: number, limit?: number, filters?: { actor_id?: string; action?: string; target_type?: string; from?: string; to?: string }) =>
      request<PaginatedResponse<AuditLog>>(withQuery("/api/admin/audit-logs", { page, limit, ...filters })),
    cleanupAuditLogs: (before_days: number) =>
      request<{ deleted: number }>("/api/admin/audit-logs/cleanup", {
        method: "POST",
        body: JSON.stringify({ before_days })
      }),
    listChannelPresets: () => request<{ items: ChannelPreset[] }>("/api/admin/channel-presets"),
    detectChannel: (base_url: string, api_key: string, provider_type?: string) =>
      request<ChannelDetectResult>("/api/admin/channels/detect", {
        method: "POST",
        body: JSON.stringify({ base_url, api_key, provider_type })
      }),
    listChannels: (page?: number, limit?: number) => request<PaginatedResponse<Channel>>(withQuery("/api/admin/channels", { page, limit })),
    getChannelDetail: (id: string) => request<ChannelDetail>(`/api/admin/channels/${id}`),
    createChannel: (payload: {
      name: string;
      provider_type: string;
      base_url: string;
      api_key: string;
      status: string;
      weight: number;
      timeout_seconds: number;
      rpm_limit: number;
      concurrency_limit: number;
      fail_threshold: number;
    }) =>
      request<Channel>("/api/admin/channels", {
        method: "POST",
        body: JSON.stringify(payload)
      }),
    updateChannel: (id: string, payload: {
      name: string;
      provider_type: string;
      base_url: string;
      api_key?: string;
      status: string;
      weight: number;
      timeout_seconds?: number;
      rpm_limit?: number;
      concurrency_limit?: number;
      fail_threshold?: number;
    }) =>
      request<Channel>(`/api/admin/channels/${id}`, {
        method: "PATCH",
        body: JSON.stringify(payload)
      }),
    disableChannel: (id: string) =>
      request<{ status: string }>(`/api/admin/channels/${id}/disable`, {
        method: "POST"
      }),
    deleteChannel: (id: string) =>
      request<{ status: string }>(`/api/admin/channels/${id}`, {
        method: "DELETE"
      }),
    testChannel: (id: string, baseURL?: string) =>
      request<{ ok: boolean; status?: number; category: string; message: string; latency_ms: number }>(`/api/admin/channels/${id}/test`, {
        method: "POST",
        body: JSON.stringify({ base_url: baseURL ?? "" })
      }),
    syncChannelModels: (id: string) =>
      request<ChannelModelSyncResult>(`/api/admin/channels/${id}/sync-models`, {
        method: "POST"
      }),
    importChannelModels: (id: string, payload: { strategy: "create_or_bind" | "bind_existing"; models: ChannelModelImportInput[] }) =>
      request<{ items: ChannelModelImportResult[] }>(`/api/admin/channels/${id}/import-models`, {
        method: "POST",
        body: JSON.stringify(payload)
      }),
    bindChannelModel: (payload: { channel_id: string; model_id: string; upstream_model_name: string }) =>
      request<{ id: string }>("/api/admin/channel-models", {
        method: "POST",
        body: JSON.stringify(payload)
      }),
    unbindChannelModel: (channelID: string, modelID: string) =>
      request<{ status: string }>(`/api/admin/channels/${channelID}/models/${modelID}`, {
        method: "DELETE"
      }),
    listAnnouncements: (page?: number, limit?: number) => request<PaginatedResponse<Announcement>>(withQuery("/api/admin/announcements", { page, limit })),
    createAnnouncement: (payload: { title: string; content: string; status: string; priority: number; pinned: boolean }) =>
      request<Announcement>("/api/admin/announcements", {
        method: "POST",
        body: JSON.stringify(payload)
      }),
    updateAnnouncement: (id: string, payload: { title: string; content: string; status: string; priority: number; pinned: boolean }) =>
      request<Announcement>(`/api/admin/announcements/${id}`, {
        method: "PATCH",
        body: JSON.stringify(payload)
      }),
    deleteAnnouncement: (id: string) =>
      request<{ status: string }>(`/api/admin/announcements/${id}`, {
        method: "DELETE"
      }),
    listRedeemCodes: (page?: number, limit?: number) => request<PaginatedResponse<RedeemCode>>(withQuery("/api/admin/redeem-codes", { page, limit })),
    createRedeemCodes: (payload: { amount: string; count: number; batch_name?: string; max_uses?: number; expires_at?: string }) =>
      request<{ items: RedeemCode[] }>("/api/admin/redeem-codes", {
        method: "POST",
        body: JSON.stringify(payload)
      }),
    disableRedeemCode: (id: string) =>
      request<{ status: string }>(`/api/admin/redeem-codes/${id}/disable`, {
        method: "POST"
      }),
    listRedeemRecords: (id: string) => request<{ items: RedeemRecord[] }>(`/api/admin/redeem-codes/${id}/records`),
    userAnnouncements: () => request<{ items: Announcement[] }>("/api/user/announcements"),
    userDashboard: () => request<UserDashboard>("/api/user/dashboard"),
    userModels: () => request<{ items: UserModelConfig[] }>("/api/user/models"),
    userAPIKeys: () => request<{ items: APIKey[] }>("/api/user/api-keys"),
    adminUserLogs: (id: string, page?: number, limit?: number, filters?: { status?: string; model?: string; from?: string; to?: string }) =>
      request<PaginatedResponse<GatewayLog>>(withQuery(`/api/admin/users/${id}/logs`, { page, limit, ...filters })),
    adminUserLedger: (id: string, page?: number, limit?: number, filters?: { type?: string; from?: string; to?: string }) =>
      request<PaginatedResponse<LedgerRecord>>(withQuery(`/api/admin/users/${id}/ledger`, { page, limit, ...filters })),
    adminUserAPIKeys: (id: string, page?: number, limit?: number) =>
      request<PaginatedResponse<APIKey>>(withQuery(`/api/admin/users/${id}/api-keys`, { page, limit })),
    adminUserSummary: (id: string) => request<{ total_charge: string; total_recharge: string }>(`/api/admin/users/${id}/summary`),
    createUserAPIKey: (payload: { name: string; allowed_endpoints?: string[] }) =>
      request<CreatedAPIKey>("/api/user/api-keys", {
        method: "POST",
        body: JSON.stringify(payload)
      }),
    updateUserAPIKey: (id: string, payload: { name?: string; status?: string; allowed_endpoints?: string[] }) =>
      request<APIKey>(`/api/user/api-keys/${id}`, {
        method: "PATCH",
        body: JSON.stringify(payload)
      }),
    deleteUserAPIKey: (id: string) =>
      request<{ status: string }>(`/api/user/api-keys/${id}`, {
        method: "DELETE"
      }),
    redeem: (code: string) =>
      request<RedeemCode>("/api/user/redeem", {
        method: "POST",
        body: JSON.stringify({ code })
      }),
    adminDashboard: () => request<AdminDashboard>("/api/admin/dashboard"),
    adminOps: () => request<OpsDashboard>("/api/admin/ops"),
    adminReportDaily: (from?: string, to?: string) => request<{ items: ReportRow[] }>(withQuery("/api/admin/reports/daily", { from, to })),
    adminReportByUser: (from?: string, to?: string) => request<{ items: ReportRow[] }>(withQuery("/api/admin/reports/by-user", { from, to })),
    adminReportByModel: (from?: string, to?: string) => request<{ items: ReportRow[] }>(withQuery("/api/admin/reports/by-model", { from, to })),
    adminReportByChannel: (from?: string, to?: string) => request<{ items: ReportRow[] }>(withQuery("/api/admin/reports/by-channel", { from, to })),
    adminLogs: (page?: number, limit?: number) => request<PaginatedResponse<GatewayLog>>(withQuery("/api/admin/gateway-requests", { page, limit })),
    adminLedger: (page?: number, limit?: number) => request<PaginatedResponse<LedgerRecord>>(withQuery("/api/admin/balance-ledger", { page, limit })),
    userLogs: () => request<{ items: UserGatewayLog[] }>("/api/user/usage/logs"),
    userLedger: () => request<{ items: UserLedgerRecord[] }>("/api/user/usage/ledger"),
    userDailyStats: () => request<{ items: DailyStat[] }>("/api/user/usage/stats/daily?days=7"),
    userModelStats: () => request<{ items: ModelStat[] }>("/api/user/usage/stats/models")
  };
}

function errorMessageFromBody(body: unknown, fallback: string) {
  if (!body || typeof body !== "object") return fallback;
  const record = body as Record<string, unknown>;
  const error = record.error;
  if (typeof error === "string" && error.trim()) return error;
  if (error && typeof error === "object") {
    const nested = error as Record<string, unknown>;
    const code = typeof nested.code === "string" ? nested.code : "";
    const message = typeof nested.message === "string" ? nested.message : "";
    return [code, message].filter(Boolean).join("：") || fallback;
  }
  if (typeof record.message === "string" && record.message.trim()) return record.message;
  return fallback;
}

function isRetryableError(err: unknown) {
  if (!(err instanceof Error)) return false;
  return err.message === "请求超时，请稍后重试" || err.message === "Failed to fetch" || err.message.includes("NetworkError");
}
