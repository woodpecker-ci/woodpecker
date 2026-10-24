<template>
  <div
    v-if="mode !== 'none'"
    class="operating-mode-banner flex w-full items-center justify-between px-4 py-1.5 text-xs font-semibold tracking-wider uppercase border-b transition-colors shadow-xs"
    :class="modeClass"
    data-testid="sfds-operating-mode-banner"
  >
    <div class="flex items-center space-x-2">
      <span class="inline-block h-2 w-2 rounded-full animate-pulse" :class="dotClass" />
      <span>Symphonix CareCI &bull; Mode: {{ modeTitle }}</span>
    </div>
    <div class="flex items-center space-x-3 text-xs normal-case opacity-90">
      <span>{{ modeDescription }}</span>
      <span class="border-l pl-3 font-mono">SFDS-HELP-PLATFORM-001</span>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue';

// Modes supported by SFDS: 'live' | 'simulation' | 'replay' | 'audit' | 'none'
const currentMode = ref<'live' | 'simulation' | 'replay' | 'audit'>('simulation');

const mode = computed(() => currentMode.value);

const modeTitle = computed(() => {
  switch (currentMode.value) {
    case 'live':
      return 'Live Clinical Operations';
    case 'simulation':
      return 'Simulation Sandbox';
    case 'replay':
      return 'Historical Replay';
    case 'audit':
      return 'DCB0129 / CAID Regulatory Audit';
    default:
      return 'Standard Mode';
  }
});

const modeDescription = computed(() => {
  switch (currentMode.value) {
    case 'live':
      return 'Active hospital deployment context. Patient safety controls enforced.';
    case 'simulation':
      return 'Synthetic EHR & BulletTrain Client Registry sandbox. Safe for algorithmic testing.';
    case 'replay':
      return 'Deterministic event log execution. State modifications read-only.';
    case 'audit':
      return 'Clinical Safety Officer inspection mode. WORM SHA-256 hash chains verified.';
    default:
      return '';
  }
});

const modeClass = computed(() => {
  switch (currentMode.value) {
    case 'live':
      return 'bg-emerald-900/90 text-emerald-100 border-emerald-700';
    case 'simulation':
      return 'bg-indigo-950/90 text-indigo-200 border-indigo-800';
    case 'replay':
      return 'bg-amber-950/90 text-amber-200 border-amber-800';
    case 'audit':
      return 'bg-cyan-950/90 text-cyan-200 border-cyan-800';
    default:
      return 'bg-neutral-800 text-neutral-200 border-neutral-700';
  }
});

const dotClass = computed(() => {
  switch (currentMode.value) {
    case 'live':
      return 'bg-emerald-400';
    case 'simulation':
      return 'bg-indigo-400';
    case 'replay':
      return 'bg-amber-400';
    case 'audit':
      return 'bg-cyan-400';
    default:
      return 'bg-neutral-400';
  }
});
</script>

<style scoped>
.text-2xs {
  font-size: 0.65rem;
}
</style>
