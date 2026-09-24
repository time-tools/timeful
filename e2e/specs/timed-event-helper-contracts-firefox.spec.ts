import { expect, test, type APIRequestContext } from "@playwright/test"
import { resolveCookieConsentEnabled } from "../config/tooling"
import {
  dismissConsent,
  seedCanonicalTimedEvent,
  type CanonicalTimedSeedInput,
} from "../helpers/timed-event-helpers"
import { dismissConsentIfPresent } from "../inspect/src/scenarios/helpers"
import { getInspectionToolingMode } from "../inspect/src/config"

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
  const previousFlag = process.env.VITE_ENABLE_COOKIE_CONSENT
  process.env.VITE_ENABLE_COOKIE_CONSENT = "true"

  try {
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
  } finally {
    if (previousFlag === undefined) {
      delete process.env.VITE_ENABLE_COOKIE_CONSENT
    } else {
      process.env.VITE_ENABLE_COOKIE_CONSENT = previousFlag
    }
  }
})

test("skips consent probing when the feature flag is disabled", async ({
  page,
}) => {
  const previousFlag = process.env.VITE_ENABLE_COOKIE_CONSENT
  process.env.VITE_ENABLE_COOKIE_CONSENT = "false"

  try {
    await page.setContent(
      '<!doctype html><html><body><button type="button">Accept all</button></body></html>',
    )

    await dismissConsent(page)

    await expect(
      page.getByRole("button", { name: "Accept all", exact: true }),
    ).toBeVisible()
  } finally {
    if (previousFlag === undefined) {
      delete process.env.VITE_ENABLE_COOKIE_CONSENT
    } else {
      process.env.VITE_ENABLE_COOKIE_CONSENT = previousFlag
    }
  }
})

test("dismisses both legacy and app consent actions", async ({ page }) => {
  const previousFlag = process.env.VITE_ENABLE_COOKIE_CONSENT
  const previousMode = process.env.FRONTEND_TOOLING_MODE
  process.env.VITE_ENABLE_COOKIE_CONSENT = "true"
  process.env.FRONTEND_TOOLING_MODE = "development"

  try {
    await page.setContent(
      '<!doctype html><html><body><button type="button">Agree</button><button type="button">Accept all</button></body></html>',
    )
    await page.evaluate(() => {
      for (const button of document.querySelectorAll("button")) {
        button.addEventListener("click", () => {
          button.remove()
        })
      }
    })

    await dismissConsentIfPresent(page, 1000)

    await expect(page.getByRole("button")).toHaveCount(0)
  } finally {
    if (previousFlag === undefined) {
      delete process.env.VITE_ENABLE_COOKIE_CONSENT
    } else {
      process.env.VITE_ENABLE_COOKIE_CONSENT = previousFlag
    }
    if (previousMode === undefined) {
      delete process.env.FRONTEND_TOOLING_MODE
    } else {
      process.env.FRONTEND_TOOLING_MODE = previousMode
    }
  }
})

test("uses an explicit inspection tooling mode", () => {
  const previousMode = process.env.FRONTEND_TOOLING_MODE

  try {
    delete process.env.FRONTEND_TOOLING_MODE
    expect(getInspectionToolingMode()).toBe("development")

    process.env.FRONTEND_TOOLING_MODE = "staging"
    expect(getInspectionToolingMode()).toBe("staging")
  } finally {
    if (previousMode === undefined) {
      delete process.env.FRONTEND_TOOLING_MODE
    } else {
      process.env.FRONTEND_TOOLING_MODE = previousMode
    }
  }
})

test("matches the app shell to the cookie consent feature flag", async ({
  page,
}) => {
  await page.goto("/")

  const bannerHeading = page.getByRole("heading", {
    name: "We value your privacy",
  })

  if (resolveCookieConsentEnabled()) {
    await expect(bannerHeading).toBeVisible()
  } else {
    await expect(bannerHeading).toHaveCount(0)
  }
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
