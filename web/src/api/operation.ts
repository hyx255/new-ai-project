import { apiClient } from "./client";
import type { ApiResponse } from "@/types/api";
import type {
  CreateOperationRequest,
  CreateOperationResponse,
  OperationWithExecutionsResponse,
  OperationListResponse,
} from "@/types/operation";

/** POST /api/operations */
export async function createOperation(
  data: CreateOperationRequest
): Promise<ApiResponse<CreateOperationResponse>> {
  return apiClient.post<CreateOperationResponse>("/operations", data);
}

/** GET /api/operations/:id */
export async function getOperation(
  id: string
): Promise<ApiResponse<OperationWithExecutionsResponse>> {
  return apiClient.get<OperationWithExecutionsResponse>(`/operations/${id}`);
}

/** GET /api/devices/:id/operations */
export async function listDeviceOperations(
  deviceId: string,
  page: number = 1,
  pageSize: number = 20
): Promise<ApiResponse<OperationListResponse>> {
  return apiClient.get<OperationListResponse>(
    `/devices/${deviceId}/operations?page=${page}&page_size=${pageSize}`
  );
}