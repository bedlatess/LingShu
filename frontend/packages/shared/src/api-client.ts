import type { APIKey, AdminDashboard, Announcement, AuditLog, Channel, CreatedAPIKey, GatewayLog, HealthResponse, LedgerRecord, LoginResponse, ModelConfig, RedeemCode, SystemSetting, User, UserDashboard, UserModelPrice } from "./types";

const env = (import.meta as unknown as { env?: Record<string, string | undefined> }).env;
const apiBaseURL = env?.VITE_API_BASE_URL ?? "http://localhost:8080";

export async function getHealth(): Promise<HealthResponse> {
  const response = await fetch(`${apiBaseURL}/healthz`);
  if (!response.ok) {
    throw new Error(`health check failed: ${response.status}`);
  }
  return response.json();
}

export function createAPI(token?: string) {
  async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
    const headers = new Headers(init.headers);
    headers.set("Content-Type", "application/json");
    if (token) {
      headers.set("Authorization", `Bearer ${token}`);
    }
    const response = await fetch(`${apiBaseURL}${path}`, { ...init, headers });
    if (!response.ok) {
      const body = await response.json().catch(() => ({ error: response.statusText }));
      throw new Error(body.error ?? response.statusText);
    }
    return response.json();
  }

  return {
    login: (login: string, password: string) =>
      request<LoginResponse>("/api/auth/login", {
        method: "POST",
        body: JSON.stringify({ login, password })
      }),
    me: () => request<User>("/api/auth/me"),
    listUsers: () => request<{ items: User[] }>("/api/admin/users"),
    createUser: (payload: { username: string; email?: string; password: string; role: "admin" | "user" }) =>
      request<User>("/api/admin/users", {
        method: "POST",
        body: JSON.stringify(payload)
      }),
    auditCount: () => request<{ count: number }>("/api/admin/audit-count"),
    listAPIKeys: () => request<{ items: APIKey[] }>("/api/admin/api-keys"),
    createAPIKey: (payload: { user_id: string; name: string }) =>
      request<APIKey & { plaintext: string }>("/api/admin/api-keys", {
        method: "POST",
        body: JSON.stringify(payload)
      }),
    adjustUserBalance: (userID: string, payload: { amount: string; remark: string }) =>
      request<User>(`/api/admin/users/${userID}/balance`, {
        method: "POST",
        body: JSON.stringify(payload)
      }),
    listModels: () => request<{ items: ModelConfig[] }>("/api/admin/models"),
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
    listSettings: () => request<{ items: SystemSetting[] }>("/api/admin/settings"),
    patchSettings: (items: { key: string; value: string }[]) =>
      request<{ items: SystemSetting[] }>("/api/admin/settings", {
        method: "PATCH",
        body: JSON.stringify({ items })
      }),
    listAuditLogs: () => request<{ items: AuditLog[] }>("/api/admin/audit-logs"),
    listChannels: () => request<{ items: Channel[] }>("/api/admin/channels"),
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
    bindChannelModel: (payload: { channel_id: string; model_id: string; upstream_model_name: string }) =>
      request<{ id: string }>("/api/admin/channel-models", {
        method: "POST",
        body: JSON.stringify(payload)
      }),
    listAnnouncements: () => request<{ items: Announcement[] }>("/api/admin/announcements"),
    createAnnouncement: (payload: { title: string; content: string; status: string; priority: number; pinned: boolean }) =>
      request<Announcement>("/api/admin/announcements", {
        method: "POST",
        body: JSON.stringify(payload)
      }),
    listRedeemCodes: () => request<{ items: RedeemCode[] }>("/api/admin/redeem-codes"),
    createRedeemCodes: (payload: { amount: string; count: number; batch_name?: string; max_uses?: number }) =>
      request<{ items: RedeemCode[] }>("/api/admin/redeem-codes", {
        method: "POST",
        body: JSON.stringify(payload)
      }),
    userAnnouncements: () => request<{ items: Announcement[] }>("/api/user/announcements"),
    userDashboard: () => request<UserDashboard>("/api/user/dashboard"),
    userModels: () => request<{ items: UserModelPrice[] }>("/api/user/models"),
    userAPIKeys: () => request<{ items: APIKey[] }>("/api/user/api-keys"),
    createUserAPIKey: (payload: { name: string }) =>
      request<CreatedAPIKey>("/api/user/api-keys", {
        method: "POST",
        body: JSON.stringify(payload)
      }),
    updateUserAPIKey: (id: string, payload: { name?: string; status?: string }) =>
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
    adminLogs: () => request<{ items: GatewayLog[] }>("/api/admin/gateway-requests"),
    adminLedger: () => request<{ items: LedgerRecord[] }>("/api/admin/balance-ledger"),
    userLogs: () => request<{ items: GatewayLog[] }>("/api/user/usage/logs"),
    userLedger: () => request<{ items: LedgerRecord[] }>("/api/user/usage/ledger"),
    userDailyStats: () => request<{ items: any[] }>("/api/user/usage/stats/daily?days=7"),
    userModelStats: () => request<{ items: any[] }>("/api/user/usage/stats/models")
  };
}
