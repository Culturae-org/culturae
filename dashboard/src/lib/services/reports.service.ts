import { apiGet } from "../api-client";
import { REPORTS_ENDPOINTS } from "../api/endpoints";
import type { PaginatedResponse } from "../types/api.types";
import type { Report, ReportStatus } from "../types/reports.types";
import { handleApiError } from "../utils/api-helpers";
import { BaseService } from "./base.service";

class ReportsService extends BaseService {
  async list(params?: {
    page?: number;
    limit?: number;
    status?: string;
  }): Promise<PaginatedResponse<Report>> {
    const url = this.buildPaginatedUrl(REPORTS_ENDPOINTS.LIST, { ...params });
    const response = await apiGet(url);
    if (!response.ok) await handleApiError(response, "Failed to fetch reports");
    const json = await response.json();
    const pagination = json.pagination ?? {};
    const limit = pagination.limit ?? json.limit ?? params?.limit ?? 20;
    const total = pagination.total ?? json.total ?? 0;
    const page = params?.page ?? 1;
    return {
      data: json.data ?? json.reports ?? [],
      page,
      limit,
      total,
      total_pages: Math.ceil(total / limit) || 1,
    };
  }

  async updateStatus(
    id: string,
    status: ReportStatus,
    notes?: string,
  ): Promise<void> {
    return this.patch<void>(REPORTS_ENDPOINTS.UPDATE_STATUS(id), {
      status,
      ...(notes !== undefined && { resolution_notes: notes }),
    });
  }

  async getById(id: string): Promise<Report> {
    return this.get<Report>(REPORTS_ENDPOINTS.GET(id));
  }
}

export const reportsService = new ReportsService();
export default reportsService;
