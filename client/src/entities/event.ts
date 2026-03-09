export type EventResponse = {
  id: number;
  performer_id: number;
  venue_id: number;
  name: string;
  date_start: string;
  is_sold_out: boolean;
  date_sale_start?: string;
  info_header?: string;
  info_body?: string;
  popularity: number;
};

export type GetEventsResponse = {
  events: EventResponse[];
  count: number;
};

export type GetEventsFilterResponse = {
  events: EventResponse[];
  count: number;
};

export type EventSearchParams = {
  query?: string;
  performer_id?: number;
  venue_id?: number;
  date_from?: string;
  date_to?: string;
  limit?: number;
  offset?: number;
};
