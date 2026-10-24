<template>
  <div class="view-level-selector flex items-center bg-black/20 rounded-md p-0.5 text-xs">
    <button
      v-for="lvl in levels"
      :key="lvl.id"
      type="button"
      class="px-2 py-1 rounded transition-colors text-xs font-medium cursor-pointer"
      :class="
        currentLevel === lvl.id
          ? 'bg-wp-primary-100 text-white font-bold shadow-xs'
          : 'text-neutral-300 hover:text-white hover:bg-white/10'
      "
      :title="lvl.description"
      @click="setLevel(lvl.id)"
    >
      {{ lvl.label }}
    </button>
  </div>
</template>

<script lang="ts" setup>
import { onMounted, ref } from 'vue';

export type ViewLevel = 'essential' | 'standard' | 'advanced' | 'expert';

const levels: { id: ViewLevel; label: string; description: string }[] = [
  { id: 'essential', label: 'Essential', description: 'Clinical Safety Officer view: safety gates, pass/fail, and hazard compliance' },
  { id: 'standard', label: 'Standard', description: 'Standard CI view: pipeline DAG, steps, and test results' },
  { id: 'advanced', label: 'Advanced', description: 'MLOps view: self-healing metrics, memory escalation, and latency' },
  { id: 'expert', label: 'Expert', description: 'Platform engineering view: raw container traces, kernel signals, and Tink crypto' },
];

const currentLevel = ref<ViewLevel>('standard');

function setLevel(lvl: ViewLevel) {
  currentLevel.value = lvl;
  localStorage.setItem('symphonix_view_level', lvl);
  document.documentElement.setAttribute('data-sfds-view-level', lvl);
}

onMounted(() => {
  const saved = localStorage.getItem('symphonix_view_level') as ViewLevel;
  if (saved && levels.some((l) => l.id === saved)) {
    setLevel(saved);
  } else {
    setLevel('standard');
  }
});
</script>

<style scoped>
.text-2xs {
  font-size: 0.68rem;
}
</style>
