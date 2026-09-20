import { expect, test, type Page } from "@playwright/test"
import { signInNewAccount } from "../helpers/account-auth"

interface EventIDs {
  shortId: string
  longId: string
}

interface PluginResponse {
  type: string
  command: string
  requestId: string
  ok: boolean
  payload?: {
    slots?: Record<
      string,
      {
        name: string
        availability: string[]
        ifNeeded: string[]
      }
    >
  }
  error?: {
    message: string
  }
}

async function sendPluginMessage(
  page: Page,
  requestId: string,
  payload: Record<string, unknown>,
): Promise<PluginResponse> {
  return page.evaluate(
    async ({ requestId, payload }) => {
      return new Promise<PluginResponse>((resolve, reject) => {
        const timeout = window.setTimeout(() => {
          window.removeEventListener("message", listener)
          reject(
            new Error(`Timed out waiting for plugin response ${requestId}`),
          )
        }, 10_000)
        const listener = (event: MessageEvent<PluginResponse>) => {
          if (
            event.data.type === "FILL_CALENDAR_EVENT_RESPONSE" &&
            event.data.requestId === requestId
          ) {
            window.clearTimeout(timeout)
            window.removeEventListener("message", listener)
            resolve(event.data)
          }
        }
        window.addEventListener("message", listener)
        window.postMessage(
          {
            type: "FILL_CALENDAR_EVENT",
            requestId,
            payload,
          },
          "*",
        )
      })
    },
    { requestId, payload },
  )
}

test("anonymous poll preserves the plugin slot contract", async ({
  page,
  request,
}) => {
  const created = await request.post("/api/events", {
    data: {
      name: "Plugin slot contract",
      type: "specific_dates",
      daysOnly: false,
      activeSlots: ["2026-01-05T14:00:00Z", "2026-01-05T14:15:00Z"],
      eventTimezone: "GMT",
      slotGeneration: {
        startTimeLocal: "14:00",
        endTimeLocal: "14:30",
        timeIncrementMinutes: 15,
      },
      timedRecurrence: {
        kind: "specific_dates",
        selectedDays: ["2026-01-05"],
        selectedDaysOfWeek: [],
        startOnMonday: false,
      },
    },
  })
  expect(created.status()).toBe(201)
  const { eventId } = (await created.json()) as { eventId: string }
  expect(eventId).toMatch(/^[0-9A-HJKMNPQRSTVWXYZ]{8}$/)

  const idsResponse = await request.get(`/api/events/${eventId}/ids`)
  expect(idsResponse.status()).toBe(200)
  const ids = (await idsResponse.json()) as EventIDs
  expect(ids).toEqual({ shortId: eventId, longId: eventId })

  const eventLoad = page.waitForResponse(
    (response) =>
      response.url().includes(`/api/events/${ids.shortId}`) &&
      response.status() === 200,
  )
  await page.goto(`/e/${ids.shortId}`)
  await eventLoad

  // The plugin message listener registers at mount, but the ScheduleOverlap
  // time grid mounts later; set-slots validates against the grid's valid time
  // ranges, so wait for the grid itself before posting any plugin message.
  await expect(
    page.getByTestId("schedule-overlap-time-grid-scroller"),
  ).toBeVisible()

  const setResponse = await sendPluginMessage(page, "set-slots", {
    type: "set-slots",
    timezone: "GMT",
    guestName: "Ada",
    slots: [
      {
        start: "2026-01-05T14:00:00",
        end: "2026-01-05T14:30:00",
        status: "available",
      },
    ],
  })
  expect(setResponse.ok, JSON.stringify(setResponse)).toBe(true)
  expect(setResponse).toMatchObject({
    type: "FILL_CALENDAR_EVENT_RESPONSE",
    command: "set-slots",
    requestId: "set-slots",
    ok: true,
  })

  const getResponse = await sendPluginMessage(page, "get-slots", {
    type: "get-slots",
    timezone: "GMT",
  })
  expect(getResponse.ok, JSON.stringify(getResponse)).toBe(true)
  expect(getResponse).toMatchObject({
    type: "FILL_CALENDAR_EVENT_RESPONSE",
    command: "get-slots",
    requestId: "get-slots",
    ok: true,
  })
  const slots = Object.values(getResponse.payload?.slots ?? {})
  expect(slots).toHaveLength(1)
  expect(slots[0]).toMatchObject({
    name: "Ada",
    ifNeeded: [],
  })
  expect(slots[0]?.availability).toHaveLength(2)
})

test("blind availability plugin flow manages only the non-owner's own responses", async ({
  page,
  request,
}) => {
  const created = await request.post("/api/events", {
    data: {
      name: "Blind plugin responses",
      type: "specific_dates",
      daysOnly: false,
      blindAvailabilityEnabled: true,
      activeSlots: ["2026-01-05T14:00:00Z", "2026-01-05T14:15:00Z"],
      eventTimezone: "GMT",
      slotGeneration: {
        startTimeLocal: "14:00",
        endTimeLocal: "14:30",
        timeIncrementMinutes: 15,
      },
      timedRecurrence: {
        kind: "specific_dates",
        selectedDays: ["2026-01-05"],
        selectedDaysOfWeek: [],
        startOnMonday: false,
      },
    },
  })
  expect(created.status()).toBe(201)
  const { eventId } = (await created.json()) as { eventId: string }

  const otherResponse = await request.post(`/api/events/${eventId}/response`, {
    data: {
      createResponse: true,
      name: "Zoe",
      availability: ["2026-01-05T14:00:00Z"],
    },
  })
  expect(otherResponse.status()).toBe(200)

  await page.goto(`/e/${eventId}`)
  await expect(
    page.getByTestId("schedule-overlap-time-grid-scroller"),
  ).toBeVisible()

  const first = await sendPluginMessage(page, "set-slots", {
    type: "set-slots",
    timezone: "GMT",
    guestName: "Ada",
    slots: [
      {
        start: "2026-01-05T14:00:00",
        end: "2026-01-05T14:15:00",
        status: "available",
      },
    ],
  })
  expect(first.ok, JSON.stringify(first)).toBe(true)

  const second = await sendPluginMessage(page, "set-slots", {
    type: "set-slots",
    timezone: "GMT",
    guestName: "Grace",
    slots: [
      {
        start: "2026-01-05T14:15:00",
        end: "2026-01-05T14:30:00",
        status: "available",
      },
    ],
  })
  expect(second.ok, JSON.stringify(second)).toBe(true)

  const afterCreate = await sendPluginMessage(page, "get-slots", {
    type: "get-slots",
    timezone: "GMT",
  })
  expect(afterCreate.ok, JSON.stringify(afterCreate)).toBe(true)
  const createdSlots = Object.values(afterCreate.payload?.slots ?? {})
  expect(createdSlots.map((slot) => slot.name).sort()).toEqual(["Ada", "Grace"])
  expect(createdSlots.find((slot) => slot.name === "Zoe")).toBeUndefined()

  const update = await sendPluginMessage(page, "set-slots", {
    type: "set-slots",
    timezone: "GMT",
    guestName: "Grace",
    slots: [
      {
        start: "2026-01-05T14:00:00",
        end: "2026-01-05T14:30:00",
        status: "available",
      },
    ],
  })
  expect(update.ok, JSON.stringify(update)).toBe(true)

  const afterUpdate = await sendPluginMessage(page, "get-slots", {
    type: "get-slots",
    timezone: "GMT",
  })
  expect(afterUpdate.ok, JSON.stringify(afterUpdate)).toBe(true)
  const updatedSlots = Object.values(afterUpdate.payload?.slots ?? {})
  expect(
    updatedSlots.find((slot) => slot.name === "Grace")?.availability,
  ).toHaveLength(2)
  expect(
    updatedSlots.find((slot) => slot.name === "Ada")?.availability,
  ).toHaveLength(1)
})

test("blind availability plugin flow creates an owner response for an unmatched guest name", async ({
  page,
}) => {
  await signInNewAccount(page.request, "blind-owner-plugin")
  const eventCreated = await page.request.post("/api/events", {
    data: {
      name: "Blind owner plugin responses",
      type: "specific_dates",
      daysOnly: false,
      blindAvailabilityEnabled: true,
      activeSlots: ["2026-01-05T14:00:00Z", "2026-01-05T14:15:00Z"],
      eventTimezone: "GMT",
      slotGeneration: {
        startTimeLocal: "14:00",
        endTimeLocal: "14:30",
        timeIncrementMinutes: 15,
      },
      timedRecurrence: {
        kind: "specific_dates",
        selectedDays: ["2026-01-05"],
        selectedDaysOfWeek: [],
        startOnMonday: false,
      },
    },
  })
  expect(eventCreated.status()).toBe(201)
  const { eventId } = (await eventCreated.json()) as { eventId: string }

  const adaResponse = await page.request.post(
    `/api/events/${eventId}/response`,
    {
      data: {
        createResponse: true,
        name: "Ada",
        availability: ["2026-01-05T14:00:00Z"],
      },
    },
  )
  expect(adaResponse.status()).toBe(200)
  const { responseId: adaResponseId } = (await adaResponse.json()) as {
    responseId: string
  }

  await page.goto(`/e/${eventId}`)
  await expect(
    page.getByTestId("schedule-overlap-time-grid-scroller"),
  ).toBeVisible()
  // Select Ada so reusing the selected response would rename it instead.
  await page.evaluate(
    ([id, responseId]) => {
      localStorage.setItem(`timeful.selectedResponse.${id}`, responseId)
    },
    [eventId, adaResponseId],
  )

  const setResponse = await sendPluginMessage(page, "set-slots", {
    type: "set-slots",
    timezone: "GMT",
    guestName: "Grace",
    slots: [
      {
        start: "2026-01-05T14:15:00",
        end: "2026-01-05T14:30:00",
        status: "available",
      },
    ],
  })
  expect(setResponse.ok, JSON.stringify(setResponse)).toBe(true)

  const getResponse = await sendPluginMessage(page, "get-slots", {
    type: "get-slots",
    timezone: "GMT",
  })
  expect(getResponse.ok, JSON.stringify(getResponse)).toBe(true)
  const slots = Object.values(getResponse.payload?.slots ?? {})
  expect(slots.map((slot) => slot.name).sort()).toEqual(["Ada", "Grace"])
  expect(slots.find((slot) => slot.name === "Ada")?.availability).toHaveLength(
    1,
  )
  expect(
    slots.find((slot) => slot.name === "Grace")?.availability,
  ).toHaveLength(1)
})
