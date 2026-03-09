import { expect, test } from "@playwright/test";

test("login with existing user (requires env)", async ({ page }) => {
  const identifier = process.env.E2E_USER_IDENTIFIER;
  const password = process.env.E2E_USER_PASSWORD;
  const isEmail = process.env.E2E_USER_IS_EMAIL !== "false";

  test.skip(!identifier || !password, "Set E2E_USER_IDENTIFIER and E2E_USER_PASSWORD to run this test.");

  await page.goto("/auth");

  const emailToggle = page.getByLabel("Login with email");
  if (isEmail) {
    await emailToggle.check();
  } else {
    await emailToggle.uncheck();
  }

  await page.getByLabel("Email or login").fill(identifier ?? "");
  await page.getByLabel("Password").fill(password ?? "");
  await page.getByRole("button", { name: "Sign in" }).click();

  await expect(page.getByText("Signed in successfully.")).toBeVisible();
});
