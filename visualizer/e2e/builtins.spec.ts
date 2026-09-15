import { expect, test } from "@playwright/test"
import {
  expectAnswerTexts,
  openVisualizer,
  runQuery,
  setProgramAndQuery,
} from "./helpers"

const seedProgram = "seed(ok)."

test.describe("built-ins and multi-goal queries", () => {
  test.beforeEach(async ({ page }) => {
    await openVisualizer(page)
  })

  test("evaluates arithmetic, comparisons, unification, and structural non-unification", async ({ page }) => {
    await setProgramAndQuery(
      page,
      seedProgram,
      String.raw`?- X is 2 + 3 * 4, X =:= 14, a \= b, Y = alice.`,
    )

    const response = await runQuery(page)
    expect(response.status()).toBe(200)
    await expect(page.locator("#solution-count")).toHaveText("1")
    await expectAnswerTexts(page, ["X = 14, Y = alice"])
    await expect(page.locator("#trace-list .trace-builtin")).toHaveCount(4)
  })

  test("keeps numeric inequality distinct from structural non-unification", async ({ page }) => {
    await setProgramAndQuery(
      page,
      seedProgram,
      String.raw`?- 4 =\= 5, a \= b.`,
    )

    const response = await runQuery(page)
    expect(response.status()).toBe(200)
    await expect(page.locator("#solution-count")).toHaveText("1")
    await expectAnswerTexts(page, ["true."])
  })

  test("renders false when fail/0 eliminates the branch", async ({ page }) => {
    await setProgramAndQuery(page, seedProgram, "?- true, fail.")

    const response = await runQuery(page)
    expect(response.status()).toBe(200)
    await expect(page.locator("#solution-count")).toHaveText("0")
    await expect(page.locator("#answers .answer-false")).toHaveText("false.")
  })
})
