import { expect, test } from "@playwright/test"
import {
  expectAnswerTexts,
  loadExample,
  openVisualizer,
  programInput,
  queryInput,
  runQuery,
} from "./helpers"

test.describe("example picker", () => {
  test.beforeEach(async ({ page }) => {
    await openVisualizer(page)
  })

  test("runs the list basics example with backtracking", async ({ page }) => {
    await loadExample(page, "list-basics")

    await expect(programInput(page)).toHaveValue(/contains_item\(Item, \[Item\|_\]\)\./)
    await expect(queryInput(page)).toHaveValue(
      "?- contains_item(Planet, [mercury, venus, earth]).",
    )

    const response = await runQuery(page)
    expect(response.status()).toBe(200)
    await expect(page.locator("#solution-count")).toHaveText("3")
    await expectAnswerTexts(page, [
      "Planet = mercury",
      "Planet = venus",
      "Planet = earth",
    ])
    await expect(page.locator("#trace-list .trace-backtrack")).not.toHaveCount(0)
  })

  test("runs the recursive reverse example", async ({ page }) => {
    await loadExample(page, "list-algorithms")

    const response = await runQuery(page)
    expect(response.status()).toBe(200)
    await expect(page.locator("#solution-count")).toHaveText("1")
    await expectAnswerTexts(page, ["Reversed = [south, east, north]"])
  })

  test("runs sorted merge with numeric comparisons", async ({ page }) => {
    await loadExample(page, "sorted-merge")

    const response = await runQuery(page)
    expect(response.status()).toBe(200)
    await expect(page.locator("#solution-count")).toHaveText("1")
    await expectAnswerTexts(page, ["Result = [1, 2, 4, 6, 7, 10, 12]"])
    await expect(page.locator("#trace-list .trace-builtin")).not.toHaveCount(0)
  })
})
