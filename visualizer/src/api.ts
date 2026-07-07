import type { SolveResponse, TraceEvent, TraceEventType } from "./types"

const eventTypes = new Set<TraceEventType>([
  "goal",
  "try_clause",
  "unified",
  "rule_expanded",
  "failed",
  "backtrack",
  "solution"
])

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null
}

function isStringArray(value: unknown): value is string[] {
  return Array.isArray(value) && value.every((item) => typeof item === "string")
}

function isTraceEvent(value: unknown): value is TraceEvent {
  if (!isRecord(value)) {
    return false
  }

  const hasValidClause =
    value.clause === undefined || typeof value.clause === "string"
  const hasValidExpandedGoals =
    value.expandedGoals === undefined || isStringArray(value.expandedGoals)

  return (
    typeof value.type === "string" &&
    eventTypes.has(value.type as TraceEventType) &&
    typeof value.depth === "number" &&
    typeof value.description === "string" &&
    hasValidClause &&
    hasValidExpandedGoals
  )
}

function toStringRecord(value: Record<string, unknown>): Record<string, string> {
  const result: Record<string, string> = {}
  
  for (const [key, item] of Object.entries(value)) {
    result[key] = String(item)
  }

  return result
}

export async function fetchDemoTrace(): Promise<SolveResponse> {
  const response = await fetch("/api/demo")

  if (!response.ok) {
    throw new Error(`The Go server returned ${response.status}.`)
  }

  const body: unknown = await response.json()

  if (!isRecord(body) || !Array.isArray(body.events)) {
    throw new Error("The API response did not include a trace.")
  }

  if (!body.events.every(isTraceEvent)) {
    throw new Error("The API returned an invalid trace event.")
  }

  const answers = Array.isArray(body.answers)
    ? body.answers.filter(isRecord).map(toStringRecord)
    : []

  return {
    events: body.events,
    answers
  }
}