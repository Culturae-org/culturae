import { DATASETS_ENDPOINTS } from "@/lib/api/endpoints";
import type { PaginatedResponse } from "@/lib/types/api.types";
import type {
  Continent,
  Country,
  DatasetUpdateInfo,
  GeographyDataset,
  GeographyDatasetStats,
  GeographyImportResult,
  Region,
} from "@/lib/types/geography.types";
import { BaseService } from "./base.service";

interface CountryListParams {
  page?: number;
  limit?: number;
  dataset_id?: string;
  search?: string;
  sort_by?: string;
  sort_order?: "asc" | "desc";
  q?: string;
}

interface RegionListParams extends CountryListParams {
  continent_code?: string;
}

interface ImportOptions {
  manifest_url: string;
  dataset_type: "geography";
  set_as_default?: boolean;
  force?: boolean;
}

class GeographyService extends BaseService {
  async getDatasets(activeOnly = false): Promise<GeographyDataset[]> {
    const params = { active_only: activeOnly ? "true" : "false" };
    const response = await this.getPaginated<GeographyDataset>(
      DATASETS_ENDPOINTS.LIST,
      params,
    );
    return response?.data ?? [];
  }

  async getDataset(id: string): Promise<GeographyDataset> {
    return this.get<GeographyDataset>(DATASETS_ENDPOINTS.GET(id));
  }

  async getDatasetBySlug(slug: string): Promise<GeographyDataset> {
    return this.get<GeographyDataset>(DATASETS_ENDPOINTS.GET_BY_SLUG(slug));
  }

  async getDefaultDataset(): Promise<GeographyDataset> {
    return this.get<GeographyDataset>(DATASETS_ENDPOINTS.GET_DEFAULT);
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

  async createDataset(
    datasetData: Partial<GeographyDataset>,
  ): Promise<GeographyDataset> {
    return this.post<GeographyDataset>(DATASETS_ENDPOINTS.IMPORT, {
      ...datasetData,
      dataset_type: "geography",
    });
  }

  async updateDataset(
    id: string,
    updates: Partial<GeographyDataset>,
  ): Promise<GeographyDataset> {
    return this.put<GeographyDataset>(DATASETS_ENDPOINTS.UPDATE(id), updates);
  }

  async listCountries(
    datasetId: string,
    params: Record<string, unknown> = {},
  ): Promise<PaginatedResponse<Country>> {
    const endpoint = DATASETS_ENDPOINTS.LIST_COUNTRIES(datasetId);
    return this.getPaginated<Country>(endpoint, params);
  }

  async updateCountry(
    datasetId: string,
    slug: string,
    updates: Partial<Country>,
  ): Promise<Country> {
    const endpoint = DATASETS_ENDPOINTS.UPDATE_COUNTRY(datasetId, slug);
    return this.put<Country>(endpoint, updates);
  }

  async searchCountries(
    datasetId: string,
    query: string,
    params: Record<string, unknown> = {},
  ): Promise<PaginatedResponse<Country>> {
    const endpoint = DATASETS_ENDPOINTS.SEARCH_COUNTRIES(datasetId);
    return this.getPaginated<Country>(endpoint, { ...params, search: query });
  }

  async getCountries(
    datasetId: string,
    params: Record<string, unknown> = {},
  ): Promise<PaginatedResponse<Country>> {
    return this.listCountries(datasetId, params);
  }

  async getCountryByCode(code: string): Promise<Country | null> {
    const response = await this.get<Country | null>(
      `${DATASETS_ENDPOINTS.GET}?code=${code}`,
    );
    return response;
  }

  async getCountriesByContinent(
    datasetId: string,
    continent: string,
  ): Promise<Country[]> {
    const endpoint = DATASETS_ENDPOINTS.COUNTRIES_BY_CONTINENT(
      datasetId,
      continent,
    );
    const response = await this.get<PaginatedResponse<Country>>(endpoint);
    return response?.data ?? [];
  }

  async getCountriesByRegion(
    datasetId: string,
    region: string,
  ): Promise<Country[]> {
    const endpoint = DATASETS_ENDPOINTS.COUNTRIES_BY_REGION(datasetId, region);
    const response = await this.get<PaginatedResponse<Country>>(endpoint);
    return response?.data ?? [];
  }

  async getContinents(datasetId: string): Promise<Continent[]> {
    const endpoint = DATASETS_ENDPOINTS.LIST_CONTINENTS(datasetId);
    const response = await this.get<Continent[]>(endpoint);
    return response ?? [];
  }

  async updateContinent(
    datasetId: string,
    slug: string,
    updates: Partial<Continent>,
  ): Promise<Continent> {
    const endpoint = DATASETS_ENDPOINTS.UPDATE_CONTINENT(datasetId, slug);
    return this.put<Continent>(endpoint, updates);
  }

  async getRegions(datasetId: string): Promise<Region[]> {
    const endpoint = DATASETS_ENDPOINTS.LIST_REGIONS(datasetId);
    const response = await this.get<Region[]>(endpoint);
    return response ?? [];
  }

  async updateRegion(
    datasetId: string,
    slug: string,
    updates: Partial<Region>,
  ): Promise<Region> {
    const endpoint = DATASETS_ENDPOINTS.UPDATE_REGION(datasetId, slug);
    return this.put<Region>(endpoint, updates);
  }

  async getDatasetStats(id: string): Promise<GeographyDatasetStats> {
    const response = await this.get<GeographyDatasetStats>(
      DATASETS_ENDPOINTS.GET_STATS(id),
    );
    return response;
  }

  async importDataset(
    importData: ImportOptions,
  ): Promise<GeographyImportResult> {
    return this.post<GeographyImportResult>(DATASETS_ENDPOINTS.IMPORT, {
      ...importData,
      dataset_type: "geography",
    });
  }

  async checkAllUpdates(): Promise<DatasetUpdateInfo[]> {
    return this.post<DatasetUpdateInfo[]>(DATASETS_ENDPOINTS.CHECK_UPDATES, {});
  }
}

export const geographyService = new GeographyService();
export default geographyService;
