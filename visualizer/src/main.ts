import "./styles.css"
import { solveProlog } from "./api"
import { examplePresets, type ExamplePreset } from "./examples"
import type { SolveResponse, TraceEvent, TraceEventType } from "./types"

function getElement<T extends HTMLElement>(id: string): T {
  const element = document.getElementById(id)

  if (!element) {
    throw new Error(`Missing element: #${id}`)
  }

  return element as T
}

const programInput = getElement<HTMLTextAreaElement>("program-input")
const queryInput = getElement<HTMLTextAreaElement>("query-input")
const exampleSelect = getElement<HTMLSelectElement>("example-select")
const exampleDescription = getElement<HTMLParagraphElement>("example-description")

const loadExampleButton = getElement<HTMLButtonElement>("load-example")
const runButton = getElement<HTMLButtonElement>("run-demo")
const resetButton = getElement<HTMLButtonElement>("reset-demo")
const previousButton = getElement<HTMLButtonElement>("previous-step")
const nextButton = getElement<HTMLButtonElement>("next-step")
const playButton = getElement<HTMLButtonElement>("play-trace")
const playbackSpeed = getElement<HTMLSelectElement>("playback-speed")
const traceScrubber = getElement<HTMLInputElement>("trace-scrubber")

const stepCounter = getElement<HTMLSpanElement>("step-counter")
const eventKind = getElement<HTMLSpanElement>("event-kind")
const eventDescription = getElement<HTMLParagraphElement>("event-description")
const currentGoal = getElement<HTMLElement>("current-goal")
const currentClause = getElement<HTMLElement>("current-clause")
const expandedGoalsList = getElement<HTMLDivElement>("expanded-goals")
const bindingsList = getElement<HTMLDivElement>("bindings")
const traceList = getElement<HTMLOListElement>("trace-list")
const traceEmpty = getElement<HTMLDivElement>("trace-empty")
const traceScroll = getElement<HTMLDivElement>("trace-scroll")
const answersList = getElement<HTMLDivElement>("answers")
const solutionCount = getElement<HTMLElement>("solution-count")
const traceCount = getElement<HTMLElement>("trace-count")
const errorBox = getElement<HTMLDivElement>("error-message")
const errorDetail = getElement<HTMLSpanElement>("error-detail")

const labels: Record<TraceEventType, string> = {
  goal: "Goal",
  builtin: "Built-in",
  try_clause: "Try clause",
  unified: "Unified",
  rule_expanded: "Rule expanded",
  failed: "Failed",
  backtrack: "Backtrack",
  solution: "Solution",
}

type RunStatus = "idle" | "loading" | "success" | "error"

const state: {
  events: TraceEvent[]
  answers: SolveResponse["answers"]
  currentIndex: number
  playTimer?: number
  status: RunStatus
  selectedExample: string
} = {
  events: [],
  answers: [],
  currentIndex: -1,
  status: "idle",
  selectedExample: examplePresets[0]?.id ?? "",
}

populateExamples()
loadExample(examplePresets[0])

loadExampleButton.addEventListener("click", () => {
  const preset = selectedPreset()
  if (preset) {
    loadExample(preset)
  }
})

exampleSelect.addEventListener("change", () => {
  const preset = selectedPreset()
  exampleDescription.textContent = preset?.description ?? ""
})

runButton.addEventListener("click", runQuery)
resetButton.addEventListener("click", resetCurrentExample)
previousButton.addEventListener("click", previousStep)
nextButton.addEventListener("click", nextStep)
playButton.addEventListener("click", togglePlayback)
playbackSpeed.addEventListener("change", restartPlaybackAtNewSpeed)
traceScrubber.addEventListener("input", scrubTrace)

document.addEventListener("keydown", (event) => {
  if ((event.ctrlKey || event.metaKey) && event.key === "Enter") {
    event.preventDefault()
    void runQuery()
  }
})

render()

function populateExamples(): void {
  const fragment = document.createDocumentFragment()

  for (const preset of examplePresets) {
    const option = document.createElement("option")
    option.value = preset.id
    option.textContent = preset.name
    fragment.append(option)
  }

  exampleSelect.replaceChildren(fragment)
  exampleSelect.value = state.selectedExample
}

function selectedPreset(): ExamplePreset | undefined {
  return examplePresets.find((preset) => preset.id === exampleSelect.value)
}

function loadExample(preset: ExamplePreset): void {
  stopPlayback()

  state.selectedExample = preset.id
  exampleSelect.value = preset.id
  exampleDescription.textContent = preset.description
  programInput.value = preset.program
  queryInput.value = preset.query

  clearRunState()
  clearError()
  render()
}

function resetCurrentExample(): void {
  const preset =
    examplePresets.find((item) => item.id === state.selectedExample) ??
    examplePresets[0]

  if (preset) {
    loadExample(preset)
  }
}

async function runQuery(): Promise<void> {
  stopPlayback()

  const program = programInput.value.trim()
  const query = queryInput.value.trim()

  if (program.length === 0) {
    showError("Program is required.")
    programInput.focus()
    return
  }

  if (query.length === 0) {
    showError("Query is required.")
    queryInput.focus()
    return
  }

  state.status = "loading"
  state.events = []
  state.answers = []
  state.currentIndex = -1

  clearError()
  render()

  try {
    const data = await solveProlog({ program, query })

    state.events = data.events
    state.answers = data.answers
    state.currentIndex = data.events.length > 0 ? 0 : -1
    state.status = "success"
  } catch (error) {
    state.status = "error"
    showError(error instanceof Error ? error.message : "Could not load the trace.")
  }

  render()
}

function clearRunState(): void {
  stopPlayback()
  state.events = []
  state.answers = []
  state.currentIndex = -1
  state.status = "idle"
}

function previousStep(): void {
  stopPlayback()

  if (state.currentIndex > 0) {
    state.currentIndex -= 1
    render()
    revealCurrentTraceRow()
  }
}

function nextStep(): void {
  stopPlayback()

  if (state.currentIndex < state.events.length - 1) {
    state.currentIndex += 1
    render()
    revealCurrentTraceRow()
  }
}

function playbackDelay(): number {
  const parsed = Number.parseInt(playbackSpeed.value, 10)
  return Number.isFinite(parsed) ? parsed : 750
}

function togglePlayback(): void {
  if (state.playTimer !== undefined) {
    stopPlayback()
    render()
    return
  }

  if (state.events.length === 0) {
    return
  }

  if (state.currentIndex >= state.events.length - 1) {
    state.currentIndex = 0
  }

  startPlayback()
}

function startPlayback(): void {
  state.playTimer = window.setInterval(() => {
    if (state.currentIndex >= state.events.length - 1) {
      stopPlayback()
      render()
      return
    }

    state.currentIndex += 1
    render()
    revealCurrentTraceRow()
  }, playbackDelay())

  render()
}

function restartPlaybackAtNewSpeed(): void {
  if (state.playTimer === undefined) {
    return
  }

  stopPlayback()
  startPlayback()
}

function stopPlayback(): void {
  if (state.playTimer !== undefined) {
    window.clearInterval(state.playTimer)
    state.playTimer = undefined
  }
}

function scrubTrace(): void {
  stopPlayback()

  if (state.events.length === 0) {
    return
  }

  state.currentIndex = Number.parseInt(traceScrubber.value, 10) - 1
  render()
  revealCurrentTraceRow()
}

function formatTraceText(text: string): string {
  // $1_Y becomes Y#1 so standardized-apart variables are easier to read.
  return text.replace(
    /\$(\d+)_([A-Za-z][A-Za-z0-9_]*)/g,
    (_match, id: string, name: string) => `${name}#${id}`,
  )
}

function render(): void {
  const hasTrace = state.events.length > 0 && state.currentIndex >= 0
  const atFirstStep = state.currentIndex <= 0
  const atLastStep = state.currentIndex >= state.events.length - 1
  const loading = state.status === "loading"

  runButton.disabled = loading
  resetButton.disabled = loading
  loadExampleButton.disabled = loading
  exampleSelect.disabled = loading
  runButton.textContent = loading ? "Running…" : "Run query"

  previousButton.disabled = !hasTrace || atFirstStep
  nextButton.disabled = !hasTrace || atLastStep
  playButton.disabled = !hasTrace
  playbackSpeed.disabled = !hasTrace
  playButton.textContent = state.playTimer === undefined ? "Play" : "Pause"

  traceScrubber.disabled = !hasTrace
  traceScrubber.max = String(Math.max(1, state.events.length))
  traceScrubber.value = String(Math.max(1, state.currentIndex + 1))

  solutionCount.textContent = state.status === "success" ? String(state.answers.length) : "—"
  traceCount.textContent = state.status === "success" ? String(state.events.length) : "—"

  renderAnswers()
  renderTrace()

  if (!hasTrace) {
    renderEmptyInspector()
    return
  }

  const event = state.events[state.currentIndex]

  stepCounter.textContent = `Step ${state.currentIndex + 1} of ${state.events.length}`
  eventKind.textContent = labels[event.type]
  eventKind.dataset.type = event.type
  eventDescription.textContent = event.description
  currentGoal.textContent = event.goal ? formatTraceText(event.goal) : "—"
  currentClause.textContent = event.clause ? formatTraceText(event.clause) : "—"

  renderExpandedGoals(event.expandedGoals)
  renderBindings(event.bindings)
}

function renderEmptyInspector(): void {
  eventKind.dataset.type = ""
  currentGoal.textContent = "—"
  currentClause.textContent = "—"
  renderExpandedGoals()
  renderBindings()

  switch (state.status) {
    case "loading":
      stepCounter.textContent = "Working…"
      eventKind.textContent = "Running"
      eventDescription.textContent = "Parsing the program and exploring the query."
      break
    case "error":
      stepCounter.textContent = "No trace loaded"
      eventKind.textContent = "Error"
      eventDescription.textContent = "Fix the program or query, then run it again."
      break
    default:
      stepCounter.textContent = "No trace loaded"
      eventKind.textContent = "Waiting"
      eventDescription.textContent =
        "Run a query to receive execution events from the Go interpreter."
  }
}

function renderExpandedGoals(goals?: string[]): void {
  expandedGoalsList.replaceChildren()

  if (!goals || goals.length === 0) {
    expandedGoalsList.append(createEmptyState("No rule expansion at this step."))
    return
  }

  for (const goal of goals) {
    const item = document.createElement("code")
    item.className = "expanded-goal"
    item.textContent = formatTraceText(goal)
    expandedGoalsList.append(item)
  }
}

function renderBindings(bindings?: Record<string, string>): void {
  bindingsList.replaceChildren()

  const entries = Object.entries(bindings ?? {})
    .filter(([variable]) => !variable.startsWith("$"))
    .sort(([left], [right]) => left.localeCompare(right))

  if (entries.length === 0) {
    bindingsList.append(createEmptyState("No query variable bindings yet."))
    return
  }

  for (const [variable, value] of entries) {
    const row = document.createElement("div")
    row.className = "binding-row"

    const variableElement = document.createElement("code")
    variableElement.textContent = variable

    const equalsElement = document.createElement("span")
    equalsElement.textContent = "="

    const valueElement = document.createElement("code")
    valueElement.textContent = formatTraceText(value)

    row.append(variableElement, equalsElement, valueElement)
    bindingsList.append(row)
  }
}

function renderTrace(): void {
  traceList.replaceChildren()
  traceEmpty.hidden = state.events.length > 0

  if (state.events.length === 0) {
    return
  }

  const fragment = document.createDocumentFragment()

  state.events.forEach((event, index) => {
    const item = document.createElement("li")
    item.className = "trace-item"

    const row = document.createElement("button")
    const safeDepth = Math.max(0, Math.min(event.depth, 12))

    row.type = "button"
    row.className = `trace-row trace-${event.type}`
    row.dataset.index = String(index)
    row.style.setProperty("--depth", String(safeDepth))
    row.setAttribute("aria-label", `Jump to step ${index + 1}: ${labels[event.type]}`)

    if (index === state.currentIndex) {
      row.classList.add("is-current")
      row.setAttribute("aria-current", "step")
    }

    if (index > state.currentIndex) {
      row.classList.add("is-future")
    }

    row.addEventListener("click", () => {
      stopPlayback()
      state.currentIndex = index
      render()
    })

    const heading = document.createElement("div")
    heading.className = "trace-heading"

    const step = document.createElement("span")
    step.className = "trace-step"
    step.textContent = String(index + 1).padStart(2, "0")

    const tag = document.createElement("span")
    tag.className = "trace-tag"
    tag.textContent = labels[event.type]

    const summary = document.createElement("strong")
    summary.textContent = event.description

    heading.append(step, tag, summary)

    const details = document.createElement("p")
    details.className = "trace-details"

    const context: string[] = []

    if (event.goal) {
      context.push(`Goal: ${formatTraceText(event.goal)}`)
    }

    if (event.clause) {
      context.push(`Clause: ${formatTraceText(event.clause)}`)
    }

    if (event.expandedGoals && event.expandedGoals.length > 0) {
      context.push(`Expands to: ${event.expandedGoals.map(formatTraceText).join(", ")}`)
    }

    details.textContent =
      context.length > 0 ? context.join("  •  ") : "No additional context for this event."

    row.append(heading, details)
    item.append(row)
    fragment.append(item)
  })

  traceList.append(fragment)
}

function renderAnswers(): void {
  answersList.replaceChildren()

  if (state.status === "idle") {
    answersList.append(createEmptyState("Run a query to see its solutions."))
    return
  }

  if (state.status === "loading") {
    answersList.append(createEmptyState("Solving…"))
    return
  }

  if (state.status === "error") {
    answersList.append(createEmptyState("No result was produced."))
    return
  }

  if (state.answers.length === 0) {
    const line = document.createElement("p")
    line.className = "answer answer-false"
    line.textContent = "false."
    answersList.append(line)
    return
  }

  state.answers.forEach((answer, index) => {
    const line = document.createElement("div")
    line.className = "answer"

    const answerNumber = document.createElement("span")
    answerNumber.className = "answer-number"
    answerNumber.textContent = `#${index + 1}`

    const answerValue = document.createElement("code")
    const bindings = Object.entries(answer)
      .filter(([variable]) => !variable.startsWith("$"))
      .map(([variable, value]) => `${variable} = ${formatTraceText(value)}`)
      .join(", ")

    answerValue.textContent = bindings || "true."
    line.append(answerNumber, answerValue)
    answersList.append(line)
  })
}

function createEmptyState(message: string): HTMLParagraphElement {
  const empty = document.createElement("p")
  empty.className = "empty-state"
  empty.textContent = message
  return empty
}

function revealCurrentTraceRow(): void {
  const current = traceList.querySelector<HTMLElement>(
    `.trace-row[data-index="${state.currentIndex}"]`,
  )

  if (!current) {
    return
  }

  const rowTop = current.offsetTop
  const rowBottom = rowTop + current.offsetHeight
  const viewportTop = traceScroll.scrollTop
  const viewportBottom = viewportTop + traceScroll.clientHeight

  if (rowTop < viewportTop || rowBottom > viewportBottom) {
    current.scrollIntoView({ block: "nearest", behavior: "smooth" })
  }
}

function showError(message: string): void {
  errorBox.hidden = false
  errorDetail.textContent = message
}

function clearError(): void {
  errorBox.hidden = true
  errorDetail.textContent = ""
}
