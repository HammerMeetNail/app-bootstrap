const { test, expect } = require('@playwright/test');
const {
  buildUser,
  register,
  clearMailpit,
  waitForEmail,
  extractTokenFromEmail,
} = require('./helpers');

test('password reset flow', async ({ page, request }, testInfo) => {
  const user = buildUser(testInfo, 'reset');
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

  await page.goto('/#forgot-password');
  await page.fill('#forgot-password-form #email', user.email);
  await page.getByRole('button', { name: 'Send reset link' }).click();
  await expect(page.getByRole('heading', { name: 'Reset link sent' })).toBeVisible();

  const resetEmail = await waitForEmail(request, {
    to: user.email,
    subject: 'Reset your',
  });
  const resetToken = extractTokenFromEmail(resetEmail, 'reset-password');

  await page.goto(`/#reset-password?token=${resetToken}`);
  await page.fill('#reset-password-form #password', 'NewPassword1');
  await page.getByRole('button', { name: 'Reset password' }).click();

  await expect(page.getByRole('heading', { name: 'Your notes' })).toBeVisible();
});
