<template>
  <aside
    v-if="isOpen"
    ref="panelRef"
    data-testid="signalbox-panel"
    class="signalbox-panel flex h-full w-96 xl:w-[420px] shrink-0 flex-col border-l border-slate-800 bg-slate-900 text-slate-100 shadow-2xl overflow-hidden transition-all duration-300 font-sans"
    aria-label="Symphonix SignalBox Clinical Assurance & Governance Panel"
  >
    <!-- Header -->
    <div class="flex items-center justify-between border-b border-slate-800 px-5 py-4 shrink-0 bg-slate-900/90">
      <div class="flex items-center space-x-3">
        <div class="flex h-8 w-8 items-center justify-center rounded-lg bg-cyan-950 text-cyan-400 border border-cyan-500/40 shrink-0">
          <svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
          </svg>
        </div>
        <div>
          <h2 class="text-sm font-bold tracking-wide text-white uppercase">SignalBox Governance</h2>
          <p class="text-xs text-slate-400">CAID / DCB0129-DCB0160 Active</p>
        </div>
      </div>
      <button
        data-testid="signalbox-close"
        class="rounded-md p-1.5 text-slate-400 hover:bg-slate-800 hover:text-white transition"
        title="Close Panel"
        @click="close"
      >
        <svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
        </svg>
      </button>
    </div>

    <!-- Scrollable Content -->
    <div class="flex-1 space-y-5 overflow-y-auto p-5 text-xs">
      <!-- Clinical Safety Officer (CSO) Signoff Box -->
      <section class="rounded-lg border border-emerald-500/30 bg-emerald-950/20 p-4">
        <div class="flex items-center justify-between mb-2">
          <span class="font-bold text-emerald-400 uppercase tracking-wide flex items-center gap-1.5 text-xs">
            <span class="h-2 w-2 rounded-full bg-emerald-400 animate-pulse"></span>
            CSO Clinical Gate Passed
          </span>
          <span class="rounded bg-emerald-900/60 px-2 py-0.5 font-mono text-xs text-emerald-300 border border-emerald-600/40">
            ACCOUNTABLE_HUMAN
          </span>
        </div>
        <p class="text-slate-300 text-xs">
          Signoff by: <span class="font-mono text-emerald-300 font-semibold">{{ stats.csoSignoff.user }}</span>
        </p>
        <p class="text-slate-400 text-xs mt-0.5">
          Timestamp: <span class="font-mono text-slate-300">{{ stats.csoSignoff.timestamp }}</span>
        </p>
        <div class="mt-2 text-xs text-slate-400 font-mono break-all bg-slate-950 p-2 rounded border border-slate-800">
          Merkle Root: {{ stats.csoSignoff.merkleRoot }}
        </div>
      </section>

      <!-- Key Assurance Metrics Grid -->
      <section class="grid grid-cols-2 gap-3">
        <div class="rounded-lg border border-slate-800 bg-slate-950/60 p-3.5">
          <div class="text-slate-400 text-xs uppercase tracking-wider font-semibold">Golden Dataset</div>
          <div class="mt-1 text-2xl font-bold font-mono text-cyan-400">
            {{ stats.goldenDatasetAccuracy }}%
          </div>
          <div class="text-xs text-slate-400 mt-0.5">Target: &ge; 99.9% zero-miss</div>
        </div>

        <div class="rounded-lg border border-slate-800 bg-slate-950/60 p-3.5">
          <div class="text-slate-400 text-xs uppercase tracking-wider font-semibold">Zero-PHI Sentry</div>
          <div class="mt-1 text-2xl font-bold font-mono text-indigo-400">
            {{ stats.zeroPhiSentryStatus }}
          </div>
          <div class="text-xs text-slate-400 mt-0.5">{{ stats.redactionsCount }} IDs redacted</div>
        </div>
      </section>

      <!-- Multi-Tier Self-Healing Telemetry -->
      <section class="rounded-lg border border-slate-800 bg-slate-950/60 p-4">
        <h3 class="font-bold text-sky-400 uppercase tracking-wide mb-3 flex items-center gap-1.5 text-xs">
          <svg class="h-4 w-4 text-sky-400 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          Multi-Tier Autonomous Healing
        </h3>
        <div class="grid grid-cols-3 gap-2 text-center">
          <div class="bg-slate-900 p-2.5 rounded-lg border border-slate-800">
            <div class="font-mono text-base font-bold text-cyan-400">{{ stats.selfHealingStats.oomEscalations }}</div>
            <div class="text-xs text-slate-400 mt-0.5">L1 OOM Scaled</div>
          </div>
          <div class="bg-slate-900 p-2.5 rounded-lg border border-slate-800">
            <div class="font-mono text-base font-bold text-amber-400">{{ stats.selfHealingStats.flakyRetries }}</div>
            <div class="text-xs text-slate-400 mt-0.5">L2 AI Retries</div>
          </div>
          <div class="bg-slate-900 p-2.5 rounded-lg border border-slate-800">
            <div class="font-mono text-base font-bold text-emerald-400">{{ stats.selfHealingStats.quarantinedAgents }}</div>
            <div class="text-xs text-slate-400 mt-0.5">L3 Quarantines</div>
          </div>
        </div>
      </section>

      <!-- Dual-Audience Explainability (CDR 1.0.0) -->
      <section class="rounded-lg border border-slate-800 bg-slate-950/60 p-4">
        <div class="flex items-center justify-between mb-3">
          <h3 class="font-bold text-slate-200 uppercase tracking-wide text-xs">Dual-Audience CDR 1.0.0</h3>
          <div class="flex rounded-md bg-slate-900 p-0.5 border border-slate-800">
            <button
              data-testid="toggle-clinician-view"
              class="px-2.5 py-1 rounded text-xs font-semibold transition"
              :class="audienceView === 'clinician' ? 'bg-cyan-600 text-white shadow' : 'text-slate-400 hover:text-slate-200'"
              @click="setAudienceView('clinician')"
            >
              Clinician View
            </button>
            <button
              data-testid="toggle-patient-view"
              class="px-2.5 py-1 rounded text-xs font-semibold transition"
              :class="audienceView === 'patient' ? 'bg-teal-600 text-white shadow' : 'text-slate-400 hover:text-slate-200'"
              @click="setAudienceView('patient')"
            >
              Explain to Patient
            </button>
          </div>
        </div>

        <!-- Clinician View -->
        <div
          v-if="audienceView === 'clinician'"
          data-testid="cdr-view-clinician"
          class="space-y-3 bg-slate-900/80 p-3 rounded-lg border border-slate-800"
        >
          <div class="flex items-center justify-between">
            <span class="font-semibold text-cyan-400 text-xs">{{ stats.cdrSample.condition }}</span>
            <span class="text-xs font-mono text-slate-400">WORM SHA-256 AUDITED</span>
          </div>
          <div>
            <div class="text-xs font-semibold uppercase text-slate-400 mb-1">Triggering Rules & Guidelines</div>
            <ul class="list-disc list-inside space-y-1 text-slate-300 text-xs">
              <li v-for="(rule, idx) in stats.cdrSample.clinicianRules" :key="idx">{{ rule }}</li>
            </ul>
          </div>
          <div class="border-t border-slate-800 pt-2 text-xs text-slate-400 font-mono">
            <span data-testid="triage-cdr-record-id">record_id: {{ stats.csoSignoff.merkleRoot }}</span>
          </div>
        </div>

        <!-- Explain to Patient View -->
        <div
          v-else
          data-testid="cdr-view-patient"
          class="space-y-3 bg-teal-950/20 p-3 rounded-lg border border-teal-800/40 text-slate-200"
        >
          <div data-testid="cdr-patient-summary" class="rounded-lg bg-teal-900/30 p-3 border border-teal-600/30 text-teal-200 text-xs leading-relaxed font-sans font-medium">
            {{ stats.cdrSample.patientSummary }}
          </div>

          <div>
            <div class="text-xs font-semibold uppercase text-teal-400 mb-1">Ruled-Out Dangers</div>
            <ul data-testid="cdr-patient-excluded" class="list-disc list-inside space-y-1 text-slate-300 text-xs">
              <li v-for="(danger, idx) in stats.cdrSample.ruledOutDangers" :key="idx">{{ danger }}</li>
            </ul>
          </div>

          <div>
            <div class="text-xs font-semibold uppercase text-teal-400 mb-1">Key Observations</div>
            <ul data-testid="cdr-patient-facts" class="list-disc list-inside space-y-1 text-slate-300 text-xs">
              <li v-for="(fact, idx) in stats.cdrSample.keyObservations" :key="idx">{{ fact }}</li>
            </ul>
          </div>

          <div data-testid="cdr-patient-uncertainty" class="text-xs text-amber-300/90 bg-amber-950/20 p-2.5 rounded-lg border border-amber-600/30">
            Notice: {{ stats.cdrSample.uncertaintyNotice }}
          </div>
        </div>
      </section>

      <!-- Authority & DID Reference -->
      <section class="text-xs text-slate-400 border-t border-slate-800 pt-3 space-y-1">
        <div>DID: <span class="font-mono text-slate-300">{{ stats.did }}</span></div>
        <div>Covenant: <span class="font-mono text-slate-300">{{ stats.careCovenant }}</span></div>
        <div>Assurance: <span class="font-mono text-slate-300">{{ stats.assuranceProfile }}</span></div>
      </section>
    </div>
  </aside>
</template>
</template>

<script lang="ts" setup>
import { onClickOutside } from '@vueuse/core';
import { ref } from 'vue';

import useSignalBox from '~/compositions/useSignalBox';

const signalBox = useSignalBox();
const { isOpen, close, stats, audienceView, setAudienceView } = signalBox;

const panelRef = ref<HTMLElement>();
onClickOutside(panelRef, () => {
  if (isOpen.value) {
    close();
  }
}, { ignore: ['.signalbox-toggle-btn'] });
</script>

<style scoped>
.signalbox-panel {
  animation: slideIn 0.25s ease-out;
}

@keyframes slideIn {
  from {
    transform: translateX(100%);
  }
  to {
    transform: translateX(0);
  }
}
</style>
