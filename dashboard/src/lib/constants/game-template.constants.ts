import type {
  CreateGameTemplateRequest,
  ScoreMode,
} from "@/lib/types/game-template.types";

export const SCORE_MODE_OPTIONS: {
  value: ScoreMode;
  label: string;
  description: string;
}[] = [
  {
    value: "classic",
    label: "Classic",
    description: "Flat points for each correct answer. Highest total wins.",
  },
  {
    value: "time_bonus",
    label: "Time Bonus",
    description:
      "Correct answers earn points + a speed bonus for fast replies.",
  },
  {
    value: "fastest_wins",
    label: "Fastest Wins",
    description: "Only the first correct player earns points per question.",
  },
];

export const FLAG_VARIANT_OPTIONS: { value: string; label: string }[] = [
  { value: "flag_to_name_4", label: "Flag → Country (×4)" },
  { value: "flag_to_name_2", label: "Flag → Country (×2)" },
  { value: "name_to_flag_4", label: "Country → Flag (×4)" },
  { value: "name_to_flag_2", label: "Country → Flag (×2)" },
  { value: "capital_to_flag_4", label: "Capital → Flag (×4)" },
  { value: "capital_to_flag_2", label: "Capital → Flag (×2)" },
  { value: "capital_to_name_4", label: "Capital → Country (×4)" },
  { value: "capital_to_name_2", label: "Capital → Country (×2)" },
  { value: "flag_to_capital", label: "Flag → Capital (text)" },
  { value: "flag_to_text", label: "Flag → Country (text)" },
  { value: "mix", label: "Mix (all types)" },
];

export const FLAG_VARIANT_LABELS: Record<string, string> = Object.fromEntries(
  FLAG_VARIANT_OPTIONS.map((o) => [o.value, o.label]),
);

export const MODE_LABELS: Record<string, string> = {
  solo: "Solo",
  "1v1": "1v1 Duel",
  multi: "Multiplayer",
};

export const CATEGORY_LABELS: Record<string, string> = {
  general: "General Knowledge",
  flags: "Flags",
  geography: "Geography",
};

export const SCORE_MODE_LABELS: Record<string, string> = {
  classic: "Classic",
  time_bonus: "Time Bonus",
  fastest_wins: "Fastest Wins",
};

export const QUESTION_TYPE_OPTIONS: {
  value: string;
  label: string;
  description: string;
}[] = [
  {
    value: "mcq_4",
    label: "MCQ — 4 choices",
    description: "Single-choice questions shown with 4 options",
  },
  {
    value: "single_choice_2",
    label: "MCQ — 2 choices (single)",
    description:
      "Single-choice questions trimmed to 2 options (1 correct + 1 wrong)",
  },
  {
    value: "true_false",
    label: "True / False",
    description: "Only True/False questions",
  },
  {
    value: "mcq_2_mix",
    label: "MCQ 2 choices + True/False",
    description: "Mix of single-choice (2 options) and True/False questions",
  },
  {
    value: "text_input",
    label: "Text input",
    description: "Player types the answer",
  },
];

export const QUESTION_TYPE_LABELS: Record<string, string> = {
  ...Object.fromEntries(QUESTION_TYPE_OPTIONS.map((o) => [o.value, o.label])),
  any: "Any (mix of formats)",
  mcq: "MCQ (any)",
  mcq_2: "MCQ — 2 choices",
};

export const FIXED_PLAYER_COUNT_MODES = ["solo", "1v1"];

export const MODE_DEFAULTS: Record<
  string,
  Partial<CreateGameTemplateRequest>
> = {
  solo: {
    min_players: 1,
    max_players: 1,
    score_mode: "time_bonus",
    time_bonus: true,
  },
  "1v1": {
    min_players: 2,
    max_players: 2,
    score_mode: "fastest_wins",
    time_bonus: false,
  },
  multi: {
    min_players: 2,
    max_players: 4,
    score_mode: "time_bonus",
    time_bonus: true,
  },
  custom: {},
};
