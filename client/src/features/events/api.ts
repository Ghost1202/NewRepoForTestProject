import { searchingApi } from "@/shared/api";
import { EventSearchParams, GetEventsFilterResponse, GetEventsResponse } from "@/entities/event";

export const getPopularEvents = async (params?: EventSearchParams) => {
  const response = await searchingApi.get<GetEventsResponse>("/events", { params, authMode: "none" });
  return response.data;
};

export const searchEvents = async (params: EventSearchParams) => {
  const response = await searchingApi.get<GetEventsFilterResponse>("/events/search", { params, authMode: "none" });
  return response.data;
};
