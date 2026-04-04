import { test, expect } from "@playwright/test";
import { login } from "../util/helper";

test.beforeEach(async ({ page }) => {
  // Suppress the MFA encouragement modal
  await page.addInitScript(() => {
    window.localStorage.setItem("mfaEncouragementDismissed", "1");
  });

  // Enter credentials and log in
  await login(page, "admin@seatsurfing.local", "Sea!surf1ng");

  // Ensure we've reached the dashboard
  await expect(page).toHaveURL(/search\/$/);
});

test("crud booking", async ({ page }) => {
  await expect(page.getByText("Loading...")).not.toBeVisible();
  await page.getByRole("combobox").selectOption({ label: "Sample Floor" });
  await expect(page.getByText("Loading...")).not.toBeVisible();
  await page.getByText("Desk 1", { exact: true }).click();
  await expect(
    page.getByRole("dialog").getByText("Book a space"),
  ).toBeVisible();
  await page.getByRole("button", { name: "Confirm booking" }).click();
  await expect(
    page.getByRole("dialog").getByText("Your booking has been confirmed!"),
  ).toBeVisible();
  await page.getByRole("button", { name: "My bookings" }).click();
  await expect(page).toHaveURL(/bookings\/$/);
  await expect(page.getByText("Loading...")).not.toBeVisible();
  await page.getByLabel("Calendar", { exact: true }).click(); // switch to list view
  await page
    .getByText(/Sample Floor/)
    .first()
    .click();
  await page.getByRole("button", { name: "Cancel booking" }).click();
  await page.getByText("No bookings.");
  await page.getByRole("link", { name: "Book a space" }).click();
  await expect(page).toHaveURL(/search\/$/);
});
