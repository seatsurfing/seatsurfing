import { Page, expect } from "@playwright/test";

const uiURL = process.env.UI_URL ? process.env.UI_URL : "http://localhost:8080";

export async function login(
  page: Page,
  email: string,
  password: string,
): Promise<void> {
  await page.goto(uiURL + "/ui/login/");
  await expect(page).toHaveURL(/login\/$/);
  await page.getByPlaceholder("you@company.com").fill(email);
  await page
    .locator("form[name='password-login'] input[type='password']")
    .fill(password);
  await page.getByRole("button", { name: "➤" }).click();
}

// Cancels all bookings of the currently logged-in user via the API.
// Used to make tests independent of leftover bookings from a previous
// failed/retried run (e.g. a different browser project sharing the same
// backend within one test run).
export async function cancelAllBookings(page: Page): Promise<void> {
  const accessToken = await page.evaluate(() =>
    window.localStorage.getItem("accessToken"),
  );
  const headers = { Authorization: `Bearer ${accessToken}` };
  const res = await page.request.get(uiURL + "/booking/", { headers });
  const bookings = await res.json();
  for (const booking of bookings) {
    await page.request.delete(uiURL + "/booking/" + booking.id, { headers });
  }
}
