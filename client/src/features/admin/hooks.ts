import { useMutation } from "@tanstack/react-query";
import {
  addBundlePromos,
  addEarlyPromos,
  addPromos,
  createEvent,
  searchBundlePromos,
  searchEarlyPromos,
  searchPromos,
  updateEvent,
} from "./api";
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

export const useCreateEvent = () =>
  useMutation<CreateEventResp, Error, CreateEventReq>({
    mutationFn: createEvent,
  });

export const useUpdateEvent = () =>
  useMutation<void, Error, UpdateEventReq>({
    mutationFn: updateEvent,
  });

export const useAddPromos = () =>
  useMutation<AddPromoResp, Error, AddPromoReq>({
    mutationFn: addPromos,
  });

export const useAddEarlyPromos = () =>
  useMutation<AddEarlyResp, Error, AddEarlyReq>({
    mutationFn: addEarlyPromos,
  });

export const useAddBundlePromos = () =>
  useMutation<AddBundleResp, Error, AddBundleReq>({
    mutationFn: addBundlePromos,
  });

export const useSearchPromos = () =>
  useMutation<PromoSearchRes, Error, PromoSearchReq>({
    mutationFn: searchPromos,
  });

export const useSearchEarlyPromos = () =>
  useMutation<EarlySearchRes, Error, EarlySearchReq>({
    mutationFn: searchEarlyPromos,
  });

export const useSearchBundlePromos = () =>
  useMutation<BundleSearchRes, Error, BundleSearchReq>({
    mutationFn: searchBundlePromos,
  });
