import { useQuery } from "@tanstack/react-query";
import { EventSearchParams, GetEventsFilterResponse, GetEventsResponse } from "@/entities/event";
import { getPopularEvents, searchEvents } from "./api";

export const usePopularEvents = (params?: EventSearchParams) => {
  return useQuery<GetEventsResponse>({
    queryKey: ["events", "popular", params],
    queryFn: () => getPopularEvents(params),
  });
};

export const useSearchEvents = (params: EventSearchParams, enabled: boolean) => {
  return useQuery<GetEventsFilterResponse>({
    queryKey: ["events", "search", params],
    queryFn: () => searchEvents(params),
    enabled,
  });
};
