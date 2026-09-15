import { expect, type Page, type Response } from "@playwright/test"

export const programInput = (page: Page) => page.locator("#program-input")
export const queryInput = (page: Page) => page.locator("#query-input")
export const runButton = (page: Page) => page.locator("#run-demo")
export const answers = (page: Page) => page.locator("#answers .answer code")

export async function openVisualizer(page: Page): Promise<void> {
  await page.goto("/")
  await expect(page.getByRole("heading", { name: "See how Prolog finds an answer." })).toBeVisible()
  await expect(programInput(page)).toHaveValue(/parent\(alice, bob\)\./)
  await expect(queryInput(page)).toHaveValue("?- grandparent(alice, Who).")
}

export async function setProgramAndQuery(
  page: Page,
  program: string,
  query: string,
): Promise<void> {
  await programInput(page).fill(program)
  await queryInput(page).fill(query)
}

export async function runQuery(page: Page): Promise<Response> {
  const responsePromise = page.waitForResponse(
    (response) =>
      response.url().includes("/api/solve") &&
      response.request().method() === "POST",
  )

  await runButton(page).click()
  const response = await responsePromise
  await expect(runButton(page)).toHaveText("Run query")
  return response
}

export async function expectAnswerTexts(
  page: Page,
  expected: string[],
): Promise<void> {
  await expect(answers(page)).toHaveText(expected)
}

export async function loadExample(page: Page, value: string): Promise<void> {
  await page.locator("#example-select").selectOption(value)
  await page.locator("#load-example").click()
}
