/** Operation type */
export type OperationType = "QUERY_STATUS" | "SET_VOLUME" | "RESTART";

export const OPERATION_TYPE_LABELS: Record<OperationType, string> = {
  QUERY_STATUS: "状态查询",
  SET_VOLUME: "音量调节",
  RESTART: "重启设备",
};

/** Operation status */
export type OperationStatus = "PENDING" | "IN_PROGRESS" | "COMPLETED" | "FAILED" | "PARTIAL_SUCCESS";

export const OPERATION_STATUS_LABELS: Record<OperationStatus, string> = {
  PENDING: "待执行",
  IN_PROGRESS: "执行中",
  COMPLETED: "已完成",
  FAILED: "失败",
  PARTIAL_SUCCESS: "部分成功",
};

/** Execution status */
export type ExecutionStatus = "PENDING" | "RUNNING" | "SUCCESS" | "FAILED" | "RETRYING" | "SKIPPED";

export const EXECUTION_STATUS_LABELS: Record<ExecutionStatus, string> = {
  PENDING: "待执行",
  RUNNING: "执行中",
  SUCCESS: "成功",
  FAILED: "失败",
  RETRYING: "重试中",
  SKIPPED: "已跳过",
};

/** Operation entity */
export interface Operation {
  id: string;
  device_id?: string;
  type: OperationType;
  parameters?: Record<string, unknown>;
  status: OperationStatus;
  total_count: number;
  success_count: number;
  failed_count: number;
  skipped_count: number;
  created_at: string;
  updated_at: string;
  finished_at?: string;
}

/** Execution entity */
export interface Execution {
  id: string;
  operation_id: string;
  device_id: string;
  status: ExecutionStatus;
  retry_count: number;
  max_retries: number;
  request?: Record<string, unknown>;
  result?: Record<string, unknown>;
  error?: string;
  started_at?: string;
  finished_at?: string;
  created_at: string;
  updated_at: string;
}

/** Create operation request */
export interface CreateOperationRequest {
  device_id: string;
  type: OperationType;
  parameters?: Record<string, unknown>;
}

/** Create operation response */
export interface CreateOperationResponse {
  operation: Operation;
  execution: Execution;
}

/** Operation with executions response */
export interface OperationWithExecutionsResponse {
  operation: Operation;
  executions: Execution[];
}

/** Operation list response */
export interface OperationListResponse {
  items: Operation[];
  total: number;
  page: number;
  page_size: number;
}