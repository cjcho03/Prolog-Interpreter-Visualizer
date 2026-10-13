import { expect, test } from "@playwright/test"
import {
  loadExample,
  openVisualizer,
  queryInput,
  runQuery,
} from "./helpers"

async function setRangeValue(page: import("@playwright/test").Page, value: number) {
  await page.locator("#trace-scrubber").evaluate((element, nextValue) => {
    const input = element as HTMLInputElement
    input.value = String(nextValue)
    input.dispatchEvent(new Event("input", { bubbles: true }))
  }, value)
}

test.describe("trace controls and editor shortcuts", () => {
  test.beforeEach(async ({ page }) => {
    await openVisualizer(page)
  })

  test("steps, scrubs, jumps, and plays through the trace", async ({ page }) => {
    const response = await runQuery(page)
    expect(response.status()).toBe(200)

    const traceCount = Number(await page.locator("#trace-count").textContent())
    expect(traceCount).toBeGreaterThan(5)

    await expect(page.locator("#step-counter")).toHaveText(`Step 1 of ${traceCount}`)

    await page.locator("#next-step").click()
    await expect(page.locator("#step-counter")).toHaveText(`Step 2 of ${traceCount}`)

    await page.locator("#previous-step").click()
    await expect(page.locator("#step-counter")).toHaveText(`Step 1 of ${traceCount}`)

    await page.locator('.trace-row[data-index="4"]').click()
    await expect(page.locator("#step-counter")).toHaveText(`Step 5 of ${traceCount}`)

    await setRangeValue(page, 3)
    await expect(page.locator("#step-counter")).toHaveText(`Step 3 of ${traceCount}`)

    await page.locator("#playback-speed").selectOption("400")
    await setRangeValue(page, traceCount - 1)
    await page.locator("#play-trace").click()
    await expect(page.locator("#step-counter")).toHaveText(
      `Step ${traceCount} of ${traceCount}`,
      { timeout: 2_000 },
    )
    await expect(page.locator("#play-trace")).toHaveText("Play")
  })

  test("runs with Ctrl+Enter", async ({ page }) => {
    const responsePromise = page.waitForResponse(
      (response) =>
        response.url().includes("/api/solve") &&
        response.request().method() === "POST",
    )

    await queryInput(page).press("Control+Enter")
    const response = await responsePromise

    expect(response.status()).toBe(200)
    await expect(page.locator("#solution-count")).toHaveText("2")
  })

  test("reset restores the selected example and clears the previous run", async ({ page }) => {
    await loadExample(page, "list-basics")
    await queryInput(page).fill("?- contains_item(Color, [cyan, magenta]).")
    await runQuery(page)

    await page.locator("#reset-demo").click()

    await expect(queryInput(page)).toHaveValue(
      "?- contains_item(Planet, [mercury, venus, earth]).",
    )
    await expect(page.locator("#solution-count")).toHaveText("—")
    await expect(page.locator("#trace-count")).toHaveText("—")
    await expect(page.locator("#event-kind")).toHaveText("Waiting")
    await expect(page.locator("#trace-empty")).toBeVisible()
  })
})
