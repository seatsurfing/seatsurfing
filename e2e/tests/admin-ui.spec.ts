import { test, expect } from '@playwright/test';

const adminUiURL = process.env.ADMIN_UI_URL ? process.env.ADMIN_UI_URL : 'http://localhost:3000';

test.beforeEach(async ({ page }) => {
  // Open login page
  await page.goto(adminUiURL + '/admin/login');
  await expect(page).toHaveURL(/login$/);

  // Enter credentials
  await page.getByPlaceholder('Email address').fill('admin@seatsurfing.local');
  await page.getByPlaceholder('Password').fill('12345678');
  await page.getByRole('button', { name: '➤' }).click();

  // Ensure we've reached the dashboard
  await expect(page).toHaveURL(/dashboard$/);
});

test('crud location', async ({ page }) => {
  const name = 'Location ' + Math.random().toString().substr(2);

  // Navigate to "Areas"
  await page.getByRole('link', { name: 'Areas' }).click();
  await expect(page).toHaveURL(/locations$/);

  // Add a new area
  await page.getByRole('link', { name: 'Add' }).click();
  await expect(page).toHaveURL(/locations\/add$/);

  // Fill the basic information
  await page.getByPlaceholder('Name').fill(name);
  await page.getByPlaceholder('Description').fill(name);
  await page.locator('#check-limitConcurrentBookings').check();
  await page.getByRole('spinbutton').fill('5');
  await page.locator('input[type="file"]').setInputFiles('../server/res/floorplan.jpg');
  await page.getByRole('button', { name: 'Save' }).click();

  // Add one space
  await page.getByRole('button', { name: 'Add space' }).click();
  await page.locator('.space-dragger').getByRole('textbox').fill('Test 1');

  // Add another space
  await page.getByRole('button', { name: 'Add space' }).click();
  await page.locator('.space-dragger').getByRole('textbox').nth(1).fill('Test 2');

  // Save & go back to area list
  await page.getByRole('button', { name: 'Save' }).click();
  await page.getByRole('link', { name: 'Back' }).click();
  await expect(page).toHaveURL(/locations$/);

  // Re-open area from list
  await page.getByRole('cell', { name: name }).click();
  await expect(page).toHaveURL(/locations\/.+$/);

  // Delete area
  page.on('dialog', dialog => dialog.accept());
  await page.getByRole('button', { name: 'Delete' }).click();

  // Check that area is not included in list anymore
  await expect(page).toHaveURL(/locations$/);
  await expect(page.getByRole('cell', { name: name })).toHaveCount(0);
});
