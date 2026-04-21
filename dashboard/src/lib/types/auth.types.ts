import type { AdminUser } from "./user.types";

export interface LoginCredentials {
  identifier: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  refresh_token: string;
  token_expires_at?: number;
  token_type?: string;
  user: AdminUser;
}

export interface RefreshTokenRequest {
  refresh_token: string;
}

export interface RefreshTokenResponse {
  token: string;
  refresh_token?: string;
  token_expires_at?: number;
  token_type?: string;
  expires_at?: number;
}

export interface SessionData {
  refreshToken: string | null;
  accessToken: string | null;
  expiresAt: string | null;
  user?: AdminUser | null;
}
