export type Comment = {
  event_id: number;
  user_id: number;
  nick: string;
  text: string;
  rating: number;
};

export type GetCommentsRequest = {
  event_id: number;
  limit?: number;
  offset?: number;
};

export type GetCommentsResponse = {
  event_id: number;
  limit: number;
  offset: number;
  items: Comment[];
};

export type GetRatingRequest = {
  event_id: number;
};

export type GetRatingResponse = {
  event_id: number;
  avg: number;
  count: number;
};

export type SetCommentRequest = {
  event_id: number;
  rating: number;
  text: string;
  nick: string;
};
