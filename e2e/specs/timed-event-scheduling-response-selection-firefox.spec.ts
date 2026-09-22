import { expect, test } from "@playwright/test"
import { Temporal } from "temporal-polyfill"
import type { Locator, Page } from "@playwright/test"
import {
  buildSpecificDateSeed,
  dragSelectGridRange,
  openEventPage,
  rowIndexForTime,
  seedCanonicalTimedEvent,
  waitForScheduleOverlapMounted,
} from "../helpers/timed-event-helpers"

const SELECTED_DAY = "2026-05-28"

async function seedSchedulingSelectionEvent(
  page: Page,
  name: string,
): Promise<string> {
  const seed = await seedCanonicalTimedEvent(
    page.request,
    buildSpecificDateSeed({
      name,
      selectedDays: [SELECTED_DAY],
      activeSlots: [
        `${SELECTED_DAY}T00:00:00Z`,
        `${SELECTED_DAY}T01:00:00Z`,
        `${SELECTED_DAY}T02:00:00Z`,
        `${SELECTED_DAY}T03:00:00Z`,
      ],
      eventTimezone: "UTC",
      startTimeLocal: "00:00:00",
      endTimeLocal: "04:00:00",
      timeIncrementMinutes: 60,
    }),
  )

  for (const response of [
    { name: "Guest One", availability: [`${SELECTED_DAY}T00:00:00Z`] },
    { name: "Guest Two", availability: [`${SELECTED_DAY}T01:00:00Z`] },
  ]) {
    const created = await page.request.post(
      `/api/events/${seed.eventId}/response`,
      {
        data: {
          guest: true,
          createResponse: true,
          name: response.name,
          email: "",
          availability: response.availability,
          ifNeeded: [],
        },
      },
    )
    expect(created.ok()).toBeTruthy()
  }

  await openEventPage(page, seed.shortId)
  await waitForScheduleOverlapMounted(page)
  return seed.eventId
}

function slotAt(page: Page, hour: number): Locator {
  return page.locator(
    `#drag-section .timeslot[data-row="${String(rowIndexForTime(hour, 0, 60))}"][data-col="0"]`,
  )
}

async function unavailableTimeGridColor(page: Page): Promise<string> {
  return page.evaluate(() => {
    const probe = document.createElement("div")
    probe.style.backgroundColor = "var(--timeful-unavailable-bg-time-grid)"
    document.body.appendChild(probe)
    const value = getComputedStyle(probe).backgroundColor
    probe.remove()
    return value
  })
}

async function backgroundColorOf(locator: Locator): Promise<string> {
  return locator.evaluate((element) => getComputedStyle(element).backgroundColor)
}

function guestCheckbox(page: Page, name: string): Locator {
  return page
    .locator(".respondent-row")
    .filter({ hasText: name })
    .locator(".respondent-control__checkbox")
}

function guestControl(page: Page, name: string): Locator {
  return page
    .locator(".respondent-row")
    .filter({ hasText: name })
    .locator(".respondent-control")
}

test("selecting a response while scheduling filters the grid without leaving scheduling", async ({
  page,
}) => {
  await seedSchedulingSelectionEvent(
    page,
    `Scheduling selection filter ${Temporal.Now.instant().epochMilliseconds}`,
  )

  const unavailable = await unavailableTimeGridColor(page)
  const guestOneSlot = slotAt(page, 0)
  const guestTwoSlot = slotAt(page, 1)
  const guestTwoBefore = await backgroundColorOf(guestTwoSlot)
  expect(guestTwoBefore).toBe(await backgroundColorOf(guestOneSlot))
  expect(guestTwoBefore).not.toBe(unavailable)

  await page.getByRole("button", { name: /^Schedule event$/i }).click()
  await dragSelectGridRange(page, {
    startRow: rowIndexForTime(0, 0, 60),
    startCol: 0,
    endRow: rowIndexForTime(0, 0, 60),
    endCol: 0,
  })

  await test.step("filters to the selected response without leaving scheduling", async () => {
    const row = page
      .locator(".respondent-row")
      .filter({ hasText: "Guest One" })
    await row.hover()
    await guestCheckbox(page, "Guest One").click()

    await expect(guestControl(page, "Guest One")).toHaveAttribute(
      "aria-pressed",
      "true",
    )
    await expect(page.getByRole("button", { name: /^Cancel$/i })).toBeVisible()
    await expect(page.getByRole("button", { name: /^Schedule$/i })).toBeEnabled()
    await expect(guestTwoSlot).toHaveCSS("background-color", unavailable)
    await expect(guestOneSlot).not.toHaveCSS("background-color", unavailable)
  })

  await test.step("restores every response when the last selection clears", async () => {
    await guestCheckbox(page, "Guest One").click()

    await expect(guestControl(page, "Guest One")).toHaveAttribute(
      "aria-pressed",
      "false",
    )
    await expect(page.getByRole("button", { name: /^Cancel$/i })).toBeVisible()
    await expect(guestTwoSlot).toHaveCSS("background-color", guestTwoBefore)
  })
})

test("cancelling scheduling preserves the selected responses", async ({
  page,
}) => {
  await seedSchedulingSelectionEvent(
    page,
    `Scheduling selection cancel ${Temporal.Now.instant().epochMilliseconds}`,
  )

  const unavailable = await unavailableTimeGridColor(page)
  const guestTwoSlot = slotAt(page, 1)

  await test.step("selects a response on the timed event page", async () => {
    const row = page
      .locator(".respondent-row")
      .filter({ hasText: "Guest One" })
    await row.hover()
    await guestCheckbox(page, "Guest One").click()

    await expect(guestControl(page, "Guest One")).toHaveAttribute(
      "aria-pressed",
      "true",
    )
    await expect(guestTwoSlot).toHaveCSS("background-color", unavailable)
  })

  await test.step("entering scheduling preserves the selection", async () => {
    await page.getByRole("button", { name: /^Schedule event$/i }).click()

    await expect(guestControl(page, "Guest One")).toHaveAttribute(
      "aria-pressed",
      "true",
    )
    await expect(page.getByRole("button", { name: /^Cancel$/i })).toBeVisible()
    await expect(guestTwoSlot).toHaveCSS("background-color", unavailable)
  })

  await test.step("cancelling keeps the filtered subset", async () => {
    await page.getByRole("button", { name: /^Cancel$/i }).click()

    await expect(
      page.getByRole("button", { name: /^Schedule event$/i }),
    ).toBeVisible()
    await expect(guestControl(page, "Guest One")).toHaveAttribute(
      "aria-pressed",
      "true",
    )
    await expect(guestTwoSlot).toHaveCSS("background-color", unavailable)
  })
})

test("confirming a schedule clears the selected responses", async ({
  page,
}) => {
  await seedSchedulingSelectionEvent(
    page,
    `Scheduling selection confirm ${Temporal.Now.instant().epochMilliseconds}`,
  )

  const unavailable = await unavailableTimeGridColor(page)
  const guestTwoSlot = slotAt(page, 1)

  await page.getByRole("button", { name: /^Schedule event$/i }).click()
  await dragSelectGridRange(page, {
    startRow: rowIndexForTime(0, 0, 60),
    startCol: 0,
    endRow: rowIndexForTime(0, 0, 60),
    endCol: 0,
  })
  const row = page.locator(".respondent-row").filter({ hasText: "Guest One" })
  await row.hover()
  await guestCheckbox(page, "Guest One").click()
  await expect(guestTwoSlot).toHaveCSS("background-color", unavailable)
  await expect(page.getByRole("button", { name: /^Schedule$/i })).toBeEnabled()

  const scheduleRequest = page.waitForRequest(
    (request) =>
      request.method() === "PUT" && request.url().includes("/schedule"),
  )
  await page.getByRole("button", { name: /^Schedule$/i }).click()
  await page.getByText("Timeful", { exact: true }).click()
  await scheduleRequest

  await expect(guestControl(page, "Guest One")).toHaveAttribute(
    "aria-pressed",
    "false",
  )
  await expect(
    page.getByRole("button", { name: /^Reschedule event$/i }),
  ).toBeVisible()
  await expect(guestTwoSlot).not.toHaveCSS("background-color", unavailable)
})

test("clearing the saved schedule clears the selected responses", async ({
  page,
}) => {
  const eventId = await seedSchedulingSelectionEvent(
    page,
    `Scheduling selection clear ${Temporal.Now.instant().epochMilliseconds}`,
  )
  const scheduleResponse = await page.request.put(
    `/api/events/${eventId}/schedule`,
    {
      data: {
        startDate: `${SELECTED_DAY}T00:00:00Z`,
        endDate: `${SELECTED_DAY}T01:00:00Z`,
      },
    },
  )
  expect(scheduleResponse.ok()).toBeTruthy()
  await page.reload()
  await waitForScheduleOverlapMounted(page)

  const unavailable = await unavailableTimeGridColor(page)
  const guestTwoSlot = slotAt(page, 1)

  await page.getByRole("button", { name: /^Reschedule event$/i }).click()
  const row = page.locator(".respondent-row").filter({ hasText: "Guest One" })
  await row.hover()
  await guestCheckbox(page, "Guest One").click()
  await expect(guestTwoSlot).toHaveCSS("background-color", unavailable)

  await page.getByRole("button", { name: /^Clear$/i }).click()

  await expect(guestControl(page, "Guest One")).toHaveAttribute(
    "aria-pressed",
    "false",
  )
  await expect(
    page.getByRole("button", { name: /^Schedule event$/i }),
  ).toBeVisible()
  await expect(guestTwoSlot).not.toHaveCSS("background-color", unavailable)
})
