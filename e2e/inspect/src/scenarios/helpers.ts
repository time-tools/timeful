import type { Page } from "@playwright/test"

import { resolveCookieConsentEnabled } from "../../../config/tooling"
import { settlePage } from "../../../helpers/settle.js"
import { getInspectionToolingMode } from "../config.js"
import { runWithPhaseTimeout } from "../page.js"
import type { AppLabel } from "../types.js"

export const DEFAULT_EVENT_PATH = "/e/dEeaF"
const SUPPORTED_EVENT_WAIT_UNTIL = [
  "commit",
  "domcontentloaded",
  "load",
  "networkidle",
] as const

type EventWaitUntil = (typeof SUPPORTED_EVENT_WAIT_UNTIL)[number]

export function resolveComparatorEventPath(defaultPath = DEFAULT_EVENT_PATH) {
  const overridePath = process.env.COMPARATOR_EVENT_PATH?.trim()
  if (!overridePath) {
    return defaultPath
  }

  return overridePath.startsWith("/") ? overridePath : `/${overridePath}`
}

export const resolveInspectionEventPath = resolveComparatorEventPath

export function resolveComparatorEventWaitUntil(
  defaultWaitUntil: EventWaitUntil = "domcontentloaded",
): EventWaitUntil {
  const overrideWaitUntil = process.env.COMPARATOR_EVENT_WAIT_UNTIL?.trim()
  if (!overrideWaitUntil) {
    return defaultWaitUntil
  }

  if (SUPPORTED_EVENT_WAIT_UNTIL.includes(overrideWaitUntil as EventWaitUntil)) {
    return overrideWaitUntil as EventWaitUntil
  }

  throw new Error(
    `Unsupported COMPARATOR_EVENT_WAIT_UNTIL=${JSON.stringify(overrideWaitUntil)}. Expected one of ${SUPPORTED_EVENT_WAIT_UNTIL.join(", ")}`,
  )
}

export const resolveInspectionEventWaitUntil = resolveComparatorEventWaitUntil

export async function gotoComparatorEventUrl(
  page: Page,
  url: string,
  phaseLabel = "event-route",
) {
  const waitUntil = resolveComparatorEventWaitUntil()

  console.error(
    `[comparator] event-goto:start phase=${phaseLabel} url=${url} waitUntil=${waitUntil}`,
  )
  await runWithPhaseTimeout(
    `event goto ${phaseLabel} ${url}`,
    page.goto(url, { waitUntil }),
  )
  console.error(`[comparator] event-goto:done phase=${phaseLabel} url=${url}`)
}

export async function clickExactText(page: Page, selector: string, text: string) {
  await page.locator(selector).filter({ hasText: text }).first().click({ force: true })
}

export async function clickContainsText(page: Page, selector: string, text: string) {
  await page.locator(selector).filter({ hasText: text }).first().click({ force: true })
}

export async function openNewEventDialog(page: Page) {
  await clickExactText(page, "button", "Create event")
  await page.waitForSelector('input[placeholder="Name your event ..."]')
}

function getConsentActions(page: Page) {
  return [
    {
      label: "agree",
      locator: page.getByRole("button", { name: /^agree$/i }),
    },
    {
      label: "accept all",
      locator: page.getByRole("button", { name: /^accept all$/i }),
    },
  ]
}

async function waitForConsentAction(
  page: Page,
  timeoutMs: number,
): Promise<boolean> {
  return await Promise.race(
    getConsentActions(page).map(({ locator }) =>
      locator
        .waitFor({ state: "visible", timeout: timeoutMs })
        .then(() => true)
        .catch(() => false),
    ),
  )
}

export async function dismissConsentIfPresent(page: Page, timeoutMs = 3_000) {
  if (!resolveCookieConsentEnabled(getInspectionToolingMode())) {
    return
  }

  const consentActions = getConsentActions(page)

  try {
    const actionAppeared = await waitForConsentAction(page, timeoutMs)
    if (!actionAppeared) {
      return
    }

    for (const { label, locator } of consentActions) {
      const count = await locator.count()
      if (count > 1) {
        throw new Error(
          `Expected at most one ${label} consent action, found ${count}`,
        )
      }
      if (count === 0 || !(await locator.isVisible())) {
        continue
      }

      await locator.click({ force: true })
      await settlePage(page, 500)
    }
  } finally {
    await page.evaluate(() => {
      document.getElementById("qc-cmp2-container")?.remove()
    })
  }
}

export const SHARED_EVENT_GUEST_NAME = "sdjkf"

export async function prepareSharedEventGridPage(
  page: Page,
  label: AppLabel,
  eventPath: string,
  showBestTimes: boolean,
) {
  const eventId = eventPath.split("/").filter(Boolean).at(-1)

  await page.addInitScript(
    ({ nextShowBestTimes }) => {
      localStorage.showBestTimes = String(nextShowBestTimes)
    },
    { nextShowBestTimes: showBestTimes },
  )

  const eventUrl = new URL(eventPath, label.url).toString()
  await gotoComparatorEventUrl(page, eventUrl, "shared-grid-initial")
  await dismissConsentIfPresent(page)

  if (eventId) {
    await page.evaluate(
      async ({ shortId, guestName }) => {
        try {
          const response = await fetch(`/api/events/${shortId}/ids`)
          if (!response.ok) return
          const ids = (await response.json()) as { longId?: string }
          if (ids.longId) {
            localStorage[`${ids.longId}.guestName`] = guestName
          }
        } catch {
          // Ignore event-id bootstrap failures and fall back to the default guest state.
        }
      },
      { shortId: eventId, guestName: SHARED_EVENT_GUEST_NAME },
    )

    await gotoComparatorEventUrl(page, eventUrl, "shared-grid-reload")
    await dismissConsentIfPresent(page)
  }

  await page.waitForTimeout(1_500)
}
