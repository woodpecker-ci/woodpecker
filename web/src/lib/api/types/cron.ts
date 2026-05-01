export interface Cron {
  id: number;
  name: string;
  branch: string;
  schedule: string;
  enabled: boolean;
  next_exec: number;
  variables: Record<string, string>;
}
