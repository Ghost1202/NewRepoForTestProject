import { expect, test } from "@playwright/test";

test("home loads and events are visible", async ({ page, request, baseURL }) => {
  const apiUrl = `${baseURL}/api/searching/events?limit=3&offset=0`;
  const response = await request.get(apiUrl);
  expect(response.ok()).toBeTruthy();

  const data = (await response.json()) as { events?: unknown[] };
  const eventsCount = data.events?.length ?? 0;
  expect(eventsCount).toBeGreaterThan(0);

  await page.goto("/");
  await page.waitForResponse((resp) => resp.url().includes("/api/searching/events") && resp.status() === 200);

  const cards = page.getByTestId("event-card");
  await expect(cards.first()).toBeVisible();
});
