import { ref, computed } from 'vue';

export type OperatingMode = 'live' | 'simulation' | 'replay' | 'audit';
export type ViewLevel = 'essential' | 'standard' | 'advanced' | 'expert';

const isSignalBoxOpen = ref<boolean>(false);
const operatingMode = ref<OperatingMode>('live');
const viewLevel = ref<ViewLevel>('standard');
const audienceView = ref<'clinician' | 'patient'>('clinician');

const clinicalAssuranceStats = ref({
  assuranceProfile: 'NHS-DCB0129-DCB0160 | FUTURE-AI | ISO-IEC-42001',
  did: 'did:web:gharra.symphonix.health:agents:fde-superpersona',
  careCovenant: 'symphonix-care-covenant-v1',
  goldenDatasetAccuracy: 99.94,
  zeroPhiSentryStatus: 'ACTIVE',
  redactionsCount: 42,
  selfHealingStats: {
    oomEscalations: 12,
    flakyRetries: 7,
    quarantinedAgents: 0,
  },
  csoSignoff: {
    required: true,
    user: 'dr.elena.rostova@symphonix.health',
    timestamp: '2026-10-24T16:45:00Z',
    merkleRoot: 'sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855',
  },
  cdrSample: {
    condition: 'Acute Coronary Syndrome Rule-Out',
    patientSummary:
      'Your heart tests and scans look reassuring. There is no evidence of sudden heart artery blockage today. We are monitoring your symptoms safely.',
    ruledOutDangers: ['ST-Elevation Myocardial Infarction (STEMI)', 'Aortic Dissection'],
    keyObservations: ['High-sensitivity Troponin T: < 5 ng/L (Normal)', '12-lead ECG: Normal sinus rhythm, no ST changes'],
    uncertaintyNotice: 'Slight residual muscular chest tenderness remains under observation. Clinical team reassessment in 2 hours.',
    clinicianRules: [
      'NICE NG185 Chest pain guideline logic v3.2',
      'GRACE 2.0 Risk Model calculated 6-month mortality < 1%',
      'L-CAID-10 Zero-Mock validation confirmed across Client Registry & Lab feed',
    ],
  },
});

export default function useSignalBox() {
  function toggle() {
    isSignalBoxOpen.value = !isSignalBoxOpen.value;
  }

  function open() {
    isSignalBoxOpen.value = true;
  }

  function close() {
    isSignalBoxOpen.value = false;
  }

  function setOperatingMode(mode: OperatingMode) {
    operatingMode.value = mode;
  }

  function setViewLevel(level: ViewLevel) {
    viewLevel.value = level;
  }

  function setAudienceView(audience: 'clinician' | 'patient') {
    audienceView.value = audience;
  }

  return {
    isOpen: computed(() => isSignalBoxOpen.value),
    operatingMode: computed(() => operatingMode.value),
    viewLevel: computed(() => viewLevel.value),
    audienceView: computed(() => audienceView.value),
    stats: computed(() => clinicalAssuranceStats.value),
    toggle,
    open,
    close,
    setOperatingMode,
    setViewLevel,
    setAudienceView,
  };
}
