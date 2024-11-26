export interface Agent {
  id: number;
  name: string;
  owner_id: number;
  org_id: number;
  token: string;
  created: number;
  updated: number;
  last_contact: number;
  platform: string;
  backend: string;
  capacity: number;
  version: string;
  no_schedule: boolean;
  custom_labels: Record<string, string>;
}
