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
  await expect(page).toHaveURL(/search\/(\?.*)?$/);
  await expect(page.getByText("Loading …")).not.toBeVisible();

  // Navigate to "Administration"
  await page.getByRole("link", { name: "Administration" }).click();
  await expect(page).toHaveURL(/admin\/dashboard\/$/);
});

test("crud location", async ({ page }) => {
  const name = "Location " + Math.random().toString().substr(2);

  // Navigate to "Areas"
  await page.getByRole("link", { name: "Areas" }).click();
  await expect(page).toHaveURL(/locations\/$/);

  // Add a new area
  await page.getByRole("link", { name: "Add" }).click();
  await expect(page).toHaveURL(/locations\/add\/$/);

  // Fill the basic information
  await page.getByPlaceholder("Name").fill(name);
  await page.getByPlaceholder("Description").fill(name);
  await page.locator("#check-limitConcurrentBookings").check();
  await page.getByRole("spinbutton").first().fill("5");
  await page
    .locator('input[type="file"]')
    .setInputFiles("../server/res/floorplan.jpg");
  await page.getByRole("button", { name: "Save" }).click();

  // Add one space
  await page.getByRole("button", { name: "Add space" }).click();
  await page.locator(".space-dragger").getByRole("textbox").fill("Test 1");

  // Add another space
  await page.getByRole("button", { name: "Add space" }).click();
  await page
    .locator(".space-dragger")
    .getByRole("textbox")
    .nth(1)
    .fill("Test 2");

  // Save & go back to area list
  await page.getByRole("button", { name: "Save" }).click();
  await expect(page.getByText("Record saved.")).toBeVisible();
  await expect(page).toHaveURL(/locations\/.+\/$/);
  await page.getByRole("link", { name: "Back" }).click();
  await expect(page).toHaveURL(/locations\/$/);

  // Re-open area from list
  await page.getByRole("cell", { name: name }).click();
  await expect(page).toHaveURL(/locations\/.+\/$/);

  // Delete area
  await page.getByRole("button", { name: "Delete" }).click();
  await page.getByRole("dialog").getByRole("button", { name: "OK" }).click();

  // Check that area is not included in list anymore
  await expect(page).toHaveURL(/locations\/$/);
  await expect(page.getByRole("cell", { name: name })).toHaveCount(0);
});

test("space blocked period", async ({ page }) => {
  const name = "Location " + Math.random().toString().substr(2);

  await page.getByRole("link", { name: "Areas" }).click();
  await page.getByRole("link", { name: "Add" }).click();
  await expect(page).toHaveURL(/locations\/add\/$/);
  await page.getByPlaceholder("Name").fill(name);
  await page
    .locator('input[type="file"]')
    .setInputFiles("../server/res/floorplan.jpg");
  await page.getByRole("button", { name: "Save" }).click();

  // Add a space and block it from today on
  await page.getByRole("button", { name: "Add space" }).click();
  await page.locator(".space-dragger").getByRole("textbox").fill("Blocked 1");
  await page
    .locator(".space-dragger")
    .first()
    .click({ position: { x: 5, y: 5 } });
  await page.getByRole("button", { name: "Edit", exact: true }).click();
  const dialog = page.getByRole("dialog");
  await expect(dialog.getByText("Edit space")).toBeVisible();
  await expect(dialog.locator("#space-blocked-from")).toBeDisabled();
  await dialog.getByLabel("Blocked from").check();
  await expect(dialog.locator("#space-blocked-from")).toBeEnabled();
  await expect(dialog.getByLabel("Blocked until")).not.toBeChecked();
  await dialog.getByRole("button", { name: "OK" }).click();
  await page.getByRole("button", { name: "Save" }).click();
  await expect(page.getByText("Record saved.")).toBeVisible();
  await expect(page).toHaveURL(/locations\/.+\/$/);

  // Verify the stored period via the API
  const locationId = page.url().split("/").filter(Boolean).pop();
  const { accessToken, today } = await page.evaluate(() => {
    const d = new Date();
    return {
      accessToken: window.localStorage.getItem("accessToken"),
      today:
        d.getFullYear() +
        "-" +
        String(d.getMonth() + 1).padStart(2, "0") +
        "-" +
        String(d.getDate()).padStart(2, "0"),
    };
  });
  const res = await page.request.get(
    new URL("/location/" + locationId + "/space/", page.url()).toString(),
    { headers: { Authorization: `Bearer ${accessToken}` } },
  );
  const spaces = await res.json();
  expect(spaces).toHaveLength(1);
  expect(spaces[0].blockedFrom).toBe(today);
  expect(spaces[0].blockedUntil).toBe("");

  // Re-open area and delete it
  await page.getByRole("link", { name: "Back" }).click();
  await page.getByRole("cell", { name: name }).click();
  await expect(page).toHaveURL(/locations\/.+\/$/);
  await page.getByRole("button", { name: "Delete" }).click();
  await page.getByRole("dialog").getByRole("button", { name: "OK" }).click();
  await expect(page).toHaveURL(/locations\/$/);
});

test("auth events", async ({ page }) => {
  // Navigate to "Audit"
  await page.getByRole("link", { name: "Audit" }).click();
  await expect(page).toHaveURL(/audit\//);
  await expect(page.getByText("Loading …")).not.toBeVisible();

  // The login from beforeEach must show up as a successful event
  const table = page.locator("#datatable");
  await expect(
    table.getByRole("cell", { name: "admin@seatsurfing.local" }).first(),
  ).toBeVisible();
  await expect(table.getByText("Successful").first()).toBeVisible();

  // Open the details modal
  await table
    .getByRole("cell", { name: "admin@seatsurfing.local" })
    .first()
    .click();
  await expect(page.getByRole("dialog").getByText("Outcome")).toBeVisible();
  await page
    .getByRole("dialog")
    .locator(".modal-footer")
    .getByRole("button", { name: "Close" })
    .click();
  await expect(page.getByRole("dialog")).toHaveCount(0);

  // Filter for failures only
  await page.locator("#outcome-select").selectOption("failure");
  await page.getByRole("button", { name: "Search" }).click();
  await expect(page.getByText("Loading …")).not.toBeVisible();
  await expect(page.locator("#datatable").getByText("Successful")).toHaveCount(
    0,
  );
});

test("filter bookings for next 7 days", async ({ page }) => {
  // Navigate to "Bookings"
  await page.getByRole("link", { name: "Bookings", exact: true }).click();
  await expect(page).toHaveURL(/admin\/bookings\/$/);

  // Select period "Next 7 days" and search
  await page.getByLabel("Period").selectOption({ label: "Next 7 days" });
  await page.getByRole("button", { name: "Search" }).click();
  await expect(page).toHaveURL(/filter=next_7_days/);
  await expect(page.getByText("Loading …")).not.toBeVisible();
  await expect(page.getByLabel("Period")).toHaveValue("next_7_days");
});
