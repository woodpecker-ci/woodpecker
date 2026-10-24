<template>
  <footer
    data-testid="clinical-disclaimer-footer"
    class="border-t border-slate-700/50 bg-slate-900/90 text-slate-400 px-4 py-2 text-[11px] flex flex-wrap items-center justify-between gap-2 shadow-inner"
  >
    <div class="flex items-center space-x-2">
      <span class="inline-flex items-center justify-center h-4 w-4 rounded-full bg-cyan-900/80 text-cyan-300 font-bold text-[10px]">
        i
      </span>
      <span class="font-medium text-slate-300">
        Reference information about how this software works. Not clinical advice. Always follow your local clinical protocols.
      </span>
    </div>

    <div class="flex items-center space-x-3 text-[10px] text-slate-500 font-mono">
      <span>DID: did:web:gharra.symphonix.health:agents:fde-superpersona</span>
      <span>•</span>
      <span class="text-cyan-400">NHS-DCB0129 / DCB0160</span>
      <span>•</span>
      <span class="text-emerald-400">ACCOUNTABLE_HUMAN_REQUIRED</span>
    </div>
  </footer>
</template>

<script lang="ts" setup>
// Symphonix Design System Mandatory Clinical Disclaimer Footer
</script>
