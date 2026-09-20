import { expect, test, type Page } from "@playwright/test"
import { expectHintAboveGrid } from "../helpers/availability-hint-helpers"
import {
  buildSpecificDateSeed,
  openEventPage,
  seedCanonicalTimedEvent,
  waitForScheduleOverlapMounted,
} from "../helpers/timed-event-helpers"
import { Temporal } from "temporal-polyfill"

const MOBILE_BANNER_TEXT =
  "Add availability (at the bottom of the screen) to show when you're available for this event."
const MOBILE_EDITING_HINT_TEXT =
  'Tap and drag on the grid below to add your "available" times in green.'

test.describe.configure({ mode: "serial" })

async function expectNoBottomBarOptionsButton(page: Page) {
  const cancelButton = page.locator(".mobile-editing-cancel-button")
  await expect(cancelButton).toBeVisible()
  const actionsRow = cancelButton.locator("..")
  await expect(
    actionsRow.getByRole("button", { name: "Options", exact: true }),
  ).toHaveCount(0)
}

test("mobile editing with no responses shows Collapse disabled times in row 2 and no Options button", async ({
  page,
  request,
}, testInfo) => {
  test.skip(
    testInfo.project.name !== "chromium-mobile",
    "Mobile editing toolbar layout assertions",
  )

  const now = Temporal.Now.instant()
  const today = now.toZonedDateTimeISO("UTC").toPlainDate().toString()

  const seed = await seedCanonicalTimedEvent(
    request,
    buildSpecificDateSeed({
      name: `Mobile no responses edit ${String(now.epochMilliseconds)}`,
      selectedDays: [today],
      activeSlots: [`${today}T09:00:00.000Z`, `${today}T10:00:00.000Z`],
      eventTimezone: "UTC",
      startTimeLocal: "09:00",
      endTimeLocal: "17:00",
      timeIncrementMinutes: 60,
    }),
  )

  await openEventPage(page, seed.shortId)

  // Row 2 shows the inline Collapse disabled times switch above the grid.
  const collapseDisabledTimesToggle = page.locator(
    "#mobile-collapse-disabled-times-toggle",
  )
  await expect(collapseDisabledTimesToggle).toBeVisible()

  const gridRow = page.locator(".schedule-overlap-time-grid__body-row").first()
  await expect(gridRow).toBeVisible()
  const toggleBox = await collapseDisabledTimesToggle.boundingBox()
  const gridBox = await gridRow.boundingBox()
  if (toggleBox === null || gridBox === null) {
    throw new Error(
      "Expected the Collapse disabled times toggle and grid to have boxes",
    )
  }
  expect(toggleBox.y + toggleBox.height).toBeLessThanOrEqual(gridBox.y)

  // Enter editing through the bottom-bar availability action.
  await page.locator("#mobile-primary-availability-btn").click()
  await page.getByRole("button", { name: "Manually", exact: true }).click()

  await expect(page.locator(".mobile-editing-cancel-button")).toBeVisible()
  await expect(page.locator(".mobile-editing-save-button")).toBeVisible()

  // The toggle stays in the toolbar row 2 while editing, and the bottom bar
  // has no Options button next to Cancel/Save.
  await expect(collapseDisabledTimesToggle).toBeVisible()
  await expectNoBottomBarOptionsButton(page)
})

test("mobile editing with responses keeps Show best times in row 2 and More options in row 3 and no Options button", async ({
  page,
}, testInfo) => {
  test.skip(
    testInfo.project.name !== "chromium-mobile",
    "Mobile editing toolbar layout assertions",
  )

  const now = Temporal.Now.instant()
  const today = now.toZonedDateTimeISO("UTC").toPlainDate().toString()

  const seed = await seedCanonicalTimedEvent(
    page.request,
    buildSpecificDateSeed({
      name: `Mobile responses edit ${String(now.epochMilliseconds)}`,
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
        name: "Mobile Editing Guest",
        email: "",
        availability: [`${today}T09:00:00.000Z`, `${today}T10:00:00.000Z`],
        ifNeeded: [],
      },
    },
  )
  expect(guestResponse.ok()).toBeTruthy()

  await openEventPage(page, seed.shortId)

  // Row 2: Show best times. Row 3: More options.
  const bestTimesToggle = page.locator(".v-switch", {
    has: page.locator("#mobile-show-best-times-toggle"),
  })
  await expect(bestTimesToggle).toBeVisible()
  const moreOptionsButton = page
    .locator("#event-options-menu-activator")
    .first()
  await expect(moreOptionsButton).toBeVisible()

  // Enter editing through the bottom-bar availability action.
  await page.locator("#mobile-primary-availability-btn").click()
  await expect(page.locator(".mobile-editing-cancel-button")).toBeVisible()
  await expect(page.locator(".mobile-editing-save-button")).toBeVisible()

  // Both rows stay in the toolbar while editing, and the bottom bar has no
  // Options button next to Cancel/Save.
  await expect(bestTimesToggle).toBeVisible()
  await expect(moreOptionsButton).toBeVisible()
  await expectNoBottomBarOptionsButton(page)

  // The bottom action bar buttons share edges with the elevated panel above:
  // Delete aligns with the left edge of Calendar options, and Save aligns
  // with the right edge of the Available/If needed row.
  const calendarOptionsButton = page.locator(".calendar-options-button")
  const availabilityToggleRow = calendarOptionsButton.locator("xpath=..")
  const deleteButton = page
    .locator(".mobile-event-action-bar")
    .getByRole("button", { name: "Delete" })
  const saveButton = page.locator(".mobile-editing-save-button")
  await expect(calendarOptionsButton).toBeVisible()
  await expect(deleteButton).toBeVisible()
  await expect(saveButton).toBeVisible()
  const [calendarOptionsBox, toggleRowBox, deleteBox, saveBox] =
    await Promise.all([
      calendarOptionsButton.boundingBox(),
      availabilityToggleRow.boundingBox(),
      deleteButton.boundingBox(),
      saveButton.boundingBox(),
    ])
  if (
    calendarOptionsBox === null ||
    toggleRowBox === null ||
    deleteBox === null ||
    saveBox === null
  ) {
    throw new Error("Expected the action bar and editing panel to have boxes")
  }
  expect(Math.abs(deleteBox.x - calendarOptionsBox.x)).toBeLessThanOrEqual(1)
  expect(
    Math.abs(saveBox.x + saveBox.width - (toggleRowBox.x + toggleRowBox.width)),
  ).toBeLessThanOrEqual(1)
})

test("mobile viewer without a response sees the add availability hint above the grid", async ({
  page,
  request,
}, testInfo) => {
  test.skip(
    testInfo.project.name !== "chromium-mobile",
    "Mobile availability hint layout assertions",
  )

  const now = Temporal.Now.instant()
  const today = now.toZonedDateTimeISO("UTC").toPlainDate().toString()

  const seed = await seedCanonicalTimedEvent(
    request,
    buildSpecificDateSeed({
      name: `Mobile add availability hint ${String(now.epochMilliseconds)}`,
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

  const hint = page.getByTestId("add-availability-hint")
  await expect(hint).toBeVisible()
  await expect(hint).toHaveText(MOBILE_BANNER_TEXT)
  await expect(hint.getByRole("button")).toHaveCount(0)
  await expectHintAboveGrid(hint, page)
})

test("mobile no-response Add availability uses the solid desktop primary treatment", async ({
  page,
  request,
}, testInfo) => {
  test.skip(
    testInfo.project.name !== "chromium-mobile",
    "Mobile primary action styling",
  )

  const now = Temporal.Now.instant()
  const today = now.toZonedDateTimeISO("UTC").toPlainDate().toString()

  const seed = await seedCanonicalTimedEvent(
    request,
    buildSpecificDateSeed({
      name: `Mobile add availability fill ${String(now.epochMilliseconds)}`,
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

  const addAvailabilityButton = page.locator("#mobile-primary-availability-btn")
  await expect(addAvailabilityButton).toHaveText(/Add availability/i)

  const styles = await addAvailabilityButton.evaluate((element) => {
    const computed = window.getComputedStyle(element)
    return {
      backgroundColor: computed.backgroundColor,
      color: computed.color,
      boxShadow: computed.boxShadow,
    }
  })

  expect(styles.backgroundColor).toBe("rgb(0, 153, 76)")
  expect(styles.color).toBe("rgb(255, 255, 255)")
  expect(styles.boxShadow).toBe("rgba(0, 0, 0, 0.14) 0px 2px 6px 0px")
})

test("mobile editing shows the instruction at the top instead of the bottom overlay", async ({
  page,
  request,
}, testInfo) => {
  test.skip(
    testInfo.project.name !== "chromium-mobile",
    "Mobile availability hint layout assertions",
  )

  const now = Temporal.Now.instant()
  const today = now.toZonedDateTimeISO("UTC").toPlainDate().toString()

  const seed = await seedCanonicalTimedEvent(
    request,
    buildSpecificDateSeed({
      name: `Mobile editing hint ${String(now.epochMilliseconds)}`,
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
  await page.locator("#mobile-primary-availability-btn").click()
  await page.getByRole("button", { name: "Manually", exact: true }).click()

  const hint = page.getByTestId("availability-editing-hint")
  await expect(hint).toBeVisible()
  await expect(hint).toHaveText(MOBILE_EDITING_HINT_TEXT)
  await expectHintAboveGrid(hint, page)
  await expect(
    page
      .locator(".schedule-overlap-mobile-overlay")
      .getByText(MOBILE_EDITING_HINT_TEXT),
  ).toHaveCount(0)
})
