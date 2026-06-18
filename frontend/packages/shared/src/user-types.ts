export type { APIKey, Announcement, CreatedAPIKey, HealthResponse, LoginResponse, RedeemCode, Role, User } from "./types";

export interface UserDashboard {
  balance: string;
  today_charge: string;
  month_charge: string;
  total_charge: string;
  total_recharge: string;
  frozen: string;
  available_models: number;
  today_requests: number;
}

export interface UserGatewayLog {
  request_id: string;
  model_id: string;
  status: string;
  http_status: number;
  total_tokens: number;
  charge: string;
  created_at: string;
}

export interface UserLedgerRecord {
  type: string;
  amount: string;
  balance_before: string;
  balance_after: string;
  remark: string;
  created_at: string;
}

export interface UserModelConfig {
  id: string;
  public_name: string;
  type: string;
  group: string;
  billing_mode: string;
  status: string;
  sort_order: number;
}

export interface UserDailyStat {
  day: string;
  requests: number;
  successes: number;
  failures: number;
  total_tokens: number;
  charge: string;
}

export interface UserModelStat {
  model_id: string;
  requests: number;
  total_tokens: number;
  charge: string;
}
