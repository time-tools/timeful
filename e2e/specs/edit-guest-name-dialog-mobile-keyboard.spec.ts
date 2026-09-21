import { expect, test, type Page } from "@playwright/test"
import {
  buildSpecificDateSeed,
  openEventPage,
  seedCanonicalTimedEvent,
  waitForScheduleOverlapMounted,
} from "../helpers/timed-event-helpers"
import {
  expectWithinVisibleViewport,
  readVisibleViewportRect,
  setVisualKeyboardInset,
  simulateMobileKeyboard,
} from "../helpers/visual-keyboard"
import { Temporal } from "temporal-polyfill"

const GUEST_NAME = "Mobile Name Dialog Guest"
const SHORT_VISIBLE_HEIGHT_PX = 160

async function openEditGuestNameDialog(page: Page): Promise<void> {
  await page.locator("#mobile-primary-availability-btn").click()

  const chip = page.locator(".editing-availability-as__guest-chip")
  await expect(chip).toContainText(GUEST_NAME)
  await chip.click()

  await expect(page.getByText("Edit guest name", { exact: true })).toBeVisible()
}

test("the Edit guest name dialog stays above the mobile keyboard", async ({
  page,
}, testInfo) => {
  test.skip(
    testInfo.project.name !== "chromium-mobile",
    "Mobile keyboard geometry assertions",
  )

  await simulateMobileKeyboard(page, SHORT_VISIBLE_HEIGHT_PX)

  const now = Temporal.Now.instant()
  const today = now.toZonedDateTimeISO("UTC").toPlainDate().toString()

  const seed = await seedCanonicalTimedEvent(
    page.request,
    buildSpecificDateSeed({
      name: `Guest rename keyboard ${String(now.epochMilliseconds)}`,
      selectedDays: [today],
      activeSlots: [`${today}T09:00:00.000Z`, `${today}T10:00:00.000Z`],
      eventTimezone: "UTC",
      startTimeLocal: "09:00",
      endTimeLocal: "17:00",
      timeIncrementMinutes: 60,
    }),
  )

  // The page-scoped request context retains the HttpOnly Event Visitor Control
  // Credential, so this browser owns the response without legacy localStorage.
  const guestResponse = await page.request.post(
    `/api/events/${seed.eventId}/response`,
    {
      data: {
        guest: true,
        createResponse: true,
        name: GUEST_NAME,
        email: "",
        availability: [`${today}T09:00:00.000Z`, `${today}T10:00:00.000Z`],
        ifNeeded: [],
      },
    },
  )
  expect(guestResponse.ok()).toBeTruthy()

  await openEventPage(page, seed.shortId)
  await waitForScheduleOverlapMounted(page)
  await openEditGuestNameDialog(page)

  const dialog = page.getByRole("dialog").filter({ hasText: "Edit guest name" })
  const nameInput = dialog.getByLabel("Guest name (required)")
  const saveButton = dialog.getByRole("button", { name: "Save", exact: true })
  const cancelButton = dialog.getByRole("button", {
    name: "Cancel",
    exact: true,
  })
  const dialogCard = page
    .locator(".v-dialog .v-overlay__content")
    .filter({ hasText: "Edit guest name" })

  await test.step("the name field keeps focus with the keyboard open", async () => {
    await nameInput.click()
    await expect(nameInput).toBeFocused()
  })

  await test.step("the field and actions stay inside the visible viewport", async () => {
    const visibleViewport = await readVisibleViewportRect(page)

    for (const control of [nameInput, saveButton, cancelButton]) {
      await expectWithinVisibleViewport(control, visibleViewport)
    }
    await expect(saveButton).toBeInViewport()
  })

  await test.step("the action stays reachable when the visible area is shorter than the dialog", async () => {
    const pageHeight = page.viewportSize()?.height
    if (pageHeight === undefined) {
      throw new Error("Expected the page to expose a viewport height")
    }
    await setVisualKeyboardInset(page, pageHeight - SHORT_VISIBLE_HEIGHT_PX)

    const visibleViewport = await readVisibleViewportRect(page)
    expect(visibleViewport.height).toBe(SHORT_VISIBLE_HEIGHT_PX)

    const cardBox = await dialogCard.boundingBox()
    if (cardBox === null) {
      throw new Error(
        "Expected the edit guest name dialog card to have a bounding box",
      )
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

    await expectWithinVisibleViewport(saveButton, visibleViewport)
  })

  await test.step("closing the keyboard restores the normal layout", async () => {
    await setVisualKeyboardInset(page, 0)

    const viewportHeight = page.viewportSize()?.height
    if (viewportHeight === undefined) {
      throw new Error("Expected the page to expose a viewport height")
    }
    const cardBox = await dialogCard.boundingBox()
    if (cardBox === null) {
      throw new Error(
        "Expected the edit guest name dialog card to have a bounding box",
      )
    }

    expect(cardBox.y).toBeGreaterThanOrEqual(0)
    expect(cardBox.y + cardBox.height).toBeLessThanOrEqual(viewportHeight)
    expect(
      Math.abs(cardBox.y + cardBox.height / 2 - viewportHeight / 2),
    ).toBeLessThanOrEqual(2)
    await expect(saveButton).toBeVisible()
  })

  await test.step("the dialog still saves the renamed guest", async () => {
    await nameInput.fill("Renamed Mobile Guest")
    await saveButton.click()

    await expect(
      page.getByText("Edit guest name", { exact: true }),
    ).toHaveCount(0)
  })
})
