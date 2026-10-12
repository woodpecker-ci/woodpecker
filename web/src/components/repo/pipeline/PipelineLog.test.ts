import { enableAutoUnmount, flushPromises, shallowMount } from '@vue/test-utils';
import { encode } from 'js-base64';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { reactive, ref } from 'vue';
import { createI18n } from 'vue-i18n';

import PipelineLog from '~/components/repo/pipeline/PipelineLog.vue';
import useUserConfig from '~/compositions/useUserConfig';
import type { PipelineLog as ApiLog, Pipeline, PipelineConfig, PipelineStep, PipelineWorkflow } from '~/lib/api/types';

const route = reactive({ params: {}, query: {}, hash: '' });
const getLogs = vi.fn<() => Promise<ApiLog[]>>();
const closeStream = vi.fn();
let streamLine: ((line: ApiLog) => void) | undefined;
const streamLogs = vi.fn((_repo: number, _pipeline: number, _step: number, onLine: (line: ApiLog) => void) => {
  streamLine = onLine;
  return { close: closeStream };
});

vi.mock('vue-router', () => ({
  useRoute: () => route,
}));

vi.mock('~/compositions/useApiClient', () => ({
  default: () => ({
    getLogs,
    streamLogs,
  }),
}));

const i18n = createI18n({
  legacy: false,
  locale: 'en',
  fallbackLocale: 'en',
  missingWarn: false,
  fallbackWarn: false,
  messages: { en: {} },
});

const repo = ref({ id: 1, owner: 'owner', name: 'repo' });
const repoPermissions = ref({ pull: true, push: true, admin: false });
const pipelineConfigs = ref<PipelineConfig[]>([{ hash: 'h', name: 'default', data: '' }]);

enableAutoUnmount(afterEach);

beforeEach(() => {
  vi.useFakeTimers();
  vi.clearAllMocks();
  getLogs.mockResolvedValue([]);
  streamLine = undefined;
  route.hash = '';
  pipelineConfigs.value = [];
  useUserConfig().setUserConfig('collapseLogGroupsByDefault', false);
});

afterEach(() => {
  vi.clearAllTimers();
  vi.useRealTimers();
});

function mountLog(pipeline: Pipeline, stepId: number) {
  return shallowMount(PipelineLog, {
    props: { pipeline, stepId },
    global: {
      plugins: [i18n],
      provide: { repo, 'repo-permissions': repoPermissions, 'pipeline-configs': pipelineConfigs },
      stubs: { 'router-link': true, RouterLink: true, IconButton: false },
    },
  });
}

function makeStep(pid: number): PipelineStep {
  return {
    id: pid,
    uuid: `uuid-${pid}`,
    pipeline_id: 1,
    pid,
    ppid: 1,
    name: `step-${pid}`,
    state: 'success',
    exit_code: 0,
  };
}

// The server omits `children` for a stepless workflow, so the shape needs a cast.
interface LooseWorkflow extends Omit<PipelineWorkflow, 'children'> {
  children?: PipelineStep[] | null;
}

function makeWorkflow(id: number, children: PipelineStep[] | null | undefined): LooseWorkflow {
  const workflow: LooseWorkflow = {
    id,
    pipeline_id: 1,
    pid: id,
    name: `workflow-${id}`,
    state: 'success',
    started: 1,
    finished: 2,
  };
  if (children !== undefined) {
    workflow.children = children;
  }
  return workflow;
}

function makePipeline(workflows: LooseWorkflow[]): Pipeline {
  return {
    id: 1,
    number: 1,
    status: 'success',
    event: 'push',
    commit: 'abcdef1234567890',
    branch: 'main',
    workflows,
  } as unknown as Pipeline;
}

function apiLogs(lines: string[], start = 0): ApiLog[] {
  return lines.map((text, index) => ({
    id: start + index + 1,
    step_id: 7,
    time: start + index,
    line: start + index,
    data: encode(text),
    type: 0,
  }));
}

async function flushLogs() {
  await flushPromises();
  await vi.advanceTimersByTimeAsync(500);
  await flushPromises();
}

async function loadLog(lines: string[], running = false) {
  getLogs.mockResolvedValue(apiLogs(lines));
  const step: PipelineStep = {
    ...makeStep(7),
    started: 1,
    ...(running ? { state: 'running' } : { finished: 2 }),
  };
  const wrapper = mountLog(makePipeline([makeWorkflow(1, [step])]), 7);
  await flushLogs();
  return wrapper;
}

describe('pipelineLog', () => {
  it('finds a step past a workflow whose children key is absent', () => {
    const pipeline = makePipeline([makeWorkflow(1, undefined), makeWorkflow(2, [makeStep(7)])]);

    const wrapper = mountLog(pipeline, 7);

    expect(wrapper.exists()).toBe(true);
  });

  it('handles a null children payload', () => {
    const pipeline = makePipeline([makeWorkflow(1, null), makeWorkflow(2, [makeStep(7)])]);

    expect(() => mountLog(pipeline, 7)).not.toThrow();
  });

  it('handles a pipeline where every workflow is stepless', () => {
    const pipeline = makePipeline([makeWorkflow(1, undefined), makeWorkflow(2, [])]);

    expect(() => mountLog(pipeline, 7)).not.toThrow();
  });

  it('keeps ordinary, legacy and malformed markers as numbered output', async () => {
    pipelineConfigs.value = [{ hash: 'h', name: 'default', data: encode('steps:\n  - commands:\n      - make test') }];
    const lines = [
      'ordinary output',
      '+ make test',
      '▶ make test',
      'prefix ▶  make test',
      ' + make test',
      '▶  ',
      '▶  \t ',
      '▶\t Build',
      '::group::Build',
      '::endgroup::',
    ];

    const wrapper = await loadLog(lines);

    expect(wrapper.findAll('.sticky')).toHaveLength(0);
    expect(wrapper.find('button[title="repo.pipeline.actions.collapse_all"]').exists()).toBe(false);
    expect(wrapper.findAll('a[href^="#L"]').map((line) => line.text())).toEqual(
      lines.map((_, index) => String(index + 1)),
    );
    for (const [index, text] of lines.entries()) {
      expect(wrapper.get(`a#L${index + 1}`).element.nextElementSibling?.textContent).toBe(`${text}\n`);
    }
  });

  it('preserves configured command headings, matrix commands and initialization', async () => {
    pipelineConfigs.value = [
      {
        hash: 'h',
        name: 'default',
        data: encode(`steps:\n  - commands:\n      - make build\n      - echo \${TARGET}`),
      },
    ];

    const wrapper = await loadLog(['initializing', '▶  make build', 'built', '▶  echo linux', 'linux']);

    expect(wrapper.findAll('.sticky span[id]').map((heading) => heading.text())).toEqual(['make build', 'echo linux']);
    await wrapper.get('button[title="repo.pipeline.actions.collapse_all"]').trigger('click');
    expect(wrapper.get('a#L1').element.nextElementSibling?.textContent).toBe('initializing\n');
    expect(wrapper.find('a#L3').exists()).toBe(false);
    expect(wrapper.find('a#L5').exists()).toBe(false);
    await wrapper.get('button[title="repo.pipeline.actions.expand_all"]').trigger('click');
    expect(wrapper.findAll('a[href^="#L"]').map((line) => line.text())).toEqual(['1', '2', '3', '4', '5']);
  });

  it('preserves whitespace-padded configured command markers', async () => {
    pipelineConfigs.value = [{ hash: 'h', name: 'default', data: encode('steps:\n  - commands:\n      - npm test') }];
    useUserConfig().setUserConfig('collapseLogGroupsByDefault', true);

    const wrapper = await loadLog(['Preparing pipeline', '  ▶  npm test  ', 'existing command output']);

    expect(wrapper.findAll('.sticky')).toHaveLength(1);
    expect(wrapper.get('.sticky span#L2').element.textContent).toBe('▶  npm test  \n');
    expect(wrapper.findAll('a[href^="#L"]').map((line) => line.text())).toEqual(['1']);
    await wrapper.get('button[title="repo.pipeline.actions.expand_all"]').trigger('click');
    expect(wrapper.findAll('a[href^="#L"]').map((line) => line.text())).toEqual(['1', '2', '3']);
    expect(wrapper.get('a#L2').element.nextElementSibling?.textContent).toBe('  ▶  npm test  \n');
    expect(wrapper.get('a#L3').element.nextElementSibling?.textContent).toBe('existing command output\n');
    await wrapper.get('.sticky').trigger('click');
    expect(wrapper.findAll('a[href^="#L"]').map((line) => line.text())).toEqual(['1']);
  });

  describe('arbitrary log groups', () => {
    it.each([false, true])('renders titles and numbered text with configs present: %s', async (hasConfig) => {
      pipelineConfigs.value = hasConfig
        ? [{ hash: 'h', name: 'default', data: encode('steps:\n  - commands:\n      - unrelated command') }]
        : [];
      const lines = [
        'Preparing plugin',
        '▶  Build <assets> & bundle',
        'Compiled café',
        '▶  Test suite',
        'Tests passed',
      ];

      const wrapper = await loadLog(lines);

      expect(getLogs).toHaveBeenCalledWith(1, 1, 7);
      expect(wrapper.findAll('.sticky span[id]').map((heading) => heading.text())).toEqual([
        'Build <assets> & bundle',
        'Test suite',
      ]);
      expect(wrapper.findAll('.sticky span[id]').map((heading) => heading.attributes('id'))).toEqual(['L2', 'L4']);
      expect(wrapper.findAll('a[href^="#L"]').map((line) => line.text())).toEqual(['1', '2', '3', '4', '5']);
      for (const [index, text] of lines.entries()) {
        expect(wrapper.get(`a#L${index + 1}`).element.nextElementSibling?.textContent).toBe(`${text}\n`);
      }
    });

    it('starts only at a nonempty heading and ends at the next heading or end of log', async () => {
      const wrapper = await loadLog(['▶  ', '▶  \t', '▶  B', 'build output', '▶  T', 'test output', 'last output']);

      expect(wrapper.findAll('.sticky span[id]').map((heading) => heading.text())).toEqual(['B', 'T']);
      await wrapper.get('.sticky').trigger('click');

      expect(wrapper.findAll('a[href^="#L"]').map((line) => line.text())).toEqual(['1', '2', '5', '6', '7']);
    });

    it('collapses and expands an individual group without hiding the next group or preamble', async () => {
      const wrapper = await loadLog(['Preparing', '▶  Build', 'built', '▶  Test', 'passed']);

      await wrapper.get('.sticky').trigger('click');

      expect(wrapper.findAll('a[href^="#L"]').map((line) => line.text())).toEqual(['1', '4', '5']);
      expect(wrapper.get('.sticky span').text()).toBe('Build');
      await wrapper.get('.sticky').trigger('click');
      expect(wrapper.findAll('a[href^="#L"]').map((line) => line.text())).toEqual(['1', '2', '3', '4', '5']);
    });

    it('collapses and expands all groups while retaining initialization', async () => {
      const wrapper = await loadLog(['Preparing', '▶  Build', 'built', '▶  Test', 'passed']);

      await wrapper.get('button[title="repo.pipeline.actions.collapse_all"]').trigger('click');

      expect(wrapper.findAll('a[href^="#L"]').map((line) => line.text())).toEqual(['1']);
      expect(wrapper.findAll('.sticky')).toHaveLength(2);
      await wrapper.get('button[title="repo.pipeline.actions.expand_all"]').trigger('click');
      expect(wrapper.findAll('a[href^="#L"]').map((line) => line.text())).toEqual(['1', '2', '3', '4', '5']);
    });

    it.each([true, false])('honors the finished-log default-collapse preference: %s', async (collapse) => {
      useUserConfig().setUserConfig('collapseLogGroupsByDefault', collapse);

      const wrapper = await loadLog(['Preparing', '▶  Build', 'built']);

      expect(wrapper.findAll('.sticky')).toHaveLength(1);
      expect(wrapper.find('a#L1').exists()).toBe(true);
      expect(wrapper.find('a#L3').exists()).toBe(!collapse);
    });

    it.each(['#L2', '#L3'])('expands the initially linked group despite the default preference: %s', async (hash) => {
      useUserConfig().setUserConfig('collapseLogGroupsByDefault', true);
      route.hash = hash;

      const wrapper = await loadLog(['Preparing', '▶  Build', 'built', '▶  Test', 'passed']);

      expect(wrapper.findAll('.sticky')).toHaveLength(2);
      expect(wrapper.find('a#L3').exists()).toBe(true);
      expect(wrapper.find('a#L5').exists()).toBe(false);
      expect(wrapper.get(hash === '#L2' ? '.sticky' : 'a#L3').classes()).toContain(
        hash === '#L2' ? 'bg-blue-900' : 'bg-blue-600/30',
      );
    });

    it('expands a collapsed group when the route hash changes', async () => {
      useUserConfig().setUserConfig('collapseLogGroupsByDefault', true);
      const wrapper = await loadLog(['▶  Build', 'built', '▶  Test', 'passed']);
      expect(wrapper.find('a#L4').exists()).toBe(false);

      route.hash = '#L4';
      await flushPromises();

      expect(wrapper.get('a#L4').classes()).toContain('bg-blue-600/30');
      expect(wrapper.find('a#L2').exists()).toBe(false);
    });

    it('preserves a running group collapse state when output and a new group stream in', async () => {
      useUserConfig().setUserConfig('collapseLogGroupsByDefault', true);
      const wrapper = await loadLog([], true);
      expect(streamLogs).toHaveBeenCalledWith(1, 1, 7, expect.any(Function));
      expect(getLogs).not.toHaveBeenCalled();
      expect(streamLine).toBeDefined();
      for (const line of apiLogs(['Preparing', '▶  Build', 'first output'])) {
        streamLine?.(line);
      }
      await flushLogs();
      expect(wrapper.find('a#L3').exists()).toBe(true);
      await wrapper.get('.sticky').trigger('click');

      for (const line of apiLogs(['later output', '▶  Test suite', 'test output'], 3)) {
        streamLine?.(line);
      }
      await flushLogs();

      expect(wrapper.findAll('.sticky span[id]').map((heading) => heading.text())).toEqual(['Build', 'Test suite']);
      expect(wrapper.findAll('a[href^="#L"]').map((line) => line.text())).toEqual(['1', '5', '6']);
      await wrapper.get('.sticky').trigger('click');
      expect(wrapper.get('a#L4').element.nextElementSibling?.textContent).toBe('later output\n');
      expect(wrapper.findAll('a[href^="#L"]').map((line) => line.text())).toEqual(['1', '2', '3', '4', '5', '6']);
      wrapper.unmount();
      expect(closeStream).toHaveBeenCalledOnce();
    });
  });
});
