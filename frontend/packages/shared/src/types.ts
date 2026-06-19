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
  rpm_limit?: number;
  concurrency_limit?: number;
  token_revoked_at?: string;
  created_at: string;
}

export interface APIKey {
  id: string;
  user_id: string;
  mask: string;
  name: string;
  status: string;
  allowed_endpoints: string[];
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
  cache_creation_price_per_1k: string;
  cache_read_price_per_1k: string;
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

export interface CleanupResult {
  table: string;
  deleted: number;
  started_at: string;
  ended_at: string;
  err?: string;
}

export interface CleanupHistoryEntry {
  id: string;
  started_at: string;
  ended_at: string;
  results: CleanupResult[];
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
  bound_count: number;
  last_latency_ms: number;
  last_success_at?: string;
}

export interface ChannelPreset {
  key: string;
  label: string;
  base_url: string;
  format: string;
  note: string;
}

export interface ChannelDetectResult {
  format: string;
  normalized_base_url: string;
  sample_models: string[];
  probe_url?: string;
}

export interface ProviderModel {
  id: string;
  type: string;
  owned: string;
}

export interface ChannelModelImportInput {
  upstream_name: string;
  public_name: string;
  type: string;
  billing_mode: string;
  input_price_per_1k: string;
  output_price_per_1k: string;
  price_per_call?: string;
  rate_multiplier: string;
  status?: string;
  sort_order?: number;
}

export interface ChannelModelImportResult {
  model_id: string;
  public_name: string;
  upstream_model_name: string;
  binding_id: string;
  created: boolean;
  bound: boolean;
}

export interface ChannelModelSyncResult {
  upstream_models: ProviderModel[];
  existing_bindings: ChannelDetailBinding[];
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
  expires_at?: string;
  created_at: string;
}

export interface RedeemRecord {
  id: string;
  user_id: string;
  username: string;
  amount: string;
  client_ip: string;
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
  total_users?: number;
  total_channels?: number;
  healthy_channels?: number;
  total_models?: number;
  enabled_models?: number;
  today_successes?: number;
  today_failures?: number;
  active_api_keys?: number;
  total_requests?: number;
}

export interface OpsSummary {
  rpm: number;
  tpm: number;
  requests_24h: number;
  error_rate_24h: string;
  p50_latency_ms: number;
  p95_latency_ms: number;
  avg_first_token_ms: number;
  channel_switches: number;
}

export interface OpsTrendPoint {
  bucket: string;
  requests: number;
  failures: number;
  total_tokens: number;
  charge: string;
  avg_latency_ms: number;
  p95_latency_ms: number;
}

export interface OpsChannelHealth {
  id: string;
  name: string;
  provider_type: string;
  status: string;
  health: string;
  fail_count: number;
  last_latency_ms: number;
  last_success_at?: string;
  last_error_at?: string;
  last_error_message: string;
  requests_24h: number;
  failures_24h: number;
  error_rate_24h: string;
  avg_latency_ms: number;
}

export interface OpsStatusBucket {
  status: string;
  count: number;
}

export interface OpsAlert {
  id: string;
  rule_key: string;
  severity: string;
  target_type: string;
  target_id: string;
  title: string;
  message: string;
  status: string;
  last_notified_at?: string;
  created_at: string;
  updated_at: string;
}

export interface OpsDashboard {
  summary: OpsSummary;
  trends: OpsTrendPoint[];
  channels: OpsChannelHealth[];
  statuses: OpsStatusBucket[];
  alerts: OpsAlert[];
}

export interface LoginResponse {
  token: string;
  user: User;
}

export interface PublicModel {
  id: string;
  public_name: string;
  type: string;
  group?: string;
  billing_mode: string;
  input_price_per_1m: string;
  output_price_per_1m: string;
  price_per_call?: string;
  currency: string;
}

export interface PublicSiteInfo {
  site_name: string;
  registration_enabled: boolean;
  registration_mode: string;
  contact_info: string;
  contact_email: string;
  site_logo_url: string;
  site_icp: string;
  site_police_beian: string;
  tos_url: string;
  privacy_url: string;
  brand_primary_color: string;
  captcha_enabled: boolean;
  captcha_provider: string;
  captcha_site_key: string;
  login_url: string;
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
  input_unit_price: string;
  output_unit_price: string;
  call_unit_price: string;
  status: string;
}
