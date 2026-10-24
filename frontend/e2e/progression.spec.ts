import { expect, test } from '@playwright/test';

test('real harvest XP unlocks the next seed tier and persists on the progress page', async ({ page, request }) => {
  test.setTimeout(90_000);
  const base = `${test.info().project.use.baseURL}/api/v1`;
  const username = `prog${Date.now()}`;
  const registration = await request.post(`${base}/auth/register`, { data: { username, email: `${username}@example.com`, password: 'test-password' } });
  expect(registration.status()).toBe(201);
  const account = await registration.json();
  const headers = { Authorization: `Bearer ${account.token}` };
  const created = await request.post(`${base}/gardens`, { headers, data: { name: 'Level Up Patch' } });
  expect(created.status()).toBe(201);
  const { garden } = await created.json();
  const catalog = await request.get(`${base}/plants`);
  const { plant_types } = await catalog.json();
  const seed = plant_types.find((plant: { name: string }) => plant.name === 'Lettuce');
  const purchase = await request.post(`${base}/store/buy`, { headers, data: { plant_type_id: seed.id, quantity: 4 } });
  expect(purchase.ok()).toBe(true);
  // Earn 96 XP through the same commands used by players, without editing data.
  for (let cycle = 0; cycle < 3; cycle++) {
    const planted = await request.post(`${base}/gardens/${garden.id}/plants`, { headers, data: { plant_type_id: seed.id, position: 0 } });
    expect(planted.status()).toBe(201);
    const { plant } = await planted.json();
    await expect.poll(async () => {
      const response = await request.get(`${base}/gardens/${garden.id}`, { headers });
      return (await response.json()).garden.plants[0].stage;
    }, { timeout: 15_000 }).toBe('harvestable');
    const harvested = await request.post(`${base}/gardens/${garden.id}/plants/${plant.id}/harvest`, { headers });
    expect(harvested.ok()).toBe(true);
  }
  await page.addInitScript(token => localStorage.setItem('auth_token', token), account.token);
  await page.goto('/profile');
  const progress = page.getByRole('region', { name: 'Level progress', exact: true });
  await expect(progress).toContainText('4 XP to level 2');
  await expect(page.getByRole('region', { name: 'Garden unlocks', exact: true })).toContainText('level 3');
  await expect(page.getByRole('article', { name: 'Level 2 seed unlocks', exact: true })).toContainText('Reach level 2');
  await page.goto(`/garden/${garden.id}`);
  await page.getByRole('button', { name: 'Plant seed in position 1', exact: true }).click();
  await page.getByRole('dialog').getByRole('radio').first().click();
  await page.getByRole('dialog').getByRole('button', { name: 'Plant Seed', exact: true }).click();
  await expect(page.getByRole('button', { name: 'Harvest plant', exact: true })).toBeVisible({ timeout: 15_000 });
  await page.getByRole('button', { name: 'Harvest plant', exact: true }).click();
  await expect(page.getByText(/Level up! You're now level 2/)).toBeVisible();
  await page.getByRole('link', { name: 'Progress', exact: true }).click();
  await expect(progress).toContainText('172 XP to level 3');
  await expect(page.getByRole('article', { name: 'Level 2 seed unlocks', exact: true })).toContainText('Available now');
  await page.reload();
  await expect(progress).toContainText('128 total XP earned');
  await page.evaluate('window.scrollTo(0, 0)');
  await page.screenshot({ path: test.info().outputPath('progress-desktop.png'), fullPage: true, animations: 'disabled' });

  await page.route('**/api/v1/plants', route => route.abort());
  await page.reload();
  await expect(page.getByRole('alert')).toContainText('Unable to load seed milestones');
  await expect(progress).toContainText('128 total XP earned');
  await page.unroute('**/api/v1/plants');
  await page.getByRole('button', { name: 'Retry milestones', exact: true }).click();
  await expect(page.getByRole('article', { name: 'Level 2 seed unlocks', exact: true })).toContainText('Available now');
  await page.evaluate("localStorage.setItem('darkMode', 'true')");
  await page.setViewportSize({ width: 390, height: 844 });
  await page.reload();
  await expect(progress).toContainText('172 XP to level 3');
  await expect(page.locator('html')).toHaveClass(/dark/);
  await expect.poll(() => page.evaluate('document.documentElement.scrollWidth <= window.innerWidth')).toBe(true);
  await page.evaluate('window.scrollTo(0, 0)');
  await page.screenshot({ path: test.info().outputPath('progress-mobile-dark.png'), fullPage: true, animations: 'disabled' });
});
