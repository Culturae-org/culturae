export type SystemMetrics = {
  total_users: number;
  active_users: number;
  total_sessions: number;
  active_sessions: number;
  total_api_requests: number;
  error_rate: number;
  avg_response_time_ms: number;
};

export type PodInfo = {
  pod_id: string;
  pod_type: "main" | "headless";
  status: "healthy" | "degraded" | "offline";
  is_current: boolean;
  connected_clients: number;
  online_users: number;
  active_games: number;
  last_heartbeat: string;
  started_at: string;
};

export type PodsMeta = {
  total_pods: number;
  main_pods: number;
  headless_pods: number;
  total_clients: number;
  total_users: number;
};

export type PodsDiscovery = {
  pods: PodInfo[];
  meta: PodsMeta;
};
