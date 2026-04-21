import { GAMES_ENDPOINTS } from "../api/endpoints";
import type { PaginatedResponse } from "../types/api.types";
import type {
  AdminGame,
  Game,
  GameAnswer,
  GameDailyStats,
  GameModeStats,
  GamePerformanceStats,
  GamePlayer,
  GameQuestion,
  GameStats,
} from "../types/games.types";
import { BaseService } from "./base.service";

export type {
  GameStats,
  GameModeStats,
  GameDailyStats,
  GamePerformanceStats,
  GameAnswer,
  GameQuestion,
};

class GamesService extends BaseService {
  async list(params?: {
    page?: number;
    limit?: number;
    status?: string;
    mode?: string;
    query?: string;
    archived?: string;
  }): Promise<PaginatedResponse<AdminGame>> {
    return this.getPaginated<AdminGame>(GAMES_ENDPOINTS.LIST, {
      page: params?.page,
      limit: params?.limit,
      status: params?.status,
      mode: params?.mode,
      query: params?.query,
      archived: params?.archived,
    });
  }

  async getStats(): Promise<GameStats> {
    return this.get<GameStats>(GAMES_ENDPOINTS.STATS);
  }

  async getById(id: string): Promise<Game> {
    return this.get<Game>(GAMES_ENDPOINTS.GET(id));
  }

  async getPlayers(gameId: string): Promise<GamePlayer[]> {
    return this.get<GamePlayer[]>(GAMES_ENDPOINTS.GET_PLAYERS(gameId));
  }

  async getQuestions(gameId: string): Promise<GameQuestion[]> {
    return this.get<GameQuestion[]>(GAMES_ENDPOINTS.GET_QUESTIONS(gameId));
  }

  async getAnswers(gameId: string): Promise<GameAnswer[]> {
    return this.get<GameAnswer[]>(GAMES_ENDPOINTS.GET_ANSWERS(gameId));
  }

  async cancel(gameId: string): Promise<void> {
    return this.post<void>(GAMES_ENDPOINTS.CANCEL(gameId));
  }

  async deleteGame(gameId: string): Promise<void> {
    return super.delete(GAMES_ENDPOINTS.DELETE(gameId));
  }

  async getModeStats(): Promise<GameModeStats> {
    return this.get<GameModeStats>(GAMES_ENDPOINTS.MODE_STATS);
  }

  async getDailyStats(): Promise<GameDailyStats[]> {
    return this.get<GameDailyStats[]>(GAMES_ENDPOINTS.DAILY_STATS);
  }

  async getPerformanceStats(): Promise<GamePerformanceStats> {
    return this.get<GamePerformanceStats>(GAMES_ENDPOINTS.PERFORMANCE_STATS);
  }

  async cleanup(): Promise<void> {
    return this.post<void>(GAMES_ENDPOINTS.CLEANUP);
  }

  async getUserGameHistory(
    userId: string,
    status?: string,
    mode?: string,
  ): Promise<
    Array<{
      game: Game;
      user_score: number;
      is_winner: boolean;
      players: GamePlayer[];
    }>
  > {
    const params: Record<string, string> = {};
    if (status) params.status = status;
    if (mode) params.mode = mode;

    return this.get<
      Array<{
        game: Game;
        user_score: number;
        is_winner: boolean;
        players: GamePlayer[];
      }>
    >(GAMES_ENDPOINTS.USER_GAME_HISTORY(userId), params);
  }
}

export const gamesService = new GamesService();
