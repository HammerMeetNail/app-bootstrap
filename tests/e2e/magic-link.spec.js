const { test, expect } = require('@playwright/test');
const {
  buildUser,
  register,
  clearMailpit,
  waitForEmail,
  extractTokenFromEmail,
} = require('./helpers');

test('magic link login', async ({ page, request }, testInfo) => {
  const user = buildUser(testInfo, 'magic');
  await clearMailpit(request);
  await register(page, user);

  const verificationEmail = await waitForEmail(request, {
    to: user.email,
    subject: 'Verify your',
  });
  const verifyToken = extractTokenFromEmail(verificationEmail, 'verify-email');
  await page.goto(`/#verify-email?token=${verifyToken}`);
  await expect(page.getByRole('heading', { name: 'Email verified' })).toBeVisible();

  await page.getByRole('button', { name: 'Sign out' }).click();
  await expect(page.getByRole('heading', { name: 'Capture the ideas that matter.' })).toBeVisible();

  await page.goto('/#login');
  await page.fill('#login-form #email', user.email);
  await page.getByRole('button', { name: 'Email me a magic link' }).click();

  const magicEmail = await waitForEmail(request, {
    to: user.email,
    subject: 'login link',
  });
  const token = extractTokenFromEmail(magicEmail, 'magic-link');

  await page.goto(`/#magic-link?token=${token}`);
  await expect(page.getByRole('heading', { name: 'Your notes' })).toBeVisible();
});
