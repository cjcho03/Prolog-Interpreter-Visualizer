export type TraceEventType =
  | "goal"
  | "try_fact"
  | "unified"
  | "failed"
  | "backtrack"
  | "solution";

export interface TraceEvent {
  type: TraceEventType;
  depth: number;
  goal?: string;
  fact?: string;
  bindings?: Record<string, string>;
  description: string;
}

export interface SolveResponse {
  events: TraceEvent[];
  answers: Record<string, string>[];
}