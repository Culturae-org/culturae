export interface UserProfile {
  username: string;
  email: string;
  public_id: string;
  role: string;
  level: number;
  rank: string;
  account_status: string;
  banned_until?: string;
  ban_reason?: string;
  has_avatar: boolean;
  created_at: string;
  updated_at: string;
  language: string;
  bio?: string;
  status: string;
  is_online: boolean;
  last_seen_at?: string;
  experience: number;
  elo_rating: number;
  elo_games_played: number;
  is_profile_public: boolean;
  show_online_status: boolean;
  allow_friend_requests: boolean;
  allow_party_invites: boolean;
  game_stats?: UserGameStats;
}

export interface AdminUser {
  id: string;
  username: string;
  email: string;
  public_id: string;
  role: string;
  level: number;
  rank: string;
  account_status: string;
  banned_until?: string;
  ban_reason?: string;
  has_avatar: boolean;
  created_at: string;
  updated_at: string;
  language: string;
  bio?: string;
  status: string;
  is_online: boolean;
  last_seen_at?: string;
  current_game_id?: string;
  experience: number;
  elo_rating: number;
  elo_games_played: number;
  is_profile_public: boolean;
  show_online_status: boolean;
  allow_friend_requests: boolean;
  allow_party_invites: boolean;
  game_stats?: UserGameStats;
}

export interface UserGameStats {
  total_games: number;
  games_won: number;
  games_lost: number;
  games_drawn: number;
  day_streak: number;
  best_day_streak: number;
  total_score: number;
  average_score: number;
  play_time: number;
  last_game_at?: string;
  updated_at: string;
}

export interface UserUpdateData {
  email?: string;
  username?: string;
  role?: string;
  account_status?: string;
  language?: string;
  bio?: string;
  is_profile_public?: boolean;
  show_online_status?: boolean;
  allow_friend_requests?: boolean;
  allow_party_invites?: boolean;
}

export interface UserCreateData {
  username: string;
  email: string;
  password: string;
  role?: string;
  account_status?: string;
  language?: string;
}

export interface UserPasswordUpdate {
  password: string;
}

export interface UserStatusUpdate {
  account_status: string;
}

export interface UserBanRequest {
  duration: string;
  reason?: string;
}

export interface UserLevelStats {
  level: string;
  count: number;
  percentage: number;
}

export interface UserRoleStats {
  role: string;
  count: number;
  percentage: number;
}

export interface UserCreationDate {
  date: string;
  count: number;
}

export interface UserSearchResult {
  id: string;
  username: string;
  email: string;
}

export interface ConnectionLog {
  id: string;
  user_id: string;
  session_id?: string;
  ip_address: string;
  user_agent: string;
  device_info?: Record<string, unknown>;
  location?: string;
  is_success: boolean;
  failure_reason?: string;
  created_at: string;
}

export interface Session {
  id: string;
  created: string;
  created_at: string;
  expires_at: string;
  last_used: string;
  ip_address: string;
  user_agent: string;
  device_fingerprint: string;
  is_active: boolean;
  is_revoked: boolean;
  revoked_at?: string;
  revoked_reason: string;
  variables?: Record<string, unknown>;
  updated_at: string;
}
