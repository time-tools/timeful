import { expect, test } from "@playwright/test"
import {
  buildSpecificDateSeed,
  changeDisplayTimezone,
  dragSelectGridRange,
  openEventPage,
  rowIndexForTime,
  seedCanonicalTimedEvent,
  waitForScheduleOverlapMounted,
} from "../helpers/timed-event-helpers"

test("scheduling tooltip reports the pending Timed Event Occurrence Span", async ({
  page,
}) => {
  const seeded = await seedCanonicalTimedEvent(
    page.request,
    buildSpecificDateSeed({
      name: "Scheduling tooltip span regression",
      selectedDays: ["2026-05-28"],
      activeSlots: [
        "2026-05-28T00:00:00Z",
        "2026-05-28T01:00:00Z",
        "2026-05-28T02:00:00Z",
        "2026-05-28T03:00:00Z",
      ],
      eventTimezone: "UTC",
      startTimeLocal: "00:00:00",
      endTimeLocal: "04:00:00",
      timeIncrementMinutes: 60,
    }),
  )

  await openEventPage(page, seeded.shortId)
  await waitForScheduleOverlapMounted(page)
  await page.getByRole("button", { name: /^Schedule event$/i }).click()

  const tooltip = page.locator(".tw\\:fixed.timeful-tooltip-layer")
  await dragSelectGridRange(page, {
    startRow: rowIndexForTime(0, 0, 60),
    startCol: 0,
    endRow: rowIndexForTime(2, 0, 60),
    endCol: 0,
  })

  await expect(tooltip).toBeVisible()
  await expect(tooltip).toContainText("00:00 to 03:00")
  await expect(tooltip).toContainText("Thu, May 28, 2026")

  const middleSlot = page.locator(
    `#drag-section .timeslot[data-row="${String(rowIndexForTime(1, 0, 60))}"][data-col="0"]`,
  )
  const middleBox = await middleSlot.boundingBox()
  expect(middleBox).not.toBeNull()
  if (!middleBox) {
    throw new Error("Expected a middle slot box inside the pending span")
  }
  await page.mouse.move(
    middleBox.x + middleBox.width / 2,
    middleBox.y + middleBox.height / 2,
  )

  await expect(tooltip).toBeVisible()
  await expect(tooltip).toContainText("00:00 to 03:00")

  await test.step("keeps the pending Instants across a Display Timezone change", async () => {
    await changeDisplayTimezone(page, {
      optionValue: "Asia/Dhaka",
      optionLabelPattern: /\(GMT\+6:00\)/i,
    })

    const reprojectedMiddleSlot = page.locator(
      `#drag-section .timeslot[data-row="${String(rowIndexForTime(7, 0, 60))}"][data-col="0"]`,
    )
    const reprojectedBox = await reprojectedMiddleSlot.boundingBox()
    expect(reprojectedBox).not.toBeNull()
    if (!reprojectedBox) {
      throw new Error("Expected a reprojected middle slot box")
    }
    await page.mouse.move(
      reprojectedBox.x + reprojectedBox.width / 2,
      reprojectedBox.y + reprojectedBox.height / 2,
    )

    // 00:00-03:00 UTC is preserved and reprojects to 06:00-09:00 at UTC+6.
    await expect(tooltip).toBeVisible()
    await expect(tooltip).toContainText("06:00 to 09:00")
    await expect(tooltip).toContainText("Thu, May 28, 2026")
  })

  await test.step("saves the preserved Instants", async () => {
    const scheduleRequest = page.waitForRequest(
      (request) =>
        request.method() === "PUT" && request.url().includes("/schedule"),
    )

    await page.getByRole("button", { name: /^Schedule$/i }).click()
    await page.getByText("Timeful", { exact: true }).click()

    const request = await scheduleRequest
    expect(request.postDataJSON()).toEqual({
      startDate: "2026-05-28T00:00:00Z",
      endDate: "2026-05-28T03:00:00Z",
    })
  })
})
