"use client";

interface ApiOptions extends RequestInit {
  skipAuth?: boolean;
}

let onAuthFailureFn: (() => void) | null = null;
let refreshingPromise: Promise<Response> | null = null;

export function configureApiClient(
  _refreshFn: () => Promise<boolean>,
  _clearFn: () => void,
  onAuthFailure?: () => void,
  _getAccessToken?: () => string | null,
) {
  onAuthFailureFn = onAuthFailure ?? null;
}

export async function apiClient(
  url: string,
  options: ApiOptions = {},
): Promise<Response> {
  const { skipAuth = false, ...fetchOptions } = options;

  const { headers: callerHeaders, ...restFetchOptions } = fetchOptions;
  const config: RequestInit = {
    credentials: "include",
    ...restFetchOptions,
    headers: {
      ...(callerHeaders as Record<string, string>),
    },
  };

  const body = config.body;
  const isFormData = body instanceof FormData;

  if (!isFormData) {
    config.headers = {
      "Content-Type": "application/json",
      ...(config.headers as Record<string, string>),
    };
  }

  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), 30000);
  config.signal = options.signal ?? controller.signal;

  try {
    const response = await fetch(url, config).finally(() =>
      clearTimeout(timeoutId),
    );

    if (response.status === 401 && !skipAuth) {
      if (!refreshingPromise) {
        refreshingPromise = fetch("/api/v1/auth/refresh", {
          method: "POST",
          credentials: "include",
        }).finally(() => {
          refreshingPromise = null;
        });
      }

      const refreshResponse = await refreshingPromise;

      if (refreshResponse.ok) {
        const retryController = new AbortController();
        const retryTimeoutId = setTimeout(() => retryController.abort(), 30000);
        return fetch(url, {
          ...config,
          signal: retryController.signal,
        }).finally(() => clearTimeout(retryTimeoutId));
      }

      if (onAuthFailureFn) {
        onAuthFailureFn();
      }

      return response;
    }

    return response;
  } catch (error) {
    console.error("API request failed:", error);
    throw error;
  }
}

export async function apiGet(url: string, options?: ApiOptions) {
  return apiClient(url, { ...options, method: "GET" });
}

export async function apiPost(
  url: string,
  data?: unknown,
  options?: ApiOptions,
) {
  return apiClient(url, {
    ...options,
    method: "POST",
    body: data ? JSON.stringify(data) : undefined,
  });
}

export async function apiPut(
  url: string,
  data?: unknown,
  options?: ApiOptions,
) {
  return apiClient(url, {
    ...options,
    method: "PUT",
    body: data ? JSON.stringify(data) : undefined,
  });
}

export async function apiPatch(
  url: string,
  data?: unknown,
  options?: ApiOptions,
) {
  return apiClient(url, {
    ...options,
    method: "PATCH",
    body: data ? JSON.stringify(data) : undefined,
  });
}

export async function apiDelete(url: string, options?: ApiOptions) {
  return apiClient(url, { ...options, method: "DELETE" });
}

export async function apiPostFormData(
  url: string,
  data: FormData,
  options?: ApiOptions,
) {
  return apiClient(url, {
    ...options,
    method: "POST",
    body: data,
    skipAuth: false,
  });
}

export async function apiClientBlob(url: string, options?: ApiOptions) {
  const response = await apiClient(url, {
    ...options,
    method: "POST",
  });
  if (!response.ok) {
    throw new Error("Failed to fetch blob");
  }
  return response.blob();
}
