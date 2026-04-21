import { FRIENDS_ENDPOINTS } from "../api/endpoints";
import type { PaginatedResponse } from "../types/api.types";
import type { FriendRequest, Friendship } from "../types/friends.types";
import { BaseService } from "./base.service";

class FriendsService extends BaseService {
  async getFriendsForUser(
    userId: string,
    page = 1,
    limit = 20,
  ): Promise<PaginatedResponse<Friendship>> {
    return this.getPaginated<Friendship>(
      FRIENDS_ENDPOINTS.USER_FRIENDS(userId),
      { page, limit },
    );
  }

  async getFriendRequestsForUser(
    userId: string,
    params?: {
      status?: string;
      direction?: string;
      page?: number;
      limit?: number;
    },
  ): Promise<PaginatedResponse<FriendRequest>> {
    return this.getPaginated<FriendRequest>(
      FRIENDS_ENDPOINTS.USER_FRIEND_REQUESTS(userId),
      { page: 1, limit: 100, ...params },
    );
  }
}

export const friendsService = new FriendsService();
