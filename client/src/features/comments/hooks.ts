import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { getComments, getRating, postComment } from "./api";
import { GetCommentsResponse, GetRatingResponse, SetCommentRequest } from "@/entities/comment";

export const useComments = (eventId: number | null, limit = 10, offset = 0) => {
  return useQuery<GetCommentsResponse>({
    queryKey: ["comments", eventId, limit, offset],
    queryFn: () => getComments({ event_id: eventId as number, limit, offset }),
    enabled: Boolean(eventId),
  });
};

export const useRating = (eventId: number | null) => {
  return useQuery<GetRatingResponse>({
    queryKey: ["rating", eventId],
    queryFn: () => getRating({ event_id: eventId as number }),
    enabled: Boolean(eventId),
  });
};

export const useAddComment = (eventId: number) => {
  const queryClient = useQueryClient();
  return useMutation<void, Error, SetCommentRequest>({
    mutationFn: postComment,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["comments", eventId] });
      void queryClient.invalidateQueries({ queryKey: ["rating", eventId] });
    },
  });
};
