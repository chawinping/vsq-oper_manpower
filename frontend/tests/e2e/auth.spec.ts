import { test, expect } from '@playwright/test';

test.describe('Authentication', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to login page
    await page.goto('/login');
  });

  test('should display login form', async ({ page }) => {
    await expect(page.locator('h1')).toContainText('Login');
    await expect(page.locator('input[type="text"]')).toBeVisible();
    await expect(page.locator('input[type="password"]')).toBeVisible();
  });

  test('should show error on invalid credentials', async ({ page }) => {
    // Fill in invalid credentials
    await page.fill('input[type="text"]', 'invaliduser');
    await page.fill('input[type="password"]', 'wrongpassword');
    await page.click('button[type="submit"]');

    // Should show error message
    await expect(page.locator('text=/error/i')).toBeVisible();
  });

  // TODO: Add test for successful login (requires backend to be running)
  // test('should login successfully with valid credentials', async ({ page }) => {
  //   await page.fill('input[type="text"]', 'admin');
  //   await page.fill('input[type="password"]', 'password');
  //   await page.click('button[type="submit"]');
  //   await expect(page).toHaveURL('/dashboard');
  // });
});






