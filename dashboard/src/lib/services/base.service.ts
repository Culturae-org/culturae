import { apiDelete, apiGet, apiPatch, apiPost, apiPut } from "../api-client";
import type { PaginatedResponse } from "../types/api.types";
import { buildUrl, extractApiData, handleApiError } from "../utils/api-helpers";

export abstract class BaseService {
  protected async handleResponse<T>(
    response: Response,
    dataPath?: string,
  ): Promise<T> {
    if (!response.ok) {
      await handleApiError(response, "API request failed");
    }

    if (response.status === 204 || response.status === 205) {
      return undefined as T;
    }

    const contentLength = response.headers.get("content-length");
    if (contentLength === "0" || response.status === 304) {
      return undefined as T;
    }

    const data = await response.json();
    return extractApiData<T>(data, dataPath);
  }

  protected async get<T>(
    endpoint: string,
    params?: Record<
      string,
      string | number | boolean | (string | number | boolean)[]
    >,
    dataPath?: string,
  ): Promise<T> {
    const url = buildUrl(endpoint, params);
    const response = await apiGet(url);
    return this.handleResponse<T>(response, dataPath);
  }

  protected async post<T>(
    endpoint: string,
    data?: unknown,
    dataPath?: string,
  ): Promise<T> {
    const response = await apiPost(endpoint, data);
    return this.handleResponse<T>(response, dataPath);
  }

  protected async put<T>(
    endpoint: string,
    data?: unknown,
    dataPath?: string,
  ): Promise<T> {
    const response = await apiPut(endpoint, data);
    return this.handleResponse<T>(response, dataPath);
  }

  protected async patch<T>(
    endpoint: string,
    data?: unknown,
    dataPath?: string,
  ): Promise<T> {
    const response = await apiPatch(endpoint, data);
    return this.handleResponse<T>(response, dataPath);
  }

  protected async delete<T = void>(
    endpoint: string,
    data?: unknown,
    dataPath?: string,
  ): Promise<T> {
    const response = await apiDelete(
      endpoint,
      data ? { body: JSON.stringify(data) } : undefined,
    );
    return this.handleResponse<T>(response, dataPath);
  }

  protected buildPaginatedUrl(
    baseEndpoint: string,
    params: Record<string, unknown>,
  ): string {
    const queryParams: Record<
      string,
      string | number | boolean | (string | number | boolean)[]
    > = {};

    for (const [key, value] of Object.entries(params)) {
      if (value !== undefined && value !== null && value !== "") {
        queryParams[key] = value as
          | string
          | number
          | boolean
          | (string | number | boolean)[];
      }
    }

    return buildUrl(baseEndpoint, queryParams);
  }

  protected async getPaginated<T>(
    endpoint: string,
    params: Record<string, unknown>,
  ): Promise<PaginatedResponse<T>> {
    const url = this.buildPaginatedUrl(endpoint, params);
    const response = await apiGet(url);

    if (!response.ok) {
      await handleApiError(response, "API request failed");
    }

    const json = await response.json();
    const pag = json.pagination ?? {};
    const limit = pag.limit ?? (Number(params.limit) || 10);

    return {
      data: json.data ?? [],
      page:
        pag.offset !== undefined
          ? Math.floor(pag.offset / limit) + 1
          : Number(params.page) || 1,
      limit,
      total: pag.total ?? 0,
      total_pages: limit > 0 ? Math.ceil((pag.total ?? 0) / limit) : 0,
    };
  }
}
