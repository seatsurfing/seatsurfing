import { Page, expect } from "@playwright/test";

const uiURL = process.env.UI_URL ? process.env.UI_URL : "http://localhost:8080";

export async function login(page: Page, email: string, password: string): Promise<void> {
  await page.goto(uiURL + "/ui/login/");
  await expect(page).toHaveURL(/login\/$/);
  await page.getByPlaceholder("you@company.com").fill(email);
  await page
    .locator("form[name='password-login'] input[type='password']")
    .fill(password);
  await page.getByRole("button", { name: "➤" }).click();
}
