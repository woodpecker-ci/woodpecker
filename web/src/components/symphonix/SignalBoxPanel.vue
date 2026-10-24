<template>
  <aside
    v-if="isOpen"
    ref="panelRef"
    data-testid="signalbox-panel"
    class="signalbox-panel flex h-[calc(100%-1.25rem)] my-2.5 mr-4 w-80 xl:w-96 shrink-0 flex-col rounded-xl border border-slate-800 bg-slate-900/95 text-slate-100 shadow-2xl overflow-hidden transition-all duration-300 font-sans z-10"
    aria-label="SignalBox Clinical Assurance & Governance Panel"
  >
    <!-- Panel Header -->
    <div class="flex items-center justify-between border-b border-slate-800 px-4 py-2.5 shrink-0 bg-slate-900/90">
      <div class="flex items-center space-x-2.5">
        <div class="flex h-7 w-7 items-center justify-center rounded bg-cyan-950 text-cyan-400 border border-cyan-500/40 shrink-0">
          <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
          </svg>
        </div>
        <div>
          <h2 class="text-xs font-bold tracking-wide text-white uppercase">SignalBox Governance</h2>
          <p class="text-[11px] text-slate-400">CAID / DCB0129-DCB0160 Active</p>
        </div>
      </div>
      <button
        data-testid="signalbox-close"
        class="rounded p-1 text-slate-400 hover:bg-slate-800 hover:text-white transition"
        title="Close Panel"
        @click="close"
      >
        <svg class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
        </svg>
      </button>
    </div>

    <!-- Navigation Tabs (Prevents Vertical Scroll) -->
    <div class="flex border-b border-slate-800 bg-slate-950 shrink-0 text-xs">
      <button
        type="button"
        class="flex-1 py-2 text-center transition font-semibold"
        :class="activeTab === 'gate' ? 'text-cyan-400 border-b-2 border-cyan-400' : 'text-slate-400 hover:text-slate-200 border-b-2 border-transparent'"
        @click="activeTab = 'gate'"
      >
        Clinical Gate
      </button>
      <button
        type="button"
        class="flex-1 py-2 text-center transition font-semibold"
        :class="activeTab === 'heal' ? 'text-cyan-400 border-b-2 border-cyan-400' : 'text-slate-400 hover:text-slate-200 border-b-2 border-transparent'"
        @click="activeTab = 'heal'"
      >
        Self-Healing
      </button>
      <button
        type="button"
        class="flex-1 py-2 text-center transition font-semibold"
        :class="activeTab === 'cdr' ? 'text-cyan-400 border-b-2 border-cyan-400' : 'text-slate-400 hover:text-slate-200 border-b-2 border-transparent'"
        @click="activeTab = 'cdr'"
      >
        CDR 1.0.0
      </button>
    </div>

    <!-- Tab Content Area: 100% Fits in Viewport -->
    <div class="flex-1 p-4 overflow-hidden flex flex-col justify-between text-xs">
      <!-- Tab 1: Clinical Gate & CSO -->
      <div v-if="activeTab === 'gate'" class="space-y-3.5 flex-1 flex flex-col justify-start">
        <div class="rounded-lg border border-emerald-500/30 bg-emerald-950/20 p-3">
          <div class="flex items-center justify-between mb-1.5">
            <span class="font-bold text-emerald-400 uppercase tracking-wide flex items-center gap-1.5 text-xs">
              <span class="h-2 w-2 rounded-full bg-emerald-400 animate-pulse"></span>
              CSO Clinical Gate Passed
            </span>
            <span class="rounded bg-emerald-900/60 px-1.5 py-0.5 font-mono text-[11px] text-emerald-300 border border-emerald-600/40">
              ACCOUNTABLE_HUMAN
            </span>
          </div>
          <p class="text-slate-300 text-xs">
            Signoff: <span class="font-mono text-emerald-300 font-semibold">{{ stats.csoSignoff.user }}</span>
          </p>
          <p class="text-slate-400 text-xs mt-0.5">
            Timestamp: <span class="font-mono text-slate-300">{{ stats.csoSignoff.timestamp }}</span>
          </p>
          <div class="mt-2 text-[11px] text-slate-400 font-mono break-all bg-slate-950 p-1.5 rounded border border-slate-800">
            Merkle Root: {{ stats.csoSignoff.merkleRoot }}
          </div>
        </div>

        <div class="grid grid-cols-2 gap-2.5">
          <div class="rounded-lg border border-slate-800 bg-slate-950/60 p-3">
            <div class="text-slate-400 text-xs uppercase font-semibold">Golden Dataset</div>
            <div class="mt-0.5 text-xl font-bold font-mono text-cyan-400">{{ stats.goldenDatasetAccuracy }}%</div>
            <div class="text-[11px] text-slate-400 mt-0.5">Target: &ge; 99.9% zero-miss</div>
          </div>

          <div class="rounded-lg border border-slate-800 bg-slate-950/60 p-3">
            <div class="text-slate-400 text-xs uppercase font-semibold">Zero-PHI Sentry</div>
            <div class="mt-0.5 text-xl font-bold font-mono text-indigo-400">{{ stats.zeroPhiSentryStatus }}</div>
            <div class="text-[11px] text-slate-400 mt-0.5">{{ stats.redactionsCount }} IDs Redacted</div>
          </div>
        </div>
      </div>

      <!-- Tab 2: Self-Healing Telemetry -->
      <div v-else-if="activeTab === 'heal'" class="space-y-2.5 flex-1 flex flex-col justify-start">
        <div class="rounded-lg border border-slate-800 bg-slate-950/60 p-2.5">
          <div class="flex items-center justify-between mb-2">
            <h3 class="font-bold text-sky-400 uppercase tracking-wide flex items-center gap-1.5 text-xs">
              <svg class="h-4 w-4 text-sky-400 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
              Autonomous Self-Healing
            </h3>
            <span class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-mono font-bold bg-cyan-950 text-cyan-300 border border-cyan-800/60">
              Qwen 2.5 Coder 32B (Air-Gap)
            </span>
          </div>
          <div class="grid grid-cols-3 gap-1.5 text-center">
            <div class="bg-slate-900 p-1.5 rounded border border-slate-800">
              <div class="font-mono text-sm font-bold text-cyan-400">{{ stats.selfHealingStats.oomEscalations }}</div>
              <div class="text-[10px] text-slate-400">L1 OOM Scaled</div>
            </div>
            <div class="bg-slate-900 p-1.5 rounded border border-slate-800">
              <div class="font-mono text-sm font-bold text-amber-400">{{ stats.selfHealingStats.flakyRetries }}</div>
              <div class="text-[10px] text-slate-400">L2 AI Retries</div>
            </div>
            <div class="bg-slate-900 p-1.5 rounded border border-slate-800">
              <div class="font-mono text-sm font-bold text-emerald-400">{{ stats.selfHealingStats.quarantinedAgents }}</div>
              <div class="text-[10px] text-slate-400">L3 Quarantines</div>
            </div>
          </div>
        </div>

        <!-- DCB0129 Code Provenance Gate Status -->
        <div class="rounded-lg border border-slate-800 bg-slate-950/60 p-2 text-xs text-slate-300 space-y-1">
          <div class="flex items-center justify-between text-[11px] font-semibold">
            <span class="text-slate-400 uppercase">DCB0129 Code Provenance:</span>
            <span class="text-emerald-400 font-mono">DUAL-KEY ENFORCED</span>
          </div>
          <div class="text-[11px] text-slate-400 font-mono flex items-center justify-between">
            <span>Author: <span class="text-slate-300">symphonix-healer[bot]</span></span>
            <span>DID: <span class="text-cyan-400">healer-v1</span></span>
          </div>
          <div class="text-[10px] text-indigo-300 bg-indigo-950/40 p-1 rounded border border-indigo-900/50">
            Direct merge blocked: 2 human sign-offs (incl. Certified CSO) required.
          </div>
        </div>

        <!-- 6 Scoped Self-Healing Primitives -->
        <div class="grid grid-cols-2 gap-1.5 text-[11px] font-mono">
          <div class="p-1.5 rounded bg-slate-900/80 border border-slate-800">
            <div class="text-slate-400 text-[10px] uppercase font-bold">Flaky Quarantine</div>
            <div class="text-emerald-400">Tiered (TTL 72h)</div>
            <div class="text-[9px] text-slate-500">Clinical tests immune</div>
          </div>
          <div class="p-1.5 rounded bg-slate-900/80 border border-slate-800">
            <div class="text-slate-400 text-[10px] uppercase font-bold">Predictive Sizing</div>
            <div class="text-cyan-400">P95 + 20% Headroom</div>
            <div class="text-[9px] text-slate-500">Zero Exit 137 OOM</div>
          </div>
          <div class="p-1.5 rounded bg-slate-900/80 border border-slate-800">
            <div class="text-slate-400 text-[10px] uppercase font-bold">Config Auto-Fix</div>
            <div class="text-amber-400">1-Click YAML Diff</div>
            <div class="text-[9px] text-slate-500">L-CAID-10 & Tink AEAD</div>
          </div>
          <div class="p-1.5 rounded bg-slate-900/80 border border-slate-800">
            <div class="text-slate-400 text-[10px] uppercase font-bold">Multi-Agent Gate</div>
            <div class="text-purple-400">Dev &times; Safety CSO</div>
            <div class="text-[9px] text-slate-500">Assertion relaxation check</div>
          </div>
        </div>
      </div>

      <!-- Tab 3: CDR 1.0.0 Explainability -->
      <div v-else-if="activeTab === 'cdr'" class="space-y-3 flex-1 flex flex-col justify-start">
        <div class="flex items-center justify-between">
          <h3 class="font-bold text-slate-200 uppercase tracking-wide text-xs">Dual-Audience CDR 1.0.0</h3>
          <div class="flex rounded bg-slate-950 p-0.5 border border-slate-800">
            <button
              data-testid="toggle-clinician-view"
              class="px-2 py-0.5 rounded text-xs font-semibold transition"
              :class="audienceView === 'clinician' ? 'bg-cyan-600 text-white shadow' : 'text-slate-400 hover:text-slate-200'"
              @click="setAudienceView('clinician')"
            >
              Clinician
            </button>
            <button
              data-testid="toggle-patient-view"
              class="px-2 py-0.5 rounded text-xs font-semibold transition"
              :class="audienceView === 'patient' ? 'bg-teal-600 text-white shadow' : 'text-slate-400 hover:text-slate-200'"
              @click="setAudienceView('patient')"
            >
              Patient
            </button>
          </div>
        </div>

        <!-- Clinician View -->
        <div
          v-if="audienceView === 'clinician'"
          data-testid="cdr-view-clinician"
          class="space-y-2 bg-slate-900/80 p-3 rounded-lg border border-slate-800"
        >
          <div class="flex items-center justify-between">
            <span class="font-semibold text-cyan-400 text-xs">{{ stats.cdrSample.condition }}</span>
            <span class="text-[11px] font-mono text-slate-400">WORM SHA-256</span>
          </div>
          <div>
            <div class="text-[11px] font-semibold uppercase text-slate-400 mb-0.5">Rules &amp; Guidelines</div>
            <ul class="list-disc list-inside space-y-0.5 text-slate-300 text-xs">
              <li v-for="(rule, idx) in stats.cdrSample.clinicianRules" :key="idx">{{ rule }}</li>
            </ul>
          </div>
          <div class="border-t border-slate-800 pt-1.5 text-[11px] text-slate-400 font-mono">
            <span data-testid="triage-cdr-record-id">record_id: {{ stats.csoSignoff.merkleRoot }}</span>
          </div>
        </div>

        <!-- Explain to Patient View -->
        <div
          v-else
          data-testid="cdr-view-patient"
          class="space-y-2 bg-teal-950/20 p-3 rounded-lg border border-teal-800/40 text-slate-200"
        >
          <div data-testid="cdr-patient-summary" class="rounded bg-teal-900/30 p-2 border border-teal-600/30 text-teal-200 text-xs leading-relaxed font-sans font-medium">
            {{ stats.cdrSample.patientSummary }}
          </div>

          <div>
            <div class="text-[11px] font-semibold uppercase text-teal-400 mb-0.5">Ruled-Out Dangers</div>
            <ul data-testid="cdr-patient-excluded" class="list-disc list-inside space-y-0.5 text-slate-300 text-xs">
              <li v-for="(danger, idx) in stats.cdrSample.ruledOutDangers" :key="idx">{{ danger }}</li>
            </ul>
          </div>

          <div data-testid="cdr-patient-uncertainty" class="text-xs text-amber-300/90 bg-amber-950/20 p-2 rounded border border-amber-600/30">
            Notice: {{ stats.cdrSample.uncertaintyNotice }}
          </div>
        </div>
      </div>

      <!-- Authority Footer (Inside Drawer, Compact) -->
      <div class="text-[11px] text-slate-400 border-t border-slate-800 pt-2 shrink-0 space-y-0.5">
        <div>DID: <span class="font-mono text-slate-300">{{ stats.did }}</span></div>
        <div>Assurance: <span class="font-mono text-slate-300">{{ stats.assuranceProfile }}</span></div>
      </div>
    </div>
  </aside>
</template>

<script lang="ts" setup>
import { onClickOutside } from '@vueuse/core';
import { ref } from 'vue';

import useSignalBox from '~/compositions/useSignalBox';

const signalBox = useSignalBox();
const { isOpen, close, stats, audienceView, setAudienceView } = signalBox;

const activeTab = ref<'gate' | 'heal' | 'cdr'>('gate');

const panelRef = ref<HTMLElement>();
onClickOutside(panelRef, () => {
  if (isOpen.value) {
    close();
  }
}, { ignore: ['.signalbox-toggle-btn'] });
</script>
