import { expect, type Locator, type Page } from "@playwright/test"

export async function expectHintAboveGrid(
  hint: Locator,
  page: Page,
): Promise<void> {
  const gridRow = page.locator(".schedule-overlap-time-grid__body-row").first()

  await expect
    .poll(
      async () => {
        const [hintBox, gridBox] = await Promise.all([
          hint.boundingBox(),
          gridRow.boundingBox(),
        ])
        if (hintBox === null || gridBox === null) {
          return false
        }
        return hintBox.y + hintBox.height <= gridBox.y
      },
      { message: "the hint must render above the timed grid" },
    )
    .toBe(true)
}
