/* eslint-disable no-console */

const path = require('path');

require(path.join(__dirname, '../api.js'));
require(path.join(__dirname, '../app.js'));

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
    process.exitCode = 1;
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

  test('handles plain text unchanged', () => {
    expect(App.escapeHtml('hello')).toBe('hello');
  });
});

describe('parseHash', () => {
  test('parses simple hash', () => {
    expect(App.parseHash('#home')).toEqual({ route: 'home', params: {} });
  });

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
