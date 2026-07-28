// Copyright 2024 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package scheduler contains a small, cli-local DAG runner for workflow
// items built by pipeline/frontend/builder. It sequences workflows by
// their depends_on relationships and runs ready workflows in parallel
// up to a caller-configured cap.
//
// This package is deliberately independent of the server. The server
// has its own scheduling implementation in server/queue/fifo.go with
// different requirements (persistence, agent distribution, priorities,
// cross-pipeline fairness). A future refactor may unify the two, but
// for now the cli-local runner is small enough to keep on its own.
//
// The scheduler emits workflow-level state transitions on an optional
// events channel. Step-level tracing and log lines are NOT handled
// here; those are the responsibility of the pipeline runtime tracer
// and logger that the caller plugs into its run function. This keeps
// the scheduler agnostic of rendering concerns — the same package
// backs both the TUI and the plain line-mode output paths.
package scheduler
