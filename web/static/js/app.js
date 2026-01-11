// Notes Template - Main Application

const App = {
  user: null,
  notes: [],
  editingNoteId: null,
  _lastHash: '',

  async init() {
    await API.init();
    await this.checkAuth();
    this.setupActionDelegation();
    this.setupNavigation();
    this._lastHash = window.location.hash || '#home';
    this.route();
  },

  qs(id) {
    return document.getElementById(id);
  },

  setText(el, text) {
    if (!el) return;
    el.textContent = text ?? '';
  },

  escapeHtml(text) {
    return String(text)
      .replaceAll('&', '&amp;')
      .replaceAll('<', '&lt;')
      .replaceAll('>', '&gt;')
      .replaceAll('"', '&quot;')
      .replaceAll("'", '&#39;');
  },

  parseHash(hash) {
    const raw = (hash || '').trim();
    if (!raw || raw === '#') {
      return { route: 'home', params: {} };
    }

    const trimmed = raw.startsWith('#') ? raw.slice(1) : raw;
    const [route, query] = trimmed.split('?');
    const params = {};
    if (query) {
      const search = new URLSearchParams(query);
      for (const [key, value] of search.entries()) {
        params[key] = value;
      }
    }

    return { route: route || 'home', params };
  },

  async checkAuth() {
    try {
      const response = await API.auth.me();
      this.user = response.user || null;
    } catch (error) {
      this.user = null;
    }
  },

  setupNavigation() {
    window.addEventListener('hashchange', () => this.route());
    this.renderNav();
  },

  renderNav() {
    const nav = this.qs('nav');
    if (!nav) return;

    if (this.user) {
      nav.innerHTML = `
        <a href="#app" class="nav-link">Notes</a>
        <button class="nav-link nav-button" data-action="logout">Sign out</button>
      `;
    } else {
      nav.innerHTML = `
        <a href="#login" class="nav-link">Sign in</a>
        <a href="#register" class="nav-link nav-link--primary">Create account</a>
      `;
    }
  },

  setupActionDelegation() {
    document.addEventListener('click', (event) => {
      const actionEl = event.target.closest ? event.target.closest('[data-action]') : null;
      if (!actionEl) return;
      const action = actionEl.dataset.action;
      if (!action) return;
      this.handleActionClick(action, actionEl, event);
    });

    document.addEventListener('submit', (event) => {
      const form = event.target.closest ? event.target.closest('form[data-action]') : null;
      if (!form) return;
      const action = form.dataset.action;
      if (!action) return;
      event.preventDefault();
      this.handleActionSubmit(action, form);
    });
  },

  async handleActionClick(action, target) {
    switch (action) {
      case 'logout':
        await this.logout();
        break;
      case 'resend-verification':
        await this.resendVerification();
        break;
      case 'edit-note':
        this.startEditingNote(target.dataset.noteId);
        break;
      case 'delete-note':
        await this.deleteNote(target.dataset.noteId);
        break;
      case 'cancel-edit':
        this.clearNoteForm();
        break;
      default:
        break;
    }
  },

  async handleActionSubmit(action, form) {
    switch (action) {
      case 'register':
        await this.register(form);
        break;
      case 'login':
        await this.login(form);
        break;
      case 'forgot-password':
        await this.forgotPassword(form);
        break;
      case 'reset-password':
        await this.resetPassword(form);
        break;
      case 'save-note':
        await this.saveNote(form);
        break;
      default:
        break;
    }
  },

  route() {
    const { route, params } = this.parseHash(window.location.hash);

    if (['app'].includes(route) && !this.user) {
      window.location.hash = '#login';
      return;
    }

    if (['login', 'register'].includes(route) && this.user) {
      window.location.hash = '#app';
      return;
    }

    switch (route) {
      case 'home':
        this.renderHome();
        break;
      case 'login':
        this.renderLogin();
        break;
      case 'register':
        this.renderRegister();
        break;
      case 'check-email':
        this.renderCheckEmail(params);
        break;
      case 'verify-email':
        this.verifyEmail(params.token);
        break;
      case 'magic-link':
        this.verifyMagicLink(params.token);
        break;
      case 'forgot-password':
        this.renderForgotPassword();
        break;
      case 'reset-password':
        this.renderResetPassword(params.token);
        break;
      case 'app':
        this.renderNotesApp();
        break;
      default:
        this.renderNotFound();
        break;
    }
  },

  renderHome() {
    const container = this.qs('main-container');
    if (!container) return;
    container.innerHTML = `
      <section class="hero">
        <div class="hero-content">
          <p class="eyebrow">Simple, secure notes</p>
          <h1>Capture the ideas that matter.</h1>
          <p class="hero-copy">A lightweight notes app with email-first authentication, built to be forked and customized.</p>
          <div class="hero-actions">
            <a class="button button-primary" href="#register">Create your account</a>
            <a class="button button-ghost" href="#login">Sign in</a>
          </div>
        </div>
        <div class="hero-card">
          <h3>Today</h3>
          <ul>
            <li>Ship the notes MVP</li>
            <li>Review launch checklist</li>
            <li>Plan next sprint</li>
          </ul>
        </div>
      </section>
    `;
  },

  renderLogin() {
    const container = this.qs('main-container');
    if (!container) return;
    container.innerHTML = `
      <section class="auth">
        <div class="card">
          <h2>Welcome back</h2>
          <p class="muted">Sign in with your password or request a magic link.</p>
          <form id="login-form" data-action="login">
            <label>Email
              <input type="email" id="email" name="email" required />
            </label>
            <label>Password
              <input type="password" id="password" name="password" required />
            </label>
            <button class="button button-primary" type="submit">Sign in</button>
          </form>
          <div class="auth-links">
            <button class="button button-ghost" data-action="magic-link">Email me a magic link</button>
            <a href="#forgot-password">Forgot password?</a>
          </div>
        </div>
      </section>
    `;

    const magicButton = container.querySelector('[data-action="magic-link"]');
    if (magicButton) {
      magicButton.addEventListener('click', async () => {
        const email = container.querySelector('#email')?.value?.trim() || '';
        if (!email) {
          this.toast('Enter your email first.');
          return;
        }
        try {
          await API.auth.magicLink(email);
          window.location.hash = `#check-email?type=magic-link&email=${encodeURIComponent(email)}`;
        } catch (error) {
          this.toast(error.message || 'Unable to send magic link.');
        }
      });
    }
  },

  renderRegister() {
    const container = this.qs('main-container');
    if (!container) return;
    container.innerHTML = `
      <section class="auth">
        <div class="card">
          <h2>Create your account</h2>
          <p class="muted">Start with a verified email and a secure password.</p>
          <form id="register-form" data-action="register">
            <label>Name
              <input type="text" id="username" name="username" required minlength="2" maxlength="100" />
            </label>
            <label>Email
              <input type="email" id="email" name="email" required />
            </label>
            <label>Password
              <input type="password" id="password" name="password" required minlength="8" />
            </label>
            <button class="button button-primary" type="submit">Create account</button>
          </form>
          <p class="auth-links">Already have an account? <a href="#login">Sign in</a></p>
        </div>
      </section>
    `;
  },

  renderCheckEmail(params) {
    const container = this.qs('main-container');
    if (!container) return;
    const type = params.type || 'verification';
    const email = params.email ? decodeURIComponent(params.email) : '';
    const titleMap = {
      verification: 'Verify your email',
      'magic-link': 'Check your inbox',
      reset: 'Reset link sent',
    };
    const messageMap = {
      verification: 'We sent you a verification link. Open it to activate your account.',
      'magic-link': 'Open the magic link to sign in without a password.',
      reset: 'Use the reset link to choose a new password.',
    };

    container.innerHTML = `
      <section class="auth">
        <div class="card">
          <h2>${titleMap[type] || 'Check your email'}</h2>
          <p class="muted">${messageMap[type] || 'Check your inbox for the next step.'}</p>
          ${email ? `<p class="pill">${this.escapeHtml(email)}</p>` : ''}
          ${type === 'verification' ? '<button class="button button-ghost" data-action="resend-verification">Resend verification</button>' : ''}
          <a class="button button-primary" href="#login">Back to sign in</a>
        </div>
      </section>
    `;
  },

  async verifyEmail(token) {
    const container = this.qs('main-container');
    if (!container) return;
    if (!token) {
      this.renderNotFound();
      return;
    }
    container.innerHTML = '<div class="loading-state"><div class="spinner"></div><p>Verifying...</p></div>';
    try {
      await API.auth.verifyEmail(token);
      await this.checkAuth();
      this.renderNav();
      container.innerHTML = `
        <section class="auth">
          <div class="card">
            <h2>Email verified</h2>
            <p class="muted">Your email is confirmed. You can keep going now.</p>
            <a class="button button-primary" href="#app">Open notes</a>
          </div>
        </section>
      `;
    } catch (error) {
      container.innerHTML = `
        <section class="auth">
          <div class="card">
            <h2>Verification failed</h2>
            <p class="muted">${this.escapeHtml(error.message || 'Unable to verify email.')}</p>
            <a class="button button-primary" href="#login">Back to sign in</a>
          </div>
        </section>
      `;
    }
  },

  async verifyMagicLink(token) {
    const container = this.qs('main-container');
    if (!container) return;
    if (!token) {
      this.renderNotFound();
      return;
    }
    container.innerHTML = '<div class="loading-state"><div class="spinner"></div><p>Signing you in...</p></div>';
    try {
      const response = await API.auth.verifyMagicLink(token);
      this.user = response.user || null;
      this.renderNav();
      window.location.hash = '#app';
    } catch (error) {
      container.innerHTML = `
        <section class="auth">
          <div class="card">
            <h2>Magic link failed</h2>
            <p class="muted">${this.escapeHtml(error.message || 'Unable to sign in.')}</p>
            <a class="button button-primary" href="#login">Back to sign in</a>
          </div>
        </section>
      `;
    }
  },

  renderForgotPassword() {
    const container = this.qs('main-container');
    if (!container) return;
    container.innerHTML = `
      <section class="auth">
        <div class="card">
          <h2>Reset your password</h2>
          <p class="muted">We will email you a link to reset your password.</p>
          <form id="forgot-password-form" data-action="forgot-password">
            <label>Email
              <input type="email" id="email" name="email" required />
            </label>
            <button class="button button-primary" type="submit">Send reset link</button>
          </form>
          <p class="auth-links"><a href="#login">Back to sign in</a></p>
        </div>
      </section>
    `;
  },

  renderResetPassword(token) {
    const container = this.qs('main-container');
    if (!container) return;
    if (!token) {
      this.renderNotFound();
      return;
    }
    container.innerHTML = `
      <section class="auth">
        <div class="card">
          <h2>Create a new password</h2>
          <p class="muted">Choose a strong password to secure your account.</p>
          <form id="reset-password-form" data-action="reset-password">
            <input type="hidden" id="token" name="token" value="${this.escapeHtml(token)}" />
            <label>New password
              <input type="password" id="password" name="password" required minlength="8" />
            </label>
            <button class="button button-primary" type="submit">Reset password</button>
          </form>
        </div>
      </section>
    `;
  },

  async renderNotesApp() {
    const container = this.qs('main-container');
    if (!container) return;
    container.innerHTML = `
      <section class="notes">
        <div class="notes-header">
          <div>
            <p class="eyebrow">Hello ${this.escapeHtml(this.user?.username || 'there')}</p>
            <h2>Your notes</h2>
          </div>
          <div class="notes-meta">
            <span class="pill">${this.user?.email ? this.escapeHtml(this.user.email) : ''}</span>
          </div>
        </div>
        <div class="notes-grid">
          <div class="card">
            <h3>${this.editingNoteId ? 'Edit note' : 'New note'}</h3>
            <form id="note-form" data-action="save-note">
              <label>Title
                <input type="text" id="note-title" name="title" required maxlength="200" />
              </label>
              <label>Body
                <textarea id="note-body" name="body" rows="6" required maxlength="5000"></textarea>
              </label>
              <div class="form-actions">
                <button class="button button-primary" type="submit">${this.editingNoteId ? 'Save changes' : 'Add note'}</button>
                <button class="button button-ghost" type="button" data-action="cancel-edit">Clear</button>
              </div>
            </form>
          </div>
          <div class="card">
            <div class="notes-list-header">
              <h3>Recent notes</h3>
              <span class="muted" id="notes-count">0 notes</span>
            </div>
            <div id="notes-list" class="notes-list"></div>
          </div>
        </div>
      </section>
    `;

    await this.loadNotes();
    this.renderNotes();
  },

  renderNotFound() {
    const container = this.qs('main-container');
    if (!container) return;
    container.innerHTML = `
      <section class="auth">
        <div class="card">
          <h2>Page not found</h2>
          <p class="muted">That route doesn’t exist. Try the home page.</p>
          <a class="button button-primary" href="#home">Back home</a>
        </div>
      </section>
    `;
  },

  async register(form) {
    const formData = new FormData(form);
    const username = formData.get('username')?.toString().trim();
    const email = formData.get('email')?.toString().trim();
    const password = formData.get('password')?.toString();

    try {
      const response = await API.auth.register(email, password, username);
      this.user = response.user || null;
      this.renderNav();
      window.location.hash = `#check-email?type=verification&email=${encodeURIComponent(email)}`;
    } catch (error) {
      this.toast(error.message || 'Unable to register.');
    }
  },

  async login(form) {
    const formData = new FormData(form);
    const email = formData.get('email')?.toString().trim();
    const password = formData.get('password')?.toString();

    try {
      const response = await API.auth.login(email, password);
      this.user = response.user || null;
      this.renderNav();
      window.location.hash = '#app';
    } catch (error) {
      this.toast(error.message || 'Unable to sign in.');
    }
  },

  async logout() {
    try {
      await API.auth.logout();
    } catch (error) {
      // Ignore
    }
    this.user = null;
    this.notes = [];
    this.renderNav();
    window.location.hash = '#home';
  },

  async resendVerification() {
    try {
      await API.auth.resendVerification();
      this.toast('Verification email sent.');
    } catch (error) {
      this.toast(error.message || 'Unable to resend verification.');
    }
  },

  async forgotPassword(form) {
    const formData = new FormData(form);
    const email = formData.get('email')?.toString().trim();

    try {
      await API.auth.forgotPassword(email);
      window.location.hash = `#check-email?type=reset&email=${encodeURIComponent(email)}`;
    } catch (error) {
      this.toast(error.message || 'Unable to send reset email.');
    }
  },

  async resetPassword(form) {
    const formData = new FormData(form);
    const token = formData.get('token')?.toString();
    const password = formData.get('password')?.toString();

    try {
      await API.auth.resetPassword(token, password);
      await this.checkAuth();
      this.renderNav();
      window.location.hash = '#app';
    } catch (error) {
      this.toast(error.message || 'Unable to reset password.');
    }
  },

  async loadNotes() {
    try {
      const response = await API.notes.list();
      this.notes = response.notes || [];
    } catch (error) {
      this.toast(error.message || 'Unable to load notes.');
    }
  },

  renderNotes() {
    const list = this.qs('notes-list');
    const count = this.qs('notes-count');
    if (!list || !count) return;

    list.innerHTML = '';

    if (this.notes.length === 0) {
      const empty = document.createElement('div');
      empty.className = 'empty-state';
      empty.textContent = 'No notes yet. Write your first one.';
      list.appendChild(empty);
      count.textContent = '0 notes';
      return;
    }

    this.notes.forEach((note) => {
      const item = document.createElement('div');
      item.className = 'note-item';

      const header = document.createElement('div');
      header.className = 'note-header';

      const title = document.createElement('h4');
      title.textContent = note.title;

      const actions = document.createElement('div');
      actions.className = 'note-actions';

      const edit = document.createElement('button');
      edit.className = 'button button-ghost';
      edit.type = 'button';
      edit.dataset.action = 'edit-note';
      edit.dataset.noteId = note.id;
      edit.textContent = 'Edit';

      const del = document.createElement('button');
      del.className = 'button button-ghost';
      del.type = 'button';
      del.dataset.action = 'delete-note';
      del.dataset.noteId = note.id;
      del.textContent = 'Delete';

      actions.appendChild(edit);
      actions.appendChild(del);
      header.appendChild(title);
      header.appendChild(actions);

      const body = document.createElement('p');
      body.textContent = note.body;

      item.appendChild(header);
      item.appendChild(body);

      list.appendChild(item);
    });

    count.textContent = `${this.notes.length} ${this.notes.length === 1 ? 'note' : 'notes'}`;
  },

  startEditingNote(noteId) {
    const note = this.notes.find((item) => item.id === noteId);
    if (!note) return;
    this.editingNoteId = note.id;

    const titleInput = this.qs('note-title');
    const bodyInput = this.qs('note-body');
    if (titleInput) titleInput.value = note.title;
    if (bodyInput) bodyInput.value = note.body;

    const form = this.qs('note-form');
    if (form) {
      const button = form.querySelector('button[type="submit"]');
      if (button) button.textContent = 'Save changes';
    }
  },

  clearNoteForm() {
    this.editingNoteId = null;
    const titleInput = this.qs('note-title');
    const bodyInput = this.qs('note-body');
    if (titleInput) titleInput.value = '';
    if (bodyInput) bodyInput.value = '';
    const form = this.qs('note-form');
    if (form) {
      const button = form.querySelector('button[type="submit"]');
      if (button) button.textContent = 'Add note';
    }
  },

  async saveNote(form) {
    const formData = new FormData(form);
    const title = formData.get('title')?.toString().trim();
    const body = formData.get('body')?.toString().trim();

    if (!title || !body) {
      this.toast('Title and body are required.');
      return;
    }

    try {
      if (this.editingNoteId) {
        const response = await API.notes.update(this.editingNoteId, title, body);
        const updated = response.note;
        this.notes = this.notes.map((note) => (note.id === updated.id ? updated : note));
        this.toast('Note updated.');
      } else {
        const response = await API.notes.create(title, body);
        this.notes = [response.note, ...this.notes];
        this.toast('Note added.');
      }
      this.clearNoteForm();
      this.renderNotes();
    } catch (error) {
      this.toast(error.message || 'Unable to save note.');
    }
  },

  async deleteNote(noteId) {
    if (!noteId) return;
    try {
      await API.notes.remove(noteId);
      this.notes = this.notes.filter((note) => note.id !== noteId);
      this.renderNotes();
      this.toast('Note deleted.');
    } catch (error) {
      this.toast(error.message || 'Unable to delete note.');
    }
  },

  toast(message) {
    const container = this.qs('toast-container');
    if (!container) return;

    const toast = document.createElement('div');
    toast.className = 'toast';
    toast.textContent = message;

    container.appendChild(toast);
    setTimeout(() => toast.classList.add('toast--visible'), 10);
    setTimeout(() => {
      toast.classList.remove('toast--visible');
      setTimeout(() => toast.remove(), 300);
    }, 2800);
  },
};

if (typeof window !== 'undefined') {
  window.App = App;
}
if (typeof global !== 'undefined') {
  global.App = App;
}

if (typeof document !== 'undefined') {
  document.addEventListener('DOMContentLoaded', () => {
    App.init();
  });
}
