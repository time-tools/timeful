import { expect, test, type Page } from "@playwright/test"
import {
  dismissConsent,
  waitForEventShell,
} from "../helpers/timed-event-helpers"

const payload = {
  name: "Copy feedback coverage",
  type: "specific_dates",
  daysOnly: true,
  dates: ["2026-10-05T00:00:00Z"],
}

// Clipboard permissions are Chromium-only in Playwright and Firefox cannot
// read the clipboard back, so the UI contract is verified against a stubbed
// writeText that records the text it received.
async function stubClipboard(page: Page): Promise<void> {
  await page.addInitScript(() => {
    const writeText = (text: string) => {
      ;(window as unknown as { __copiedText?: string }).__copiedText = text
      return Promise.resolve()
    }
    Object.defineProperty(navigator, "clipboard", {
      configurable: true,
      value: { writeText },
    })
  })
}

async function copiedText(page: Page): Promise<string | undefined> {
  return page.evaluate(
    () => (window as unknown as { __copiedText?: string }).__copiedText,
  )
}

test("event page copy link confirms in place and reverts", async ({ page }) => {
  await stubClipboard(page)
  const created = await page.request.post("/api/events", { data: payload })
  expect(created.status()).toBe(201)
  const { eventId } = (await created.json()) as { eventId: string }

  await page.goto(`/e/${eventId}`, { waitUntil: "domcontentloaded" })
  await dismissConsent(page)
  await waitForEventShell(page)

  const copyButton = page.locator("#copy-link-btn")
  await copyButton.click()

  await expect(copyButton).toContainText("Copied")
  expect(await copiedText(page)).toMatch(new RegExp(`/e/${eventId}$`))

  await expect(copyButton).toContainText("Copy link")
  await expect(page.getByText("Link copied to clipboard!")).toHaveCount(0)
})

test("manage access copy link confirms in place and reverts", async ({
  page,
}) => {
  await stubClipboard(page)
  const created = await page.request.post("/api/events", { data: payload })
  expect(created.status()).toBe(201)
  const { eventId } = (await created.json()) as { eventId: string }

  await page.goto(`/e/${eventId}`, { waitUntil: "domcontentloaded" })
  await dismissConsent(page)
  await page.getByRole("button", { name: "Manage access", exact: true }).click()
  const dialog = page.getByRole("dialog")
  await dialog
    .getByRole("button", { name: "Create new transfer link", exact: true })
    .click()

  const copyButton = dialog.getByRole("button", {
    name: "Copy link",
    exact: true,
  })
  const step2 = dialog.getByTestId("manage-access-step-2")
  await copyButton.click()

  await expect(
    dialog.getByRole("button", { name: "Copied", exact: true }),
  ).toBeVisible()
  await expect(step2).toHaveAttribute("aria-current", "step")
  expect(await copiedText(page)).toMatch(new RegExp(`/transfer/${eventId}/`))

  await expect(
    dialog.getByRole("button", { name: "Copy link", exact: true }),
  ).toBeVisible()
  await expect(step2).toHaveAttribute("aria-current", "step")
})
