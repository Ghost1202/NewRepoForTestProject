import { searchingApi } from "@/shared/api";
import {
  GetCommentsRequest,
  GetCommentsResponse,
  GetRatingRequest,
  GetRatingResponse,
  SetCommentRequest,
} from "@/entities/comment";

export const getComments = async (payload: GetCommentsRequest) => {
  const response = await searchingApi.post<GetCommentsResponse>(
    "/comments/list",
    payload,
    { authMode: "none" },
  );
  return response.data;
};

export const getRating = async (payload: GetRatingRequest) => {
  const response = await searchingApi.post<GetRatingResponse>(
    "/comments/rating",
    payload,
    { authMode: "none" },
  );
  return response.data;
};

export const postComment = async (payload: SetCommentRequest) => {
  const response = await searchingApi.post<void>("/comments", payload, { authMode: "required" });
  return response.data;
};
