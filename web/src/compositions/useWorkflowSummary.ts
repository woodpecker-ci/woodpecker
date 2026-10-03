import { computed } from 'vue';
import type { ComputedRef, Ref } from 'vue';
import { useI18n } from 'vue-i18n';

// Summarizes a workflow selection the same way in every form built on
// WorkflowSelect: an empty selection runs every workflow, which is easy to
// miss once a repo has enough of them that the checkbox list scrolls.
export function useWorkflowSummary(workflows: Ref<string[]> | ComputedRef<string[]>) {
  const i18n = useI18n();

  return computed(() => {
    const count = workflows.value.length;
    if (count === 0) {
      return i18n.t('repo.workflow_select.summary.all');
    }
    if (count === 1) {
      return i18n.t('repo.workflow_select.summary.one');
    }
    return i18n.t('repo.workflow_select.summary.many', { count });
  });
}
