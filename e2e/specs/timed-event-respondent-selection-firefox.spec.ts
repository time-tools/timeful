import { expect, test } from "@playwright/test"
import { Temporal } from "temporal-polyfill"
import {
  buildSpecificDateSeed,
  openEventPage,
  seedCanonicalTimedEvent,
} from "../helpers/timed-event-helpers"

test("respondent selection reveals beside the row indicator without moving the name", async ({
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

  await expect(indicator).toBeVisible()
  await expect(checkbox).toBeHidden()

  const nameBeforeHover = await name.boundingBox()
  expect(nameBeforeHover).not.toBeNull()
  if (!nameBeforeHover) {
    throw new Error("Expected the respondent name to have a bounding box")
  }

  await row.hover()
  await expect(checkbox).toBeVisible()
  await expect(indicator).toBeVisible()
  await expect(checkbox).toHaveCSS("border-top-color", "rgb(0, 153, 76)")

  const [indicatorBox, checkboxBox, nameAfterHover] = await Promise.all([
    indicator.boundingBox(),
    checkbox.boundingBox(),
    name.boundingBox(),
  ])
  expect(indicatorBox).not.toBeNull()
  expect(checkboxBox).not.toBeNull()
  expect(nameAfterHover).not.toBeNull()
  if (!indicatorBox || !checkboxBox || !nameAfterHover) {
    throw new Error("Expected the respondent control boxes to be measurable")
  }

  expect(checkboxBox.x).toBeGreaterThanOrEqual(
    indicatorBox.x + indicatorBox.width - 1,
  )
  expect(Math.abs(nameAfterHover.x - nameBeforeHover.x)).toBeLessThanOrEqual(1)
  expect(Math.abs(nameAfterHover.y - nameBeforeHover.y)).toBeLessThanOrEqual(1)

  await row.click()
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
})
