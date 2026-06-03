/* eslint-disable no-template-curly-in-string -- YAML fixtures intentionally contain literal ${VAR} matrix variables */
import { describe, expect, it } from 'vitest';

import { extractCommandMatchers, extractCmdFromTrace } from './pipelineConfig';

// Map form: steps as a keyed object
const mapFormYaml = `
steps:
  build-web:
    image: node:24
    directory: web/
    commands:
      - corepack enable
      - pnpm install --frozen-lockfile
      - pnpm build
  cross-compile:
    image: golang
    commands:
      - make cross-compile-server
`;

// List form: steps as an array with name keys
const listFormYaml = `
steps:
  - name: spellcheck
    image: node:24-alpine
    commands:
      - corepack enable
      - pnpm cspell lint
  - name: lint-editorconfig
    image: editorconfig-checker
`;

// Matrix variables in commands
const matrixYaml = `
steps:
  deploy:
    commands:
      - echo \${VERSION}
      - tag next-\${CI_COMMIT_SHA}
`;

// A step whose only command is a bare matrix variable
const bareVarYaml = `
steps:
  run:
    commands:
      - \${DYNAMIC_CMD}
      - echo hello
`;

// commands given as a single string instead of a list
const singleStringYaml = `
steps:
  build:
    commands: make build
`;

// Config that uses YAML anchors/aliases (as Woodpecker configs do)
const anchorYaml = `
variables:
  - &node_image 'node:24-alpine'
steps:
  build-web:
    image: *node_image
    commands:
      - corepack enable
      - pnpm build
    when: &when
      - event: push
  lint:
    image: *node_image
    commands:
      - pnpm lint
    when: *when
`;

describe('extractCommandMatchers', () => {
  describe('map form', () => {
    it('extracts only the named step’s commands', () => {
      const matchers = extractCommandMatchers(mapFormYaml, 'build-web');
      expect(matchers).toHaveLength(3);
      expect(matchers.some((m) => m.test('corepack enable'))).toBe(true);
      expect(matchers.some((m) => m.test('pnpm build'))).toBe(true);
    });

    it('does not leak commands from other steps', () => {
      const matchers = extractCommandMatchers(mapFormYaml, 'build-web');
      expect(matchers.some((m) => m.test('make cross-compile-server'))).toBe(false);
    });

    it('matches commands exactly, not partially', () => {
      const matchers = extractCommandMatchers(mapFormYaml, 'build-web');
      expect(matchers.some((m) => m.test('pnpm'))).toBe(false);
      expect(matchers.some((m) => m.test('pnpm install'))).toBe(false);
    });

    it('does not match pnpm dependency-output lines', () => {
      const matchers = extractCommandMatchers(mapFormYaml, 'build-web');
      expect(matchers.some((m) => m.test('@mdi/js 7.4.47'))).toBe(false);
      expect(matchers.some((m) => m.test('@kyvg/vue3-notification 3.4.2'))).toBe(false);
      expect(matchers.some((m) => m.test('ansi_up 6.0.6'))).toBe(false);
    });
  });

  describe('list form', () => {
    it('finds a step by its name key', () => {
      const matchers = extractCommandMatchers(listFormYaml, 'spellcheck');
      expect(matchers).toHaveLength(2);
      expect(matchers.some((m) => m.test('pnpm cspell lint'))).toBe(true);
    });

    it('returns empty for a step with no commands', () => {
      expect(extractCommandMatchers(listFormYaml, 'lint-editorconfig')).toHaveLength(0);
    });
  });

  describe('matrix variables', () => {
    it('turns ${VAR} into a wildcard', () => {
      const matchers = extractCommandMatchers(matrixYaml, 'deploy');
      expect(matchers.some((m) => m.test('echo 1.2.3'))).toBe(true);
      expect(matchers.some((m) => m.test('tag next-abc123'))).toBe(true);
    });

    it('wildcard does not match unrelated text', () => {
      const matchers = extractCommandMatchers(matrixYaml, 'deploy');
      expect(matchers.some((m) => m.test('pnpm build'))).toBe(false);
    });

    it('drops bare ${VAR} commands that would match everything', () => {
      const matchers = extractCommandMatchers(bareVarYaml, 'run');
      expect(matchers).toHaveLength(1);
      expect(matchers[0].test('echo hello')).toBe(true);
      expect(matchers[0].test('@mdi/js 7.4.47')).toBe(false);
    });
  });

  describe('command value forms', () => {
    it('accepts a single command string', () => {
      const matchers = extractCommandMatchers(singleStringYaml, 'build');
      expect(matchers).toHaveLength(1);
      expect(matchers[0].test('make build')).toBe(true);
    });
  });

  describe('yaml anchors / aliases', () => {
    it('extracts commands from a step using anchored values', () => {
      const matchers = extractCommandMatchers(anchorYaml, 'build-web');
      expect(matchers.some((m) => m.test('corepack enable'))).toBe(true);
      expect(matchers.some((m) => m.test('pnpm build'))).toBe(true);
    });

    it('still scopes correctly when other steps reference aliases', () => {
      const matchers = extractCommandMatchers(anchorYaml, 'lint');
      expect(matchers).toHaveLength(1);
      expect(matchers[0].test('pnpm lint')).toBe(true);
      expect(matchers.some((m) => m.test('pnpm build'))).toBe(false);
    });
  });

  describe('matrix variables in step names', () => {
    const matrixNameYaml = `
steps:
  - name: build-\${PLATFORM}
    commands:
      - make build
  - name: lint
    commands:
      - pnpm lint
`;

    it('matches an interpolated step name against a ${VAR} config name', () => {
      const matchers = extractCommandMatchers(matrixNameYaml, 'build-linux-amd64');
      expect(matchers.some((m) => m.test('make build'))).toBe(true);
    });

    it('does not pull commands from a non-matching named step', () => {
      const matchers = extractCommandMatchers(matrixNameYaml, 'build-linux-amd64');
      expect(matchers.some((m) => m.test('pnpm lint'))).toBe(false);
    });

    it('an exact name still does not over-match a different step', () => {
      const yaml = `
steps:
  - name: build-web
    commands:
      - pnpm build
  - name: build-server
    commands:
      - make server
`;
      const matchers = extractCommandMatchers(yaml, 'build-web');
      expect(matchers.some((m) => m.test('pnpm build'))).toBe(true);
      expect(matchers.some((m) => m.test('make server'))).toBe(false);
    });

    it('a pure-wildcard step name matches every step (search-all fallback)', () => {
      const yaml = `
steps:
  - name: \${ANYTHING}
    commands:
      - shared-setup
  - name: build
    commands:
      - make build
`;
      // Runtime step "build" should pick up both the wildcard step's commands and its own
      const matchers = extractCommandMatchers(yaml, 'build');
      expect(matchers.some((m) => m.test('shared-setup'))).toBe(true);
      expect(matchers.some((m) => m.test('make build'))).toBe(true);
      // And an unrelated runtime step still gets the wildcard step's commands
      const other = extractCommandMatchers(yaml, 'totally-different');
      expect(other.some((m) => m.test('shared-setup'))).toBe(true);
    });

    it('works for map-form names containing a variable', () => {
      const yaml = `
steps:
  deploy-\${ENV}:
    commands:
      - ./deploy.sh
`;
      const matchers = extractCommandMatchers(yaml, 'deploy-staging');
      expect(matchers.some((m) => m.test('./deploy.sh'))).toBe(true);
    });
  });

  describe('edge cases', () => {
    it('returns empty for an unknown step', () => {
      expect(extractCommandMatchers(mapFormYaml, 'does-not-exist')).toHaveLength(0);
    });

    it('returns empty for empty input', () => {
      expect(extractCommandMatchers('', 'build')).toHaveLength(0);
    });

    it('returns empty (no throw) for invalid YAML', () => {
      expect(extractCommandMatchers('steps:\n  build:\n    commands:\n  -bad: : :', 'build')).toHaveLength(0);
    });

    it('returns empty when there is no steps key', () => {
      expect(extractCommandMatchers('when:\n  - event: push\n', 'build')).toHaveLength(0);
    });

    it('handles regex special characters in commands', () => {
      const yaml = `
steps:
  s:
    commands:
      - pnpm install --frozen-lockfile
      - bash -c 'echo (hi) | grep h'
`;
      const matchers = extractCommandMatchers(yaml, 's');
      expect(matchers.some((m) => m.test('pnpm install --frozen-lockfile'))).toBe(true);
      expect(matchers.some((m) => m.test('pnpm installXXfrozen-lockfile'))).toBe(false);
      expect(matchers.some((m) => m.test("bash -c 'echo (hi) | grep h'"))).toBe(true);
    });
  });
});

describe('extractCmdFromTrace', () => {
  it('strips the + prefix from a plain trace line', () => {
    expect(extractCmdFromTrace('+ corepack enable')).toBe('corepack enable');
  });

  it('strips single quotes added by the Windows local backend', () => {
    expect(extractCmdFromTrace("+ 'net use'")).toBe('net use');
  });

  it('strips double quotes', () => {
    expect(extractCmdFromTrace('+ "net use"')).toBe('net use');
  });

  it('does not strip mismatched quotes', () => {
    expect(extractCmdFromTrace("+ 'net use\"")).toBe("'net use\"");
  });

  it('does not strip a quote that only wraps part of the command', () => {
    expect(extractCmdFromTrace("+ echo 'hello world'")).toBe("echo 'hello world'");
  });

  it('returns null for a non-trace line', () => {
    expect(extractCmdFromTrace('corepack enable')).toBeNull();
    expect(extractCmdFromTrace('Lockfile is up to date')).toBeNull();
    expect(extractCmdFromTrace('+ ')).toBe('');
  });
});
