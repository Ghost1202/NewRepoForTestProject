import { useState } from "react";
import { useFieldArray, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Card } from "@/shared/ui/Card/Card";
import { Input } from "@/shared/ui/Input/Input";
import { Button } from "@/shared/ui/Button/Button";
import { toRfc3339 } from "@/shared/utils/format";
import { getErrorMessage } from "@/shared/utils/error";
import {
  addBundleSchema,
  AddBundleFormValues,
  addEarlySchema,
  AddEarlyFormValues,
  addPromoSchema,
  AddPromoFormValues,
  createEventSchema,
  CreateEventFormValues,
  searchPromoSchema,
  SearchPromoValues,
  updateEventSchema,
  UpdateEventFormValues,
} from "@/features/admin/schema";
import {
  useAddBundlePromos,
  useAddEarlyPromos,
  useAddPromos,
  useCreateEvent,
  useSearchBundlePromos,
  useSearchEarlyPromos,
  useSearchPromos,
  useUpdateEvent,
} from "@/features/admin/hooks";
import styles from "./AdminDashboard.module.scss";

export const AdminDashboard = () => {
  const [eventStatus, setEventStatus] = useState<{ type: "success" | "error" | "info"; message: string } | null>(null);
  const [promoStatus, setPromoStatus] = useState<{ type: "success" | "error" | "info"; message: string } | null>(null);
  const [earlyStatus, setEarlyStatus] = useState<{ type: "success" | "error" | "info"; message: string } | null>(null);
  const [bundleStatus, setBundleStatus] = useState<{ type: "success" | "error" | "info"; message: string } | null>(null);
  const [searchPromoStatus, setSearchPromoStatus] = useState<{ type: "success" | "error" | "info"; message: string } | null>(null);
  const [searchEarlyStatus, setSearchEarlyStatus] = useState<{ type: "success" | "error" | "info"; message: string } | null>(null);
  const [searchBundleStatus, setSearchBundleStatus] = useState<{ type: "success" | "error" | "info"; message: string } | null>(null);
  const createEventMutation = useCreateEvent();
  const updateEventMutation = useUpdateEvent();
  const addPromosMutation = useAddPromos();
  const addEarlyMutation = useAddEarlyPromos();
  const addBundleMutation = useAddBundlePromos();
  const searchPromosMutation = useSearchPromos();
  const searchEarlyMutation = useSearchEarlyPromos();
  const searchBundleMutation = useSearchBundlePromos();

  const createForm = useForm<CreateEventFormValues>({
    resolver: zodResolver(createEventSchema),
    defaultValues: {
      ticket_constructor: [],
      promos: [],
      early: [],
      bundles: [],
    },
  });

  const updateForm = useForm<UpdateEventFormValues>({
    resolver: zodResolver(updateEventSchema),
  });

  const promoForm = useForm<AddPromoFormValues>({
    resolver: zodResolver(addPromoSchema),
    defaultValues: {
      promos: [],
    },
  });

  const earlyForm = useForm<AddEarlyFormValues>({
    resolver: zodResolver(addEarlySchema),
    defaultValues: {
      early: [],
    },
  });

  const bundleForm = useForm<AddBundleFormValues>({
    resolver: zodResolver(addBundleSchema),
    defaultValues: {
      bundles: [],
    },
  });

  const searchPromoForm = useForm<SearchPromoValues>({
    resolver: zodResolver(searchPromoSchema),
    defaultValues: {
      limit: 10,
      offset: 0,
    },
  });

  const searchEarlyForm = useForm<SearchPromoValues>({
    resolver: zodResolver(searchPromoSchema),
    defaultValues: {
      limit: 10,
      offset: 0,
    },
  });

  const searchBundleForm = useForm<SearchPromoValues>({
    resolver: zodResolver(searchPromoSchema),
    defaultValues: {
      limit: 10,
      offset: 0,
    },
  });

  const constructorFields = useFieldArray({ control: createForm.control, name: "ticket_constructor" });
  const promosFields = useFieldArray({ control: createForm.control, name: "promos" });
  const earlyFields = useFieldArray({ control: createForm.control, name: "early" });
  const bundlesFields = useFieldArray({ control: createForm.control, name: "bundles" });

  const addPromoFields = useFieldArray({ control: promoForm.control, name: "promos" });
  const addEarlyFields = useFieldArray({ control: earlyForm.control, name: "early" });
  const addBundlesFields = useFieldArray({ control: bundleForm.control, name: "bundles" });

  const handleCreateEvent = async (data: CreateEventFormValues) => {
    setEventStatus({ type: "info", message: "Creating event..." });
    const { ticket_constructor, ...rest } = data;
    const payload = {
      ...rest,
      date_start: toRfc3339(data.date_start),
      post_date: toRfc3339(data.post_date),
      sale_start_date: toRfc3339(data.sale_start_date),
      constructor: ticket_constructor?.length ? ticket_constructor : undefined,
      promos: data.promos?.length ? data.promos : undefined,
      early: data.early?.length
        ? data.early.map((item) => ({ ...item, valid_until: toRfc3339(item.valid_until) }))
        : undefined,
      bundles: data.bundles?.length ? data.bundles : undefined,
    };

    try {
      const response = await createEventMutation.mutateAsync(payload);
      setEventStatus({ type: "success", message: `Created event ID: ${response.event_id}` });
    } catch (error) {
      setEventStatus({ type: "error", message: getErrorMessage(error, "Failed to create event.") });
    }
  };

  const handleUpdateEvent = async (data: UpdateEventFormValues) => {
    setEventStatus({ type: "info", message: "Updating event..." });
    try {
      await updateEventMutation.mutateAsync({
        ...data,
        date_start: toRfc3339(data.date_start),
      });
      setEventStatus({ type: "success", message: `Updated event ID: ${data.event_id}` });
    } catch (error) {
      setEventStatus({ type: "error", message: getErrorMessage(error, "Failed to update event.") });
    }
  };

  const handleAddPromos = async (data: AddPromoFormValues) => {
    setPromoStatus({ type: "info", message: "Adding promos..." });
    try {
      const response = await addPromosMutation.mutateAsync(data);
      setPromoStatus({
        type: "success",
        message: `Created promo IDs: ${response.promo_ids.join(", ") || "none"}`,
      });
    } catch (error) {
      setPromoStatus({ type: "error", message: getErrorMessage(error, "Failed to add promos.") });
    }
  };

  const handleAddEarly = async (data: AddEarlyFormValues) => {
    setEarlyStatus({ type: "info", message: "Adding early promos..." });
    try {
      const payload = {
        ...data,
        early: data.early.map((item) => ({ ...item, valid_until: toRfc3339(item.valid_until) })),
      };
      const response = await addEarlyMutation.mutateAsync(payload);
      setEarlyStatus({
        type: "success",
        message: `Created early IDs: ${response.early_ids.join(", ") || "none"}`,
      });
    } catch (error) {
      setEarlyStatus({ type: "error", message: getErrorMessage(error, "Failed to add early promos.") });
    }
  };

  const handleAddBundle = async (data: AddBundleFormValues) => {
    setBundleStatus({ type: "info", message: "Adding bundle promos..." });
    try {
      const response = await addBundleMutation.mutateAsync(data);
      setBundleStatus({
        type: "success",
        message: `Created bundle IDs: ${response.bundle_ids.join(", ") || "none"}`,
      });
    } catch (error) {
      setBundleStatus({ type: "error", message: getErrorMessage(error, "Failed to add bundles.") });
    }
  };

  const handleSearchPromos = async (data: SearchPromoValues) => {
    setSearchPromoStatus({ type: "info", message: "Searching promos..." });
    try {
      const response = await searchPromosMutation.mutateAsync({
        ...data,
        event_id: Number.isFinite(data.event_id ?? NaN) ? data.event_id : undefined,
      });
      setSearchPromoStatus({ type: "success", message: `Promos found: ${response.promos.length}` });
    } catch (error) {
      setSearchPromoStatus({ type: "error", message: getErrorMessage(error, "Failed to search promos.") });
    }
  };

  const handleSearchEarly = async (data: SearchPromoValues) => {
    setSearchEarlyStatus({ type: "info", message: "Searching early promos..." });
    try {
      const response = await searchEarlyMutation.mutateAsync({
        ...data,
        event_id: Number.isFinite(data.event_id ?? NaN) ? data.event_id : undefined,
      });
      setSearchEarlyStatus({ type: "success", message: `Early promos found: ${response.early.length}` });
    } catch (error) {
      setSearchEarlyStatus({ type: "error", message: getErrorMessage(error, "Failed to search early promos.") });
    }
  };

  const handleSearchBundle = async (data: SearchPromoValues) => {
    setSearchBundleStatus({ type: "info", message: "Searching bundles..." });
    try {
      const response = await searchBundleMutation.mutateAsync({
        ...data,
        event_id: Number.isFinite(data.event_id ?? NaN) ? data.event_id : undefined,
      });
      setSearchBundleStatus({ type: "success", message: `Bundles found: ${response.bundles.length}` });
    } catch (error) {
      setSearchBundleStatus({ type: "error", message: getErrorMessage(error, "Failed to search bundles.") });
    }
  };

  return (
    <div className="container page">
      <h1>Admin dashboard</h1>
      <section className={styles.sectionGrid}>
        <Card>
          <h2>Create event</h2>
          <form className={styles.form} onSubmit={createForm.handleSubmit(handleCreateEvent)}>
            <Input label="Venue ID" type="number" {...createForm.register("venue_id", { valueAsNumber: true })} />
            <Input label="Name" {...createForm.register("name")} />
            <Input label="Date start" type="datetime-local" {...createForm.register("date_start")} />
            <Input label="Post date" type="datetime-local" {...createForm.register("post_date")} />
            <Input label="Sale start" type="datetime-local" {...createForm.register("sale_start_date")} />
            <Input
              label="Max price coef"
              type="number"
              step="0.1"
              {...createForm.register("max_price_cof", { valueAsNumber: true })}
            />
            <Input
              label="Min price coef"
              type="number"
              step="0.1"
              {...createForm.register("min_price_cof", { valueAsNumber: true })}
            />

            <div className={styles.group}>
              <h3>Ticket constructor</h3>
              {constructorFields.fields.map((field, index) => (
                <div key={field.id} className={styles.inlineGrid}>
                  <Input label="Name" {...createForm.register(`ticket_constructor.${index}.name`)} />
                  <Input label="Type" {...createForm.register(`ticket_constructor.${index}.type`)} />
                  <Input label="Price" type="number" {...createForm.register(`ticket_constructor.${index}.price`, { valueAsNumber: true })} />
                  <Input label="Rows" type="number" {...createForm.register(`ticket_constructor.${index}.rows_count`, { valueAsNumber: true })} />
                  <Input label="Seats/row" type="number" {...createForm.register(`ticket_constructor.${index}.seats_per_row`, { valueAsNumber: true })} />
                  <Button type="button" variant="ghost" onClick={() => constructorFields.remove(index)}>
                    Remove
                  </Button>
                </div>
              ))}
              <Button type="button" variant="ghost" onClick={() => constructorFields.append({ name: "", type: "", price: 0, rows_count: 0, seats_per_row: 0 })}>
                Add constructor row
              </Button>
            </div>

            <div className={styles.group}>
              <h3>Promos</h3>
              {promosFields.fields.map((field, index) => (
                <div key={field.id} className={styles.inlineGrid}>
                  <Input label="Code" {...createForm.register(`promos.${index}.code`)} />
                  <Input label="Type" {...createForm.register(`promos.${index}.type`)} />
                  <Input label="Value" type="number" {...createForm.register(`promos.${index}.value`, { valueAsNumber: true })} />
                  <Input label="Sector" {...createForm.register(`promos.${index}.sector`)} />
                  <Button type="button" variant="ghost" onClick={() => promosFields.remove(index)}>
                    Remove
                  </Button>
                </div>
              ))}
              <Button type="button" variant="ghost" onClick={() => promosFields.append({ code: "", type: "", value: 0, sector: "" })}>
                Add promo
              </Button>
            </div>

            <div className={styles.group}>
              <h3>Early promos</h3>
              {earlyFields.fields.map((field, index) => (
                <div key={field.id} className={styles.inlineGrid}>
                  <Input label="Code" {...createForm.register(`early.${index}.code`)} />
                  <Input label="Type" {...createForm.register(`early.${index}.type`)} />
                  <Input label="Value" type="number" {...createForm.register(`early.${index}.value`, { valueAsNumber: true })} />
                  <Input label="Sector" {...createForm.register(`early.${index}.sector`)} />
                  <Input label="Valid until" type="datetime-local" {...createForm.register(`early.${index}.valid_until`)} />
                  <Button type="button" variant="ghost" onClick={() => earlyFields.remove(index)}>
                    Remove
                  </Button>
                </div>
              ))}
              <Button type="button" variant="ghost" onClick={() => earlyFields.append({ code: "", type: "", value: 0, sector: "", valid_until: "" })}>
                Add early promo
              </Button>
            </div>

            <div className={styles.group}>
              <h3>Bundles</h3>
              {bundlesFields.fields.map((field, index) => (
                <div key={field.id} className={styles.inlineGrid}>
                  <Input label="Code" {...createForm.register(`bundles.${index}.code`)} />
                  <Input label="Sector" {...createForm.register(`bundles.${index}.sector`)} />
                  <Input label="Buy count" type="number" {...createForm.register(`bundles.${index}.buy_count`, { valueAsNumber: true })} />
                  <Input label="Get count" type="number" {...createForm.register(`bundles.${index}.get_count`, { valueAsNumber: true })} />
                  <Button type="button" variant="ghost" onClick={() => bundlesFields.remove(index)}>
                    Remove
                  </Button>
                </div>
              ))}
              <Button type="button" variant="ghost" onClick={() => bundlesFields.append({ code: "", sector: "", buy_count: 0, get_count: 0 })}>
                Add bundle
              </Button>
            </div>

            <Button type="submit" disabled={createEventMutation.isPending}>
              {createEventMutation.isPending ? "Creating..." : "Create event"}
            </Button>
            {eventStatus && (
              <p
                className={`status ${
                  eventStatus.type === "error"
                    ? "statusError"
                    : eventStatus.type === "success"
                      ? "statusSuccess"
                      : "statusInfo"
                }`}
              >
                {eventStatus.message}
              </p>
            )}
          </form>
        </Card>

        <Card>
          <h2>Update event</h2>
          <form className={styles.form} onSubmit={updateForm.handleSubmit(handleUpdateEvent)}>
            <Input label="Event ID" type="number" {...updateForm.register("event_id", { valueAsNumber: true })} />
            <Input label="Name" {...updateForm.register("name")} />
            <Input label="Date start" type="datetime-local" {...updateForm.register("date_start")} />
            <Input label="Max price coef" type="number" step="0.1" {...updateForm.register("max_price_cof", { valueAsNumber: true })} />
            <Input label="Min price coef" type="number" step="0.1" {...updateForm.register("min_price_cof", { valueAsNumber: true })} />
            <Button type="submit" disabled={updateEventMutation.isPending}>
              {updateEventMutation.isPending ? "Updating..." : "Update event"}
            </Button>
          </form>
        </Card>
      </section>

      <section className={styles.sectionGridThree}>
        <Card>
          <h2>Add promo</h2>
          <form className={styles.form} onSubmit={promoForm.handleSubmit(handleAddPromos)}>
            <Input label="Event ID" type="number" {...promoForm.register("event_id", { valueAsNumber: true })} />

            <div className={styles.group}>
              {addPromoFields.fields.map((field, index) => (
                <div key={field.id} className={styles.inlineGrid}>
                  <Input label="Code" {...promoForm.register(`promos.${index}.code`)} />
                  <Input label="Type" {...promoForm.register(`promos.${index}.type`)} />
                  <Input
                    label="Value"
                    type="number"
                    {...promoForm.register(`promos.${index}.value`, { valueAsNumber: true })}
                  />
                  <Input label="Sector" {...promoForm.register(`promos.${index}.sector`)} />
                  <Button type="button" variant="ghost" onClick={() => addPromoFields.remove(index)}>
                    Remove
                  </Button>
                </div>
              ))}
              <Button
                type="button"
                variant="ghost"
                onClick={() => addPromoFields.append({ code: "", type: "", value: 0, sector: "" })}
              >
                Add promo
              </Button>
            </div>

            <Button type="submit" disabled={addPromosMutation.isPending}>
              {addPromosMutation.isPending ? "Saving..." : "Add promo"}
            </Button>
            {promoStatus && (
              <p
                className={`status ${
                  promoStatus.type === "error"
                    ? "statusError"
                    : promoStatus.type === "success"
                      ? "statusSuccess"
                      : "statusInfo"
                }`}
              >
                {promoStatus.message}
              </p>
            )}
          </form>
        </Card>

        <Card>
          <h2>Add early promo</h2>
          <form className={styles.form} onSubmit={earlyForm.handleSubmit(handleAddEarly)}>
            <Input label="Event ID" type="number" {...earlyForm.register("event_id", { valueAsNumber: true })} />

            <div className={styles.group}>
              {addEarlyFields.fields.map((field, index) => (
                <div key={field.id} className={styles.inlineGrid}>
                  <Input label="Code" {...earlyForm.register(`early.${index}.code`)} />
                  <Input label="Type" {...earlyForm.register(`early.${index}.type`)} />
                  <Input
                    label="Value"
                    type="number"
                    {...earlyForm.register(`early.${index}.value`, { valueAsNumber: true })}
                  />
                  <Input label="Sector" {...earlyForm.register(`early.${index}.sector`)} />
                  <Input label="Valid until" type="datetime-local" {...earlyForm.register(`early.${index}.valid_until`)} />
                  <Button type="button" variant="ghost" onClick={() => addEarlyFields.remove(index)}>
                    Remove
                  </Button>
                </div>
              ))}
              <Button
                type="button"
                variant="ghost"
                onClick={() => addEarlyFields.append({ code: "", type: "", value: 0, sector: "", valid_until: "" })}
              >
                Add early promo
              </Button>
            </div>

            <Button type="submit" disabled={addEarlyMutation.isPending}>
              {addEarlyMutation.isPending ? "Saving..." : "Add early promo"}
            </Button>
            {earlyStatus && (
              <p
                className={`status ${
                  earlyStatus.type === "error"
                    ? "statusError"
                    : earlyStatus.type === "success"
                      ? "statusSuccess"
                      : "statusInfo"
                }`}
              >
                {earlyStatus.message}
              </p>
            )}
          </form>
        </Card>

        <Card>
          <h2>Add bundle promo</h2>
          <form className={styles.form} onSubmit={bundleForm.handleSubmit(handleAddBundle)}>
            <Input label="Event ID" type="number" {...bundleForm.register("event_id", { valueAsNumber: true })} />

            <div className={styles.group}>
              {addBundlesFields.fields.map((field, index) => (
                <div key={field.id} className={styles.inlineGrid}>
                  <Input label="Code" {...bundleForm.register(`bundles.${index}.code`)} />
                  <Input label="Sector" {...bundleForm.register(`bundles.${index}.sector`)} />
                  <Input
                    label="Buy count"
                    type="number"
                    {...bundleForm.register(`bundles.${index}.buy_count`, { valueAsNumber: true })}
                  />
                  <Input
                    label="Get count"
                    type="number"
                    {...bundleForm.register(`bundles.${index}.get_count`, { valueAsNumber: true })}
                  />
                  <Button type="button" variant="ghost" onClick={() => addBundlesFields.remove(index)}>
                    Remove
                  </Button>
                </div>
              ))}
              <Button
                type="button"
                variant="ghost"
                onClick={() => addBundlesFields.append({ code: "", sector: "", buy_count: 0, get_count: 0 })}
              >
                Add bundle
              </Button>
            </div>

            <Button type="submit" disabled={addBundleMutation.isPending}>
              {addBundleMutation.isPending ? "Saving..." : "Add bundle"}
            </Button>
            {bundleStatus && (
              <p
                className={`status ${
                  bundleStatus.type === "error"
                    ? "statusError"
                    : bundleStatus.type === "success"
                      ? "statusSuccess"
                      : "statusInfo"
                }`}
              >
                {bundleStatus.message}
              </p>
            )}
          </form>
        </Card>
      </section>

      <section className={styles.sectionGridThree}>
        <Card>
          <h2>Search promos</h2>
          <form className={styles.form} onSubmit={searchPromoForm.handleSubmit(handleSearchPromos)}>
            <Input
              label="Event ID"
              type="number"
              {...searchPromoForm.register("event_id", {
                setValueAs: (value) => (value === "" ? undefined : Number(value)),
              })}
            />
            <Input label="Limit" type="number" {...searchPromoForm.register("limit", { valueAsNumber: true })} />
            <Input label="Offset" type="number" {...searchPromoForm.register("offset", { valueAsNumber: true })} />
            <Button type="submit" disabled={searchPromosMutation.isPending}>
              {searchPromosMutation.isPending ? "Searching..." : "Search promos"}
            </Button>
            {searchPromoStatus && (
              <p
                className={`status ${
                  searchPromoStatus.type === "error"
                    ? "statusError"
                    : searchPromoStatus.type === "success"
                      ? "statusSuccess"
                      : "statusInfo"
                }`}
              >
                {searchPromoStatus.message}
              </p>
            )}
          </form>
        </Card>

        <Card>
          <h2>Search early promos</h2>
          <form className={styles.form} onSubmit={searchEarlyForm.handleSubmit(handleSearchEarly)}>
            <Input
              label="Event ID"
              type="number"
              {...searchEarlyForm.register("event_id", {
                setValueAs: (value) => (value === "" ? undefined : Number(value)),
              })}
            />
            <Input label="Limit" type="number" {...searchEarlyForm.register("limit", { valueAsNumber: true })} />
            <Input label="Offset" type="number" {...searchEarlyForm.register("offset", { valueAsNumber: true })} />
            <Button type="submit" disabled={searchEarlyMutation.isPending}>
              {searchEarlyMutation.isPending ? "Searching..." : "Search early"}
            </Button>
            {searchEarlyStatus && (
              <p
                className={`status ${
                  searchEarlyStatus.type === "error"
                    ? "statusError"
                    : searchEarlyStatus.type === "success"
                      ? "statusSuccess"
                      : "statusInfo"
                }`}
              >
                {searchEarlyStatus.message}
              </p>
            )}
          </form>
        </Card>

        <Card>
          <h2>Search bundles</h2>
          <form className={styles.form} onSubmit={searchBundleForm.handleSubmit(handleSearchBundle)}>
            <Input
              label="Event ID"
              type="number"
              {...searchBundleForm.register("event_id", {
                setValueAs: (value) => (value === "" ? undefined : Number(value)),
              })}
            />
            <Input label="Limit" type="number" {...searchBundleForm.register("limit", { valueAsNumber: true })} />
            <Input label="Offset" type="number" {...searchBundleForm.register("offset", { valueAsNumber: true })} />
            <Button type="submit" disabled={searchBundleMutation.isPending}>
              {searchBundleMutation.isPending ? "Searching..." : "Search bundles"}
            </Button>
            {searchBundleStatus && (
              <p
                className={`status ${
                  searchBundleStatus.type === "error"
                    ? "statusError"
                    : searchBundleStatus.type === "success"
                      ? "statusSuccess"
                      : "statusInfo"
                }`}
              >
                {searchBundleStatus.message}
              </p>
            )}
          </form>
        </Card>
      </section>
    </div>
  );
};
