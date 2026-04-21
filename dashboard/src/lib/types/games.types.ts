export type GameMode = "solo" | "1v1" | "multi";
export type GameStatus =
  | "waiting"
  | "ready"
  | "in_progress"
  | "completed"
  | "cancelled"
  | "abandoned";
export type PlayerStatus = "playing" | "left" | "finished" | "disconnected";

export interface GamePlayer {
  id: string;
  game_id: string;
  user_public_id: string;
  username?: string;
  score: number;
  is_ready: boolean;
  status: PlayerStatus | string;
  joined_at: string;
  user?: {
    id: string;
    username: string;
    has_avatar: boolean;
    email?: string;
    public_id?: string;
    role?: string;
    rank?: string;
    level?: number;
    xp?: number;
  };
}

export interface Game {
  id: string;
  public_id: string;
  mode: GameMode | string;
  status: GameStatus | string;
  creator_public_id: string;
  question_count: number;
  dataset_id?: string;
  winner_id?: string;
  started_at?: string;
  completed_at?: string;
  created_at: string;
  players: GamePlayer[];
}

export interface UserGameHistory {
  games: Game[];
  total: number;
}

export interface MatchmakingQueueStats {
  total_players: number;
  queues: Record<string, number>;
}

export interface AdminGame {
  id: string;
  public_id?: string;
  mode: string;
  status: string;
  creator_public_id?: string;
  question_count?: number;
  category?: string;
  flag_variant?: string;
  language?: string;
  created_at: string;
  started_at: string | null;
  completed_at: string | null;
  updated_at?: string;
  deleted_at?: string | null;
  players?: GamePlayer[];

  max_players?: number;
  current_players?: number;
  winner_id?: string | null;
  created_by?: string;
}

export interface GamePlayerSummary {
  id: string;
  username: string;
  score: number;
  joined_at: string;
}

export interface GameQuestionSummary {
  id: string;
  question: string;
  answered: boolean;
}

export interface GameAnswerSummary {
  player_id: string;
  question_id: string;
  answer: string;
  is_correct: boolean;
  answered_at: string;
}

export interface GameDetails extends Omit<AdminGame, "players"> {
  players?: GamePlayerSummary[];
  questions?: GameQuestionSummary[];
  answers?: GameAnswerSummary[];
}

export interface QuestionI18n {
  stem?: string;
  Stem?: string;
}

export interface QuestionAnswer {
  slug: string;
  is_correct: boolean;
  [key: string]: unknown;
}

export interface QuestionData {
  id: string;
  qtype: string;
  difficulty: string;
  points: number;
  i18n: Record<string, QuestionI18n>;
  answers: QuestionAnswer[];
}

export interface GeoQuestionCorrectAnswer {
  flag?: string;
  name?: Record<string, string>;
  slug?: string;
}

export interface GeoQuestionData {
  target_id?: string;
  target_slug?: string;
  target_iso2?: string;
  target_iso3?: string;
  target_name?: Record<string, string>;
  correct_answer?: GeoQuestionCorrectAnswer;
  mode?: string;
  flag?: string;
  variant?: string;
  options?: Array<{
    flag?: string;
    name?: Record<string, string>;
    slug?: string;
    is_correct?: boolean;
  }>;
}

export interface FlagAnswerData {
  question_type?: string;
  question_slug?: string;
  match_slug?: string;
  match_type?: string;
  user_answer?: {
    coordinates?: { lat: number; lng: number };
    submitted_slug?: string;
  };
  server_time_spent_ms?: number;
}

export interface GameQuestion {
  id: string;
  game_id: string;
  question_id: string | null;
  order_number: number;
  type: string;
  question: QuestionData | null;
  data?: GeoQuestionData;
}

export interface GameAnswer {
  id: string;
  game_id: string;
  player_id: string;
  question_id: string | null;
  answer_slug: string;
  is_correct: boolean;
  time_spent: number;
  points: number;
  answered_at: string;
  data?: FlagAnswerData | Record<string, unknown>;
}

export interface GameDetail {
  id: string;
  public_id: string;
  mode: string;
  status: string;
  creator_public_id: string;
  winner_id: string | null;
  question_count: number;
  category: string;
  flag_variant: string;
  language: string;
  template_id?: string | null;
  time_bonus?: boolean;
  points_per_correct?: number;
  created_at: string;
  started_at: string | null;
  completed_at: string | null;
  updated_at?: string;
  players: GamePlayer[];
  questions?: GameQuestion[];
  answers: GameAnswer[];
}

export interface GameStats {
  total_games: number;
  active_games: number;
  completed_games: number;
  abandoned_games: number;
  total_players: number;
  avg_game_duration: number | null;
  avg_players_per_game: number | null;
  most_popular_mode: string | null;
}

export interface GameModeStats {
  mode: string;
  count: number;
}

export interface GameDailyStats {
  date: string;
  total_games: number;
  completed_games: number;
  cancelled_games?: number;
  total_players?: number;
}

export interface GamePerformanceStats {
  avg_game_duration_seconds: number;
  avg_questions_per_game: number;
  avg_players_per_game: number;
  total_questions_used: number;
  most_popular_mode: string;
}

export interface GameAnswerDetail {
  id: string;
  game_id: string;
  player_id: string;
  question_id: string | null;
  question_slug: string;
  question_title: string;
  question_type: string;
  answer_slug: string;
  answer_label: string;
  correct_answer_slug: string;
  correct_answer_label: string;
  data?: Record<string, unknown>;
  is_correct: boolean;
  time_spent: number;
  points: number;
  answered_at: string;
}

export interface GameEventLog {
  id: string;
  game_id: string;
  event_type: string;
  data: Record<string, unknown>;
  occurred_at: string;
}

export type ApiPlayer = GamePlayer;
export type ApiAnswer = GameAnswer;
export type ApiGameQuestion = GameQuestion;
export type ApiGameDetail = GameDetail;
