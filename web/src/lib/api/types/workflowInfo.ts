// WorkflowInfo is one workflow a manual pipeline or cron job can select, and
// the workflows it requires. Optional dependencies are not listed: they are
// dropped when absent, so they never make a selection invalid.
export interface WorkflowInfo {
  name: string;
  depends_on?: string[];
}
