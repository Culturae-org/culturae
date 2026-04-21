import { apiPostFormData } from "../api-client";
import { AVATAR_ENDPOINTS, USERS_ENDPOINTS } from "../api/endpoints";
import type { PaginatedResponse, UsersQueryParams } from "../types/api.types";
import type {
  AdminUser,
  ConnectionLog,
  Session,
  UserBanRequest,
  UserCreateData,
  UserPasswordUpdate,
  UserStatusUpdate,
  UserUpdateData,
} from "../types/user.types";
import { BaseService } from "./base.service";

class UsersService extends BaseService {
  async getUsers(
    params?: UsersQueryParams,
  ): Promise<PaginatedResponse<AdminUser>> {
    if (params?.query?.trim()) {
      return this.getPaginated<AdminUser>(USERS_ENDPOINTS.SEARCH, {
        ...params,
      });
    }
    return this.getPaginated<AdminUser>(USERS_ENDPOINTS.LIST, { ...params });
  }

  async getUserById(id: string): Promise<AdminUser> {
    return this.get<AdminUser>(USERS_ENDPOINTS.GET(id));
  }

  async createUser(userData: UserCreateData): Promise<AdminUser> {
    return this.post<AdminUser>(USERS_ENDPOINTS.CREATE, userData);
  }

  async updateUser(id: string, updates: UserUpdateData): Promise<AdminUser> {
    return this.put<AdminUser>(USERS_ENDPOINTS.UPDATE(id), updates);
  }

  async updateUserPassword(
    id: string,
    passwordData: UserPasswordUpdate,
  ): Promise<void> {
    return this.put<void>(USERS_ENDPOINTS.UPDATE_PASSWORD(id), passwordData);
  }

  async updateUserStatus(
    id: string,
    statusData: UserStatusUpdate,
  ): Promise<AdminUser> {
    return this.patch<AdminUser>(USERS_ENDPOINTS.UPDATE_STATUS(id), statusData);
  }

  async banUser(id: string, banData: UserBanRequest): Promise<AdminUser> {
    return this.post<AdminUser>(USERS_ENDPOINTS.BAN(id), banData);
  }

  async unbanUser(id: string): Promise<AdminUser> {
    return this.post<AdminUser>(USERS_ENDPOINTS.UNBAN(id), {});
  }

  async deactivateUser(id: string): Promise<void> {
    return this.patch<void>(USERS_ENDPOINTS.DEACTIVATE(id), {});
  }

  async deleteUser(id: string): Promise<void> {
    return this.delete<void>(USERS_ENDPOINTS.DELETE(id));
  }

  async getUserCount(): Promise<number> {
    return this.get<number>(USERS_ENDPOINTS.COUNT);
  }

  async getLevelStats(): Promise<Record<string, number>> {
    return this.get<Record<string, number>>(USERS_ENDPOINTS.LEVEL_STATS);
  }

  async getRoleStats(): Promise<Record<string, number>> {
    return this.get<Record<string, number>>(USERS_ENDPOINTS.ROLE_STATS);
  }

  async getCreationDates(
    startDate?: string,
    endDate?: string,
  ): Promise<string[]> {
    const params: Record<string, string> = {};
    if (startDate) params.start_date = startDate;
    if (endDate) params.end_date = endDate;
    return this.get<string[]>(USERS_ENDPOINTS.CREATION_DATES, params);
  }

  async searchUsers(
    params?: UsersQueryParams,
  ): Promise<PaginatedResponse<AdminUser>> {
    return this.getPaginated<AdminUser>(USERS_ENDPOINTS.SEARCH, { ...params });
  }

  async getUserConnectionLogs(
    userId: string,
    params?: { page?: number; limit?: number },
  ): Promise<{
    logs: ConnectionLog[];
    total_pages: number;
    total_count: number;
  }> {
    return this.get<{
      logs: ConnectionLog[];
      total_pages: number;
      total_count: number;
    }>(`${USERS_ENDPOINTS.LIST}/${userId}/connection-logs`, params);
  }

  async getUserSessions(
    userId: string,
    params?: { page?: number; limit?: number },
  ): Promise<{
    sessions: Session[];
    total_pages: number;
    total_count: number;
  }> {
    return this.get<{
      sessions: Session[];
      total_pages: number;
      total_count: number;
    }>(`${USERS_ENDPOINTS.LIST}/${userId}/sessions`, params);
  }

  async getUserActionLogs(
    userId: string,
    params?: { page?: number; limit?: number },
  ): Promise<{ logs: unknown[]; total_pages: number; total_count: number }> {
    return this.get<{
      logs: unknown[];
      total_pages: number;
      total_count: number;
    }>(`${USERS_ENDPOINTS.LIST}/${userId}/action-logs`, params);
  }

  async uploadAvatar(userId: string, file: File): Promise<void> {
    const formData = new FormData();
    formData.append("avatar", file);
    const response = await apiPostFormData(
      AVATAR_ENDPOINTS.UPLOAD(userId),
      formData,
    );
    if (!response.ok) {
      let errorMessage = "Failed to upload avatar";
      try {
        const errorData = await response.json();
        if (errorData.error?.message) errorMessage = errorData.error.message;
      } catch {
        /* use default */
      }
      throw new Error(errorMessage);
    }
  }

  async deleteAvatar(userId: string): Promise<void> {
    return this.delete<void>(AVATAR_ENDPOINTS.DELETE(userId));
  }

  getAvatarUrl(userId: string): string {
    return AVATAR_ENDPOINTS.GET(userId);
  }
}

export const usersService = new UsersService();
export default usersService;
