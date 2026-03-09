export type TicketResponse = {
  id: number;
  event_id: number;
  venue_id: number;
  sector_name: string;
  row_no: number;
  seat_no: number;
  price: number;
  status: string;
};

export type EarlyResponse = {
  id: number;
  event_id: number;
  code: string;
  type: string;
  value: number;
  sector_name: string;
  valid_until: string;
};

export type BundleResponse = {
  id: number;
  event_id: number;
  code: string;
  sector_name: string;
  bundle_buy_count: number;
  bundle_get_count: number;
};

export type UserTicketsResponse = {
  tickets: TicketResponse[];
};

export type GetTicketsResponse = {
  tickets: TicketResponse[];
  count: number;
};

export type GetEventDiscountResponse = {
  early: EarlyResponse[];
  bundles: BundleResponse[];
};

export type BookingRequest = {
  ticket_ids: number[];
  user_email: string;
  promo_code?: string;
  early_id?: number;
  bundle_id?: number;
};

export type BookingResponse = {
  payment_url: string;
};

export type WalletChargeStatus =
  | "CHARGE_STATUS_SUCCESS"
  | "CHARGE_STATUS_TICKET_TAKEN"
  | "CHARGE_STATUS_INSUFFICIENT_FUNDS"
  | "CHARGE_STATUS_UNSPECIFIED";

export type WalletBookingResponse = {
  booking_id: string;
  status: WalletChargeStatus;
};

export type RefundRequest = {
  event_id: number;
  ticket_id: number;
};

export type TransferRequest = {
  ticket_id: number;
  to_user_id: number;
};

export type WaitlistRequest = {
  event_id: number;
  email: string;
};
