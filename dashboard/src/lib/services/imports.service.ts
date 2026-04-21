import { IMPORTS_ENDPOINTS } from "../api/endpoints";
import type { PaginatedResponse } from "../types/api.types";
import type { ImportJob, ImportQuestionLog } from "../types/datasets.types";
import { BaseService } from "./base.service";

interface ImportsQueryParams {
  page?: number;
  limit?: number;
}

interface ImportLogsParams {
  page?: number;
  limit?: number;
  action?: string;
}

class ImportsService extends BaseService {
  async getImports(
    params?: ImportsQueryParams,
  ): Promise<PaginatedResponse<ImportJob>> {
    return this.getPaginated<ImportJob>(IMPORTS_ENDPOINTS.LIST, {
      page: params?.page ?? 1,
      limit: params?.limit ?? 20,
    });
  }

  async getImportById(id: string): Promise<ImportJob> {
    return this.get<ImportJob>(IMPORTS_ENDPOINTS.GET(id));
  }

  async getLogs(
    id: string,
    params?: ImportLogsParams,
  ): Promise<{ logs: ImportQuestionLog[]; total: number }> {
    const queryParams: Record<string, string | number> = {
      page: params?.page ?? 1,
      limit: params?.limit ?? 50,
    };
    if (params?.action && params.action !== "all")
      queryParams.action = params.action;
    const response = await this.get<{
      data?: ImportQuestionLog[];
      logs?: ImportQuestionLog[];
      pagination?: { total: number };
      total?: number;
    }>(IMPORTS_ENDPOINTS.LOGS(id), queryParams);
    const logs = Array.isArray(response?.data)
      ? response.data
      : Array.isArray(response?.logs)
        ? response.logs
        : [];
    const total = response?.pagination?.total ?? response?.total ?? 0;
    return { logs, total };
  }
}

export const importsService = new ImportsService();
export default importsService;
