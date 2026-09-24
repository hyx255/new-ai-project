import { apiClient } from "./client";
import type { ApiResponse } from "@/types/api";
import type {
  DeviceType,
  SimpleList,
  CreateDeviceTypeRequest,
  UpdateDeviceTypeRequest,
} from "@/types/device";

/** GET /api/device-types */
export async function listDeviceTypes(): Promise<
  ApiResponse<SimpleList<DeviceType>>
> {
  return apiClient.get<SimpleList<DeviceType>>("/device-types");
}

/** GET /api/device-types/:id */
export async function getDeviceType(
  id: string
): Promise<ApiResponse<DeviceType>> {
  return apiClient.get<DeviceType>(`/device-types/${id}`);
}

/** POST /api/device-types */
export async function createDeviceType(
  data: CreateDeviceTypeRequest
): Promise<ApiResponse<DeviceType>> {
  return apiClient.post<DeviceType>("/device-types", data);
}

/** PUT /api/device-types/:id */
export async function updateDeviceType(
  id: string,
  data: UpdateDeviceTypeRequest
): Promise<ApiResponse<DeviceType>> {
  return apiClient.put<DeviceType>(`/device-types/${id}`, data);
}

/** DELETE /api/device-types/:id */
export async function deleteDeviceType(
  id: string
): Promise<ApiResponse<{ message: string }>> {
  return apiClient.delete<{ message: string }>(`/device-types/${id}`);
}
