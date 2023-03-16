export type Task = {
  id: number;
  data: string;
  labels: { [key: string]: string };
  dependencies: string[];
  dep_status: { [key: string]: string };
  run_on: string[];
};

export type QueueInfo = {
  pending: Task[];
  waiting_on_deps: Task[];
  running: Task[];
  stats: {
    worker_count: number;
    pending_count: number;
    waiting_on_deps_count: number;
    running_count: number;
    completed_count: number;
  };
  paused: boolean;
};
