import { test, expect } from "@playwright/test";
import { login } from "../util/helper";

test.beforeEach(async ({ page }) => {
  await page.addInitScript(() => {
    window.localStorage.setItem("mfaEncouragementDismissed", "1");
  });
  await login(page, "admin@seatsurfing.local", "Sea!surf1ng");
  await expect(page).toHaveURL(/search\/$/);
});

test("next event button jumps to booking week", async ({ page }) => {
  await expect(page.getByText("Loading …")).not.toBeVisible();
  await page.getByRole("combobox").selectOption({ label: "Sample Floor" });
  await expect(page.getByText("Loading …")).not.toBeVisible();
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
  await expect(page.getByText("Loading …")).not.toBeVisible();

  // Calendar view should be visible by default with the booking in the current week
  const toolbar = page.locator(".custom-toolbar");
  await expect(toolbar).toBeVisible();
  const nextEventBtn = toolbar.getByRole("link", { name: /next event/i });
  await expect(nextEventBtn).toBeVisible();

  await page.screenshot({ path: "test-results/next-event-initial.png" });

  // Navigate several weeks forward, away from the booking
  const nextWeekBtn = toolbar.locator("a").nth(2); // Today, Prev, Next
  for (let i = 0; i < 4; i++) {
    await nextWeekBtn.click();
  }
  const labelAfterNav = await page.locator(".toolbar-label").textContent();
  console.log("Label after navigating forward:", labelAfterNav);

  await page.screenshot({ path: "test-results/next-event-away.png" });

  // Click "Next event" - since we're past the only booking, it should wrap to it
  await nextEventBtn.click();
  await page.waitForTimeout(300);
  const labelAfterJump = await page.locator(".toolbar-label").textContent();
  console.log("Label after clicking Next event:", labelAfterJump);

  await page.screenshot({ path: "test-results/next-event-after-jump.png" });

  expect(labelAfterJump).not.toEqual(labelAfterNav);

  // Cleanup: cancel the booking we created
  await page.getByLabel("Calendar", { exact: true }).click();
  await page
    .getByText(/Sample Floor/)
    .first()
    .click();
  await page.getByRole("button", { name: "Cancel booking" }).click();
});
