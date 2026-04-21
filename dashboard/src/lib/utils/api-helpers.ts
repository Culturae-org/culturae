import { ApiError } from "../types/api.types";

export function buildQueryParams(
  params: Record<
    string,
    string | number | boolean | (string | number | boolean)[]
  >,
): URLSearchParams {
  const queryParams = new URLSearchParams();

  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== null && value !== "") {
      if (Array.isArray(value)) {
        for (const item of value) {
          queryParams.append(key, String(item));
        }
      } else {
        queryParams.append(key, String(value));
      }
    }
  }

  return queryParams;
}

export function buildUrl(
  baseUrl: string,
  params?: Record<
    string,
    string | number | boolean | (string | number | boolean)[]
  >,
): string {
  if (!params || Object.keys(params).length === 0) {
    return baseUrl;
  }

  const queryString = buildQueryParams(params).toString();
  return queryString ? `${baseUrl}?${queryString}` : baseUrl;
}

export function extractApiData<T>(response: unknown, dataPath?: string): T {
  const resp = response as Record<string, unknown>;

  if (dataPath) {
    const data = resp?.data as Record<string, unknown> | undefined;
    if (data) {
      return (data[dataPath] ?? response) as T;
    }
    return (resp[dataPath] ?? response) as T;
  }

  return (resp?.data ?? response) as T;
}

export function extractApiMessage(response: unknown): string | undefined {
  const resp = response as Record<string, unknown>;
  return resp?.message as string | undefined;
}

function parseErrorResponse(errorData: unknown): {
  message: string;
  code?: string;
} {
  const err = errorData as Record<string, unknown>;

  if (err.error && typeof err.error === "object") {
    const errorObj = err.error as Record<string, unknown>;
    return {
      message: (errorObj.message as string) || "An error occurred",
      code: errorObj.code as string,
    };
  }

  return {
    message:
      (err.error as string) ||
      (err.message as string) ||
      (err.detail as string) ||
      "An error occurred",
    code: err.code as string | undefined,
  };
}

export async function handleApiError(
  response: Response,
  defaultMessage: string,
): Promise<never> {
  let errorMessage = defaultMessage;
  let errorCode: string | undefined;

  try {
    const errorData = await response.json();
    const parsed = parseErrorResponse(errorData);
    errorMessage = parsed.message;
    errorCode = parsed.code;
  } catch {}

  throw new ApiError(errorMessage, response.status, errorCode);
}

export function validateResponse<T>(
  data: unknown,
  validator?: (data: unknown) => data is T,
): T {
  if (validator && !validator(data)) {
    throw new Error("Response data does not match expected type");
  }
  return data as T;
}

export function createPaginatedResponse<T>(
  data: T[],
  total: number,
  page: number,
  limit: number,
) {
  return {
    data,
    page,
    limit,
    total,
    total_pages: Math.ceil(total / limit),
  };
}

export const QueryBuilders = {
  pagination: (page?: number, limit?: number) => ({
    page: page || 1,
    limit: limit || 10,
  }),

  filters: (filters: Record<string, string | number | boolean>) => filters,

  search: (query?: string) => (query ? { query } : {}),

  combined: (params: {
    page?: number;
    limit?: number;
    query?: string;
    filters?: Record<string, string | number | boolean>;
  }) => ({
    ...QueryBuilders.pagination(params.page, params.limit),
    ...QueryBuilders.search(params.query),
    ...(params.filters || {}),
  }),
} as const;
