import type { ApiResponse, HealthData } from "@/types/api";

/** Health check API - calls GET /health directly (no /api prefix) */
export async function getHealth(): Promise<ApiResponse<HealthData>> {
  try {
    const response = await fetch("/health");
    const data: ApiResponse<HealthData> = await response.json();

    if (!response.ok && data.code === 0) {
      data.code = response.status;
    }

    return data;
  } catch (error) {
    return {
      code: -1,
      message: error instanceof Error ? error.message : "Network error",
    };
  }
}
