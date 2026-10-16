export interface Cron {
  id: number;
  name: string;
  branch: string;
  schedule: string;
  timezone: string;
  enabled: boolean;
  next_exec: number;
  variables: Record<string, string>;
  // workflows narrows the run to the named workflows. An empty list runs all of them.
  workflows?: string[];
}
