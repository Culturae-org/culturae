import { apiGet } from "../api-client";
import { QUESTIONS_ENDPOINTS } from "../api/endpoints";
import type { PaginatedResponse } from "../types/api.types";
import type {
  Question,
  QuestionCreateData,
  QuestionUpdateData,
  QuestionsQueryParams,
} from "../types/question.types";
import { BaseService } from "./base.service";

class QuestionsService extends BaseService {
  async getQuestions(
    params?: QuestionsQueryParams,
  ): Promise<PaginatedResponse<Question>> {
    return this.getPaginated<Question>(QUESTIONS_ENDPOINTS.LIST, { ...params });
  }

  async getQuestionById(id: string): Promise<Question> {
    return this.get<Question>(QUESTIONS_ENDPOINTS.GET(id));
  }

  async getQuestionBySlug(slug: string, datasetId?: string): Promise<Question> {
    const url = datasetId
      ? `${QUESTIONS_ENDPOINTS.GET_BY_SLUG(slug)}?dataset_id=${datasetId}`
      : QUESTIONS_ENDPOINTS.GET_BY_SLUG(slug);
    return this.get<Question>(url);
  }

  async createQuestion(questionData: QuestionCreateData): Promise<Question> {
    return this.post<Question>(QUESTIONS_ENDPOINTS.CREATE, questionData);
  }

  async updateQuestion(
    id: string,
    updates: QuestionUpdateData,
  ): Promise<Question> {
    return this.put<Question>(QUESTIONS_ENDPOINTS.UPDATE(id), updates);
  }

  async deleteQuestion(id: string): Promise<void> {
    return this.delete<void>(QUESTIONS_ENDPOINTS.DELETE(id));
  }

  async exportQuestions(): Promise<Blob> {
    const response = await apiGet(QUESTIONS_ENDPOINTS.EXPORT);
    if (!response.ok) throw new Error("Failed to export questions");
    return response.blob();
  }
}

export const questionsService = new QuestionsService();
export default questionsService;
