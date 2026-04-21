import type { UnifiedDataset } from "@/hooks/useDatasetsList";
import { DATASETS_ENDPOINTS } from "@/lib/api/endpoints";
import type { PaginatedResponse } from "@/lib/types/api.types";
import type {
  CreateDatasetRequest,
  DatasetHistoryItem,
  DatasetUpdateInfo,
  ImportDatasetRequest,
  ImportJob,
  ImportQuestionLog,
  ImportResult,
  ImportStats,
  QuestionDataset,
  UpdateDatasetRequest,
} from "@/lib/types/datasets.types";
import type { GeographyDataset } from "@/lib/types/geography.types";
import type { Question } from "@/lib/types/question.types";
import { BaseService } from "./base.service";

class DatasetsService extends BaseService {
  async getDatasets(activeOnly = false): Promise<QuestionDataset[]> {
    const params = { active_only: activeOnly ? "true" : "false" };
    const response = await this.get<UnifiedDataset[]>(
      DATASETS_ENDPOINTS.LIST,
      params,
    );
    return (response ?? []).filter(
      (d) => d.type === "questions",
    ) as unknown as QuestionDataset[];
  }

  async getGeographyDatasets(activeOnly = false): Promise<GeographyDataset[]> {
    const params = { active_only: activeOnly ? "true" : "false" };
    const response = await this.get<UnifiedDataset[]>(
      DATASETS_ENDPOINTS.LIST,
      params,
    );
    return (response ?? []).filter(
      (d) => d.type === "geography",
    ) as unknown as GeographyDataset[];
  }

  async getDatasetsPaginated(params: {
    limit: number;
    offset: number;
    active_only?: boolean;
    default_only?: boolean;
    dataset_type?: "questions" | "geography" | "all";
  }): Promise<PaginatedResponse<UnifiedDataset>> {
    return this.getPaginated<UnifiedDataset>(DATASETS_ENDPOINTS.LIST, {
      limit: params.limit,
      offset: params.offset,
      active_only: params.active_only,
      default_only: params.default_only,
      dataset_type: params.dataset_type,
    });
  }

  async getDataset(id: string): Promise<QuestionDataset> {
    return this.get<QuestionDataset>(DATASETS_ENDPOINTS.GET(id));
  }

  async getDatasetById(id: string): Promise<UnifiedDataset | null> {
    const response = await this.get<UnifiedDataset | null>(
      DATASETS_ENDPOINTS.GET(id),
    );
    return response;
  }

  async getDatasetBySlug(slug: string): Promise<UnifiedDataset> {
    return this.get<UnifiedDataset>(DATASETS_ENDPOINTS.GET_BY_SLUG(slug));
  }

  async getDefaultDataset(): Promise<QuestionDataset> {
    return this.get<QuestionDataset>(DATASETS_ENDPOINTS.GET_DEFAULT);
  }

  async createDataset(
    datasetData: CreateDatasetRequest,
  ): Promise<QuestionDataset> {
    return this.post<QuestionDataset>(DATASETS_ENDPOINTS.CREATE, datasetData);
  }

  async updateDataset(
    id: string,
    updates: UpdateDatasetRequest,
  ): Promise<QuestionDataset> {
    return this.put<QuestionDataset>(DATASETS_ENDPOINTS.UPDATE(id), updates);
  }

  async deleteDataset(id: string, force = false): Promise<void> {
    const url = force
      ? `${DATASETS_ENDPOINTS.DELETE(id)}?force=true`
      : DATASETS_ENDPOINTS.DELETE(id);
    return this.delete<void>(url);
  }

  async setDefaultDataset(id: string): Promise<void> {
    return this.post<void>(DATASETS_ENDPOINTS.SET_DEFAULT(id), {});
  }

  async checkForUpdates(manifestUrl: string): Promise<DatasetUpdateInfo> {
    return this.post<DatasetUpdateInfo>(DATASETS_ENDPOINTS.CHECK_UPDATES, {
      manifest_url: manifestUrl,
    });
  }

  async checkAllUpdates(): Promise<DatasetUpdateInfo[]> {
    return this.post<DatasetUpdateInfo[]>(DATASETS_ENDPOINTS.CHECK_UPDATES, {});
  }

  async importDataset(importData: ImportDatasetRequest): Promise<ImportResult> {
    return this.post<ImportResult>(DATASETS_ENDPOINTS.IMPORT, {
      ...importData,
      dataset_type: "questions",
    });
  }

  async getDatasetQuestions(
    id: string,
    params?: { page?: number; limit?: number },
  ): Promise<PaginatedResponse<Question>> {
    return this.getPaginated<Question>(DATASETS_ENDPOINTS.GET_QUESTIONS(id), {
      ...params,
    });
  }

  async updateDatasetStatistics(id: string): Promise<void> {
    return this.post<void>(DATASETS_ENDPOINTS.UPDATE_STATISTICS(id), {});
  }

  async getImportJobs(): Promise<ImportJob[]> {
    const response = await this.get<ImportJob[]>(
      DATASETS_ENDPOINTS.LIST.replace("/datasets", "/imports"),
    );
    return response ?? [];
  }

  async getHistory(params?: {
    page?: number;
    limit?: number;
    type?: string;
  }): Promise<PaginatedResponse<DatasetHistoryItem>> {
    return this.getPaginated<DatasetHistoryItem>(DATASETS_ENDPOINTS.HISTORY, {
      ...params,
    });
  }

  async getImportJob(id: string): Promise<ImportJob> {
    return this.get<ImportJob>(
      `${DATASETS_ENDPOINTS.LIST.replace("/datasets", "/imports")}/${id}`,
    );
  }

  async getImportJobLogs(
    id: string,
    params?: { limit?: number; offset?: number; action?: string },
  ): Promise<{ logs: ImportQuestionLog[]; total: number }> {
    const response = await this.get<{
      logs: ImportQuestionLog[];
      total: number;
    }>(
      `${DATASETS_ENDPOINTS.LIST.replace("/datasets", "/imports")}/${id}/logs`,
      params,
    );
    return { logs: response?.logs ?? [], total: response?.total ?? 0 };
  }

  async getImportStats(): Promise<ImportStats> {
    const response = await this.get<ImportStats>(
      `${DATASETS_ENDPOINTS.LIST.replace("/datasets", "/imports")}/stats`,
    );
    return response ?? {};
  }
}

export const datasetsService = new DatasetsService();
