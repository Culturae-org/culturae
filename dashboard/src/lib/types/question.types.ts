export interface Question {
  id: string;
  slug: string;
  version: string;
  kind: string;
  qtype: string;
  difficulty: string;
  estimated_seconds: number;
  shuffle_answers: boolean;
  dataset_id?: string;
  theme: {
    id: string;
    slug: string;
  };
  subthemes: Array<{
    id: string;
    slug: string;
  }>;
  tags: Array<{
    id: string;
    slug: string;
  }>;
  i18n: Record<
    string,
    {
      title: string;
      stem: string;
      explanation?: string;
    }
  >;
  answers: Array<{
    slug: string;
    is_correct: boolean;
    i18n: Record<
      string,
      {
        label: string;
      }
    >;
  }>;
  sources?: string[];
  country_id?: string;
  country_slug?: string;
  variant?: string;
  times_played?: number;
  times_correct?: number;
  success_rate?: number;
  avg_time_ms?: number;
  created_at: string;
  updated_at: string;
}

export interface QuestionCreateData {
  kind: string;
  version: string;
  slug: string;
  dataset_id?: string;
  theme: {
    slug: string;
  };
  subthemes?: Array<{
    slug: string;
  }>;
  tags?: Array<{
    slug: string;
  }>;
  qtype: string;
  difficulty: string;
  estimated_seconds: number;
  shuffle_answers?: boolean;
  i18n: Record<
    string,
    {
      title: string;
      stem: string;
      explanation?: string;
    }
  >;
  answers: Array<{
    slug: string;
    is_correct: boolean;
    i18n: Record<
      string,
      {
        label: string;
      }
    >;
  }>;
  sources?: string[];
}

export interface QuestionUpdateData {
  kind?: string;
  version?: string;
  slug?: string;
  theme?: {
    slug: string;
  };
  subthemes?: Array<{
    slug: string;
  }>;
  tags?: Array<{
    slug: string;
  }>;
  qtype?: string;
  difficulty?: string;
  estimated_seconds?: number;
  shuffle_answers?: boolean;
  i18n?: Record<
    string,
    {
      title?: string;
      stem?: string;
      explanation?: string;
    }
  >;
  answers?: Array<{
    slug: string;
    is_correct: boolean;
    i18n: Record<
      string,
      {
        label: string;
      }
    >;
  }>;
  sources?: string[];
}

export interface QuestionsQueryParams {
  page?: number;
  limit?: number;
  search?: string;
  dataset_id?: string;
  category?: string;
  difficulty?: string;
  language?: string;
  theme?: string;
  qtype?: string;
}
