export interface CountdownConfig extends Record<string, unknown> {
  pre_game_countdown_seconds: number;
  reconnect_grace_period_seconds: number;
}

export interface GameConfig extends Record<string, unknown> {
  active_ttl_minutes: number;
  finished_ttl_minutes: number;
}

export interface RankDefinition {
  name: string;
  min_level: number;
}

export interface XPConfig extends Record<string, unknown> {
  base_xp: number;
  growth_rate: number;
  solo_multiplier: number;
  onevone_multiplier: number;
  multi_multiplier: number;
  tournament_multiplier: number;
  winner_bonus: number;
  ranks: RankDefinition[];
}

export interface ELOConfig extends Record<string, unknown> {
  k_factor_low_games: number;
  k_factor_high_games: number;
  k_factor_threshold: number;
  min_rating: number;
  max_rating: number;
}

export interface RateLimitConfig extends Record<string, unknown> {
  enabled: boolean;
  apply_to_admin: boolean;
  max_requests: number;
  window_seconds: number;
}

export interface WebSocketConfig extends Record<string, unknown> {
  write_wait_seconds: number;
  pong_wait_seconds: number;
  max_message_size_kb: number;
  allowed_origins: string[];
  reconnect_grace_period_seconds: number;
  message_rate_limit: number;
  message_rate_window_seconds: number;
}

export interface AvatarConfig extends Record<string, unknown> {
  max_file_size_mb: number;
  allowed_mime_types: string[];
  allowed_extensions: string[];
}

export interface SystemConfig extends Record<string, unknown> {
  user_cache_ttl_minutes: number;
  cleanup_interval_minutes: number;
  offline_delay_seconds: number;
  game_leave_delay_seconds: number;
  analytics_active_days: number;
  analytics_archive_days: number;
  dataset_check_enabled: boolean;
  dataset_check_cron: string;
  version_check_enabled: boolean;
  session_cleanup_enabled: boolean;
  session_cleanup_cron: string;
  game_cleanup_enabled: boolean;
  game_cleanup_cron: string;
}

export interface AuthConfig extends Record<string, unknown> {
  access_token_ttl_minutes: number;
  refresh_token_ttl_days: number;
  session_ttl_days: number;
  max_concurrent_sessions: number;
  failed_login_attempts: number;
  login_lockout_minutes: number;
}
