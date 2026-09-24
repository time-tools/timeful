import { expect, test, type APIRequestContext } from "@playwright/test"
import {
  dismissConsent,
  seedCanonicalTimedEvent,
  type CanonicalTimedSeedInput,
} from "../helpers/timed-event-helpers"

const seedInput: CanonicalTimedSeedInput = {
  name: "Helper contract event",
  type: "specific_dates",
  daysOnly: true,
  dates: ["2026-10-05T00:00:00Z"],
  eventTimezone: "UTC",
  slotGeneration: {
    startTimeLocal: "09:00",
    endTimeLocal: "10:00",
    timeIncrementMinutes: 30,
  },
  timedRecurrence: {
    kind: "specific_dates",
    selectedDays: ["2026-10-05"],
    selectedDaysOfWeek: [],
    startOnMonday: false,
  },
}

function response(body: unknown) {
  return {
    ok: () => true,
    json: () => Promise.resolve(body),
  }
}

test("handles a consent action mounted after DOMContentLoaded", async ({
  page,
}) => {
  await page.setContent("<!doctype html><html><body></body></html>")
  await page.evaluate(() => {
    window.setTimeout(() => {
      const action = document.createElement("button")
      action.type = "button"
      action.textContent = "Accept all"
      action.addEventListener("click", () => {
        document.body.dataset.consentClicked = "true"
        action.remove()
      })
      document.body.append(action)
    }, 1000)
  })

  await dismissConsent(page)

  await expect(page.locator("body")).toHaveAttribute(
    "data-consent-clicked",
    "true",
  )
  await expect(
    page.getByRole("button", { name: "Accept all", exact: true }),
  ).toHaveCount(0)
})

test("uses a returned event short ID without the compatibility lookup", async () => {
  const lookupCalls: string[] = []
  const request = {
    post: () =>
      Promise.resolve(
        response({ eventId: "event-fast", shortId: "short-fast" }),
      ),
    get: (url: string) => {
      lookupCalls.push(url)
      return Promise.resolve(response({ shortId: "short-fallback" }))
    },
  } as unknown as APIRequestContext

  await expect(seedCanonicalTimedEvent(request, seedInput)).resolves.toEqual({
    eventId: "event-fast",
    shortId: "short-fast",
  })
  expect(lookupCalls).toEqual([])
})

test("falls back to the event IDs endpoint when creation omits the short ID", async () => {
  const lookupCalls: string[] = []
  const request = {
    post: () => Promise.resolve(response({ eventId: "event-fallback" })),
    get: (url: string) => {
      lookupCalls.push(url)
      return Promise.resolve(response({ shortId: "short-fallback" }))
    },
  } as unknown as APIRequestContext

  await expect(seedCanonicalTimedEvent(request, seedInput)).resolves.toEqual({
    eventId: "event-fallback",
    shortId: "short-fallback",
  })
  expect(lookupCalls).toEqual(["/api/events/event-fallback/ids"])
})
