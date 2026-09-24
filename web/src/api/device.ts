import { apiClient } from "./client";
import type { ApiResponse } from "@/types/api";
import type {
  Device,
  PaginatedList,
  CreateDeviceRequest,
  UpdateDeviceRequest,
  UpdateDeviceStatusRequest,
} from "@/types/device";

/** GET /api/devices */
export async function listDevices(
  page: number = 1,
  pageSize: number = 20,
  search?: string
): Promise<ApiResponse<PaginatedList<Device>>> {
  const params = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
  });
  if (search) {
    params.set("search", search);
  }
  return apiClient.get<PaginatedList<Device>>(`/devices?${params.toString()}`);
}

/** GET /api/devices/:id */
export async function getDevice(id: string): Promise<ApiResponse<Device>> {
  return apiClient.get<Device>(`/devices/${id}`);
}

/** POST /api/devices */
export async function createDevice(
  data: CreateDeviceRequest
): Promise<ApiResponse<Device>> {
  return apiClient.post<Device>("/devices", data);
}

/** PUT /api/devices/:id */
export async function updateDevice(
  id: string,
  data: UpdateDeviceRequest
): Promise<ApiResponse<Device>> {
  return apiClient.put<Device>(`/devices/${id}`, data);
}

/** PATCH /api/devices/:id/status */
export async function updateDeviceStatus(
  id: string,
  data: UpdateDeviceStatusRequest
): Promise<ApiResponse<Device>> {
  return apiClient.patch<Device>(`/devices/${id}/status`, data);
}
