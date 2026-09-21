import { expect, test } from "@playwright/test"
import { Temporal } from "temporal-polyfill"
import {
  buildSpecificDateSeed,
  openEventPage,
  rowIndexForTime,
  seedCanonicalTimedEvent,
} from "../helpers/timed-event-helpers"

test("respondent selection reveals after the name without moving it", async ({
  page,
}) => {
  const now = Temporal.Now.instant()
  const today = now.toZonedDateTimeISO("UTC").toPlainDate().toString()
  const slot = `${today}T09:00:00.000Z`
  const guestName = "Guest One"

  const seed = await seedCanonicalTimedEvent(
    page.request,
    buildSpecificDateSeed({
      name: `Respondent selection layout ${String(now.epochMilliseconds)}`,
      selectedDays: [today],
      activeSlots: [slot, `${today}T10:00:00.000Z`],
      eventTimezone: "UTC",
      startTimeLocal: "09:00",
      endTimeLocal: "17:00",
      timeIncrementMinutes: 60,
    }),
  )

  const guestResponse = await page.request.post(
    `/api/events/${seed.eventId}/response`,
    {
      data: {
        guest: true,
        createResponse: true,
        name: guestName,
        email: "",
        availability: [slot],
        ifNeeded: [],
      },
    },
  )
  expect(guestResponse.ok()).toBeTruthy()

  await openEventPage(page, seed.shortId)

  const row = page.locator(".respondent-row").filter({ hasText: guestName })
  await expect(row).toHaveCount(1)

  const control = row.locator(".respondent-control")
  const indicator = row.locator(".respondent-control__avatar")
  const checkbox = row.locator(".respondent-control__checkbox")
  const name = row.locator(".respondent-name-line")
  const editAction = row.locator(".respondent-edit-status")
  const legendLabel = page
    .locator(".color-legend__indicator-slot + span")
    .first()

  await expect(indicator).toBeVisible()
  await expect(checkbox).toBeHidden()

  const [nameBeforeHover, legendLabelBox] = await Promise.all([
    name.boundingBox(),
    legendLabel.boundingBox(),
  ])
  expect(nameBeforeHover).not.toBeNull()
  expect(legendLabelBox).not.toBeNull()
  if (!nameBeforeHover || !legendLabelBox) {
    throw new Error("Expected the respondent name and legend to be measurable")
  }
  expect(Math.abs(nameBeforeHover.x - legendLabelBox.x)).toBeLessThanOrEqual(1)

  await row.hover()
  await expect(checkbox).toBeVisible()
  await expect(indicator).toBeVisible()
  await expect(checkbox).toHaveCSS("border-top-color", "rgb(0, 153, 76)")

  const [indicatorBox, checkboxBox, nameAfterHover, editActionBox] =
    await Promise.all([
      indicator.boundingBox(),
      checkbox.boundingBox(),
      name.boundingBox(),
      editAction.boundingBox(),
    ])
  expect(indicatorBox).not.toBeNull()
  expect(checkboxBox).not.toBeNull()
  expect(nameAfterHover).not.toBeNull()
  expect(editActionBox).not.toBeNull()
  if (!indicatorBox || !checkboxBox || !nameAfterHover || !editActionBox) {
    throw new Error("Expected the respondent control boxes to be measurable")
  }

  expect(checkboxBox.x).toBeGreaterThanOrEqual(
    nameAfterHover.x + nameAfterHover.width - 1,
  )
  expect(checkboxBox.x + checkboxBox.width).toBeLessThanOrEqual(
    editActionBox.x + 1,
  )
  expect(Math.abs(nameAfterHover.x - nameBeforeHover.x)).toBeLessThanOrEqual(1)
  expect(Math.abs(nameAfterHover.y - nameBeforeHover.y)).toBeLessThanOrEqual(1)

  await checkbox.click()
  await expect(control).toHaveAttribute("aria-pressed", "true")
  await expect(indicator).toBeVisible()
  await expect(checkbox).toBeVisible()
  await expect(checkbox.locator("svg")).toBeVisible()

  const nameAfterSelection = await name.boundingBox()
  expect(nameAfterSelection).not.toBeNull()
  if (!nameAfterSelection) {
    throw new Error("Expected the respondent name to keep a bounding box")
  }
  expect(
    Math.abs(nameAfterSelection.x - nameBeforeHover.x),
  ).toBeLessThanOrEqual(1)
  expect(
    Math.abs(nameAfterSelection.y - nameBeforeHover.y),
  ).toBeLessThanOrEqual(1)

  await page.mouse.move(0, 0)
  await expect(checkbox).toBeVisible()
  await expect(checkbox.locator("svg")).toBeVisible()
  await expect(checkbox).toHaveCSS("border-top-color", "rgb(0, 153, 76)")

  await checkbox.click()
  await expect(control).toHaveAttribute("aria-pressed", "false")

  await page.mouse.move(0, 0)
  await expect(checkbox).toBeHidden()
})

test("respondent selection restores the timed-grid cursor on hover", async ({
  page,
}) => {
  const now = Temporal.Now.instant()
  const today = now.toZonedDateTimeISO("UTC").toPlainDate().toString()
  const timeIncrementMinutes = 60
  const guestName = "Guest One"

  const seed = await seedCanonicalTimedEvent(
    page.request,
    buildSpecificDateSeed({
      name: `Respondent cursor hover ${String(now.epochMilliseconds)}`,
      selectedDays: [today],
      activeSlots: [`${today}T09:00:00.000Z`, `${today}T10:00:00.000Z`],
      eventTimezone: "UTC",
      startTimeLocal: "09:00",
      endTimeLocal: "17:00",
      timeIncrementMinutes,
    }),
  )

  const guestResponse = await page.request.post(
    `/api/events/${seed.eventId}/response`,
    {
      data: {
        guest: true,
        createResponse: true,
        name: guestName,
        email: "",
        availability: [`${today}T09:00:00.000Z`],
        ifNeeded: [],
      },
    },
  )
  expect(guestResponse.ok()).toBeTruthy()

  await openEventPage(page, seed.shortId)

  const row = page.locator(".respondent-row").filter({ hasText: guestName })
  await expect(row).toHaveCount(1)
  const control = row.locator(".respondent-control")
  const checkbox = row.locator(".respondent-control__checkbox")

  await row.hover()
  await expect(checkbox).toBeVisible()
  await checkbox.click()
  await expect(control).toHaveAttribute("aria-pressed", "true")

  // Grid rows are indexed from midnight, so the 10:00 slot sits at row 10.
  const hoveredSlot = page.locator(
    `#drag-section .timeslot[data-row="${String(rowIndexForTime(10, 0, timeIncrementMinutes))}"][data-col="0"]`,
  )
  await hoveredSlot.scrollIntoViewIfNeeded()
  await hoveredSlot.hover()

  await expect(hoveredSlot).toHaveClass(
    /schedule-overlap-time-grid__selected-timeslot/,
  )
  await expect
    .poll(() =>
      hoveredSlot.evaluate(
        (element) => window.getComputedStyle(element, "::after").boxShadow,
      ),
    )
    .not.toBe("none")
})
