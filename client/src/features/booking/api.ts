import { bookingApi } from "@/shared/api";
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

export const getTickets = async (eventId: number) => {
  const response = await bookingApi.post<GetTicketsResponse>(
    "/tickets",
    { event_id: eventId },
    { authMode: "none" },
  );
  return response.data;
};

export const getDiscounts = async (eventId: number) => {
  const response = await bookingApi.post<GetEventDiscountResponse>(
    "/tickets/discounts",
    { event_id: eventId },
    { authMode: "none" },
  );
  return response.data;
};

export const createBooking = async (payload: BookingRequest) => {
  const response = await bookingApi.post<BookingResponse>("/bookings", payload, { authMode: "required" });
  return response.data;
};

export const createBookingWithWallet = async (payload: BookingRequest) => {
  const response = await bookingApi.post<WalletBookingResponse>("/bookings/wallet", payload, { authMode: "required" });
  return response.data;
};

export const refundBooking = async (payload: RefundRequest) => {
  const response = await bookingApi.post<void>("/bookings/refund", payload, { authMode: "required" });
  return response.data;
};

export const getUserTickets = async () => {
  const response = await bookingApi.post<UserTicketsResponse>("/user/tickets", {}, { authMode: "required" });
  return response.data;
};

export const transferTicket = async (payload: TransferRequest) => {
  const response = await bookingApi.post<void>("/user/tickets/transfer", payload, { authMode: "required" });
  return response.data;
};

export const addToWaitlist = async (payload: WaitlistRequest) => {
  const response = await bookingApi.post<void>("/user/waitlist", payload, { authMode: "required" });
  return response.data;
};
