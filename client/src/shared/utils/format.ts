export const formatDateTime = (value?: string) => {
  if (!value) {
    return "N/A";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return new Intl.DateTimeFormat("en-US", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
};

export const formatMoney = (value: number, currency = "UAH") => {
  if (!Number.isFinite(value)) {
    return "-";
  }
  const normalized = value / 100;
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency,
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(normalized);
};

export const toRfc3339 = (value: string) => {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toISOString();
};

export const toRating = (value: number) => {
  if (!Number.isFinite(value)) {
    return "0.0";
  }
  return (value / 10).toFixed(1);
};
