import { test, expect } from '@playwright/test';

test.describe('Dashboard', () => {
  test('should redirect to login if not authenticated', async ({ page }) => {
    await page.goto('/dashboard');
    // Should redirect to login page
    await expect(page).toHaveURL(/\/login/);
  });

  // TODO: Add authenticated tests (requires login setup)
  // test('should display dashboard for authenticated user', async ({ page }) => {
  //   // Login first
  //   await page.goto('/login');
  //   await page.fill('input[type="text"]', 'admin');
  //   await page.fill('input[type="password"]', 'password');
  //   await page.click('button[type="submit"]');
  //
  //   // Navigate to dashboard
  //   await page.goto('/dashboard');
  //   await expect(page.locator('h1')).toContainText('Dashboard');
  // });
});



