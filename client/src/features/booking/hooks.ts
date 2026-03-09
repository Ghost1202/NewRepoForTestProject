import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  addToWaitlist,
  createBooking,
  createBookingWithWallet,
  getDiscounts,
  getTickets,
  getUserTickets,
  refundBooking,
  transferTicket,
} from "./api";
import {
  BookingRequest,
  BookingResponse,
  GetEventDiscountResponse,
  GetTicketsResponse,
  RefundRequest,
  TransferRequest,
  UserTicketsResponse,
  WalletBookingResponse,
  WaitlistRequest,
} from "@/entities/ticket";

export const useEventTickets = (eventId: number | null) => {
  return useQuery<GetTicketsResponse>({
    queryKey: ["tickets", eventId],
    queryFn: () => getTickets(eventId as number),
    enabled: Boolean(eventId),
  });
};

export const useEventDiscounts = (eventId: number | null) => {
  return useQuery<GetEventDiscountResponse>({
    queryKey: ["discounts", eventId],
    queryFn: () => getDiscounts(eventId as number),
    enabled: Boolean(eventId),
  });
};

export const useCreateBooking = () => {
  return useMutation<BookingResponse, Error, BookingRequest>({
    mutationFn: createBooking,
  });
};

export const useCreateBookingWithWallet = () => {
  return useMutation<WalletBookingResponse, Error, BookingRequest>({
    mutationFn: createBookingWithWallet,
  });
};

export const useRefundBooking = () => {
  const queryClient = useQueryClient();
  return useMutation<void, Error, RefundRequest>({
    mutationFn: refundBooking,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["userTickets"] });
    },
  });
};

export const useUserTickets = () => {
  return useQuery<UserTicketsResponse>({
    queryKey: ["userTickets"],
    queryFn: () => getUserTickets(),
  });
};

export const useTransferTicket = () => {
  const queryClient = useQueryClient();
  return useMutation<void, Error, TransferRequest>({
    mutationFn: transferTicket,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["userTickets"] });
    },
  });
};

export const useWaitlist = () => {
  return useMutation<void, Error, WaitlistRequest>({
    mutationFn: addToWaitlist,
  });
};
