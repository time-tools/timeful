import { expect, test, type APIRequestContext } from "@playwright/test"
import { Temporal } from "temporal-polyfill"
import { expectHintAboveGrid } from "../helpers/availability-hint-helpers"
import {
  buildSpecificDateSeed,
  openEventPage,
  seedCanonicalTimedEvent,
  waitForScheduleOverlapMounted,
} from "../helpers/timed-event-helpers"

const BANNER_TEXT =
  "Add availability (in the event header) to show when you're available for this event."
const EDITING_HINT_TEXT =
  'Click and drag on the grid below to add your "available" times in green.'

async function seedEvent(request: APIRequestContext) {
  const today = Temporal.Now.plainDateISO().toString()

  return seedCanonicalTimedEvent(
    request,
    buildSpecificDateSeed({
      name: `Availability hint ${String(Temporal.Now.instant().epochMilliseconds)}`,
      selectedDays: [today],
      activeSlots: [`${today}T09:00:00.000Z`, `${today}T10:00:00.000Z`],
      eventTimezone: "UTC",
      startTimeLocal: "09:00",
      endTimeLocal: "17:00",
      timeIncrementMinutes: 60,
    }),
  )
}

test("A viewer without a response sees the add availability hint above the grid until they respond", async ({
  page,
  request,
}) => {
  const { shortId, eventId } = await seedEvent(request)
  const today = Temporal.Now.plainDateISO().toString()

  await openEventPage(page, shortId)
  await waitForScheduleOverlapMounted(page)

  const hint = page.getByTestId("add-availability-hint")
  await expect(hint).toBeVisible()
  await expect(hint).toHaveText(BANNER_TEXT)
  await expect(hint.getByRole("button")).toHaveCount(0)
  await expectHintAboveGrid(hint, page)

  const response = await page.request.post(`/api/events/${eventId}/response`, {
    data: {
      guest: true,
      createResponse: true,
      name: "Hint Respondent",
      email: "",
      availability: [`${today}T09:00:00.000Z`],
      ifNeeded: [],
    },
  })
  expect(response.ok()).toBeTruthy()

  await page.reload({ waitUntil: "domcontentloaded" })
  await waitForScheduleOverlapMounted(page)
  await expect(page.getByTestId("add-availability-hint")).toHaveCount(0)
})

test("The editing instruction renders at the top and replaces the grid strip", async ({
  page,
  request,
}) => {
  const { shortId } = await seedEvent(request)

  await openEventPage(page, shortId)
  await waitForScheduleOverlapMounted(page)
  await page.locator("#desktop-primary-availability-btn").click()
  await page.getByRole("button", { name: "Manually", exact: true }).click()

  const hint = page.getByTestId("availability-editing-hint")
  await expect(hint).toBeVisible()
  await expect(hint).toHaveText(EDITING_HINT_TEXT)
  await expectHintAboveGrid(hint, page)
  await expect(page.getByText(EDITING_HINT_TEXT)).toHaveCount(1)

  await test.step("the instruction has no dismiss control and clears when editing ends", async () => {
    await expect(hint.getByRole("button")).toHaveCount(0)

    await page.locator(".desktop-editing-cancel-button").click()
    await expect(hint).toHaveCount(0)

    await page.locator("#desktop-primary-availability-btn").click()
    await page.getByRole("button", { name: "Manually", exact: true }).click()
    await expect(hint).toBeVisible()
  })
})

test("The editing instruction renders despite a stored legacy dismissal", async ({
  page,
  request,
}) => {
  const { shortId } = await seedEvent(request)

  await page.addInitScript(() => {
    localStorage.setItem("closedHintTextedit_availability", "true")
  })
  await openEventPage(page, shortId)
  await waitForScheduleOverlapMounted(page)
  await page.locator("#desktop-primary-availability-btn").click()
  await page.getByRole("button", { name: "Manually", exact: true }).click()

  const hint = page.getByTestId("availability-editing-hint")
  await expect(hint).toBeVisible()
  await expect(hint).toHaveText(EDITING_HINT_TEXT)
  await expect(hint.getByRole("button")).toHaveCount(0)
})

test("The add availability hint clears after the viewer saves an in-app response", async ({
  page,
  request,
}) => {
  const { shortId, eventId } = await seedEvent(request)

  await openEventPage(page, shortId)
  await waitForScheduleOverlapMounted(page)
  await expect(page.getByTestId("add-availability-hint")).toBeVisible()

  await page.locator("#desktop-primary-availability-btn").click()
  await page.getByRole("button", { name: "Manually", exact: true }).click()

  const slot = page.locator(
    '#drag-section .timeslot[data-row="9"][data-col="0"]',
  )
  await slot.scrollIntoViewIfNeeded()
  await slot.click()

  const saved = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      new URL(response.url()).pathname === `/api/events/${eventId}/response`,
  )
  await page.locator(".desktop-editing-save-button").click()
  await page.getByLabel("Guest name (required)").fill("In-app Respondent")
  await page
    .getByRole("dialog")
    .filter({ hasText: "Continue as guest" })
    .getByRole("button", { name: "Save", exact: true })
    .click()
  expect((await saved).status()).toBe(200)

  await expect(page.getByTestId("add-availability-hint")).toHaveCount(0)
})
