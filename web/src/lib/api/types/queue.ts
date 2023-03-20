export type Task = {
  id: number;
  data: string;
  labels: { [key: string]: string };
  dependencies: string[];
  dep_status: { [key: string]: string };
  run_on: string[];
  agent_id: number;
};

export type QueueStats = {
  worker_count: number;
  pending_count: number;
  waiting_on_deps_count: number;
  running_count: number;
  completed_count: number;
};

export type QueueStats = {
  worker_count: number;
  pending_count: number;
  waiting_on_deps_count: number;
  running_count: number;
  completed_count: number;
};

export type QueueInfo = {
  pending: Task[];
  waiting_on_deps: Task[];
  running: Task[];
  stats: QueueStats;
  paused: boolean;
};
