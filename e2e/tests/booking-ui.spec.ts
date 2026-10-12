import { test, expect } from "@playwright/test";
import { login, cancelAllBookings } from "../util/helper";

test.beforeEach(async ({ page }) => {
  // Suppress the MFA encouragement modal
  await page.addInitScript(() => {
    window.localStorage.setItem("mfaEncouragementDismissed", "1");
  });

  // Enter credentials and log in
  await login(page, "admin@seatsurfing.local", "Sea!surf1ng");

  // Ensure we've reached the dashboard
  await expect(page).toHaveURL(/search\/(\?.*)?$/);

  // Start with a clean slate: a previous failed/retried run (or another
  // browser project sharing the same backend) may have left Desk 1 booked,
  // which would make this test fail before it even gets to clean up itself.
  await cancelAllBookings(page);
});

test("crud booking", async ({ page }) => {
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
  await page.getByLabel("Calendar", { exact: true }).click(); // switch to list view
  await page
    .getByText(/Sample Floor/)
    .first()
    .click();
  await page.getByRole("button", { name: "Cancel booking" }).click();
  await expect(page.getByText("No bookings.")).toBeVisible();
  await page.getByRole("link", { name: "Book a space" }).click();
  await expect(page).toHaveURL(/search\/$/);
});

test("direct link with multiple spaces and enter=now", async ({ page }) => {
  const uiURL = process.env.UI_URL
    ? process.env.UI_URL
    : "http://localhost:8080";
  const accessToken = await page.evaluate(() =>
    window.localStorage.getItem("accessToken"),
  );
  const headers = { Authorization: `Bearer ${accessToken}` };
  const locations = await (
    await page.request.get(uiURL + "/location/", { headers })
  ).json();
  const location = locations.find((l: any) => l.name === "Sample Floor");
  const spaces = await (
    await page.request.get(uiURL + `/location/${location.id}/space/`, {
      headers,
    })
  ).json();
  const desk1 = spaces.find((s: any) => s.name === "Desk 1");
  const desk2 = spaces.find((s: any) => s.name === "Desk 2");

  await page.goto(
    `${uiURL}/ui/search/?lid=${location.id}&sid=${desk1.id},${desk2.id}&enter=now`,
  );
  const dialog = page.getByRole("dialog");
  await expect(dialog.getByText("Book a space")).toBeVisible({
    timeout: 20000,
  });

  // The user can choose between the linked spaces only
  const spaceSelect = dialog.getByLabel("Space");
  await expect(spaceSelect.getByRole("option")).toHaveText([
    "Please select...",
    "Desk 1",
    "Desk 2",
  ]);
  await spaceSelect.selectOption({ label: "Desk 2" });

  // Enter/leave times can be adjusted in the dialog
  await expect(dialog.getByTitle("Previous day")).toBeVisible();
  await expect(dialog.getByTitle("Next day")).toBeVisible();
  await expect(dialog.getByTitle("All day")).toBeVisible();

  await dialog.getByRole("button", { name: "Confirm booking" }).click();
  await expect(
    dialog.getByText("Your booking has been confirmed!"),
  ).toBeVisible();

  const bookings = await (
    await page.request.get(uiURL + "/booking/", { headers })
  ).json();
  expect(bookings.length).toBe(1);
  expect(bookings[0].space.id).toBe(desk2.id);
});

test("direct link reloads availability when the date changes", async ({
  page,
}) => {
  const uiURL = process.env.UI_URL
    ? process.env.UI_URL
    : "http://localhost:8080";
  const accessToken = await page.evaluate(() =>
    window.localStorage.getItem("accessToken"),
  );
  const headers = { Authorization: `Bearer ${accessToken}` };
  const locations = await (
    await page.request.get(uiURL + "/location/", { headers })
  ).json();
  const location = locations.find((l: any) => l.name === "Sample Floor");
  const spaces = await (
    await page.request.get(uiURL + `/location/${location.id}/space/`, {
      headers,
    })
  ).json();
  const desk1 = spaces.find((s: any) => s.name === "Desk 1");
  const desk2 = spaces.find((s: any) => s.name === "Desk 2");

  // Book Desk 1 for 10:00-11:00 on a day two days ahead
  const day = new Date();
  day.setDate(day.getDate() + 2);
  const date = [
    day.getFullYear(),
    String(day.getMonth() + 1).padStart(2, "0"),
    String(day.getDate()).padStart(2, "0"),
  ].join("-");
  const created = await page.request.post(uiURL + "/booking/", {
    headers,
    data: {
      spaceId: desk1.id,
      enter: `${date}T10:00:00.000Z`,
      leave: `${date}T10:59:59.000Z`,
      subject: "",
    },
  });
  expect(created.status()).toBe(201);

  await page.goto(
    `${uiURL}/ui/search/?lid=${location.id}&sid=${desk1.id},${desk2.id}&enter=${date}T10:00&leave=${date}T11:00`,
  );
  const dialog = page.getByRole("dialog");
  await expect(dialog.getByText("Book a space")).toBeVisible({
    timeout: 20000,
  });

  // Desk 1 is taken: it is disabled and the available Desk 2 is preselected
  const spaceSelect = dialog.getByLabel("Space");
  await expect(spaceSelect.getByRole("option")).toHaveText([
    "Please select...",
    "Desk 1 (not available)",
    "Desk 2",
  ]);
  await expect(spaceSelect).toHaveValue(desk2.id);

  // Moving to the next day reloads availability: Desk 1 is free again and
  // the current choice is kept
  await dialog.getByTitle("Next day").click();
  await expect(spaceSelect.getByRole("option")).toHaveText([
    "Please select...",
    "Desk 1",
    "Desk 2",
  ]);
  await expect(spaceSelect).toHaveValue(desk2.id);
  await expect(
    dialog.getByRole("button", { name: "Confirm booking" }),
  ).toBeEnabled();

  await cancelAllBookings(page);
});
