import { test, expect } from "@playwright/test";
import { login, deleteAllBookingAttributes } from "../util/helper";

test.beforeEach(async ({ page }) => {
  // Suppress the MFA encouragement modal
  await page.addInitScript(() => {
    window.localStorage.setItem("mfaEncouragementDismissed", "1");
  });

  // Enter credentials and log in
  await login(page, "admin@seatsurfing.local", "Sea!surf1ng");

  // Ensure we've reached the dashboard
  await expect(page).toHaveURL(/search\/$/);
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

test("crud booking attribute", async ({ page }) => {
  const label = "Attribute " + Math.random().toString().substr(2);
  await deleteAllBookingAttributes(page);

  // Navigate to "Settings"
  await page.getByRole("link", { name: "Settings" }).click();
  await expect(page).toHaveURL(/settings\/$/);
  await expect(page.getByText("Loading …")).not.toBeVisible();

  // Add a new attribute
  await page.getByRole("button", { name: "Add attribute" }).click();
  const dialog = page.getByRole("dialog");
  await dialog.getByLabel("Label").fill(label);
  await dialog.getByLabel("Type").selectOption({ label: "Text (multi-line)" });
  await dialog.getByLabel("Required").check();
  await dialog.getByRole("button", { name: "Save" }).click();
  await expect(dialog).toHaveCount(0);

  const row = page.getByRole("row").filter({ hasText: label });
  await expect(row).toHaveCount(1);
  await expect(row).toContainText("Text (multi-line)");

  // Edit the attribute
  await row.getByRole("button", { name: "Edit" }).click();
  await dialog.getByLabel("Label").fill(label + " edited");
  await dialog.getByLabel("Type").selectOption({ label: "Number" });
  await dialog.getByRole("button", { name: "Save" }).click();
  await expect(dialog).toHaveCount(0);

  const editedRow = page
    .getByRole("row")
    .filter({ hasText: label + " edited" });
  await expect(editedRow).toHaveCount(1);
  await expect(editedRow).toContainText("Number");

  // Add attributes up to the limit
  for (let i = 0; i < 2; i++) {
    await page.getByRole("button", { name: "Add attribute" }).click();
    await dialog.getByLabel("Label").fill(label + " " + i);
    await dialog.getByRole("button", { name: "Save" }).click();
    await expect(dialog).toHaveCount(0);
  }
  await expect(
    page.getByRole("button", { name: "Add attribute" }),
  ).toBeDisabled();

  // Delete the attribute
  await editedRow.getByRole("button", { name: "Delete" }).click();
  await page.getByRole("dialog").getByRole("button", { name: "OK" }).click();
  await expect(editedRow).toHaveCount(0);
  await expect(
    page.getByRole("button", { name: "Add attribute" }),
  ).toBeEnabled();

  await deleteAllBookingAttributes(page);
});
