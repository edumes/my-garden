import { expect, test } from '@playwright/test';

// Each run signs up a new player, so the suite works against a database that
// already has data.
test('a new player signs in, creates a garden, and plants a seed', async ({ page }) => {
  const username = `e2e${Date.now()}`;
  const password = `pw-${username}`;

  await page.goto('/register');
  await page.getByPlaceholder('Choose a username').fill(username);
  await page.getByPlaceholder('your@email.com').fill(`${username}@example.com`);
  await page.getByPlaceholder('Create a password').fill(password);
  await page.getByRole('button', { name: 'Create Account' }).click();
  await expect(page).toHaveURL(/\/login$/);

  await page.getByPlaceholder('Enter your username').fill(username);
  await page.getByPlaceholder('Enter your password').fill(password);
  await page.getByRole('button', { name: 'Sign In' }).click();
  await expect(page).toHaveURL(/\/$/);
  await expect(page.getByRole('button', { name: username })).toBeVisible();

  await page.getByRole('button', { name: 'New Garden' }).first().click();
  await page.getByPlaceholder('My Beautiful Garden').fill('E2E Garden');
  await page.getByRole('dialog').getByRole('button', { name: /Create/ }).click();
  await page.getByText('E2E Garden').click();
  await expect(page).toHaveURL(/\/garden\/[\w-]+$/);

  await page.getByRole('button', { name: 'Store' }).first().click();
  await page.getByRole('dialog').getByRole('button', { name: 'Buy Seeds' }).first().click();
  await expect(page.getByText(/Successfully purchased/)).toBeVisible();
  await page.keyboard.press('Escape');

  // Empty grid cells are plain clickable divs; switch to a role-based locator
  // once they become accessible buttons.
  await page.locator('.aspect-square').first().click();
  const plantDialog = page.getByRole('dialog');
  await expect(plantDialog.getByText('Plant Seed - Position 1')).toBeVisible();
  await plantDialog.locator('h4').first().click();
  await plantDialog.getByRole('button', { name: 'Plant Seed' }).click();
  await expect(page.getByText('Seed planted successfully!')).toBeVisible();
});
