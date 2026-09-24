import type { Operation, Execution } from "./operation";

// Re-export existing types for convenience
export type { Operation, Execution };

/** Batch progress response */
export interface BatchProgress {
  operation_id: string;
  status: string;
  total_count: number;
  success_count: number;
  failed_count: number;
  skipped_count: number;
  running_count: number;
  is_complete: boolean;
}

/** Create batch operation request */
export interface CreateBatchOperationRequest {
  device_ids: string[];
  type: "QUERY_STATUS" | "SET_VOLUME" | "RESTART";
  parameters?: Record<string, unknown>;
}

/** Create batch operation response */
export interface CreateBatchOperationResponse {
  operation: Operation;
  executions: Execution[];
}

/** Batch operation detail response */
export interface BatchOperationDetailResponse {
  operation: Operation;
  executions: Execution[];
}