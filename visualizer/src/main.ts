import "./styles.css";
import { solveProlog } from "./api";
import type { SolveResponse, TraceEvent, TraceEventType } from "./types";

const demoProgram = `parent(alice, bob).
parent(alice, carol).
parent(bob, diana).
parent(carol, eli).

grandparent(X, Z) :-
    parent(X, Y),
    parent(Y, Z).`;

const demoQuery = `?- grandparent(alice, Who).`;

function getElement<T extends HTMLElement>(id: string): T {
  const element = document.getElementById(id);

  if (!element) {
    throw new Error(`Missing element: #${id}`);
  }

  return element as T;
}

const programInput = getElement<HTMLTextAreaElement>("program-input");
const queryInput = getElement<HTMLTextAreaElement>("query-input");

const runButton = getElement<HTMLButtonElement>("run-demo");
const resetButton = getElement<HTMLButtonElement>("reset-demo");
const previousButton = getElement<HTMLButtonElement>("previous-step");
const nextButton = getElement<HTMLButtonElement>("next-step");
const playButton = getElement<HTMLButtonElement>("play-trace");

const stepCounter = getElement<HTMLParagraphElement>("step-counter");
const eventKind = getElement<HTMLParagraphElement>("event-kind");
const eventDescription = getElement<HTMLParagraphElement>("event-description");
const currentGoal = getElement<HTMLParagraphElement>("current-goal");
const currentClause = getElement<HTMLParagraphElement>("current-clause");
const expandedGoalsList = getElement<HTMLDivElement>("expanded-goals");
const bindingsList = getElement<HTMLDivElement>("bindings");
const traceList = getElement<HTMLOListElement>("trace-list");
const answersList = getElement<HTMLDivElement>("answers");
const errorBox = getElement<HTMLParagraphElement>("error-message");

const labels: Record<TraceEventType, string> = {
  goal: "Goal",
  try_clause: "Try Clause",
  unified: "Unified",
  rule_expanded: "Rule Expanded",
  failed: "Failed",
  backtrack: "Backtrack",
  solution: "Solution",
};

const state: {
  events: TraceEvent[];
  answers: SolveResponse["answers"];
  currentIndex: number;
  playTimer?: number;
  loading: boolean;
} = {
  events: [],
  answers: [],
  currentIndex: -1,
  loading: false,
};

programInput.value = demoProgram;
queryInput.value = demoQuery;

runButton.addEventListener("click", runQuery);
resetButton.addEventListener("click", resetDemo);
previousButton.addEventListener("click", previousStep);
nextButton.addEventListener("click", nextStep);
playButton.addEventListener("click", togglePlayback);

render();

async function runQuery(): Promise<void> {
  stopPlayback();

  const program = programInput.value.trim();
  const query = queryInput.value.trim();

  if (program.length === 0) {
    showError("Program is required.");
    return;
  }

  if (query.length === 0) {
    showError("Query is required.");
    return;
  }

  state.loading = true;
  state.events = [];
  state.answers = [];
  state.currentIndex = -1;

  clearError();
  render();

  try {
    const data = await solveProlog({
      program,
      query,
    });

    state.events = data.events;
    state.answers = data.answers;
    state.currentIndex = data.events.length > 0 ? 0 : -1;
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "Could not load the trace.";

    showError(message);
  } finally {
    state.loading = false;
    render();
  }
}

function resetDemo(): void {
  stopPlayback();

  programInput.value = demoProgram;
  queryInput.value = demoQuery;

  state.events = [];
  state.answers = [];
  state.currentIndex = -1;

  clearError();
  render();
}

function previousStep(): void {
  stopPlayback();

  if (state.currentIndex > 0) {
    state.currentIndex -= 1;
  }

  render();
}

function nextStep(): void {
  stopPlayback();

  if (state.currentIndex < state.events.length - 1) {
    state.currentIndex += 1;
  }

  render();
}

function togglePlayback(): void {
  if (state.playTimer !== undefined) {
    stopPlayback();
    return;
  }

  if (state.events.length === 0) {
    return;
  }

  if (state.currentIndex >= state.events.length - 1) {
    state.currentIndex = 0;
  }

  state.playTimer = window.setInterval(() => {
    if (state.currentIndex >= state.events.length - 1) {
      stopPlayback();
      return;
    }

    state.currentIndex += 1;
    render();
  }, 900);

  render();
}

function stopPlayback(): void {
  if (state.playTimer !== undefined) {
    window.clearInterval(state.playTimer);
    state.playTimer = undefined;
  }
}

function formatTraceText(text: string): string {
  // $1_Y becomes Y#1 so standardized-apart variables are easier to read.
  return text.replace(
    /\$(\d+)_([A-Za-z][A-Za-z0-9_]*)/g,
    (_match, id: string, name: string) => `${name}#${id}`,
  );
}

function render(): void {
  const hasTrace = state.events.length > 0 && state.currentIndex >= 0;
  const atFirstStep = state.currentIndex <= 0;
  const atLastStep = state.currentIndex >= state.events.length - 1;

  runButton.disabled = state.loading;
  resetButton.disabled = state.loading;
  runButton.textContent = state.loading ? "Running..." : "Run query";

  previousButton.disabled = !hasTrace || atFirstStep;
  nextButton.disabled = !hasTrace || atLastStep;
  playButton.disabled = !hasTrace;
  playButton.textContent = state.playTimer === undefined ? "Play" : "Pause";

  renderAnswers();
  renderTrace();

  if (!hasTrace) {
    stepCounter.textContent = "No trace loaded.";
    eventKind.textContent = "Waiting";
    eventDescription.textContent =
      "Run a query to receive events from the Go interpreter.";
    currentGoal.textContent = "—";
    currentClause.textContent = "—";
    renderExpandedGoals();
    renderBindings();
    return;
  }

  const event = state.events[state.currentIndex];

  stepCounter.textContent = `Step ${state.currentIndex + 1} of ${state.events.length}`;
  eventKind.textContent = labels[event.type];
  eventDescription.textContent = event.description;
  currentGoal.textContent = event.goal ? formatTraceText(event.goal) : "—";
  currentClause.textContent = event.clause
    ? formatTraceText(event.clause)
    : "—";

  renderExpandedGoals(event.expandedGoals);
  renderBindings(event.bindings);
}

function renderExpandedGoals(goals?: string[]): void {
  expandedGoalsList.replaceChildren();

  if (!goals || goals.length === 0) {
    const empty = document.createElement("p");
    empty.className = "empty-state";
    empty.textContent = "This operation does not expand a rule.";
    expandedGoalsList.append(empty);
    return;
  }

  for (const goal of goals) {
    const item = document.createElement("code");
    item.className = "expanded-goal";
    item.textContent = formatTraceText(goal);
    expandedGoalsList.append(item);
  }
}

function renderBindings(bindings?: Record<string, string>): void {
  bindingsList.replaceChildren();

  const entries = Object.entries(bindings ?? {})
    .filter(([variable]) => !variable.startsWith("$"))
    .sort(([left], [right]) => left.localeCompare(right));

  if (entries.length === 0) {
    const empty = document.createElement("p");
    empty.className = "empty-state";
    empty.textContent = "No query variable bindings yet.";
    bindingsList.append(empty);
    return;
  }

  for (const [variable, value] of entries) {
    const row = document.createElement("div");
    row.className = "binding-row";

    const variableElement = document.createElement("code");
    variableElement.textContent = variable;

    const equalsElement = document.createElement("span");
    equalsElement.textContent = "=";

    const valueElement = document.createElement("code");
    valueElement.textContent = formatTraceText(value);

    row.append(variableElement, equalsElement, valueElement);
    bindingsList.append(row);
  }
}

function renderTrace(): void {
  traceList.replaceChildren();

  const fragment = document.createDocumentFragment();

  state.events.forEach((event, index) => {
    const item = document.createElement("li");
    const safeDepth = Math.max(0, Math.min(event.depth, 12));

    item.className = `trace-row trace-${event.type}`;

    if (index === state.currentIndex) {
      item.classList.add("is-current");
    }

    if (index > state.currentIndex) {
      item.classList.add("is-future");
    }

    item.style.setProperty("--depth", String(safeDepth));
    item.title = "Jump to this operation";

    item.addEventListener("click", () => {
      stopPlayback();
      state.currentIndex = index;
      render();
    });

    const heading = document.createElement("div");
    heading.className = "trace-heading";

    const tag = document.createElement("span");
    tag.className = "trace-tag";
    tag.textContent = labels[event.type];

    const summary = document.createElement("strong");
    summary.textContent = event.description;

    heading.append(tag, summary);

    const details = document.createElement("p");
    details.className = "trace-details";

    const context: string[] = [];

    if (event.goal) {
      context.push(`Goal: ${formatTraceText(event.goal)}`);
    }

    if (event.clause) {
      context.push(`Clause: ${formatTraceText(event.clause)}`);
    }

    if (event.expandedGoals && event.expandedGoals.length > 0) {
      const goals = event.expandedGoals.map(formatTraceText).join(", ");
      context.push(`Expands to: ${goals}`);
    }

    details.textContent =
      context.length > 0
        ? context.join("  •  ")
        : "No goal, clause, or expansion attached.";

    item.append(heading, details);
    fragment.append(item);
  });

  traceList.append(fragment);
}

function renderAnswers(): void {
  answersList.replaceChildren();

  if (state.answers.length === 0) {
    const empty = document.createElement("p");
    empty.className = "empty-state";
    empty.textContent = "No solutions found yet.";
    answersList.append(empty);
    return;
  }

  state.answers.forEach((answer, index) => {
    const line = document.createElement("p");
    line.className = "answer";

    const bindings = Object.entries(answer)
      .filter(([variable]) => !variable.startsWith("$"))
      .map(([variable, value]) => `${variable} = ${formatTraceText(value)}`)
      .join(", ");

    line.textContent = `Answer ${index + 1}: ${bindings || "true"}`;
    answersList.append(line);
  });
}

function showError(message: string): void {
  errorBox.hidden = false;
  errorBox.textContent = message;
}

function clearError(): void {
  errorBox.hidden = true;
  errorBox.textContent = "";
}