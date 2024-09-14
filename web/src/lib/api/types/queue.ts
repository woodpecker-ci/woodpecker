export interface Task {
  id: number;
  labels: { [key: string]: string };
  dependencies: string[];
  dep_status: { [key: string]: string };
  run_on: string[];
  agent_id: number;
}

export interface QueueStats {
  worker_count: number;
  pending_count: number;
  waiting_on_deps_count: number;
  running_count: number;
  completed_count: number;
}

export interface QueueInfo {
  pending: Task[];
  waiting_on_deps: Task[];
  running: Task[];
  stats: QueueStats;
  paused: boolean;
}
