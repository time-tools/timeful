import { expect, test, type Locator, type Page } from "@playwright/test"
import {
  buildSpecificDateSeed,
  openEventPage,
  seedCanonicalTimedEvent,
  waitForScheduleOverlapMounted,
} from "../helpers/timed-event-helpers"
import { Temporal } from "temporal-polyfill"

// Playwright cannot drive a real on-screen keyboard. Mobile Chrome resizes
// only the visual viewport when the keyboard opens, so the spec replaces
// window.visualViewport before app scripts run and leaves the layout viewport
// untouched, reproducing the geometry that hides the dialog's actions.
const SIMULATED_KEYBOARD_HEIGHT_PX = 300

interface VisualKeyboardController {
  setVisualKeyboardInset: (inset: number) => void
}

interface VisibleViewportRect {
  top: number
  height: number
}

async function simulateMobileKeyboard(page: Page): Promise<void> {
  await page.addInitScript((initialInset: number) => {
    const keyboardInset = { current: initialInset }
    const visibleViewport = new EventTarget()
    Object.defineProperties(visibleViewport, {
      width: { get: () => window.innerWidth },
      height: {
        get: () => Math.max(window.innerHeight - keyboardInset.current, 240),
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
  }, SIMULATED_KEYBOARD_HEIGHT_PX)
}

function setVisualKeyboardInset(page: Page, inset: number): Promise<void> {
  return page.evaluate((nextInset: number) => {
    const controller = (window as unknown as Record<string, unknown>)
      .setVisualKeyboardInset as VisualKeyboardController["setVisualKeyboardInset"]
    controller(nextInset)
  }, inset)
}

function readVisibleViewportRect(page: Page): Promise<VisibleViewportRect> {
  return page.evaluate(() => ({
    top: window.visualViewport?.offsetTop ?? 0,
    height: window.visualViewport?.height ?? window.innerHeight,
  }))
}

async function expectWithinVisibleViewport(
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

async function openGuestDialogFromMobileSaveFlow(page: Page): Promise<void> {
  await page.locator("#mobile-primary-availability-btn").click()
  await page.getByRole("button", { name: "Manually", exact: true }).click()

  const slot = page.locator("#drag-section .timeslot").first()
  await slot.scrollIntoViewIfNeeded()
  await slot.tap()
  await page.locator(".mobile-editing-save-button").click()

  await expect(page.getByText("Continue as guest")).toBeVisible()
}

function seedDate(now: Temporal.Instant): string {
  return now.toZonedDateTimeISO("UTC").toPlainDate().toString()
}

function specificDateSeedPayload(name: string, today: string) {
  return {
    name,
    type: "specific_dates",
    daysOnly: false,
    activeSlots: [`${today}T09:00:00.000Z`, `${today}T10:00:00.000Z`],
    eventTimezone: "UTC",
    slotGeneration: {
      startTimeLocal: "09:00",
      endTimeLocal: "17:00",
      timeIncrementMinutes: 60,
    },
    timedRecurrence: {
      kind: "specific_dates",
      selectedDays: [today],
      selectedDaysOfWeek: [],
      startOnMonday: false,
    },
  }
}

test("the Continue as guest dialog follows the visible viewport above the mobile keyboard", async ({
  page,
  request,
}, testInfo) => {
  test.skip(
    testInfo.project.name !== "chromium-mobile",
    "Mobile keyboard geometry assertions",
  )

  await simulateMobileKeyboard(page)

  const now = Temporal.Now.instant()
  const today = seedDate(now)

  const seed = await seedCanonicalTimedEvent(
    request,
    buildSpecificDateSeed({
      name: `Guest dialog keyboard ${String(now.epochMilliseconds)}`,
      selectedDays: [today],
      activeSlots: [`${today}T09:00:00.000Z`, `${today}T10:00:00.000Z`],
      eventTimezone: "UTC",
      startTimeLocal: "09:00",
      endTimeLocal: "17:00",
      timeIncrementMinutes: 60,
    }),
  )

  await openEventPage(page, seed.shortId)
  await waitForScheduleOverlapMounted(page)
  await openGuestDialogFromMobileSaveFlow(page)

  const dialog = page.getByRole("dialog")
  const nameInput = dialog.getByLabel("Guest name (required)")
  const allowOthersCheckbox = dialog.getByRole("checkbox")
  const continueButton = dialog.getByRole("button", {
    name: "Continue",
    exact: true,
  })
  const dialogCard = page
    .locator(".v-dialog .v-overlay__content")
    .filter({ hasText: "Continue as guest" })

  await test.step("the name field keeps focus with the keyboard open", async () => {
    await nameInput.click()
    await expect(nameInput).toBeFocused()
  })

  await test.step("the controls stay inside the visible viewport", async () => {
    const visibleViewport = await readVisibleViewportRect(page)

    for (const control of [nameInput, allowOthersCheckbox, continueButton]) {
      await expectWithinVisibleViewport(control, visibleViewport)
    }
    await expect(continueButton).toBeInViewport()
  })

  await test.step("closing the keyboard restores the normal layout", async () => {
    await setVisualKeyboardInset(page, 0)

    const viewportHeight = page.viewportSize()?.height
    if (viewportHeight === undefined) {
      throw new Error("Expected the page to expose a viewport height")
    }
    const cardBox = await dialogCard.boundingBox()
    if (cardBox === null) {
      throw new Error("Expected the guest dialog card to have a bounding box")
    }

    expect(cardBox.y).toBeGreaterThanOrEqual(0)
    expect(cardBox.y + cardBox.height).toBeLessThanOrEqual(viewportHeight)
    expect(
      Math.abs(cardBox.y + cardBox.height / 2 - viewportHeight / 2),
    ).toBeLessThanOrEqual(2)
    await expect(continueButton).toBeVisible()
  })
})

test("the email-collecting dialog keeps its fields and action usable above the mobile keyboard", async ({
  page,
  request,
}, testInfo) => {
  test.skip(
    testInfo.project.name !== "chromium-mobile",
    "Mobile keyboard geometry assertions",
  )

  await simulateMobileKeyboard(page)

  const now = Temporal.Now.instant()
  const today = seedDate(now)

  const created = await request.post("/api/events", {
    data: {
      ...specificDateSeedPayload(
        `Guest dialog keyboard email ${String(now.epochMilliseconds)}`,
        today,
      ),
      collectEmails: true,
    },
  })
  expect(created.ok()).toBeTruthy()
  const { eventId } = (await created.json()) as { eventId: string }
  const idsResponse = await request.get(`/api/events/${eventId}/ids`)
  expect(idsResponse.ok()).toBeTruthy()
  const { shortId } = (await idsResponse.json()) as { shortId: string }

  await openEventPage(page, shortId)
  await waitForScheduleOverlapMounted(page)
  await openGuestDialogFromMobileSaveFlow(page)

  const dialog = page.getByRole("dialog")
  const nameInput = dialog.getByLabel("Guest name (required)")
  const emailInput = dialog.getByPlaceholder("Enter your email...")
  const allowOthersCheckbox = dialog.getByRole("checkbox")
  const continueButton = dialog.getByRole("button", {
    name: "Continue",
    exact: true,
  })
  const dialogCard = page
    .locator(".v-dialog .v-overlay__content")
    .filter({ hasText: "Continue as guest" })

  await test.step("the name and email fields stay inside the visible viewport", async () => {
    const visibleViewport = await readVisibleViewportRect(page)

    await expectWithinVisibleViewport(nameInput, visibleViewport)
    await expectWithinVisibleViewport(emailInput, visibleViewport)
  })

  await test.step("the action stays reachable by scrolling the dialog content", async () => {
    const visibleViewport = await readVisibleViewportRect(page)

    const cardBox = await dialogCard.boundingBox()
    if (cardBox === null) {
      throw new Error("Expected the guest dialog card to have a bounding box")
    }
    expect(cardBox.height).toBeLessThanOrEqual(visibleViewport.height + 1)

    const scrolled = await dialogCard.evaluate((element) => {
      const candidates = [element, ...element.querySelectorAll("*")]
      const scroller = candidates.find(
        (candidate) => candidate.scrollHeight > candidate.clientHeight + 1,
      )
      if (!scroller) return false
      scroller.scrollTop = scroller.scrollHeight
      return true
    })
    expect(scrolled).toBe(true)

    await expectWithinVisibleViewport(continueButton, visibleViewport)
    await expectWithinVisibleViewport(allowOthersCheckbox, visibleViewport)
  })

  await test.step("the focused dialog still submits the guest response", async () => {
    await nameInput.fill("Keyboard Guest")
    await emailInput.fill("keyboard@example.com")
    await expect(continueButton).toBeEnabled()
    await continueButton.click()

    await expect(page.getByText("Continue as guest")).toHaveCount(0)
  })
})
