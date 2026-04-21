export interface Friendship {
  id: string;
  user1_id: string;
  user1_username: string;
  user2_id: string;
  user2_username: string;
  created_at: string;
}

export interface FriendRequest {
  id: string;
  from_user_id: string;
  from_username: string;
  to_user_id: string;
  to_username: string;
  status: string;
  created_at: string;
  updated_at: string;
}
