import { LOGS_ENDPOINTS } from "../api/endpoints";
import type { PaginatedResponse } from "../types/api.types";
import type {
  APIRequestStats,
  AdminActionLog,
  AdminActionStats,
  LogsQueryParams,
  UserActionLog,
  UserActionStats,
} from "../types/logs.types";
import type {
  ServiceStatusResponse,
  SystemMetrics,
} from "../types/stats.types";
import type { ConnectionLog, Session } from "../types/user.types";
import { BaseService } from "./base.service";

class LogsService extends BaseService {
  async getUserConnectionLogs(
    userId: string,
    params?: LogsQueryParams,
  ): Promise<ConnectionLog[]> {
    return this.get<ConnectionLog[]>(
      LOGS_ENDPOINTS.CONNECTIONS(userId),
      params as
        | Record<
            string,
            string | number | boolean | (string | number | boolean)[]
          >
        | undefined,
    );
  }

  async getUserActionLogs(
    userId: string,
    params?: LogsQueryParams,
  ): Promise<UserActionLog[]> {
    return this.get<UserActionLog[]>(
      LOGS_ENDPOINTS.USER_ACTIONS(userId),
      params as
        | Record<
            string,
            string | number | boolean | (string | number | boolean)[]
          >
        | undefined,
    );
  }

  async getUserActiveSessions(userId: string): Promise<Session[]> {
    return this.get<Session[]>(LOGS_ENDPOINTS.ACTIVE_SESSIONS(userId));
  }

  async getAllUserActionLogs(
    params?: LogsQueryParams,
  ): Promise<PaginatedResponse<UserActionLog>> {
    return this.getPaginated<UserActionLog>(LOGS_ENDPOINTS.ALL_USER_ACTIONS, {
      ...params,
    });
  }

  async getAdminActionLogs(
    params?: LogsQueryParams,
  ): Promise<PaginatedResponse<AdminActionLog>> {
    return this.getPaginated<AdminActionLog>(LOGS_ENDPOINTS.ADMIN_ACTIONS, {
      ...params,
    });
  }

  async getAllConnectionLogs(
    params?: LogsQueryParams,
  ): Promise<PaginatedResponse<ConnectionLog>> {
    return this.getPaginated<ConnectionLog>(LOGS_ENDPOINTS.ALL_CONNECTIONS, {
      ...params,
    });
  }

  async getAPIRequestStats(params?: {
    start_date?: string;
    end_date?: string;
  }): Promise<APIRequestStats> {
    return this.get<APIRequestStats>(
      LOGS_ENDPOINTS.API_REQUEST_STATS,
      params as Record<string, string | number | boolean>,
    );
  }

  async getAdminActionStats(): Promise<AdminActionStats> {
    return this.get<AdminActionStats>(LOGS_ENDPOINTS.ADMIN_STATS);
  }

  async getUserActionStats(): Promise<UserActionStats> {
    return this.get<UserActionStats>(LOGS_ENDPOINTS.USER_STATS);
  }

  async getSystemMetrics(): Promise<SystemMetrics> {
    return this.get<SystemMetrics>(LOGS_ENDPOINTS.SYSTEM_METRICS);
  }

  async getServiceStatus(): Promise<ServiceStatusResponse> {
    const result = await this.get<{
      services: Array<{
        service_name: string;
        status: string;
        last_check?: string;
        response_time_ms?: number;
        error_msg?: string;
        details?: Record<string, unknown>;
      }>;
    }>(LOGS_ENDPOINTS.SERVICE_STATUS);

    const services = result?.services ?? [];
    const statusObj: ServiceStatusResponse = {};
    for (const service of services) {
      statusObj[service.service_name as keyof ServiceStatusResponse] = {
        service_name: service.service_name,
        status: service.status as "healthy" | "unhealthy" | "degraded",
        last_check: service.last_check,
        response_time_ms: service.response_time_ms,
        error_msg: service.error_msg,
        details: service.details,
      };
    }
    return statusObj;
  }

  async getAPIRequestTimestamps(params?: {
    method?: string;
    status_code?: string;
    start_date?: string;
    end_date?: string;
  }): Promise<string[]> {
    return this.get<string[]>(
      LOGS_ENDPOINTS.API_REQUEST_TIMESTAMPS,
      params as Record<string, string | number | boolean>,
    );
  }
}

export const logsService = new LogsService();
export default logsService;
