const { test, expect } = require('@playwright/test');
const {
  buildUser,
  register,
  clearMailpit,
  waitForEmail,
  extractTokenFromEmail,
} = require('./helpers');

test('sign up, verify email, and manage notes', async ({ page, request }, testInfo) => {
  const user = buildUser(testInfo, 'notes');
  await clearMailpit(request);
  await register(page, user);

  const verificationEmail = await waitForEmail(request, {
    to: user.email,
    subject: 'Verify your',
  });
  const token = extractTokenFromEmail(verificationEmail, 'verify-email');

  await page.goto(`/#verify-email?token=${token}`);
  await expect(page.getByRole('heading', { name: 'Email verified' })).toBeVisible();

  await page.goto('/#app');
  await expect(page.getByRole('heading', { name: 'Your notes' })).toBeVisible();

  await page.fill('#note-title', 'First note');
  await page.fill('#note-body', 'This is a note body.');
  await page.getByRole('button', { name: 'Add note' }).click();
  await expect(page.getByText('First note')).toBeVisible();

  await page.getByRole('button', { name: 'Edit' }).first().click();
  await page.fill('#note-title', 'Updated note');
  await page.getByRole('button', { name: 'Save changes' }).click();
  await expect(page.getByText('Updated note')).toBeVisible();

  await page.fill('#note-title', 'XSS test');
  await page.fill('#note-body', '<img src=x onerror=alert(1)>');
  await page.getByRole('button', { name: 'Add note' }).click();
  await expect(page.getByText('<img src=x onerror=alert(1)>')).toBeVisible();
  await expect(page.locator('.note-item img')).toHaveCount(0);

  const updatedNote = page.getByText('Updated note').locator('..');
  await updatedNote.getByRole('button', { name: 'Delete' }).click();
  await expect(page.getByText('Updated note')).not.toBeVisible();
});
