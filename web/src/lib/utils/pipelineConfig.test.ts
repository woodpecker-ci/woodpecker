import { describe, expect, it } from 'vitest';

import { extractCommandMatchers } from './pipelineConfig';

// Minimal YAML fixture with a commands: block
const simpleYaml = `
steps:
  build:
    image: node:24
    commands:
      - corepack enable
      - pnpm install --frozen-lockfile
      - pnpm build
`;

// YAML with matrix variables in commands
const matrixYaml = `
steps:
  deploy:
    image: alpine
    commands:
      - echo \${VERSION}
      - ./deploy.sh \${TARGET_ENV}
`;

// YAML that contains list items outside commands: (branch filters, paths, etc.)
const noisyYaml = `
steps:
  build:
    image: node:24
    commands:
      - corepack enable
      - pnpm build
    when:
      branch:
        - \${CI_REPO_DEFAULT_BRANCH}
        - release/*
      path:
        - web/**
        - server/api/**
`;

// YAML where a commands: entry is purely a matrix variable (bare \${VAR})
const bareVarYaml = `
steps:
  run:
    image: alpine
    commands:
      - \${DYNAMIC_CMD}
      - echo hello
`;

describe('extractCommandMatchers', () => {
  describe('basic extraction', () => {
    it('returns one matcher per command', () => {
      const matchers = extractCommandMatchers(simpleYaml);
      expect(matchers).toHaveLength(3);
    });

    it('matches exact known commands', () => {
      const matchers = extractCommandMatchers(simpleYaml);
      expect(matchers.some((m) => m.test('corepack enable'))).toBe(true);
      expect(matchers.some((m) => m.test('pnpm install --frozen-lockfile'))).toBe(true);
      expect(matchers.some((m) => m.test('pnpm build'))).toBe(true);
    });

    it('does not match partial command text', () => {
      const matchers = extractCommandMatchers(simpleYaml);
      expect(matchers.some((m) => m.test('pnpm'))).toBe(false);
      expect(matchers.some((m) => m.test('pnpm install'))).toBe(false);
    });

    it('does not match arbitrary strings', () => {
      const matchers = extractCommandMatchers(simpleYaml);
      expect(matchers.some((m) => m.test('@mdi/js 7.4.47'))).toBe(false);
      expect(matchers.some((m) => m.test('@kyvg/vue3-notification 3.4.2'))).toBe(false);
      expect(matchers.some((m) => m.test('ansi_up 6.0.6'))).toBe(false);
    });

    it('returns empty array for empty string', () => {
      expect(extractCommandMatchers('')).toHaveLength(0);
    });

    it('returns empty array when no commands: block present', () => {
      const yaml = `steps:\n  build:\n    image: alpine\n`;
      expect(extractCommandMatchers(yaml)).toHaveLength(0);
    });
  });

  describe('matrix variable handling', () => {
    it('turns \${VAR} into a wildcard that matches any value', () => {
      const matchers = extractCommandMatchers(matrixYaml);
      expect(matchers.some((m) => m.test('echo 1.2.3'))).toBe(true);
      expect(matchers.some((m) => m.test('echo production'))).toBe(true);
    });

    it('wildcard does not match unrelated commands', () => {
      const matchers = extractCommandMatchers(matrixYaml);
      expect(matchers.some((m) => m.test('pnpm build'))).toBe(false);
    });

    it('skips bare \${VAR} entries that would match everything', () => {
      const matchers = extractCommandMatchers(bareVarYaml);
      // Only `echo hello` should survive; the bare ${DYNAMIC_CMD} is dropped
      expect(matchers).toHaveLength(1);
      expect(matchers[0].test('echo hello')).toBe(true);
    });

    it('dropped wildcard-only pattern does not match pnpm dep lines', () => {
      const matchers = extractCommandMatchers(bareVarYaml);
      expect(matchers.some((m) => m.test('@mdi/js 7.4.47'))).toBe(false);
    });
  });

  describe('noise filtering — non-command YAML list items', () => {
    it('does not match branch filter values', () => {
      const matchers = extractCommandMatchers(noisyYaml);
      // ${CI_REPO_DEFAULT_BRANCH} would become ^.*$ without the fix — must not match
      expect(matchers.some((m) => m.test('@mdi/js 7.4.47'))).toBe(false);
      expect(matchers.some((m) => m.test('anything at all'))).toBe(false);
    });

    it('does not match path glob values', () => {
      const matchers = extractCommandMatchers(noisyYaml);
      expect(matchers.some((m) => m.test('web/**'))).toBe(false);
      expect(matchers.some((m) => m.test('server/api/**'))).toBe(false);
      expect(matchers.some((m) => m.test('release/*'))).toBe(false);
    });

    it('still matches the real commands from a noisy config', () => {
      const matchers = extractCommandMatchers(noisyYaml);
      expect(matchers.some((m) => m.test('corepack enable'))).toBe(true);
      expect(matchers.some((m) => m.test('pnpm build'))).toBe(true);
    });
  });

  describe('regex special characters in commands', () => {
    it('handles commands with dots and dashes', () => {
      const yaml = `
steps:
  s:
    image: alpine
    commands:
      - pnpm install --frozen-lockfile
`;
      const matchers = extractCommandMatchers(yaml);
      // Must match exactly, not treat -- as regex
      expect(matchers.some((m) => m.test('pnpm install --frozen-lockfile'))).toBe(true);
      expect(matchers.some((m) => m.test('pnpm installXXfrozen-lockfile'))).toBe(false);
    });

    it('handles commands with parentheses and pipes', () => {
      const yaml = `
steps:
  s:
    image: alpine
    commands:
      - bash -c 'echo (hello) | grep h'
`;
      const matchers = extractCommandMatchers(yaml);
      expect(matchers.some((m) => m.test("bash -c 'echo (hello) | grep h'"))).toBe(true);
      expect(matchers.some((m) => m.test('something else'))).toBe(false);
    });
  });
});
