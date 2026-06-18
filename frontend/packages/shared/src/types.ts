export type Role = "admin" | "user";

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  limit: number;
}

export interface HealthResponse {
  status: string;
  database: string;
  redis: string;
}

export interface User {
  id: string;
  username: string;
  email: string;
  role: Role;
  status: "active" | "banned";
  balance: string;
  created_at: string;
}

export interface APIKey {
  id: string;
  user_id: string;
  mask: string;
  name: string;
  status: string;
  created_at: string;
}

export interface CreatedAPIKey extends APIKey {
  plaintext: string;
}

export interface ModelConfig {
  id: string;
  public_name: string;
  type: string;
  group: string;
  billing_mode: string;
  input_price_per_1k: string;
  output_price_per_1k: string;
  price_per_call: string;
  rate_multiplier: string;
  status: string;
  sort_order: number;
}

export interface SystemSetting {
  key: string;
  value: string;
  description: string;
  updated_by: string;
  updated_at: string;
}

export interface AuditLog {
  id: string;
  actor_id: string;
  action: string;
  target_type: string;
  target_id: string;
  before_snapshot?: Record<string, unknown>;
  after_snapshot?: Record<string, unknown>;
  ip: string;
  user_agent: string;
  created_at: string;
}

export interface Channel {
  id: string;
  name: string;
  provider_type: string;
  base_url: string;
  status: string;
  health: string;
  weight: number;
}

export interface Announcement {
  id: string;
  title: string;
  content: string;
  status: string;
  priority: number;
  pinned: boolean;
  created_at: string;
}

export interface RedeemCode {
  id: string;
  code?: string;
  code_prefix: string;
  batch_name: string;
  amount: string;
  status: string;
  max_uses: number;
  used_count: number;
  created_at: string;
}

export interface GatewayLog {
  request_id: string;
  user_id?: string;
  model_id: string;
  status: string;
  http_status: number;
  total_tokens: number;
  base_cost: string;
  charge: string;
  created_at: string;
}

export interface LedgerRecord {
  user_id?: string;
  type: string;
  amount: string;
  balance_before: string;
  balance_after: string;
  base_cost: string;
  rate_multiplier: string;
  remark: string;
  created_at: string;
}

export interface DailyStat {
  day: string;
  requests: number;
  successes: number;
  failures: number;
  total_tokens: number;
  base_cost: string;
  charge: string;
  gross_profit: string;
}

export interface ModelStat {
  model_id: string;
  requests: number;
  total_tokens: number;
  base_cost: string;
  charge: string;
}

export interface ModelChannelBinding {
  id: string;
  channel_id: string;
  channel_name: string;
  provider_type: string;
  base_url: string;
  upstream_model_name: string;
  status: string;
  created_at: string;
}

export interface ModelDetail {
  model: ModelConfig;
  channels: ModelChannelBinding[];
  stats: {
    requests: number;
    successes: number;
    failures: number;
    base_cost: string;
    charge: string;
    gross_profit: string;
  };
}

export interface ChannelDetailBinding {
  id: string;
  model_id: string;
  model_name: string;
  upstream_model_name: string;
  status: string;
  created_at: string;
}

export interface ChannelDetail {
  channel: Channel;
  models: ChannelDetailBinding[];
  stats: {
    requests: number;
    successes: number;
    failures: number;
    average_latency: string;
  };
}

export interface ReportRow {
  key: string;
  label: string;
  requests: number;
  successes: number;
  failures: number;
  base_cost: string;
  charge: string;
  gross_profit: string;
}

export interface AdminDashboard {
  today_requests: number;
  today_charge: string;
  today_base_cost: string;
  gross_profit: string;
  active_users: number;
  balance_total: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}

export interface UserDashboard {
  balance: string;
  today_charge: string;
  month_charge: string;
  frozen: string;
  available_models: number;
  today_requests: number;
}

export interface UserModelPrice {
  id: string;
  public_name: string;
  type: string;
  group: string;
  billing_mode: string;
  input_price_per_1k: string;
  output_price_per_1k: string;
  price_per_call: string;
  rate_multiplier: string;
  input_unit_price: string;
  output_unit_price: string;
  call_unit_price: string;
  status: string;
}
