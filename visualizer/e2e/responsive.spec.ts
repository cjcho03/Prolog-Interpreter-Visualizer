import { expect, test } from "@playwright/test"
import { openVisualizer, runQuery } from "./helpers"

test("remains usable at a narrow mobile-sized viewport", async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 })
  await openVisualizer(page)

  const response = await runQuery(page)
  expect(response.status()).toBe(200)

  await expect(page.locator("#answers")).toContainText("Who = diana")
  await expect(page.locator("#trace-list")).toBeVisible()

  const hasHorizontalPageOverflow = await page.evaluate(
    () => document.documentElement.scrollWidth > document.documentElement.clientWidth,
  )
  expect(hasHorizontalPageOverflow).toBe(false)
})
