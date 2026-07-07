export type TraceEventType =
  | "goal"
  | "try_clause"
  | "unified"
  | "rule_expanded"
  | "failed"
  | "backtrack"
  | "solution";

export interface TraceEvent {
  type: TraceEventType;
  depth: number;
  goal?: string;
  clause?: string;
  expandedGoals?: string[];
  bindings?: Record<string, string>;
  description: string;
}

export interface SolveResponse {
  events: TraceEvent[];
  answers: Record<string, string>[];
}