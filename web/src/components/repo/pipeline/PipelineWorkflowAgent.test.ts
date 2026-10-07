import { DOMWrapper, enableAutoUnmount, flushPromises, mount } from '@vue/test-utils';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ref } from 'vue';
import { createI18n } from 'vue-i18n';

import PipelineWorkflowAgent from '~/components/repo/pipeline/PipelineWorkflowAgent.vue';
import type { Agent, OrgPermissions, PipelineWorkflow, User } from '~/lib/api/types';

const getAgent = vi.fn<(id: number) => Promise<Agent>>();
const getOrgAgents = vi.fn<(orgId: number, opts?: { page?: number; perPage?: number }) => Promise<Agent[] | null>>();
const getOrgPermissions = vi.fn<(orgId: number) => Promise<OrgPermissions>>();

vi.mock('~/compositions/useApiClient', () => ({
  default: () => ({ getAgent, getOrgAgents, getOrgPermissions }),
}));

const i18n = createI18n({
  legacy: false,
  locale: 'en',
  missingWarn: false,
  fallbackWarn: false,
  messages: { en: {} },
});

const repo = ref({ id: 1, org_id: 5, owner: 'owner', name: 'repo' });

function makeAgent(overrides: Partial<Agent> = {}): Agent {
  return {
    id: 3,
    name: 'builder-03',
    owner_id: -1,
    org_id: -1,
    token: 'secret-agent-token',
    created: 1,
    updated: 1,
    last_contact: 1,
    platform: 'linux/amd64',
    backend: 'docker',
    capacity: 2,
    version: 'next',
    no_schedule: false,
    custom_labels: { zone: 'eu-1' },
    filters: {},
    ...overrides,
  };
}

function makeWorkflow(overrides: Partial<PipelineWorkflow> = {}): PipelineWorkflow {
  return { id: 1, pipeline_id: 1, pid: 1, name: 'docker', state: 'running', agent_id: 3, ...overrides };
}

function mountAgent(workflow: PipelineWorkflow | undefined = makeWorkflow()) {
  return mount(PipelineWorkflowAgent, {
    props: { workflow },
    attachTo: document.body,
    global: { plugins: [i18n], provide: { repo } },
  });
}

type Wrapper = ReturnType<typeof mountAgent>;

const infoButton = (wrapper: Wrapper) => wrapper.get('button[aria-haspopup="dialog"]');
// the popover is teleported to <body> so it can escape the log box (overflow, stacking contexts)
const popover = () => new DOMWrapper(document.body).find('[role="dialog"]');

function deferred<T>() {
  let resolveFn!: (value: T) => void;
  let rejectFn!: (reason: Error) => void;
  const promise = new Promise<T>((resolve, reject) => {
    resolveFn = resolve;
    rejectFn = reject;
  });
  return { promise, resolve: resolveFn, reject: rejectFn };
}

enableAutoUnmount(afterEach);

beforeEach(() => {
  vi.clearAllMocks();
  window.WOODPECKER_USER = { id: 1, login: 'admin', admin: true } as User;
  window.WOODPECKER_USER_REGISTERED_AGENTS = true;
  getAgent.mockResolvedValue(makeAgent());
  getOrgAgents.mockResolvedValue([]);
  getOrgPermissions.mockResolvedValue({ member: true, admin: true });
});

afterEach(() => {
  window.WOODPECKER_USER = undefined;
  window.WOODPECKER_USER_REGISTERED_AGENTS = undefined;
});

describe('pipelineWorkflowAgent', () => {
  describe('permissions', () => {
    it('grays out the icon for anonymous users without calling any api', async () => {
      window.WOODPECKER_USER = undefined;
      const wrapper = mountAgent();
      await flushPromises();

      expect(infoButton(wrapper).attributes('disabled')).toBeDefined();
      expect(infoButton(wrapper).attributes('title')).toBe('repo.pipeline.agent.no_permission');
      expect(getAgent).not.toHaveBeenCalled();
      expect(getOrgPermissions).not.toHaveBeenCalled();
    });

    it('grays out the icon for a user that is neither instance nor org admin', async () => {
      window.WOODPECKER_USER = { id: 2, login: 'dev', admin: false } as User;
      getOrgPermissions.mockResolvedValue({ member: true, admin: false });
      const wrapper = mountAgent();
      await flushPromises();

      expect(getOrgPermissions).toHaveBeenCalledWith(5);
      expect(infoButton(wrapper).attributes('disabled')).toBeDefined();
      expect(infoButton(wrapper).attributes('title')).toBe('repo.pipeline.agent.no_permission');
      expect(getAgent).not.toHaveBeenCalled();
      expect(getOrgAgents).not.toHaveBeenCalled();
    });

    it('does not ask for org permissions when user registered agents are disabled', async () => {
      window.WOODPECKER_USER = { id: 2, login: 'dev', admin: false } as User;
      window.WOODPECKER_USER_REGISTERED_AGENTS = false;
      const wrapper = mountAgent();
      await flushPromises();

      expect(getOrgPermissions).not.toHaveBeenCalled();
      expect(infoButton(wrapper).attributes('disabled')).toBeDefined();
    });

    it('lets an org admin look up an agent of the org through the org agent list, page by page', async () => {
      window.WOODPECKER_USER = { id: 2, login: 'dev', admin: false } as User;
      const fullPage = Array.from({ length: 50 }, (_, i) => makeAgent({ id: 100 + i, org_id: 5 }));
      getOrgAgents
        .mockResolvedValueOnce(fullPage)
        .mockResolvedValueOnce([makeAgent({ id: 3, org_id: 5, name: 'org-builder' })]);
      const wrapper = mountAgent();
      await flushPromises();

      expect(getAgent).not.toHaveBeenCalled();
      expect(getOrgAgents).toHaveBeenNthCalledWith(1, 5, { page: 1, perPage: 50 });
      expect(getOrgAgents).toHaveBeenNthCalledWith(2, 5, { page: 2, perPage: 50 });
      expect(infoButton(wrapper).attributes('disabled')).toBeUndefined();
      await infoButton(wrapper).trigger('click');
      expect(popover().text()).toContain('org-builder');
    });

    it('grays out the icon for an org admin when the agent is not one of the org agents', async () => {
      window.WOODPECKER_USER = { id: 2, login: 'dev', admin: false } as User;
      getOrgAgents.mockResolvedValue([makeAgent({ id: 9, org_id: 5 })]);
      const wrapper = mountAgent();
      await flushPromises();

      expect(getOrgAgents).toHaveBeenCalledOnce();
      expect(infoButton(wrapper).attributes('disabled')).toBeDefined();
      expect(infoButton(wrapper).attributes('title')).toBe('repo.pipeline.agent.no_permission');
    });
  });

  describe('agent assignment', () => {
    it('grays out the icon while no agent is assigned', async () => {
      const wrapper = mountAgent(makeWorkflow({ state: 'pending', agent_id: undefined }));
      await flushPromises();

      expect(infoButton(wrapper).attributes('disabled')).toBeDefined();
      expect(infoButton(wrapper).attributes('title')).toBe('repo.pipeline.agent.not_assigned');
      expect(getAgent).not.toHaveBeenCalled();
    });

    it('loads the agent as soon as a pending workflow gets one assigned', async () => {
      const wrapper = mountAgent(makeWorkflow({ state: 'pending', agent_id: undefined }));
      await flushPromises();

      await wrapper.setProps({ workflow: makeWorkflow({ state: 'running', agent_id: 3 }) });
      await flushPromises();

      expect(getAgent).toHaveBeenCalledWith(3);
      expect(infoButton(wrapper).attributes('disabled')).toBeUndefined();
      expect(infoButton(wrapper).attributes('title')).toBe('repo.pipeline.agent.title');
    });

    it('refreshes the shown agent when the workflow state changes', async () => {
      const wrapper = mountAgent();
      await flushPromises();
      await infoButton(wrapper).trigger('click');

      getAgent.mockResolvedValue(makeAgent({ name: 'renamed-builder' }));
      await wrapper.setProps({ workflow: makeWorkflow({ state: 'success', finished: 2 }) });
      await flushPromises();

      expect(getAgent).toHaveBeenCalledTimes(2);
      expect(popover().text()).toContain('renamed-builder');
    });

    it('does not refetch when only steps of the workflow change', async () => {
      const wrapper = mountAgent();
      await flushPromises();

      await wrapper.setProps({ workflow: makeWorkflow({ children: [] }) });
      await flushPromises();

      expect(getAgent).toHaveBeenCalledOnce();
    });

    it('grays out the icon when the agent cannot be loaded anymore', async () => {
      getAgent.mockRejectedValue(new Error('Not Found'));
      const wrapper = mountAgent(makeWorkflow({ state: 'success' }));
      await flushPromises();

      expect(infoButton(wrapper).attributes('disabled')).toBeDefined();
      expect(infoButton(wrapper).attributes('title')).toBe('repo.pipeline.agent.unavailable');
    });

    it('ignores a late answer for a previously selected workflow', async () => {
      const slow = deferred<Agent>();
      getAgent.mockReturnValueOnce(slow.promise);
      const wrapper = mountAgent();

      getAgent.mockResolvedValue(makeAgent({ id: 4, name: 'test-agent' }));
      await wrapper.setProps({ workflow: makeWorkflow({ id: 2, name: 'test', agent_id: 4 }) });
      await flushPromises();
      slow.resolve(makeAgent());
      await flushPromises();

      await infoButton(wrapper).trigger('click');
      expect(popover().text()).toContain('test-agent');
      expect(popover().text()).not.toContain('builder-03');
    });

    it('ignores a late failure for a previously selected workflow', async () => {
      const slow = deferred<Agent>();
      getAgent.mockReturnValueOnce(slow.promise);
      const wrapper = mountAgent();

      await wrapper.setProps({ workflow: makeWorkflow({ id: 2, agent_id: 4 }) });
      await flushPromises();
      slow.reject(new Error('Not Found'));
      await flushPromises();

      expect(infoButton(wrapper).attributes('disabled')).toBeUndefined();
    });

    it('drops the previous agent when the workflow gets another agent assigned', async () => {
      const wrapper = mountAgent();
      await flushPromises();
      await infoButton(wrapper).trigger('click');

      const slow = deferred<Agent>();
      getAgent.mockReturnValueOnce(slow.promise);
      await wrapper.setProps({ workflow: makeWorkflow({ agent_id: 4 }) });
      expect(popover().exists()).toBe(false);
      expect(infoButton(wrapper).attributes('disabled')).toBeDefined();

      slow.resolve(makeAgent({ id: 4, name: 'replacement' }));
      await flushPromises();
      await infoButton(wrapper).trigger('click');
      expect(popover().text()).toContain('replacement');
    });

    it('keeps the popover open with the current data while refreshing the same agent', async () => {
      const wrapper = mountAgent();
      await flushPromises();
      await infoButton(wrapper).trigger('click');

      getAgent.mockReturnValueOnce(deferred<Agent>().promise);
      await wrapper.setProps({ workflow: makeWorkflow({ state: 'success', finished: 2 }) });
      expect(popover().text()).toContain('builder-03');
    });

    it('closes the popover when another workflow gets selected', async () => {
      const wrapper = mountAgent();
      await flushPromises();
      await infoButton(wrapper).trigger('click');

      await wrapper.setProps({ workflow: makeWorkflow({ id: 2, name: 'test' }) });
      await flushPromises();

      expect(popover().exists()).toBe(false);
    });
  });

  describe('popover', () => {
    it('shows workflow, agent, platform and backend, labels only on demand and never the token', async () => {
      const wrapper = mountAgent();
      await flushPromises();
      expect(popover().exists()).toBe(false);

      await infoButton(wrapper).trigger('click');
      const text = popover().text();
      expect(infoButton(wrapper).attributes('aria-expanded')).toBe('true');
      expect(text).toContain('docker');
      expect(text).toContain('builder-03');
      expect(text).toContain('linux/amd64');
      expect(text).not.toContain('zone');
      expect(text).not.toContain('secret-agent-token');

      await popover().get('button').trigger('click');
      expect(popover().text()).toContain('zone');
      expect(popover().text()).toContain('eu-1');

      await popover().get('button').trigger('click');
      expect(popover().text()).not.toContain('zone');
    });

    it('falls back to the agent id for an agent without name and says when it has no labels', async () => {
      getAgent.mockResolvedValue(makeAgent({ name: '', custom_labels: {} }));
      const wrapper = mountAgent();
      await flushPromises();
      await infoButton(wrapper).trigger('click');

      expect(popover().text()).toContain('#3');
      await popover().get('button').trigger('click');
      expect(popover().text()).toContain('repo.pipeline.agent.no_labels');
    });

    it('closes on Escape and on a click outside', async () => {
      const wrapper = mountAgent();
      await flushPromises();

      await infoButton(wrapper).trigger('click');
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
      await flushPromises();
      expect(popover().exists()).toBe(false);

      await infoButton(wrapper).trigger('click');
      // onClickOutside handles one click per macrotask
      await new Promise((resolve) => setTimeout(resolve, 0));
      document.body.dispatchEvent(new MouseEvent('pointerdown', { bubbles: true }));
      document.body.dispatchEvent(new MouseEvent('click', { bubbles: true }));
      await flushPromises();
      expect(popover().exists()).toBe(false);
    });

    it('places the popover below the icon inside the viewport and limits its height to the space left', async () => {
      Object.assign(window, { innerWidth: 416, innerHeight: 300 });
      let buttonTop = 10;
      let toolbar: Element | null = null;
      vi.spyOn(Element.prototype, 'getBoundingClientRect').mockImplementation(function (this: Element) {
        if (this.matches('button[aria-haspopup="dialog"]')) {
          return DOMRect.fromRect({ x: 300, y: buttonTop, width: 32, height: 32 });
        }
        if (this === toolbar) {
          // the toolbar the info icon sits in, its right edge aligns the popover
          return DOMRect.fromRect({ x: 200, y: buttonTop, width: 200, height: 32 });
        }
        return DOMRect.fromRect({ x: 0, y: 0, width: 0, height: 0 });
      });
      const wrapper = mountAgent();
      // button -> component root -> toolbar
      toolbar = infoButton(wrapper).element.parentElement?.parentElement ?? null;
      await flushPromises();
      await infoButton(wrapper).trigger('click');
      await flushPromises();

      const style = (popover().element as HTMLElement).style;
      expect(popover().classes()).toContain('fixed');
      expect(style.top).toBe('50px');
      expect(style.right).toBe('16px');
      expect((popover().get('[data-scroll]').element as HTMLElement).style.maxHeight).toBe('234px');
      expect((popover().get('[data-arrow]').element as HTMLElement).style.right).toBe('78px');

      buttonTop = 110;
      window.dispatchEvent(new Event('resize'));
      await flushPromises();
      expect(style.top).toBe('150px');
      expect((popover().get('[data-scroll]').element as HTMLElement).style.maxHeight).toBe('134px');

      vi.restoreAllMocks();
    });

    it('stays open when clicking inside the popover', async () => {
      const wrapper = mountAgent();
      await flushPromises();
      await infoButton(wrapper).trigger('click');
      await new Promise((resolve) => setTimeout(resolve, 0));

      popover().element.dispatchEvent(new MouseEvent('pointerdown', { bubbles: true }));
      popover().element.dispatchEvent(new MouseEvent('click', { bubbles: true }));
      await flushPromises();
      expect(popover().exists()).toBe(true);
    });
  });
});
