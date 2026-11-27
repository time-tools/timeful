import { expect, test, type Locator } from "@playwright/test"
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

test("desktop scheduling shows the Grid Pointer only on hover", async ({
  page,
}) => {
  const seeded = await seedCanonicalTimedEvent(
    page.request,
    buildSpecificDateSeed({
      name: "Scheduling Grid Pointer regression",
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

  const guestResponse = await page.request.post(
    `/api/events/${seeded.eventId}/response`,
    {
      data: {
        guest: true,
        createResponse: true,
        name: "Guest One",
        email: "",
        availability: ["2026-05-28T00:00:00Z"],
        ifNeeded: [],
      },
    },
  )
  expect(guestResponse.ok()).toBeTruthy()

  await openEventPage(page, seeded.shortId)
  await waitForScheduleOverlapMounted(page)
  await page.getByRole("button", { name: /^Schedule event$/i }).click()

  const slotAt = (row: number) =>
    page.locator(
      `#drag-section .timeslot[data-row="${String(row)}"][data-col="0"]`,
    )
  const cursor = page.locator(".schedule-overlap-time-grid__selected-timeslot")
  const centerOf = async (row: number) => {
    const slot = slotAt(row)
    await slot.scrollIntoViewIfNeeded()
    const box = await slot.boundingBox()
    expect(box).not.toBeNull()
    if (!box) throw new Error("Expected a scheduling grid slot box")
    return {
      x: box.x + box.width / 2,
      y: box.y + box.height / 2,
    }
  }

  await test.step("shows the cursor at the hovered cell", async () => {
    const hoverPoint = await centerOf(rowIndexForTime(0, 0, 60))
    await page.mouse.move(hoverPoint.x, hoverPoint.y)

    await expect(slotAt(rowIndexForTime(0, 0, 60))).toHaveClass(
      /schedule-overlap-time-grid__selected-timeslot/,
    )
  })

  await test.step("hides the cursor from a click and re-shows it on the next move", async () => {
    const clickRow = rowIndexForTime(1, 0, 60)
    const clickPoint = await centerOf(clickRow)
    await page.mouse.move(clickPoint.x, clickPoint.y)
    await expect(slotAt(clickRow)).toHaveClass(
      /schedule-overlap-time-grid__selected-timeslot/,
    )

    await page.mouse.down()
    await page.mouse.up()

    await expect(cursor).toHaveCount(0)

    await page.mouse.move(clickPoint.x + 3, clickPoint.y + 3)
    await expect(slotAt(clickRow)).toHaveClass(
      /schedule-overlap-time-grid__selected-timeslot/,
    )
  })

  await test.step("hides the cursor from drag through release", async () => {
    await dragSelectGridRange(page, {
      startRow: rowIndexForTime(1, 0, 60),
      startCol: 0,
      endRow: rowIndexForTime(2, 0, 60),
      endCol: 0,
    })

    await expect(cursor).toHaveCount(0)
  })

  await test.step("re-shows the cursor on the next move inside the released cell", async () => {
    const releasePoint = await centerOf(rowIndexForTime(2, 0, 60))
    await page.mouse.move(releasePoint.x + 3, releasePoint.y + 3)

    await expect(slotAt(rowIndexForTime(2, 0, 60))).toHaveClass(
      /schedule-overlap-time-grid__selected-timeslot/,
    )
  })
})

test("event header reports the saved Timed Event Occurrence Span", async ({
  page,
}) => {
  const seeded = await seedCanonicalTimedEvent(
    page.request,
    buildSpecificDateSeed({
      name: "Header occurrence span regression",
      selectedDays: ["2026-05-28"],
      activeSlots: [
        "2026-05-28T00:00:00Z",
        "2026-05-28T01:00:00Z",
        "2026-05-28T02:00:00Z",
      ],
      eventTimezone: "UTC",
      startTimeLocal: "00:00:00",
      endTimeLocal: "04:00:00",
      timeIncrementMinutes: 60,
    }),
  )

  const scheduleResponse = await page.request.put(
    `/api/events/${seeded.eventId}/schedule`,
    {
      data: {
        startDate: "2026-05-28T00:00:00Z",
        endDate: "2026-05-28T03:00:00Z",
      },
    },
  )
  expect(scheduleResponse.ok()).toBeTruthy()

  await openEventPage(page, seeded.shortId)
  await waitForScheduleOverlapMounted(page)

  const span = page.locator("#event-header-scheduled-span")

  await test.step("shows the saved span below the event title", async () => {
    await expect(span).toBeVisible()
    await expect(span).toContainText("Thu, May 28, 2026")
    await expect(span).toContainText("00:00 \u2013 03:00")
  })

  await test.step("follows a Display Timezone change", async () => {
    await changeDisplayTimezone(page, {
      optionValue: "Asia/Dhaka",
      optionLabelPattern: /\(GMT\+6:00\)/i,
    })

    // 00:00-03:00 UTC reprojects to 06:00-09:00 at UTC+6.
    await expect(span).toContainText("Thu, May 28, 2026")
    await expect(span).toContainText("06:00 \u2013 09:00")
  })

  await test.step("keeps the saved span while a different pending selection is unsaved", async () => {
    await page.getByRole("button", { name: /^Reschedule event$/i }).click()

    // 07:00-09:00 at UTC+6 differs from the saved 06:00-09:00 span.
    await dragSelectGridRange(page, {
      startRow: rowIndexForTime(7, 0, 60),
      startCol: 0,
      endRow: rowIndexForTime(8, 0, 60),
      endCol: 0,
    })

    await expect(span).toContainText("06:00 \u2013 09:00")
    await expect(span).not.toContainText("07:00")

    await page.getByRole("button", { name: /^Clear$/i }).click()
    await expect(span).toHaveCount(0)
  })

  await test.step("shows the span again after scheduling without a reload", async () => {
    await page.getByRole("button", { name: /^Schedule event$/i }).click()
    await dragSelectGridRange(page, {
      startRow: rowIndexForTime(6, 0, 60),
      startCol: 0,
      endRow: rowIndexForTime(7, 0, 60),
      endCol: 0,
    })
    await page.getByRole("button", { name: /^Schedule$/i }).click()
    await page.getByText("Timeful", { exact: true }).click()

    await expect(span).toBeVisible()
    await expect(span).toContainText("Thu, May 28, 2026")
    await expect(span).toContainText("06:00 \u2013 08:00")
  })
})

test("desktop scheduled time occupies a full header action row", async ({
  page,
}) => {
  const eventName = "Scheduled header alignment"
  const day = "2026-05-28"
  const startDate = `${day}T11:00:00Z`
  const endDate = `${day}T12:00:00Z`
  const seeded = await seedCanonicalTimedEvent(
    page.request,
    buildSpecificDateSeed({
      name: eventName,
      selectedDays: [day],
      activeSlots: [startDate, endDate],
      eventTimezone: "UTC",
      startTimeLocal: "11:00:00",
      endTimeLocal: "13:00:00",
      timeIncrementMinutes: 60,
    }),
  )
  const response = await page.request.post(
    `/api/events/${seeded.eventId}/response`,
    {
      data: {
        guest: true,
        createResponse: true,
        name: "Guest",
        email: "",
        availability: [startDate],
        ifNeeded: [],
      },
    },
  )
  expect(response.ok()).toBeTruthy()

  const title = page.getByText(eventName, { exact: true })
  const spanRow = page.locator("#event-header-scheduled-span")
  const spanText = spanRow.getByText("11:00 – 12:00", { exact: false })
  const metadata = page.locator("#event-header-meta-row")
  const options = page.getByRole("button", {
    name: "More options",
    exact: true,
  })
  const box = async (locator: Locator) => {
    await expect(locator).toBeVisible()
    const bounds = await locator.boundingBox()
    if (!bounds)
      throw new Error("Expected a visible header element to have a box")
    return bounds
  }
  const centerY = (bounds: { y: number; height: number }) =>
    bounds.y + bounds.height / 2

  await test.step("the unscheduled header leaves no reserved span row", async () => {
    await openEventPage(page, seeded.shortId)
    await waitForScheduleOverlapMounted(page)
    await expect(spanRow).toHaveCount(0)
    expect((await box(metadata)).y).toBeCloseTo((await box(options)).y)
  })

  const schedule = await page.request.put(
    `/api/events/${seeded.eventId}/schedule`,
    {
      data: { startDate, endDate },
    },
  )
  expect(schedule.ok()).toBeTruthy()
  await page.reload()
  await waitForScheduleOverlapMounted(page)

  await test.step("the scheduled row stays compact on a phone", async () => {
    await page.setViewportSize({ width: 375, height: 900 })
    const row = await box(spanRow)
    const text = await box(spanText)
    const titleBounds = await box(title)
    expect(row.height).toBeCloseTo(text.height)
    expect(row.y - (titleBounds.y + titleBounds.height)).toBeCloseTo(4)
    expect((await box(metadata)).y).toBeGreaterThan(row.y + row.height)
  })

  await test.step("the scheduled row and metadata align with the desktop controls", async () => {
    await page.setViewportSize({ width: 1066, height: 850 })
    const row = await box(spanRow)
    const optionsBounds = await box(options)
    const reschedule = await box(
      page.getByRole("button", { name: "Reschedule event", exact: true }),
    )
    expect(row.height, "scheduled row height").toBeCloseTo(optionsBounds.height)
    expect(row.y, "scheduled row top").toBeCloseTo(optionsBounds.y)
    expect(centerY(await box(spanText)), "scheduled time center").toBeCloseTo(
      centerY(row),
    )
    for (const name of ["Edit event", "Copy link", "Manage access"]) {
      const action = await box(
        metadata.getByRole("button", { name, exact: true }),
      )
      expect(centerY(action), `${name} center`).toBeCloseTo(centerY(reschedule))
    }
  })
})
