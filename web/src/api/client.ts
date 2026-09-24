import type { ApiResponse } from "@/types/api";

const BASE_URL = import.meta.env.VITE_API_BASE_URL || "/api";
const TIMEOUT = 10000;

/** Unified API client for Backend communication */
class ApiClient {
  private baseUrl: string;
  private timeout: number;

  constructor(baseUrl: string = BASE_URL, timeout: number = TIMEOUT) {
    this.baseUrl = baseUrl;
    this.timeout = timeout;
  }

  async request<T>(
    method: string,
    path: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    const url = `${this.baseUrl}${path}`;

    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.timeout);

    try {
      const response = await fetch(url, {
        method,
        ...options,
        signal: controller.signal,
        headers: {
          "Content-Type": "application/json",
          ...options.headers,
        },
      });

      clearTimeout(timeoutId);

      const data: ApiResponse<T> = await response.json();

      // If HTTP error but code is still 0 (success), use HTTP status as code
      if (!response.ok && data.code === 0) {
        data.code = response.status;
      }

      return data;
    } catch (error) {
      clearTimeout(timeoutId);

      if (error instanceof DOMException && error.name === "AbortError") {
        return {
          code: -1,
          message: "请求超时，请稍后重试",
        };
      }

      return {
        code: -1,
        message: error instanceof Error ? error.message : "网络异常，请检查连接",
      };
    }
  }

  get<T>(path: string): Promise<ApiResponse<T>> {
    return this.request<T>("GET", path);
  }

  post<T>(path: string, body?: unknown): Promise<ApiResponse<T>> {
    return this.request<T>("POST", path, {
      body: body ? JSON.stringify(body) : undefined,
    });
  }

  put<T>(path: string, body?: unknown): Promise<ApiResponse<T>> {
    return this.request<T>("PUT", path, {
      body: body ? JSON.stringify(body) : undefined,
    });
  }

  patch<T>(path: string, body?: unknown): Promise<ApiResponse<T>> {
    return this.request<T>("PATCH", path, {
      body: body ? JSON.stringify(body) : undefined,
    });
  }

  delete<T>(path: string): Promise<ApiResponse<T>> {
    return this.request<T>("DELETE", path);
  }
}

/**
 * Extract user-friendly error message from API response.
 * Priority: business code → backend message → HTTP status fallback → network/timeout fallback
 */
export function extractErrorMessage(resp: ApiResponse): string {
  // Network/timeout errors (code = -1)
  if (resp.code === -1) {
    return resp.message || "网络异常";
  }

  // Business error with message from backend
  if (resp.code !== 0 && resp.message) {
    return resp.message;
  }

  // HTTP status fallback (should not reach here if backend follows contract)
  if (resp.code >= 500) {
    return "服务端错误";
  }
  if (resp.code >= 400) {
    return "请求错误";
  }

  return "未知错误";
}

export const apiClient = new ApiClient();
