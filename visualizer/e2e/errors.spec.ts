import { expect, test } from "@playwright/test"
import {
  openVisualizer,
  programInput,
  queryInput,
  runButton,
  runQuery,
  setProgramAndQuery,
} from "./helpers"

test.describe("error handling", () => {
  test.beforeEach(async ({ page }) => {
    await openVisualizer(page)
  })

  test("shows program parser errors from the Go API", async ({ page }) => {
    await setProgramAndQuery(
      page,
      "parent(alice, bob",
      "?- parent(alice, bob).",
    )

    const response = await runQuery(page)
    expect(response.status()).toBe(400)
    await expect(page.locator("#error-message")).toBeVisible()
    await expect(page.locator("#error-detail")).toContainText("program parse error")
    await expect(page.locator("#event-kind")).toHaveText("Error")
  })

  test("shows query parser errors from the Go API", async ({ page }) => {
    await queryInput(page).fill("grandparent(alice, Who).")

    const response = await runQuery(page)
    expect(response.status()).toBe(400)
    await expect(page.locator("#error-detail")).toContainText("query parse error")
  })

  test("validates empty input before sending a request", async ({ page }) => {
    await programInput(page).fill("")
    await runButton(page).click()

    await expect(page.locator("#error-detail")).toHaveText("Program is required.")
    await expect(programInput(page)).toBeFocused()

    await programInput(page).fill("seed(ok).")
    await queryInput(page).fill("")
    await runButton(page).click()

    await expect(page.locator("#error-detail")).toHaveText("Query is required.")
    await expect(queryInput(page)).toBeFocused()
  })

  test("shows a useful message when the API cannot be reached", async ({ page }) => {
    await page.route("**/api/solve", async (route) => {
      await route.abort()
    })

    await runButton(page).click()

    await expect(page.locator("#error-message")).toBeVisible()
    await expect(page.locator("#error-detail")).toContainText(
      "Could not reach the Prolog API",
    )
  })
})
