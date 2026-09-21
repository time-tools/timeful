import { expect, type Locator, type Page } from "@playwright/test"

// Playwright cannot drive a real on-screen keyboard. Mobile Chrome resizes
// only the visual viewport when the keyboard opens, so the helper replaces
// window.visualViewport before app scripts run and leaves the layout viewport
// untouched, reproducing the geometry that hides a dialog's actions.
export const SIMULATED_KEYBOARD_HEIGHT_PX = 300

// The default floor keeps the simulation within plausible phone geometry.
export const SIMULATED_KEYBOARD_MIN_VISIBLE_HEIGHT_PX = 240

interface VisualKeyboardController {
  setVisualKeyboardInset: (inset: number) => void
}

export interface VisibleViewportRect {
  top: number
  height: number
}

export async function simulateMobileKeyboard(
  page: Page,
  minimumVisibleHeightPx = SIMULATED_KEYBOARD_MIN_VISIBLE_HEIGHT_PX,
): Promise<void> {
  await page.addInitScript(
    ({ initialInset, minimumHeight }) => {
      const keyboardInset = { current: initialInset }
      const visibleViewport = new EventTarget()
      Object.defineProperties(visibleViewport, {
        width: { get: () => window.innerWidth },
        height: {
          get: () =>
            Math.max(window.innerHeight - keyboardInset.current, minimumHeight),
        },
        offsetTop: { get: () => 0 },
        offsetLeft: { get: () => 0 },
        scale: { get: () => 1 },
        pageTop: { get: () => 0 },
        pageLeft: { get: () => 0 },
      })

      Object.defineProperty(window, "visualViewport", {
        configurable: true,
        value: visibleViewport,
      })
      ;(window as unknown as Record<string, unknown>).setVisualKeyboardInset = (
        inset: number,
      ) => {
        keyboardInset.current = inset
        visibleViewport.dispatchEvent(new Event("resize"))
      }
    },
    {
      initialInset: SIMULATED_KEYBOARD_HEIGHT_PX,
      minimumHeight: minimumVisibleHeightPx,
    },
  )
}

export function setVisualKeyboardInset(
  page: Page,
  inset: number,
): Promise<void> {
  return page.evaluate((nextInset: number) => {
    const controller = (window as unknown as Record<string, unknown>)
      .setVisualKeyboardInset as VisualKeyboardController["setVisualKeyboardInset"]
    controller(nextInset)
  }, inset)
}

export function readVisibleViewportRect(
  page: Page,
): Promise<VisibleViewportRect> {
  return page.evaluate(() => ({
    top: window.visualViewport?.offsetTop ?? 0,
    height: window.visualViewport?.height ?? window.innerHeight,
  }))
}

export async function expectWithinVisibleViewport(
  locator: Locator,
  visibleViewport: VisibleViewportRect,
): Promise<void> {
  const box = await locator.boundingBox()
  if (box === null) {
    throw new Error("Expected the element to have a bounding box")
  }
  expect(box.y).toBeGreaterThanOrEqual(visibleViewport.top - 1)
  expect(box.y + box.height).toBeLessThanOrEqual(
    visibleViewport.top + visibleViewport.height + 1,
  )
}
