import { apiPost } from "../api-client";
import { AUTH_ENDPOINTS } from "../api/endpoints";
import type { ApiResponse } from "../types/api.types";
import type {
  LoginCredentials,
  LoginResponse,
  RefreshTokenRequest,
  RefreshTokenResponse,
} from "../types/auth.types";

class AuthService {
  async login(credentials: LoginCredentials): Promise<LoginResponse> {
    const response = await apiPost(AUTH_ENDPOINTS.LOGIN, credentials, {
      skipAuth: true,
    });

    if (!response.ok) {
      const error = await response
        .json()
        .catch(() => ({ error: { message: "Login failed" } }));
      const message = error.error?.message || error.error || "Login failed";
      throw new Error(message);
    }

    const data: ApiResponse<LoginResponse> = await response.json();

    if (!data.data) {
      throw new Error("Invalid response format");
    }

    return data.data;
  }

  async logout(): Promise<boolean> {
    try {
      const response = await apiPost(AUTH_ENDPOINTS.LOGOUT);
      return response.ok;
    } catch (error) {
      console.error("Logout error:", error);
      return false;
    }
  }

  async refreshToken(
    refreshToken: string,
  ): Promise<RefreshTokenResponse | null> {
    try {
      const response = await apiPost(
        AUTH_ENDPOINTS.REFRESH,
        { refresh_token: refreshToken } as RefreshTokenRequest,
        { skipAuth: true },
      );

      if (!response.ok) {
        return null;
      }

      const data: ApiResponse<RefreshTokenResponse> = await response.json();

      if (!data.data) {
        return null;
      }

      return data.data;
    } catch (error) {
      console.error("Refresh token error:", error);
      return null;
    }
  }
}

export const authService = new AuthService();
export default authService;
