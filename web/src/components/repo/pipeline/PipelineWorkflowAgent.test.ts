import { DOMWrapper, enableAutoUnmount, flushPromises, mount, RouterLinkStub } from '@vue/test-utils';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ref } from 'vue';
import { createI18n } from 'vue-i18n';

import PipelineWorkflowAgent from '~/components/repo/pipeline/PipelineWorkflowAgent.vue';
import { ApiRequestError } from '~/lib/api/client';
import type { AgentSnapshot, OrgPermissions, PipelineWorkflow, User } from '~/lib/api/types';

const getWorkflowAgent =
  vi.fn<(repoId: number, pipelineNumber: number, workflowId: number) => Promise<AgentSnapshot>>();

const getOrgPermissions = vi.fn<(orgId: number) => Promise<OrgPermissions>>();

vi.mock('~/compositions/useApiClient', () => ({
  default: () => ({ getWorkflowAgent, getOrgPermissions }),
}));

const i18n = createI18n({
  legacy: false,
  locale: 'en',
  missingWarn: false,
  fallbackWarn: false,
  messages: { en: {} },
});

const repo = ref({ id: 1, org_id: 5, owner: 'owner', name: 'repo' });

function makeAgent(overrides: Partial<AgentSnapshot> = {}): AgentSnapshot {
  return {
    id: 3,
    org_id: -1,
    name: 'builder-03',
    platform: 'linux/amd64',
    backend: 'docker',
    custom_labels: { zone: 'eu-1' },
    ...overrides,
  };
}

function makeWorkflow(overrides: Partial<PipelineWorkflow> = {}): PipelineWorkflow {
  return { id: 1, pipeline_id: 1, pid: 1, name: 'docker', state: 'running', agent_id: 3, ...overrides };
}

function mountAgent(workflow: PipelineWorkflow | undefined = makeWorkflow()) {
  return mount(PipelineWorkflowAgent, {
    props: { workflow, pipelineNumber: 7 },
    attachTo: document.body,
    global: { plugins: [i18n], provide: { repo }, stubs: { RouterLink: RouterLinkStub } },
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
  window.WOODPECKER_USER = { id: 1, login: 'admin', admin: true, org_id: 1 } as User;
  window.WOODPECKER_USER_REGISTERED_AGENTS = true;
  getWorkflowAgent.mockResolvedValue(makeAgent());
  getOrgPermissions.mockResolvedValue({ member: false, admin: false });
});

afterEach(() => {
  window.WOODPECKER_USER = undefined;
  window.WOODPECKER_USER_REGISTERED_AGENTS = undefined;
});

describe('pipelineWorkflowAgent', () => {
  describe('permissions', () => {
    it('grays out the icon for anonymous users without asking the server', async () => {
      window.WOODPECKER_USER = undefined;
      const wrapper = mountAgent();
      await flushPromises();

      expect(infoButton(wrapper).attributes('disabled')).toBeDefined();
      expect(infoButton(wrapper).attributes('title')).toBe('repo.pipeline.agent.no_permission');
      expect(getWorkflowAgent).not.toHaveBeenCalled();
    });

    it('shows the agent to every logged in user', async () => {
      window.WOODPECKER_USER = { id: 2, login: 'dev', admin: false } as User;
      getWorkflowAgent.mockResolvedValue(makeAgent({ org_id: 6, name: 'other-org-builder' }));
      const wrapper = mountAgent();
      await flushPromises();

      expect(getWorkflowAgent).toHaveBeenCalledWith(1, 7, 1);
      await infoButton(wrapper).trigger('click');
      expect(popover().text()).toContain('other-org-builder');
    });
  });

  describe('agent assignment', () => {
    it('grays out the icon while no agent is assigned', async () => {
      const wrapper = mountAgent(makeWorkflow({ state: 'pending', agent_id: undefined }));
      await flushPromises();

      expect(infoButton(wrapper).attributes('disabled')).toBeDefined();
      expect(infoButton(wrapper).attributes('title')).toBe('repo.pipeline.agent.not_assigned');
      expect(getWorkflowAgent).not.toHaveBeenCalled();
    });

    it('loads the agent as soon as a pending workflow gets one assigned', async () => {
      const wrapper = mountAgent(makeWorkflow({ state: 'pending', agent_id: undefined }));
      await flushPromises();

      await wrapper.setProps({ workflow: makeWorkflow({ state: 'running', agent_id: 3 }) });
      await flushPromises();

      expect(getWorkflowAgent).toHaveBeenCalledWith(1, 7, 1);
      expect(infoButton(wrapper).attributes('disabled')).toBeUndefined();
      expect(infoButton(wrapper).attributes('title')).toBe('repo.pipeline.agent.title');
    });

    it('does not refetch the snapshot for state or step updates of the same assignment', async () => {
      const wrapper = mountAgent();
      await flushPromises();

      await wrapper.setProps({ workflow: makeWorkflow({ children: [] }) });
      await wrapper.setProps({ workflow: makeWorkflow({ state: 'success', finished: 2 }) });
      await flushPromises();

      expect(getWorkflowAgent).toHaveBeenCalledOnce();
    });

    it('grays out the icon when no agent information was recorded', async () => {
      getWorkflowAgent.mockRejectedValue(new ApiRequestError('Not Found', 404));
      const wrapper = mountAgent(makeWorkflow({ state: 'success' }));
      await flushPromises();

      expect(infoButton(wrapper).attributes('disabled')).toBeDefined();
      expect(infoButton(wrapper).attributes('title')).toBe('repo.pipeline.agent.unavailable');
    });

    it('ignores a late answer for a previously selected workflow', async () => {
      const slow = deferred<AgentSnapshot>();
      getWorkflowAgent.mockReturnValueOnce(slow.promise);
      const wrapper = mountAgent();

      getWorkflowAgent.mockResolvedValue(makeAgent({ id: 4, name: 'test-agent' }));
      await wrapper.setProps({ workflow: makeWorkflow({ id: 2, name: 'test', agent_id: 4 }) });
      await flushPromises();
      slow.resolve(makeAgent());
      await flushPromises();

      await infoButton(wrapper).trigger('click');
      expect(popover().text()).toContain('test-agent');
      expect(popover().text()).not.toContain('builder-03');
    });

    it('ignores a late failure for a previously selected workflow', async () => {
      const slow = deferred<AgentSnapshot>();
      getWorkflowAgent.mockReturnValueOnce(slow.promise);
      const wrapper = mountAgent();

      await wrapper.setProps({ workflow: makeWorkflow({ id: 2, agent_id: 4 }) });
      await flushPromises();
      slow.reject(new ApiRequestError('Not Found', 404));
      await flushPromises();

      expect(infoButton(wrapper).attributes('disabled')).toBeUndefined();
    });

    it('drops the previous agent when the workflow gets another agent assigned', async () => {
      const wrapper = mountAgent();
      await flushPromises();
      await infoButton(wrapper).trigger('click');

      const slow = deferred<AgentSnapshot>();
      getWorkflowAgent.mockReturnValueOnce(slow.promise);
      await wrapper.setProps({ workflow: makeWorkflow({ agent_id: 4 }) });
      expect(popover().exists()).toBe(false);
      expect(infoButton(wrapper).attributes('disabled')).toBeDefined();

      slow.resolve(makeAgent({ id: 4, name: 'replacement' }));
      await flushPromises();
      await infoButton(wrapper).trigger('click');
      expect(popover().text()).toContain('replacement');
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
    it('shows workflow, agent, platform and backend, and labels only on demand', async () => {
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

      await popover().get('button').trigger('click');
      expect(popover().text()).toContain('zone');
      expect(popover().text()).toContain('eu-1');

      await popover().get('button').trigger('click');
      expect(popover().text()).not.toContain('zone');
    });

    it('falls back to the agent id for an agent without name and says when it has no labels', async () => {
      getWorkflowAgent.mockResolvedValue(makeAgent({ name: '', custom_labels: {} }));
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

  describe('agent link', () => {
    async function openPopover(agent: AgentSnapshot) {
      getWorkflowAgent.mockResolvedValue(agent);
      const wrapper = mountAgent();
      await flushPromises();
      await infoButton(wrapper).trigger('click');
      return wrapper;
    }

    const agentLink = (wrapper: Wrapper) => wrapper.findComponent(RouterLinkStub);

    it.each([
      [-1, true],
      [-1, false],
      [1, true],
      [1, false],
      [5, true],
      [5, false],
    ] as const)('links agent org %i to admin settings when user registration is %s', async (orgId, registration) => {
      window.WOODPECKER_USER_REGISTERED_AGENTS = registration;
      const wrapper = await openPopover(makeAgent({ org_id: orgId }));

      expect(agentLink(wrapper).props('to')).toStrictEqual({ name: 'admin-settings-agents', query: { agent: 3 } });
      expect(agentLink(wrapper).attributes('target')).toBe('_blank');
      expect(agentLink(wrapper).text()).toBe('builder-03');
      expect(getOrgPermissions).not.toHaveBeenCalled();
    });

    it('does not link a global agent for other users', async () => {
      window.WOODPECKER_USER = { id: 2, login: 'dev', admin: false, org_id: 2 } as User;
      const wrapper = await openPopover(makeAgent());

      expect(agentLink(wrapper).exists()).toBe(false);
      expect(popover().text()).toContain('builder-03');
    });

    it('links an org agent to the org agents settings for org admins', async () => {
      window.WOODPECKER_USER = { id: 2, login: 'dev', admin: false, org_id: 2 } as User;
      getOrgPermissions.mockResolvedValue({ member: true, admin: true });
      const wrapper = await openPopover(makeAgent({ org_id: 5 }));

      expect(getOrgPermissions).toHaveBeenCalledWith(5);
      expect(agentLink(wrapper).props('to')).toStrictEqual({
        name: 'org-settings-agents',
        params: { orgId: 5 },
        query: { agent: 3 },
      });
      expect(agentLink(wrapper).attributes('target')).toBe('_blank');
    });

    it('does not link an org agent for org members without admin', async () => {
      window.WOODPECKER_USER = { id: 2, login: 'dev', admin: false, org_id: 2 } as User;
      getOrgPermissions.mockResolvedValue({ member: true, admin: false });
      const wrapper = await openPopover(makeAgent({ org_id: 5 }));

      expect(agentLink(wrapper).exists()).toBe(false);
    });

    it('links an agent of the own user org to the user agents settings', async () => {
      window.WOODPECKER_USER = { id: 2, login: 'dev', admin: false, org_id: 2 } as User;
      const wrapper = await openPopover(makeAgent({ org_id: 2 }));

      expect(getOrgPermissions).not.toHaveBeenCalled();
      expect(agentLink(wrapper).props('to')).toStrictEqual({ name: 'user-agents', query: { agent: 3 } });
    });

    it('does not link org or user agents when user agent registration is disabled, like the settings tabs', async () => {
      window.WOODPECKER_USER = { id: 2, login: 'dev', admin: false, org_id: 2 } as User;
      window.WOODPECKER_USER_REGISTERED_AGENTS = false;
      getOrgPermissions.mockResolvedValue({ member: true, admin: true });
      const wrapper = await openPopover(makeAgent({ org_id: 2 }));
      expect(agentLink(wrapper).exists()).toBe(false);

      const other = await openPopover(makeAgent({ org_id: 5 }));
      expect(agentLink(other).exists()).toBe(false);
      expect(getOrgPermissions).not.toHaveBeenCalled();
    });
  });
});
