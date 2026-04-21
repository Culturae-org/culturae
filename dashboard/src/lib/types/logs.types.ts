export interface UserActionLog {
  ID: string;
  UserID: string;
  Username: string;
  Action: string;
  Resource: string;
  ResourceID?: string;
  IPAddress?: string;
  UserAgent?: string;
  Details?: Record<string, unknown>;
  IsSuccess: boolean;
  ErrorMsg?: string;
  CreatedAt: string;
}

export interface AdminActionLog {
  ID: string;
  AdminID: string;
  AdminName: string;
  Action: string;
  Resource: string;
  ResourceID?: string;
  IPAddress?: string;
  UserAgent?: string;
  Details?: Record<string, unknown>;
  IsSuccess: boolean;
  ErrorMsg?: string;
  CreatedAt: string;
}

export interface APIRequestLog {
  ID: string;
  Method: string;
  Path: string;
  StatusCode: number;
  UserID?: string;
  IPAddress?: string;
  UserAgent?: string;
  RequestSize: number;
  ResponseSize: number;
  Duration: number;
  IsError: boolean;
  ErrorMsg?: string;
  CreatedAt: string;
}

export interface LogsQueryParams
  extends Record<
    string,
    string | number | boolean | (string | number | boolean)[] | undefined
  > {
  page?: number;
  limit?: number;
  user_id?: string;
  start_date?: string;
  end_date?: string;
  action?: string;
  method?: string;
  status_code?: number;
}

export interface APIRequestStats {
  total_requests: number;
  error_rate: number;
  requests_by_status: Record<string, number>;
  requests_by_method: Record<string, number>;
  requests_by_path: Record<string, number>;
  avg_response_time_ms: number;
  daily_average: number;
}

export interface AdminActionStats {
  total_actions: number;
  success_rate: number;
  actions_by_type: Record<string, number>;
  actions_by_resource: Record<string, number>;
  actions_by_admin: Record<string, number>;
  daily_average: number;
}

export interface UserActionStats {
  total_actions: number;
  actions_by_type: Record<string, number>;
  top_users: Array<{
    user_id: string;
    username: string;
    action_count: number;
  }>;
}
