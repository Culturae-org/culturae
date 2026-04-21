export const API_BASE = "/api/v1";

export const AUTH_ENDPOINTS = {
  LOGIN: `${API_BASE}/auth/login-admin`,
  LOGOUT: `${API_BASE}/auth/logout`,
  REFRESH: `${API_BASE}/auth/refresh`,
} as const;

export const REALTIME_ENDPOINT = "/api/v1/realtime";

export const PROFILE_ENDPOINTS = {
  GET: `${API_BASE}/profile`,
  UPDATE: `${API_BASE}/profile`,
  REGENERATE_ID: `${API_BASE}/profile/regenerate-id`,
} as const;

export const ADMIN_PROFILE_ENDPOINTS = {
  ME: `${API_BASE}/admin/me`,
} as const;

export const USERS_ENDPOINTS = {
  LIST: `${API_BASE}/admin/users`,
  GET: (id: string) => `${API_BASE}/admin/users/${id}`,
  CREATE: `${API_BASE}/admin/users`,
  UPDATE: (id: string) => `${API_BASE}/admin/users/${id}`,
  UPDATE_PASSWORD: (id: string) => `${API_BASE}/admin/users/${id}/password`,
  UPDATE_STATUS: (id: string) => `${API_BASE}/admin/users/${id}/status`,
  DEACTIVATE: (id: string) => `${API_BASE}/admin/users/${id}/deactivate`,
  DELETE: (id: string) => `${API_BASE}/admin/users/${id}`,
  REGENERATE_PUBLIC_ID: (id: string) =>
    `${API_BASE}/admin/users/${id}/regenerate-id`,
  BAN: (id: string) => `${API_BASE}/admin/users/${id}/ban`,
  UNBAN: (id: string) => `${API_BASE}/admin/users/${id}/unban`,

  COUNT: `${API_BASE}/admin/users/count`,
  COUNT_ONLINE: `${API_BASE}/admin/users/count/online`,
  COUNT_ACTIVE_WEEKLY: `${API_BASE}/admin/users/count/active/weekly`,
  LEVEL_STATS: `${API_BASE}/admin/users/level-stats`,
  ROLE_STATS: `${API_BASE}/admin/users/role-stats`,
  CREATION_DATES: `${API_BASE}/admin/users/creation-dates`,
  SEARCH: `${API_BASE}/admin/users/search`,
} as const;

export const AVATAR_ENDPOINTS = {
  UPLOAD: (userId: string) => `${API_BASE}/admin/avatar/${userId}`,
  DELETE: (userId: string) => `${API_BASE}/admin/avatar/${userId}`,
  GET: (userId: string, timestamp?: number) => {
    const url = `${API_BASE}/admin/avatar/${userId}`;
    return timestamp ? `${url}?t=${timestamp}` : url;
  },
} as const;

export const LOGS_ENDPOINTS = {
  CONNECTIONS: (userId: string) =>
    `${API_BASE}/admin/logs/connections/${userId}`,
  USER_ACTIONS: (userId: string) =>
    `${API_BASE}/admin/logs/user-actions/${userId}`,
  ACTIVE_SESSIONS: (userId: string) =>
    `${API_BASE}/admin/logs/active-sessions/${userId}`,

  ALL_USER_ACTIONS: `${API_BASE}/admin/logs/user-actions`,
  ADMIN_ACTIONS: `${API_BASE}/admin/logs/admin-actions`,
  ALL_CONNECTIONS: `${API_BASE}/admin/logs/connections`,

  SYSTEM_METRICS: `${API_BASE}/admin/logs/system-metrics`,
  SERVICE_STATUS: `${API_BASE}/admin/logs/service-status`,

  API_REQUEST_STATS: `${API_BASE}/admin/logs/api-requests/stats`,
  API_REQUEST_TIMESTAMPS: `${API_BASE}/admin/logs/api-requests/timestamps`,
  ADMIN_STATS: `${API_BASE}/admin/logs/admin-actions/stats`,
  USER_STATS: `${API_BASE}/admin/logs/user-actions/stats`,
} as const;

export const QUESTIONS_ENDPOINTS = {
  LIST: `${API_BASE}/admin/questions`,
  GET: (id: string) => `${API_BASE}/admin/questions/${id}`,
  GET_BY_SLUG: (slug: string) => `${API_BASE}/admin/questions/slug/${slug}`,
  CREATE: `${API_BASE}/admin/questions`,
  UPDATE: (id: string) => `${API_BASE}/admin/questions/${id}`,
  DELETE: (id: string) => `${API_BASE}/admin/questions/${id}`,
  EXPORT: `${API_BASE}/admin/questions/export`,
} as const;

export const IMPORTS_ENDPOINTS = {
  LIST: `${API_BASE}/admin/imports`,
  STATS: `${API_BASE}/admin/imports/stats`,
  GET: (id: string) => `${API_BASE}/admin/imports/${id}`,
  LOGS: (id: string) => `${API_BASE}/admin/imports/${id}/logs`,
} as const;

export const DATASETS_ENDPOINTS = {
  LIST: `${API_BASE}/admin/datasets`,
  LIST_QUESTIONS: (datasetType: string) =>
    datasetType
      ? `${API_BASE}/admin/datasets?dataset_type=${datasetType}`
      : `${API_BASE}/admin/datasets`,
  GET: (id: string) => `${API_BASE}/admin/datasets/${id}`,
  GET_BY_SLUG: (slug: string) => `${API_BASE}/admin/datasets/slug/${slug}`,
  CREATE: `${API_BASE}/admin/datasets`,
  UPDATE: (id: string) => `${API_BASE}/admin/datasets/${id}`,
  DELETE: (id: string) => `${API_BASE}/admin/datasets/${id}`,
  SET_DEFAULT: (id: string) => `${API_BASE}/admin/datasets/${id}/set-default`,
  GET_DEFAULT: `${API_BASE}/admin/datasets/default`,
  CHECK_UPDATES: `${API_BASE}/admin/datasets/check-updates`,
  HISTORY: `${API_BASE}/admin/datasets/history`,
  IMPORT: `${API_BASE}/admin/datasets/import`,
  GET_QUESTIONS: (id: string) => `${API_BASE}/admin/datasets/${id}/questions`,
  GET_STATS: (id: string) => `${API_BASE}/admin/datasets/${id}/stats`,
  GET_GEO_STATS: (id: string) => `${API_BASE}/admin/geography/${id}/stats`,
  UPDATE_STATISTICS: (id: string) =>
    `${API_BASE}/admin/datasets/${id}/update-stats`,

  LIST_COUNTRIES: (id: string) => `${API_BASE}/admin/geography/${id}/countries`,
  SEARCH_COUNTRIES: (id: string) =>
    `${API_BASE}/admin/geography/${id}/countries/search`,
  GET_COUNTRY: (id: string, slug: string) =>
    `${API_BASE}/admin/geography/${id}/countries/${slug}`,
  UPDATE_COUNTRY: (id: string, slug: string) =>
    `${API_BASE}/admin/geography/${id}/countries/${slug}`,
  COUNTRIES_BY_CONTINENT: (id: string, continent: string) =>
    `${API_BASE}/admin/geography/${id}/countries/continent/${continent}`,
  COUNTRIES_BY_REGION: (id: string, region: string) =>
    `${API_BASE}/admin/geography/${id}/countries/region/${region}`,

  LIST_CONTINENTS: (id: string) =>
    `${API_BASE}/admin/geography/${id}/continents`,
  GET_CONTINENT: (id: string, slug: string) =>
    `${API_BASE}/admin/geography/${id}/continents/${slug}`,
  UPDATE_CONTINENT: (id: string, slug: string) =>
    `${API_BASE}/admin/geography/${id}/continents/${slug}`,

  LIST_REGIONS: (id: string) => `${API_BASE}/admin/geography/${id}/regions`,
  GET_REGION: (id: string, slug: string) =>
    `${API_BASE}/admin/geography/${id}/regions/${slug}`,
  UPDATE_REGION: (id: string, slug: string) =>
    `${API_BASE}/admin/geography/${id}/regions/${slug}`,
  REGIONS_BY_CONTINENT: (id: string, continent: string) =>
    `${API_BASE}/admin/geography/${id}/regions/continent/${continent}`,

  GET_FLAG: (countryCode: string) =>
    `${API_BASE}/admin/geography/flags/${countryCode}`,
  GET_FLAG_URL: (id: string, countryCode: string) =>
    `${API_BASE}/admin/geography/${id}/flags/${countryCode}/url`,
} as const;

export const GAMES_ENDPOINTS = {
  LIST: `${API_BASE}/admin/games`,
  STATS: `${API_BASE}/admin/games/stats`,

  GET: (id: string) => `${API_BASE}/admin/games/${id}`,
  GET_PLAYERS: (id: string) => `${API_BASE}/admin/games/${id}/players`,
  GET_QUESTIONS: (id: string) => `${API_BASE}/admin/games/${id}/questions`,
  GET_ANSWERS: (id: string) => `${API_BASE}/admin/games/${id}/answers`,
  GET_EVENTS: (id: string) => `${API_BASE}/admin/games/${id}/events`,
  CANCEL: (id: string) => `${API_BASE}/admin/games/${id}/cancel`,
  ARCHIVE: (id: string) => `${API_BASE}/admin/games/${id}/archive`,
  UNARCHIVE: (id: string) => `${API_BASE}/admin/games/${id}/unarchive`,
  DELETE: (id: string) => `${API_BASE}/admin/games/${id}`,

  MODE_STATS: `${API_BASE}/admin/games/stats/modes`,
  DAILY_STATS: `${API_BASE}/admin/games/stats/daily`,
  USER_STATS: (userId: string) =>
    `${API_BASE}/admin/games/stats/users/${userId}`,
  PERFORMANCE_STATS: `${API_BASE}/admin/games/stats/performance`,
  CLEANUP: `${API_BASE}/admin/games/cleanup`,
  MAINTENANCE: `${API_BASE}/admin/games/maintenance`,

  USER_GAME_HISTORY: (userId: string) =>
    `${API_BASE}/admin/games/users/${userId}`,
} as const;

export const GAME_TEMPLATES_ENDPOINTS = {
  LIST: `${API_BASE}/admin/game-templates`,
  GET: (id: string) => `${API_BASE}/admin/game-templates/${id}`,
  CREATE: `${API_BASE}/admin/game-templates`,
  SEED_DEFAULTS: `${API_BASE}/admin/game-templates/seed-defaults`,
  UPDATE: (id: string) => `${API_BASE}/admin/game-templates/${id}`,
  DELETE: (id: string) => `${API_BASE}/admin/game-templates/${id}`,
} as const;

export const MATCHMAKING_ENDPOINTS = {
  STATS: `${API_BASE}/admin/matchmaking/stats`,
  CLEAR_QUEUE: (mode: string) => `${API_BASE}/admin/matchmaking/queue/${mode}`,
} as const;

export const OPENAPI_ENDPOINTS = {
  SPEC: `${API_BASE}/admin/openapi.yaml`,
} as const;

// Not use yet, waiting for asyncAPI scalar
// export const ASYNCAPI_ENDPOINTS = {
//   SPEC: `${API_BASE}/admin/asyncapi.yaml`,
// } as const;

export const FRIENDS_ENDPOINTS = {
  USER_FRIENDS: (userId: string) => `${API_BASE}/admin/friends/${userId}`,
  USER_FRIEND_REQUESTS: (userId: string) =>
    `${API_BASE}/admin/friends/requests/${userId}`,
} as const;

export const REPORTS_ENDPOINTS = {
  LIST: `${API_BASE}/admin/reports`,
  UPDATE_STATUS: (id: string) => `${API_BASE}/admin/reports/${id}/status`,
  GET: (id: string) => `${API_BASE}/admin/reports/${id}`,
} as const;

export const SETTINGS_ENDPOINTS = {
  MAINTENANCE: `${API_BASE}/admin/settings/maintenance`,
  CACHE_CLEAR: `${API_BASE}/admin/settings/cache/clear`,
  RATE_LIMIT: `${API_BASE}/admin/settings/rate-limit`,
  WEBSOCKET: `${API_BASE}/admin/settings/websocket`,
  AVATAR: `${API_BASE}/admin/settings/avatar`,
  XP_CONFIG: `${API_BASE}/admin/settings/xp`,
  ELO_CONFIG: `${API_BASE}/admin/settings/elo`,
  GAME_CONFIG: `${API_BASE}/admin/settings/games`,
  GAME_COUNTDOWN_CONFIG: `${API_BASE}/admin/settings/games-countdown`,
  SYSTEM_CONFIG: `${API_BASE}/admin/settings/system`,
  AUTH_CONFIG: `${API_BASE}/admin/settings/auth`,
  VERSION_STATUS: `${API_BASE}/admin/settings/version-status`,
} as const;

export const INFO_ENDPOINT = `${API_BASE}/admin/info`;
