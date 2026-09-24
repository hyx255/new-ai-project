import { apiClient } from "./client";
import type { ApiResponse } from "@/types/api";
import type {
  CreateBatchOperationRequest,
  CreateBatchOperationResponse,
  BatchOperationDetailResponse,
  BatchProgress,
} from "@/types/batch-operation";

/** POST /api/batch-operations */
export async function createBatchOperation(
  data: CreateBatchOperationRequest
): Promise<ApiResponse<CreateBatchOperationResponse>> {
  return apiClient.post<CreateBatchOperationResponse>("/batch-operations", data);
}

/** GET /api/batch-operations/:id */
export async function getBatchOperation(
  id: string
): Promise<ApiResponse<BatchOperationDetailResponse>> {
  return apiClient.get<BatchOperationDetailResponse>(`/batch-operations/${id}`);
}

/** GET /api/batch-operations/:id/progress */
export async function getBatchProgress(
  id: string
): Promise<ApiResponse<BatchProgress>> {
  return apiClient.get<BatchProgress>(`/batch-operations/${id}/progress`);
}