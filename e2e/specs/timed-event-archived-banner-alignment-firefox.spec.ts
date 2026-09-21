import { expect, test, type Locator } from "@playwright/test"
import { Temporal } from "temporal-polyfill"
import {
  buildSpecificDateSeed,
  openEventPage,
  seedCanonicalTimedEvent,
} from "../helpers/timed-event-helpers"

const BANNER_TEXT = "This event is archived and read-only."

async function expectLeftEdgesAligned(
  banner: Locator,
  title: Locator,
): Promise<void> {
  await expect
    .poll(
      async () => {
        const [bannerBox, titleBox] = await Promise.all([
          banner.boundingBox(),
          title.boundingBox(),
        ])
        if (bannerBox === null || titleBox === null) {
          return Number.POSITIVE_INFINITY
        }
        return Math.abs(bannerBox.x - titleBox.x)
      },
      {
        message:
          "the archived banner's left edge must align with the event title's left edge",
      },
    )
    .toBeLessThanOrEqual(1)
}

async function expectActionBelowSentence(
  sentence: Locator,
  action: Locator,
): Promise<void> {
  await expect
    .poll(
      async () => {
        const [sentenceBox, actionBox] = await Promise.all([
          sentence.boundingBox(),
          action.boundingBox(),
        ])
        if (sentenceBox === null || actionBox === null) {
          return false
        }
        return actionBox.y >= sentenceBox.y + sentenceBox.height - 1
      },
      {
        message:
          "the banner action must stack below the sentence inside the strip",
      },
    )
    .toBe(true)
}

test("The archived read-only banner aligns with the event title's left edge", async ({
  page,
}) => {
  const today = Temporal.Now.plainDateISO().toString()
  const { shortId, eventId } = await seedCanonicalTimedEvent(
    page.request,
    buildSpecificDateSeed({
      name: `Archived banner alignment ${String(Temporal.Now.instant().epochMilliseconds)}`,
      selectedDays: [today],
      activeSlots: [`${today}T09:00:00.000Z`, `${today}T10:00:00.000Z`],
      eventTimezone: "UTC",
      startTimeLocal: "09:00",
      endTimeLocal: "17:00",
      timeIncrementMinutes: 15,
    }),
  )
  const archiveResponse = await page.request.post(
    `/api/events/${eventId}/archive`,
    { data: { archive: true } },
  )
  expect(archiveResponse.status()).toBe(200)

  await openEventPage(page, shortId)

  const banner = page.locator(".v-alert").filter({ hasText: BANNER_TEXT })
  await expect(banner).toBeVisible()
  const title = page.locator("#event-header-title")
  await expect(title).toBeVisible()

  await test.step("desktop viewport", async () => {
    await expectLeftEdgesAligned(banner, title)
  })

  await test.step("phone-width viewport", async () => {
    await page.setViewportSize({ width: 375, height: 900 })
    await expectLeftEdgesAligned(banner, title)
  })
})

test("The archived read-only banner places its Unarchive event action inside the strip", async ({
  page,
}) => {
  const today = Temporal.Now.plainDateISO().toString()
  const { shortId, eventId } = await seedCanonicalTimedEvent(
    page.request,
    buildSpecificDateSeed({
      name: `Archived banner action ${String(Temporal.Now.instant().epochMilliseconds)}`,
      selectedDays: [today],
      activeSlots: [`${today}T09:00:00.000Z`, `${today}T10:00:00.000Z`],
      eventTimezone: "UTC",
      startTimeLocal: "09:00",
      endTimeLocal: "17:00",
      timeIncrementMinutes: 15,
    }),
  )
  const archiveResponse = await page.request.post(
    `/api/events/${eventId}/archive`,
    { data: { archive: true } },
  )
  expect(archiveResponse.status()).toBe(200)

  await openEventPage(page, shortId)

  const banner = page.locator(".v-alert").filter({ hasText: BANNER_TEXT })
  await expect(banner).toBeVisible()
  const sentence = banner.getByText(BANNER_TEXT, { exact: true })
  const action = banner.getByRole("button", {
    name: "Unarchive event",
    exact: true,
  })
  await expect(action).toBeVisible()

  await test.step("desktop viewport", async () => {
    await expectActionBelowSentence(sentence, action)
  })

  await test.step("phone-width viewport", async () => {
    await page.setViewportSize({ width: 375, height: 900 })
    await expectActionBelowSentence(sentence, action)
  })
})
