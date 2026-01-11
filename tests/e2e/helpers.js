const crypto = require('crypto');
const { expect } = require('@playwright/test');

const MAILPIT_BASE_URL = process.env.MAILPIT_BASE_URL || 'http://mailpit:8025';
const MAILPIT_WAIT_TIMEOUT_MS = Number.parseInt(process.env.MAILPIT_WAIT_TIMEOUT_MS || '30000', 10);

function buildUser(testInfo, prefix, options = {}) {
  const workerIndex = testInfo && Number.isInteger(testInfo.workerIndex) ? testInfo.workerIndex : 0;
  const rawId = testInfo && testInfo.testId ? testInfo.testId : crypto.randomUUID();
  const safeId = String(rawId).toLowerCase().replace(/[^a-z0-9]/g, '');
  const baseId = safeId.slice(-8) || Date.now().toString(36).slice(-8);
  const safePrefix = String(prefix || 'user').toLowerCase().replace(/[^a-z0-9]/g, '').slice(0, 10) || 'user';
  const base = `${safePrefix}${workerIndex}${baseId}`;
  return {
    username: options.username || base,
    email: options.email || `${base}@test.com`,
    password: options.password || 'Password1',
  };
}

async function register(page, user) {
  await page.goto('/#register', { waitUntil: 'domcontentloaded' });
  await expect(page.getByRole('heading', { name: 'Create your account' })).toBeVisible();
  await page.locator('#register-form #username').fill(user.username);
  await page.locator('#register-form #email').fill(user.email);
  await page.locator('#register-form #password').fill(user.password);
  await page.getByRole('button', { name: 'Create account' }).click();
  await expect(page.getByRole('heading', { name: /Verify your email/i })).toBeVisible();
}

async function loginWithCredentials(page, user) {
  await page.goto('/#login', { waitUntil: 'domcontentloaded' });
  await page.locator('#login-form #email').fill(user.email);
  await page.locator('#login-form #password').fill(user.password);
  await page.getByRole('button', { name: 'Sign in' }).click();
  await expect(page.getByRole('heading', { name: 'Your notes' })).toBeVisible();
}

async function clearMailpit(request) {
  const response = await request.delete(`${MAILPIT_BASE_URL}/api/v1/messages`);
  if (!response.ok()) {
    return;
  }
}

function getMessageId(message) {
  return message.ID || message.id || message.Id || null;
}

function getMessageSubject(message) {
  return message.Subject || message.subject || '';
}

function getMessageRecipients(message) {
  const to = message.To || message.to || message.Recipients || message.recipients || [];
  if (Array.isArray(to)) {
    return to.map((entry) => {
      if (!entry) return '';
      if (typeof entry === 'string') return entry;
      return entry.Address || entry.address || entry.Email || entry.email || entry.Mailbox || '';
    }).filter(Boolean);
  }
  if (typeof to === 'string') {
    return [to];
  }
  return [];
}

function getMessageCreated(message) {
  const created = message.Created || message.created || message.Date || message.date || '';
  const parsed = Date.parse(created);
  return Number.isNaN(parsed) ? 0 : parsed;
}

function pickLatestMessage(messages) {
  if (messages.length === 0) return null;
  const sorted = [...messages].sort((a, b) => getMessageCreated(b) - getMessageCreated(a));
  return sorted[0];
}

function getMessageBody(message) {
  return message.Text || message.text || message.HTML || message.html || message.Body || message.body || '';
}

async function waitForEmail(request, { to, subject, timeout = MAILPIT_WAIT_TIMEOUT_MS } = {}) {
  const start = Date.now();
  const lowerTo = String(to || '').toLowerCase();
  const lowerSubject = subject ? String(subject).toLowerCase() : '';

  while (Date.now() - start < timeout) {
    const response = await request.get(`${MAILPIT_BASE_URL}/api/v1/messages`);
    if (response.ok()) {
      let data = null;
      try {
        data = await response.json();
      } catch (error) {
        data = null;
      }

      const messages = (data && (data.messages || data.Messages || data.items)) || [];
      const filtered = messages.filter((message) => {
        const recipients = getMessageRecipients(message).map((recipient) => recipient.toLowerCase());
        const matchesRecipient = !lowerTo || recipients.some((recipient) => recipient.includes(lowerTo));
        const matchesSubject = !lowerSubject || getMessageSubject(message).toLowerCase().includes(lowerSubject);
        return matchesRecipient && matchesSubject;
      });

      const match = pickLatestMessage(filtered);
      if (match) {
        const messageId = getMessageId(match);
        if (!messageId) {
          throw new Error('Mailpit message missing ID');
        }

        const messageResponse = await request.get(`${MAILPIT_BASE_URL}/api/v1/message/${messageId}`);
        if (messageResponse.ok()) {
          return messageResponse.json();
        }
      }
    }

    await new Promise((resolve) => setTimeout(resolve, 500));
  }

  throw new Error(`Timed out waiting for email${to ? ` to ${to}` : ''}${subject ? ` with subject ${subject}` : ''}`);
}

function extractTokenFromEmail(message, route) {
  const body = getMessageBody(message);
  const tokenMatch = body.match(new RegExp(`#${route}\\?token=([a-f0-9]+)`, 'i'));
  if (!tokenMatch) {
    throw new Error(`Unable to find ${route} token in email`);
  }
  return tokenMatch[1];
}

module.exports = {
  buildUser,
  register,
  loginWithCredentials,
  clearMailpit,
  waitForEmail,
  extractTokenFromEmail,
};
