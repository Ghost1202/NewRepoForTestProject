import { ChangeEvent, useState } from "react";
import { useNavigate } from "react-router-dom";
import { usePopularEvents, useSearchEvents } from "@/features/events/hooks";
import { EventResponse } from "@/entities/event";
import { Card } from "@/shared/ui/Card/Card";
import { Input } from "@/shared/ui/Input/Input";
import { Button } from "@/shared/ui/Button/Button";
import { Badge } from "@/shared/ui/Badge/Badge";
import { formatDateTime } from "@/shared/utils/format";
import styles from "./HomePage.module.scss";

type SearchState = {
  query: string;
  performerId: string;
  venueId: string;
  dateFrom: string;
  dateTo: string;
};

export const HomePage = () => {
  const navigate = useNavigate();
  const [filters, setFilters] = useState<SearchState>({
    query: "",
    performerId: "",
    venueId: "",
    dateFrom: "",
    dateTo: "",
  });
  const [searchParams, setSearchParams] = useState<Record<string, string> | null>(null);

  const popularQuery = usePopularEvents({ limit: 9, offset: 0 });
  const toNumber = (value?: string) => {
    if (!value) return undefined;
    const numeric = Number(value);
    return Number.isFinite(numeric) ? numeric : undefined;
  };

  const searchQuery = useSearchEvents(
    {
      query: searchParams?.query || undefined,
      performer_id: toNumber(searchParams?.performer_id),
      venue_id: toNumber(searchParams?.venue_id),
      date_from: searchParams?.date_from || undefined,
      date_to: searchParams?.date_to || undefined,
      limit: 12,
      offset: 0,
    },
    Boolean(searchParams),
  );

  const isSearching = Boolean(searchParams);
  const events = isSearching ? searchQuery.data?.events : popularQuery.data?.events;
  const isLoading = isSearching ? searchQuery.isLoading : popularQuery.isLoading;

  const handleChange = (key: keyof SearchState) => (event: ChangeEvent<HTMLInputElement>) => {
    setFilters((prev) => ({ ...prev, [key]: event.target.value }));
  };

  const handleSearch = () => {
    const payload = {
      query: filters.query.trim(),
      performer_id: filters.performerId.trim(),
      venue_id: filters.venueId.trim(),
      date_from: filters.dateFrom.trim(),
      date_to: filters.dateTo.trim(),
    };

    const hasAny = Object.values(payload).some((value) => value.length > 0);
    if (!hasAny) {
      setSearchParams(null);
      return;
    }
    setSearchParams(payload);
  };

  const openEvent = (eventItem: EventResponse) => {
    navigate(`/event/${eventItem.id}`, { state: { event: eventItem } });
  };

  return (
    <div className="container page">
      <section className={styles.hero}>
        <div className={styles.heroCopy}>
          <Badge tone="muted" className={styles.heroBadge}>
            Searching service
          </Badge>
          <h1 className={styles.heroTitle}>Find events worth the trip</h1>
          <p className={styles.heroLead}>
            Browse popular shows or search with filters. Results are powered by the Event Searching
            Service and cached in Redis.
          </p>
        </div>
        <Card className={styles.searchCard}>
          <div className={styles.searchGrid}>
            <Input label="Search text" value={filters.query} onChange={handleChange("query")} />
            <Input label="Performer ID" value={filters.performerId} onChange={handleChange("performerId")} />
            <Input label="Venue ID" value={filters.venueId} onChange={handleChange("venueId")} />
            <Input label="Date from (RFC3339)" value={filters.dateFrom} onChange={handleChange("dateFrom")} />
            <Input label="Date to (RFC3339)" value={filters.dateTo} onChange={handleChange("dateTo")} />
          </div>
          <div className={styles.searchActions}>
            <Button onClick={handleSearch}>{isSearching ? "Update search" : "Search"}</Button>
            {isSearching && (
              <Button variant="ghost" onClick={() => setSearchParams(null)}>
                Clear
              </Button>
            )}
          </div>
        </Card>
      </section>

      <section className="section">
        <div className={styles.sectionHeader}>
          <h2>{isSearching ? "Search results" : "Popular events"}</h2>
          <span className="pill">{events?.length ?? 0} events</span>
        </div>

        {isLoading && <p>Loading events...</p>}
        {!isLoading && events && events.length === 0 && <p>No events found.</p>}

        <div className={`${styles.cards} grid grid-3`}>
          {events?.map((eventItem) => (
            <button
              key={eventItem.id}
              type="button"
              className={styles.cardButton}
              data-testid="event-card"
              onClick={() => openEvent(eventItem)}
            >
              <Card className={styles.eventCard}>
                <div className={styles.cardHeader}>
                  <h3>{eventItem.name}</h3>
                  {eventItem.is_sold_out && <Badge tone="warning">Sold out</Badge>}
                </div>
                <p className={styles.meta}>Start: {formatDateTime(eventItem.date_start)}</p>
                <p className={styles.meta}>Venue ID: {eventItem.venue_id}</p>
                <p className={styles.meta}>Popularity: {eventItem.popularity}</p>
              </Card>
            </button>
          ))}
        </div>
      </section>
    </div>
  );
};
