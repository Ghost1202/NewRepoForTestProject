import { eventApi } from "@/shared/api";
import {
  AddBundleReq,
  AddBundleResp,
  AddEarlyReq,
  AddEarlyResp,
  AddPromoReq,
  AddPromoResp,
  BundleSearchReq,
  BundleSearchRes,
  CreateEventReq,
  CreateEventResp,
  EarlySearchReq,
  EarlySearchRes,
  PromoSearchReq,
  PromoSearchRes,
  UpdateEventReq,
} from "@/entities/promo";

export const createEvent = async (payload: CreateEventReq) => {
  const response = await eventApi.post<CreateEventResp>("/events", payload, { authMode: "required" });
  return response.data;
};

export const updateEvent = async (payload: UpdateEventReq) => {
  const response = await eventApi.put<void>("/events", payload, { authMode: "required" });
  return response.data;
};

export const addPromos = async (payload: AddPromoReq) => {
  const response = await eventApi.post<AddPromoResp>("/promos/promo", payload, { authMode: "required" });
  return response.data;
};

export const addEarlyPromos = async (payload: AddEarlyReq) => {
  const response = await eventApi.post<AddEarlyResp>("/promos/early", payload, { authMode: "required" });
  return response.data;
};

export const addBundlePromos = async (payload: AddBundleReq) => {
  const response = await eventApi.post<AddBundleResp>("/promos/bundle", payload, { authMode: "required" });
  return response.data;
};

export const searchPromos = async (payload: PromoSearchReq) => {
  const response = await eventApi.post<PromoSearchRes>("/promos/promo/search", payload, { authMode: "required" });
  return response.data;
};

export const searchEarlyPromos = async (payload: EarlySearchReq) => {
  const response = await eventApi.post<EarlySearchRes>("/promos/early/search", payload, { authMode: "required" });
  return response.data;
};

export const searchBundlePromos = async (payload: BundleSearchReq) => {
  const response = await eventApi.post<BundleSearchRes>("/promos/bundle/search", payload, { authMode: "required" });
  return response.data;
};
