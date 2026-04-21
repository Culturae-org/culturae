export type ReportStatus = "pending" | "in_progress" | "resolved";

export interface Report {
  id: string;
  user_id: string;
  question_id?: string;
  game_question_id?: string;
  reason: string;
  message: string;
  resolution_notes?: string;
  status: ReportStatus;
  created_at: string;
  updated_at: string;
  user?: {
    username: string;
    avatar_url?: string;
  };
  question?: {
    question: {
      en: string;
    };
    category: string;
  };
  game_question?: {
    id: string;
    game_id: string;
    order_number: number;
  };
}

export interface ReportsResponse {
  reports: Report[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

export interface ReportFilters {
  status?: ReportStatus;
}
