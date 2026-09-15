import { expect, test } from "@playwright/test"
import {
  expectAnswerTexts,
  openVisualizer,
  queryInput,
  runQuery,
} from "./helpers"

test.describe("family example", () => {
  test("returns both grandparent solutions and a trace", async ({ page }) => {
    await openVisualizer(page)

    const response = await runQuery(page)
    expect(response.status()).toBe(200)

    await expect(page.locator("#solution-count")).toHaveText("2")
    await expectAnswerTexts(page, ["Who = diana", "Who = eli"])

    const traceCount = Number(await page.locator("#trace-count").textContent())
    expect(traceCount).toBeGreaterThan(0)
    await expect(page.locator("#step-counter")).toContainText("Step 1 of")
    await expect(page.locator("#event-kind")).toHaveText("Goal")
    await expect(page.locator("#trace-list .trace-row")).toHaveCount(traceCount)
  })

  test("supports anonymous variables without exposing internal names", async ({ page }) => {
    await openVisualizer(page)
    await queryInput(page).fill("?- parent(alice, _).")

    const response = await runQuery(page)
    expect(response.status()).toBe(200)

    await expect(page.locator("#solution-count")).toHaveText("2")
    await expectAnswerTexts(page, ["true.", "true."])
    await expect(page.locator("#answers")).not.toContainText("$anon_")
  })
})
