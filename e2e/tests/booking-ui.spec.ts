import { test, expect } from '@playwright/test';

const bookingUiUrl = process.env.BOOKING_UI_URL ? process.env.BOOKING_UI_URL : 'http://localhost:3001';

test.beforeEach(async ({ page }) => {
  // Open login page
  await page.goto(bookingUiUrl + '/ui/login');
  await expect(page).toHaveURL(/login$/);

  // Enter credentials
  await page.getByPlaceholder('you@company.com').fill('admin@seatsurfing.local');
  await page.getByPlaceholder('Password').fill('12345678');
  await page.getByRole('button', { name: '➤' }).click();

  // Ensure we've reached the dashboard
  await expect(page).toHaveURL(/search$/);
});

test('crud booking', async ({ page }) => {
  await page.getByRole('combobox').selectOption({label: 'Sample Floor'});
  await page.getByText('Desk 1', { exact: true }).click();
  await page.getByRole('button', { name: 'Confirm booking' }).click();
  await page.getByRole('button', { name: 'My bookings' }).click();
  await expect(page).toHaveURL(/bookings$/);
  await page.getByText(/Sample Floor/).click();
  await page.getByRole('button', { name: 'Cancel booking' }).click();
  await page.getByText('No bookings.');
  await page.getByRole('link', { name: 'Book a space' }).click();
  await expect(page).toHaveURL(/search$/);
});
