import { expect, test } from "@playwright/test"
import { Temporal } from "temporal-polyfill"
import {
  buildSpecificDateSeed,
  buildUtcSpecificTimesRangeInstants,
  openEventPage,
  seedCanonicalTimedEvent,
} from "../helpers/timed-event-helpers"

test("mobile pager chevron buttons are the same size when several dates are picked", async ({
  page,
  request,
  browserName,
}, testInfo) => {
  test.skip(
    browserName !== "chromium" || testInfo.project.name !== "chromium-mobile",
    "The mobile-only 32px button size only applies below the sm breakpoint",
  )

  const today = Temporal.Now.plainDateISO()
  const selectedDays = Array.from({ length: 8 }, (_, index) =>
    today.add({ days: index }).toString(),
  )
  const activeSlots = selectedDays.flatMap((day) =>
    buildUtcSpecificTimesRangeInstants({
      day,
      startHour: 9,
      startMinute: 0,
      endHour: 10,
      endMinute: 0,
      incrementMinutes: 60,
    }),
  )

  const { shortId } = await seedCanonicalTimedEvent(
    request,
    buildSpecificDateSeed({
      name: "Mobile pager button size regression",
      selectedDays,
      activeSlots,
      eventTimezone: "UTC",
      startTimeLocal: "09:00",
      endTimeLocal: "17:00",
      timeIncrementMinutes: 60,
    }),
  )
  await openEventPage(page, shortId)

  const prevButton = page.getByRole("button", { name: "Previous page" })
  const nextButton = page.getByRole("button", { name: "Next page" })

  await expect(nextButton).toBeVisible()
  await expect(prevButton).not.toBeVisible()
  await nextButton.click()
  await expect(prevButton).toBeVisible()
  await expect(nextButton).toBeVisible()

  const [prevBox, nextBox] = await Promise.all([
    prevButton.boundingBox(),
    nextButton.boundingBox(),
  ])
  expect(prevBox).not.toBeNull()
  expect(nextBox).not.toBeNull()
  if (!prevBox || !nextBox) throw new Error("Expected visible pager buttons")

  expect(prevBox.width).toBe(nextBox.width)
  expect(prevBox.height).toBe(nextBox.height)
  expect(prevBox.width).toBe(32)
  expect(prevBox.height).toBe(32)
})
