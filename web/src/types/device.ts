/** Device capabilities */
export type Capability = "QUERY_STATUS" | "SET_VOLUME" | "RESTART";

export const ALL_CAPABILITIES: Capability[] = [
  "QUERY_STATUS",
  "SET_VOLUME",
  "RESTART",
];

export const CAPABILITY_LABELS: Record<Capability, string> = {
  QUERY_STATUS: "状态查询",
  SET_VOLUME: "音量调节",
  RESTART: "重启",
};

/** Device status */
export type DeviceStatus = "REGISTERED" | "ACTIVE" | "OFFLINE" | "DISABLED";

export const STATUS_LABELS: Record<DeviceStatus, string> = {
  REGISTERED: "已注册",
  ACTIVE: "在线",
  OFFLINE: "离线",
  DISABLED: "已禁用",
};

/** DeviceType entity */
export interface DeviceType {
  id: string;
  name: string;
  vendor: string;
  model: string;
  description: string;
  capabilities: Capability[];
  created_at: string;
  updated_at: string;
}

/** Device entity */
export interface Device {
  id: string;
  name: string;
  device_type_id: string;
  address: string;
  status: DeviceStatus;
  last_online_at: string | null;
  created_at: string;
  updated_at: string;
  device_type?: DeviceType;
}

/** Paginated list response */
export interface PaginatedList<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
}

/** Simple list response (no pagination) */
export interface SimpleList<T> {
  items: T[];
  total: number;
}

/** Create DeviceType request */
export interface CreateDeviceTypeRequest {
  name: string;
  vendor: string;
  model: string;
  description: string;
  capabilities: Capability[];
}

/** Update DeviceType request */
export interface UpdateDeviceTypeRequest {
  name: string;
  vendor: string;
  model: string;
  description: string;
  capabilities: Capability[];
}

/** Create Device request */
export interface CreateDeviceRequest {
  name: string;
  device_type_id: string;
  address: string;
}

/** Update Device request */
export interface UpdateDeviceRequest {
  name: string;
  device_type_id: string;
  address: string;
}

/** Update Device status request */
export interface UpdateDeviceStatusRequest {
  action: "enable" | "disable";
}
