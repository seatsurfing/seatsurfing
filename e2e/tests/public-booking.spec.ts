import { test, expect, Page } from "@playwright/test";
import * as fs from "fs";
import { login } from "../util/helper";

const uiURL = process.env.UI_URL ? process.env.UI_URL : "http://localhost:8080";

async function authHeaders(page: Page): Promise<{ [key: string]: string }> {
  const accessToken = await page.evaluate(() =>
    window.localStorage.getItem("accessToken"),
  );
  return { Authorization: `Bearer ${accessToken}` };
}

async function setSetting(
  page: Page,
  headers: { [key: string]: string },
  name: string,
  value: string,
): Promise<void> {
  const res = await page.request.put(uiURL + "/setting/" + name, {
    headers,
    data: { value },
  });
  expect(res.status()).toBe(204);
}

test.beforeEach(async ({ page }) => {
  // Suppress the MFA encouragement modal
  await page.addInitScript(() => {
    window.localStorage.setItem("mfaEncouragementDismissed", "1");
  });
  await login(page, "admin@seatsurfing.local", "Sea!surf1ng");
  await expect(page).toHaveURL(/search\/$/);
});

test("public booking select space on map", async ({ page, browser }) => {
  const suffix = Math.random().toString().substring(2);
  const locationName = "Public " + suffix;
  const headers = await authHeaders(page);

  // Set up a location with a floor plan and two public-bookable spaces
  let res = await page.request.post(uiURL + "/group/", {
    headers,
    data: { name: "Approvers " + suffix },
  });
  expect(res.status()).toBe(201);
  const groupId = res.headers()["x-object-id"];
  res = await page.request.post(uiURL + "/location/", {
    headers,
    data: { name: locationName, enabled: true, mapScale: 1 },
  });
  expect(res.status()).toBe(201);
  const locationId = res.headers()["x-object-id"];
  res = await page.request.post(uiURL + "/location/" + locationId + "/map", {
    headers,
    data: fs.readFileSync("../server/res/floorplan.jpg"),
  });
  expect(res.status()).toBe(204);
  for (const [name, x] of [
    ["Public Desk A", 50],
    ["Public Desk B", 250],
  ] as [string, number][]) {
    res = await page.request.post(
      uiURL + "/location/" + locationId + "/space/",
      {
        headers,
        data: {
          name,
          x,
          y: 50,
          width: 100,
          height: 50,
          rotation: 0,
          enabled: true,
          shape: "rect",
          fontSize: "normal",
          approverGroupIds: [groupId],
          publicBookingEnabled: true,
        },
      },
    );
    expect(res.status()).toBe(201);
  }
  await setSetting(page, headers, "public_booking_enabled", "1");
  await setSetting(page, headers, "public_booking_show_map", "1");

  try {
    // Open the public booking page as an anonymous visitor
    const context = await browser.newContext();
    const publicPage = await context.newPage();
    await publicPage.goto(uiURL + "/ui/book/");
    await expect(publicPage.getByText("Loading …")).not.toBeVisible();
    await publicPage
      .getByRole("combobox", { name: "Space" })
      .selectOption({ label: locationName + " / Public Desk A" });
    // The floor plan is hidden by default and opened in a modal
    await expect(publicPage.getByTestId("public-booking-map")).toHaveCount(0);
    await publicPage.getByRole("button", { name: "Floor plan" }).click();
    await expect(publicPage.getByTestId("public-booking-map")).toBeVisible();

    // Select a space by clicking it on the map
    await expect(
      publicPage.getByRole("button", { name: "Public Desk A" }),
    ).toHaveAttribute("aria-pressed", "true");
    await publicPage.getByRole("button", { name: "Public Desk B" }).click();
    await expect(publicPage.getByRole("dialog")).toHaveCount(0);
    await expect(
      publicPage
        .getByRole("combobox", { name: "Space" })
        .locator("option:checked"),
    ).toHaveText(locationName + " / Public Desk B");
    await context.close();

    // Without the setting, the map is not shown
    await setSetting(page, headers, "public_booking_show_map", "0");
    const context2 = await browser.newContext();
    const publicPage2 = await context2.newPage();
    await publicPage2.goto(uiURL + "/ui/book/");
    await expect(publicPage2.getByText("Loading …")).not.toBeVisible();
    await expect(
      publicPage2.getByRole("combobox", { name: "Space" }),
    ).toBeVisible();
    await expect(
      publicPage2.getByRole("button", { name: "Floor plan" }),
    ).toHaveCount(0);
    await context2.close();
  } finally {
    await setSetting(page, headers, "public_booking_show_map", "0");
    await setSetting(page, headers, "public_booking_enabled", "0");
    await page.request.delete(uiURL + "/location/" + locationId, { headers });
    await page.request.delete(uiURL + "/group/" + groupId, { headers });
  }
});
