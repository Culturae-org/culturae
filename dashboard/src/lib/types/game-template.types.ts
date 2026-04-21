export type ScoreMode = "classic" | "time_bonus" | "fastest_wins";
export type TemplateMode = "solo" | "1v1" | "multi" | "";
export type TemplateCategory = "general" | "flags" | "geography" | "";
export type FlagVariant =
  | "flag_to_name_4"
  | "flag_to_name_2"
  | "name_to_flag_4"
  | "name_to_flag_2"
  | "capital_to_flag_4"
  | "capital_to_flag_2"
  | "capital_to_name_4"
  | "capital_to_name_2"
  | "flag_to_capital"
  | "flag_to_text"
  | "mix"
  | "";

export interface GameTemplate {
  id: string;
  name: string;
  description: string;
  name_i18n?: Record<string, string>;
  description_i18n?: Record<string, string>;
  slug: string;
  mode: TemplateMode;
  min_players: number;
  max_players: number;
  question_count: number;
  dataset_id?: string;
  category: TemplateCategory;
  flag_variant: FlagVariant;
  question_type: string;
  continent: string;
  include_territories: boolean;
  language: string;
  points_per_correct: number;
  time_bonus: boolean;
  score_mode: ScoreMode;
  xp_multiplier?: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateGameTemplateRequest {
  name: string;
  description?: string;
  name_i18n?: Record<string, string>;
  description_i18n?: Record<string, string>;
  slug: string;
  mode?: TemplateMode;
  min_players: number;
  max_players: number;
  question_count: number;
  dataset_id?: string;
  category?: TemplateCategory;
  flag_variant?: FlagVariant;
  question_type?: string;
  continent?: string;
  include_territories?: boolean;
  points_per_correct?: number;
  time_bonus?: boolean;
  score_mode?: ScoreMode;
  xp_multiplier?: number;
  is_active?: boolean;
}

export interface UpdateGameTemplateRequest {
  name?: string;
  description?: string;
  name_i18n?: Record<string, string>;
  description_i18n?: Record<string, string>;
  slug?: string;
  mode?: TemplateMode;
  min_players?: number;
  max_players?: number;
  question_count?: number;
  dataset_id?: string;
  category?: TemplateCategory;
  flag_variant?: FlagVariant;
  question_type?: string;
  continent?: string;
  include_territories?: boolean;
  points_per_correct?: number;
  time_bonus?: boolean;
  score_mode?: ScoreMode;
  xp_multiplier?: number;
  is_active?: boolean;
}

export interface GameTemplatesQueryParams {
  page?: number;
  limit?: number;
  query?: string;
  active_only?: boolean;
  mode?: TemplateMode;
  category?: TemplateCategory;
}
