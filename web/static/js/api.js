// Notes Template - API Client

const API = {
  csrfToken: null,
  retryCount: 0,
  maxRetries: 2,

  async init() {
    await this.fetchCSRFToken();
  },

  async fetchCSRFToken() {
    try {
      const response = await fetch('/api/csrf');
      if (!response.ok) {
        throw new Error('Failed to fetch CSRF token');
      }
      const data = await response.json();
      this.csrfToken = data.token;
    } catch (error) {
      console.error('Failed to fetch CSRF token:', error);
      if (this.retryCount < 1) {
        this.retryCount++;
        await new Promise((resolve) => setTimeout(resolve, 1000));
        return this.fetchCSRFToken();
      }
    }
  },

  async request(method, path, body = null, options = {}) {
    const headers = {
      'Content-Type': 'application/json',
    };

    if (this.csrfToken && ['POST', 'PUT', 'DELETE', 'PATCH'].includes(method)) {
      headers['X-CSRF-Token'] = this.csrfToken;
    }

    const fetchOptions = {
      method,
      headers,
      credentials: 'same-origin',
    };

    if (body && method !== 'GET') {
      fetchOptions.body = JSON.stringify(body);
    }

    const timeout = options.timeout || 30000;
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), timeout);
    fetchOptions.signal = controller.signal;

    try {
      const response = await fetch(path, fetchOptions);
      clearTimeout(timeoutId);

      const contentType = response.headers.get('content-type');
      let data = null;
      if (contentType && contentType.includes('application/json')) {
        const text = await response.text();
        data = text ? JSON.parse(text) : {};
      }

      if (!response.ok) {
        if (response.status === 401) {
          const message = data?.error || 'Session expired. Please log in again.';
          throw new APIError(message, response.status, data);
        }
        if (response.status === 403) {
          const isCSRFError = typeof data?.error === 'string' && data.error.toLowerCase().includes('csrf token');
          if (isCSRFError && !options.retried) {
            await this.fetchCSRFToken();
            return this.request(method, path, body, { ...options, retried: true });
          }
          throw new APIError(data?.error || 'Access denied.', response.status, data);
        }
        if (response.status >= 500) {
          throw new APIError(data?.error || 'Server error. Please try again later.', response.status, data);
        }
        throw new APIError(data?.error || 'Request failed', response.status, data);
      }

      return data;
    } catch (error) {
      clearTimeout(timeoutId);

      if (error.name === 'AbortError') {
        throw new APIError('Request timed out. Please check your connection.', 0);
      }
      if (error instanceof APIError) {
        throw error;
      }
      if (!navigator.onLine) {
        throw new APIError('No internet connection. Please check your network.', 0);
      }
      throw new APIError('Connection error. Please try again.', 0);
    }
  },

  auth: {
    async register(email, password, username) {
      return API.request('POST', '/api/auth/register', {
        email,
        password,
        username,
      });
    },

    async login(email, password) {
      return API.request('POST', '/api/auth/login', { email, password });
    },

    async logout() {
      return API.request('POST', '/api/auth/logout');
    },

    async me() {
      return API.request('GET', '/api/auth/me');
    },

    async changePassword(currentPassword, newPassword) {
      return API.request('POST', '/api/auth/password', {
        current_password: currentPassword,
        new_password: newPassword,
      });
    },

    async verifyEmail(token) {
      return API.request('POST', '/api/auth/verify-email', { token });
    },

    async resendVerification() {
      return API.request('POST', '/api/auth/resend-verification');
    },

    async magicLink(email) {
      return API.request('POST', '/api/auth/magic-link', { email });
    },

    async verifyMagicLink(token) {
      const params = new URLSearchParams({ token });
      return API.request('GET', `/api/auth/magic-link/verify?${params.toString()}`);
    },

    async forgotPassword(email) {
      return API.request('POST', '/api/auth/forgot-password', { email });
    },

    async resetPassword(token, password) {
      return API.request('POST', '/api/auth/reset-password', { token, password });
    },
  },

  notes: {
    async list() {
      return API.request('GET', '/api/notes');
    },

    async create(title, body) {
      return API.request('POST', '/api/notes', { title, body });
    },

    async update(id, title, body) {
      return API.request('PUT', `/api/notes/${id}`, { title, body });
    },

    async remove(id) {
      return API.request('DELETE', `/api/notes/${id}`);
    },
  },
};

class APIError extends Error {
  constructor(message, status, data) {
    super(message);
    this.name = 'APIError';
    this.status = status;
    this.data = data;
  }
}

if (typeof window !== 'undefined') {
  window.API = API;
}
if (typeof global !== 'undefined') {
  global.API = API;
}
