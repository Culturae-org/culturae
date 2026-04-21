export interface SystemMetrics {
  total_users: number;
  active_users: number;
  total_sessions: number;
  active_sessions: number;
  total_api_requests: number;
  error_rate: number;
  avg_response_time_ms: number;
}

export interface DatabaseDetails {
  idle_connections: number;
  in_use_connections: number;
  max_open_connections: number;
  open_connections: number;
  total_connections: number;
  wait_count: number;
  wait_duration: string;
}

export interface RedisDetails {
  cache_hit_rate: number;
  connected_clients: string;
  evicted_keys: string;
  expired_keys: string;
  keyspace_hits: string;
  keyspace_misses: string;
  total_commands: string;
  total_connections: string;
  uptime_seconds: string;
  used_memory_bytes: string;
  version: string;
}

export interface MinioDetails {
  bucket_name: string;
  location: string;
  total_objects: number;
  total_size_bytes: number;
  total_size_human: string;
}

export type ServiceDetails =
  | DatabaseDetails
  | RedisDetails
  | MinioDetails
  | Record<string, unknown>;

export interface ServiceStatus {
  service_name: string;
  status: "healthy" | "unhealthy" | "degraded";
  last_check?: string;
  response_time_ms?: number;
  error_msg?: string;
  details?: ServiceDetails;
}

export interface ServiceStatusResponse {
  database?: ServiceStatus;
  redis?: ServiceStatus;
  minio?: ServiceStatus;
  backend?: ServiceStatus;
}
