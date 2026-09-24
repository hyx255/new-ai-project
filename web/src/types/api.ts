/** Unified API response structure from Backend */
export interface ApiResponse<T = unknown> {
  code: number;
  message: string;
  data?: T;
}

/** Health check response data */
export interface HealthData {
  status: "ok" | "degraded";
  database: "healthy" | "unhealthy" | "not connected";
  trace_id?: string;
}

/** Error codes matching Backend errors package */
export enum ErrorCode {
  Success = 0,
  SystemError = 1000,
  ValidationError = 1001,
  BusinessError = 1002,
  Unauthorized = 1003,
  Forbidden = 1004,
  NotFound = 1005,
  InternalError = 1999,
}

