import { expect, test } from '@playwright/test';

// Each run signs up a new player, so the suite works against a database that
// already has data.
test('a new player completes the garden loop and keeps rewards after reload', async ({ page }) => {
  test.setTimeout(120_000);
  const username = `e2e${Date.now()}`;
  const password = `pw-${username}`;

  await page.goto('/register');
  await page.getByPlaceholder('Choose a username').fill(username);
  await page.getByPlaceholder('your@email.com').fill(`${username}@example.com`);
  await page.getByPlaceholder('Create a password').fill(password);
  await page.getByRole('button', { name: 'Create Account' }).click();
  await expect(page).toHaveURL(/\/$/);
  await expect(page.getByRole('button', { name: username })).toBeVisible();

  await expect(page.getByRole('region', { name: 'Your adventure' })).toContainText('First quest: plant a seed');
  await page.getByRole('button', { name: 'New Garden' }).first().click();
  await page.getByPlaceholder('My Beautiful Garden').fill('E2E Garden');
  await page.getByRole('dialog').getByRole('button', { name: /Create/ }).click();
  await expect(page).toHaveURL(/\/garden\/[\w-]+$/);

  await expect(page.getByRole('region', { name: 'First planting quest' })).toContainText('Garden created!');
  await page.getByRole('button', { name: 'Buy first seeds' }).click();
  const purchaseResponse = page.waitForResponse(r => r.url().endsWith('/store/buy') && r.request().method() === 'POST');
  await page.getByRole('dialog').getByRole('button', { name: 'Buy Seeds' }).first().click();
  const purchase = await (await purchaseResponse).json();
  await expect(page.getByText(/Successfully purchased/)).toBeVisible();
  await page.keyboard.press('Escape');
  await expect(page.getByRole('region', { name: 'First planting quest' })).toContainText('Seeds ready!');

  await page.getByRole('button', { name: 'Plant seed in position 1', exact: true }).click();
  const plantDialog = page.getByRole('dialog');
  await expect(plantDialog.getByText('Plant Seed - Position 1')).toBeVisible();
  await plantDialog.getByRole('radio').first().click();
  await plantDialog.getByRole('button', { name: 'Plant Seed' }).click();
  await expect(page.getByText('Seed planted successfully!')).toBeVisible();
  await expect(page.getByRole('region', { name: 'First planting quest' })).toHaveCount(0);
  await expect(page.getByText(/Growth rate:/)).toBeVisible();
  await expect(page.getByRole('button', { name: 'Harvest plant', exact: true })).toBeVisible({ timeout: 90_000 });
  await page.setViewportSize({ width: 390, height: 844 });
  await page.screenshot({ path: test.info().outputPath('garden-ripe-mobile.png'), fullPage: true, animations: 'disabled' });
  await expect.poll(() => page.evaluate('document.documentElement.scrollWidth <= window.innerWidth')).toBe(true);
  const harvestResponse = page.waitForResponse(r => r.url().endsWith('/harvest') && r.request().method() === 'POST');
  await page.getByRole('button', { name: 'Harvest plant', exact: true }).click();
  const { reward } = await (await harvestResponse).json();
  expect(reward.coins).toBeGreaterThan(purchase.total_cost);
  expect(reward.experience).toBeGreaterThan(0);
  await expect(page.getByText(/Harvested!/)).toBeVisible();
  await expect(page.getByRole('button', { name: 'Plant seed in position 1', exact: true })).toBeVisible();
  const stats = page.getByText('Gardener Stats').locator('..').locator('..');
  await expect(stats).toContainText(`${reward.experience + 2} XP`);
  await expect(stats.getByText(String(purchase.user_coins + reward.coins), { exact: true })).toBeVisible();
  const rewardedStats = await stats.innerText();
  await page.reload();
  await expect.poll(() => stats.innerText()).toBe(rewardedStats);
  await expect(page.getByRole('button', { name: 'Plant seed in position 1', exact: true })).toBeVisible();
  await page.getByRole('button', { name: 'Plant seed in position 1', exact: true }).click();
  await expect(page.getByRole('dialog').getByText("You don't have any seeds!")).toBeVisible();
  await page.keyboard.press('Escape');

  // A failed refresh keeps the last state visible and recovers on retry.
  const gardenUrl = `**/api/v1/gardens/${page.url().split('/').pop()}`;
  await page.route(gardenUrl, route => route.abort());
  await expect(page.getByRole('alert')).toContainText('Garden updates unavailable');
  await expect(page.getByRole('button', { name: 'Plant seed in position 1', exact: true })).toBeDisabled();
  await page.unroute(gardenUrl);
  await page.getByRole('button', { name: 'Retry', exact: true }).click();
  await expect(page.getByRole('alert')).toHaveCount(0);
  await expect(page.getByRole('button', { name: 'Plant seed in position 1', exact: true })).toBeEnabled();
  await page.setViewportSize({ width: 1280, height: 720 });
  await page.screenshot({ path: test.info().outputPath('garden-desktop.png'), fullPage: true, animations: 'disabled' });
  await page.setViewportSize({ width: 390, height: 844 });
  await expect(page.getByRole('button', { name: 'Plant seed in position 1', exact: true })).toBeVisible();
  await page.screenshot({ path: test.info().outputPath('garden-mobile.png'), fullPage: true, animations: 'disabled' });
});
