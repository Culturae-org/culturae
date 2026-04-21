import { GAME_TEMPLATES_ENDPOINTS } from "../api/endpoints";
import type { PaginatedResponse } from "../types/api.types";
import type {
  CreateGameTemplateRequest,
  GameTemplate,
  GameTemplatesQueryParams,
  UpdateGameTemplateRequest,
} from "../types/game-template.types";
import { BaseService } from "./base.service";

class GameTemplatesService extends BaseService {
  async getTemplates(
    params?: GameTemplatesQueryParams,
  ): Promise<PaginatedResponse<GameTemplate>> {
    const { query, ...rest } = params ?? {};
    return this.getPaginated<GameTemplate>(GAME_TEMPLATES_ENDPOINTS.LIST, {
      ...rest,
      ...(query ? { q: query } : {}),
    });
  }

  async getTemplate(id: string): Promise<GameTemplate> {
    return this.get<GameTemplate>(GAME_TEMPLATES_ENDPOINTS.GET(id));
  }

  async createTemplate(data: CreateGameTemplateRequest): Promise<GameTemplate> {
    return this.post<GameTemplate>(GAME_TEMPLATES_ENDPOINTS.CREATE, data);
  }

  async updateTemplate(
    id: string,
    data: UpdateGameTemplateRequest,
  ): Promise<GameTemplate> {
    return this.patch<GameTemplate>(GAME_TEMPLATES_ENDPOINTS.UPDATE(id), data);
  }

  async deleteTemplate(id: string): Promise<void> {
    return this.delete<void>(GAME_TEMPLATES_ENDPOINTS.DELETE(id));
  }

  async seedDefaultTemplates(): Promise<{ created: number }> {
    return this.post<{ created: number }>(
      GAME_TEMPLATES_ENDPOINTS.SEED_DEFAULTS,
      {},
    );
  }
}

export const gameTemplatesService = new GameTemplatesService();
export default gameTemplatesService;
