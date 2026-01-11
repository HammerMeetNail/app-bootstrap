(function () {
  function describe(name, fn) {
    console.log(`\n${name}`);
    fn();
  }

  function test(name, fn) {
    try {
      fn();
      console.log(`  ✓ ${name}`);
    } catch (error) {
      console.error(`  ✗ ${name}`);
      console.error(`    ${error.message}`);
    }
  }

  function expect(value) {
    return {
      toBe(expected) {
        if (value !== expected) {
          throw new Error(`Expected ${value} to be ${expected}`);
        }
      },
      toEqual(expected) {
        const actual = JSON.stringify(value);
        const target = JSON.stringify(expected);
        if (actual !== target) {
          throw new Error(`Expected ${actual} to equal ${target}`);
        }
      },
    };
  }

  describe('escapeHtml', () => {
    test('escapes special characters', () => {
      expect(App.escapeHtml('<script>')).toBe('&lt;script&gt;');
    });
  });

  describe('parseHash', () => {
    test('parses hash params', () => {
      expect(App.parseHash('#reset-password?token=abc')).toEqual({
        route: 'reset-password',
        params: { token: 'abc' },
      });
    });
  });

  describe('API Client Structure', () => {
    test('auth namespace exists', () => {
      expect(typeof API.auth).toBe('object');
    });

    test('notes namespace exists', () => {
      expect(typeof API.notes).toBe('object');
    });
  });
})();
