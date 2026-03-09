export type Promo = {
  code: string;
  type: string;
  value: number;
  sector?: string;
};

export type Early = {
  code: string;
  type: string;
  value: number;
  sector?: string;
  valid_until: string;
};

export type Bundle = {
  code: string;
  sector?: string;
  buy_count: number;
  get_count: number;
};

export type TicketConstructor = {
  name: string;
  type: string;
  price: number;
  rows_count: number;
  seats_per_row: number;
};

export type CreateEventReq = {
  venue_id: number;
  name: string;
  date_start: string;
  post_date: string;
  sale_start_date: string;
  max_price_cof: number;
  min_price_cof: number;
  constructor?: TicketConstructor[];
  promos?: Promo[];
  early?: Early[];
  bundles?: Bundle[];
};

export type CreateEventResp = {
  event_id: number;
};

export type UpdateEventReq = {
  event_id: number;
  name: string;
  date_start: string;
  max_price_cof: number;
  min_price_cof: number;
};

export type AddPromoReq = {
  event_id: number;
  promos: Promo[];
};

export type AddPromoResp = {
  promo_ids: number[];
};

export type AddEarlyReq = {
  event_id: number;
  early: Early[];
};

export type AddEarlyResp = {
  early_ids: number[];
};

export type AddBundleReq = {
  event_id: number;
  bundles: Bundle[];
};

export type AddBundleResp = {
  bundle_ids: number[];
};

export type PromoSearchReq = {
  event_id?: number;
  limit: number;
  offset: number;
};

export type PromoSearchRes = {
  promos: Promo[];
};

export type EarlySearchReq = {
  event_id?: number;
  limit: number;
  offset: number;
};

export type EarlySearchRes = {
  early: Early[];
};

export type BundleSearchReq = {
  event_id?: number;
  limit: number;
  offset: number;
};

export type BundleSearchRes = {
  bundles: Bundle[];
};
