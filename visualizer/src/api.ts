import type { SolveRequest, SolveResponse, TraceEvent, TraceEventType } from "./types"

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

function isStringRecord(value: unknown): value is Record<string, string> {
  if (!isRecord(value)) {
    return false
  }

  return Object.values(value).every((item) => typeof item === "string")
}

function isTraceEvent(value: unknown): value is TraceEvent {
  if (!isRecord(value)) {
    return false
  }

  const hasValidGoal =
    value.goal === undefined || typeof value.goal === "string"
  const hasValidClause =
    value.clause === undefined || typeof value.clause === "string"
  const hasValidExpandedGoals =
    value.expandedGoals === undefined || isStringArray(value.expandedGoals)
  const hasValidBindngs =
    value.bindings === undefined || isStringRecord(value.bindings)

  return (
    typeof value.type === "string" &&
    eventTypes.has(value.type as TraceEventType) &&
    typeof value.depth === "number" &&
    typeof value.description === "string" &&
    hasValidGoal &&
    hasValidClause &&
    hasValidExpandedGoals &&
    hasValidBindngs
  )
}

function toStringRecord(value: Record<string, unknown>): Record<string, string> {
  const result: Record<string, string> = {}
  
  for (const [key, item] of Object.entries(value)) {
    result[key] = String(item)
  }

  return result
}

function parseSolveResponse(body: unknown): SolveResponse {
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
    answers,
  }
}

async function readErrorMessage(response: Response): Promise<string> {
  try {
    const body: unknown = await response.json()

    if (isRecord(body) && typeof body.error === "string") {
      return body.error
    }
  } catch {
    // Fall through to generic message
  }
  return `The Go server returned ${response.status}.`
}

export async function fetchDemoTrace(): Promise<SolveResponse> {
  const response = await fetch("/api/demo")

  if (!response.ok) {
    throw new Error(await readErrorMessage(response))
  }

  const body: unknown = await response.json()

  return parseSolveResponse(body)
}

export async function solveProlog(request: SolveRequest): Promise<SolveResponse> {
  const response = await fetch("/api/solve", {
    method: "POST",
    headers: {
      "Content-Type": "application/json"
    },
    body: JSON.stringify(request)
  })

  if (!response.ok) {
    throw new Error(await readErrorMessage(response))
  }

  const body: unknown = await response.json()

  return parseSolveResponse(body)
}