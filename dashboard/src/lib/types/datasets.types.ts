export interface ImportJob {
  id: string;
  manifest_url: string;
  dataset: string;
  version: string;
  started_at: string;
  finished_at?: string;
  success: boolean;
  added: number;
  updated: number;
  skipped: number;
  errors: number;
  message?: string;
  flags_started_at?: string;
  flags_finished_at?: string;
  flags_svg_count: number;
  flags_png512_count: number;
  flags_png1024_count: number;
}

export interface ImportQuestionLog {
  id: string;
  job_id: string;
  line: number;
  slug: string;
  action: "created" | "updated" | "skipped" | "error";
  message?: string;
}

export interface QuestionDataset {
  id: string;
  slug: string;
  name: string;
  description: string;
  version: string;
  manifest_url?: string;
  manifest_data?: DatasetManifest;
  source: "cultpedia" | "custom" | "imported";
  import_job_id?: string;
  import_job?: ImportJob;
  imported_at: string;
  question_count: number;
  theme_count: number;
  is_active: boolean;
  is_default: boolean;
  created_at: string;
  updated_at: string;
}

export interface DatasetManifest {
  schema_version: string;
  dataset: string;
  version: string;
  updated_at: string;
  description?: string;
  license?: string;
  includes: string[];
  counts: {
    questions: number;
    themes: number;
    subthemes: number;
    tags: number;
  };
}

export interface ImportStats {
  total_jobs: number;
  successful_jobs: number;
  failed_jobs: number;
  total_questions_added: number;
  total_questions_updated: number;
  latest_job?: {
    id: string;
    dataset: string;
    started_at: string;
    success: boolean;
  };
}

export interface DatasetUpdateInfo {
  has_update: boolean;
  current_version: string;
  latest_version: string;
  updated_at?: string;
  manifest?: DatasetManifest;
  name?: string;
  dataset_id?: string;
}

export interface CreateDatasetRequest {
  slug: string;
  name: string;
  description?: string;
  version: string;
  source: "cultpedia" | "custom" | "imported";
  is_default?: boolean;
}

export interface UpdateDatasetRequest {
  name?: string;
  description?: string;
  is_active?: boolean;
  is_default?: boolean;
}

export interface ImportDatasetRequest {
  manifest_url: string;
  set_as_default?: boolean;
}

export interface ImportResult {
  success: boolean;
  message: string;
  questions_added: number;
  questions_updated: number;
  questions_skipped: number;
  themes_added: number;
  subthemes_added: number;
  tags_added: number;
  errors: string[];
  job_id?: string;
}

export interface DatasetHistoryItem {
  id: string;
  manifest_url: string;
  dataset: string;
  version: string;
  type: "questions" | "geography";
  started_at: string;
  finished_at?: string;
  success: boolean;
  added: number;
  updated: number;
  skipped: number;
  errors: number;
  message?: string;
  flags_started_at?: string;
  flags_finished_at?: string;
  flags_svg_count: number;
  flags_png512_count: number;
  flags_png1024_count: number;
}
