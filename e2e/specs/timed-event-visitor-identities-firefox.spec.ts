import { expect, test } from "@playwright/test"

test("Event Visitor Identity survives reload and edits two independently selected responses", async ({
  page,
}) => {
  const created = await page.request.post("/api/events", {
    data: {
      name: "Visitor identities",
      type: "specific_dates",
      daysOnly: false,
      activeSlots: ["2026-10-05T14:00:00Z", "2026-10-05T14:15:00Z"],
      eventTimezone: "GMT",
      slotGeneration: {
        startTimeLocal: "14:00",
        endTimeLocal: "14:30",
        timeIncrementMinutes: 15,
      },
      timedRecurrence: {
        kind: "specific_dates",
        selectedDays: ["2026-10-05"],
        selectedDaysOfWeek: [],
        startOnMonday: false,
      },
    },
  })
  expect(created.status()).toBe(201)
  const { eventId, eventVisitorId } = (await created.json()) as {
    eventId: string
    eventVisitorId: string
  }
  const responseIds: string[] = []
  for (const name of ["Ada", "Grace"]) {
    const response = await page.request.post(
      `/api/events/${eventId}/response`,
      {
        data: {
          createResponse: true,
          name,
          availability: ["2026-10-05T14:00:00Z"],
        },
      },
    )
    expect(response.status()).toBe(200)
    responseIds.push(
      ((await response.json()) as { responseId: string }).responseId,
    )
  }
  await page.goto(`/e/${eventId}`)
  await expect(
    page.getByRole("button", { name: "Edit Ada", exact: true }),
  ).toBeVisible()
  expect(
    await page.evaluate(
      (id) => localStorage.getItem(`timeful.eventVisitor.${id}`),
      eventId,
    ),
  ).toBe(eventVisitorId)
  expect(await page.evaluate(() => document.cookie)).not.toContain(
    "timeful_evcc_",
  )
  await page.reload()
  await expect(
    page.getByRole("button", { name: "Edit Grace", exact: true }),
  ).toBeVisible()
  expect(
    await page.evaluate(
      (id) => localStorage.getItem(`timeful.eventVisitor.${id}`),
      eventId,
    ),
  ).toBe(eventVisitorId)
  for (const [index, name] of ["Ada", "Grace"].entries()) {
    await page
      .getByRole("button", { name: `Edit ${name}`, exact: true })
      .click()
    const mutation = page.waitForRequest(
      (request) =>
        request.method() === "POST" &&
        request.url().includes(`/api/events/${eventId}/response`),
    )
    await page.getByRole("button", { name: "Save", exact: true }).click()
    const sent = await mutation
    expect(sent.postDataJSON()).toMatchObject({
      responseId: responseIds[index],
      createResponse: false,
      name,
    })
    await expect(
      page.getByRole("button", { name: `Edit ${name}`, exact: true }),
    ).toBeVisible()
  }
  const final = await page.request.get(`/api/events/${eventId}`)
  const event = (await final.json()) as { responses: Record<string, unknown> }
  expect(Object.keys(event.responses).sort()).toEqual(responseIds.sort())
  await page.request.post("/api/auth/sign-out")
  await page.reload()
  await expect(
    page.getByRole("button", { name: "Edit Ada", exact: true }),
  ).toBeVisible()
  expect(
    await page.evaluate(
      (id) => localStorage.getItem(`timeful.eventVisitor.${id}`),
      eventId,
    ),
  ).toBe(eventVisitorId)
})

test("Editing an existing response exposes deletion under the selected responseId contract", async ({
  page,
}) => {
  const created = await page.request.post("/api/events", {
    data: {
      name: "Visitor identity deletion",
      type: "specific_dates",
      daysOnly: false,
      activeSlots: ["2026-10-05T14:00:00Z", "2026-10-05T14:15:00Z"],
      eventTimezone: "GMT",
      slotGeneration: {
        startTimeLocal: "14:00",
        endTimeLocal: "14:30",
        timeIncrementMinutes: 15,
      },
      timedRecurrence: {
        kind: "specific_dates",
        selectedDays: ["2026-10-05"],
        selectedDaysOfWeek: [],
        startOnMonday: false,
      },
    },
  })
  expect(created.status()).toBe(201)
  const { eventId, eventVisitorId } = (await created.json()) as {
    eventId: string
    eventVisitorId: string
  }
  const response = await page.request.post(`/api/events/${eventId}/response`, {
    data: {
      createResponse: true,
      name: "Ada",
      availability: ["2026-10-05T14:00:00Z"],
    },
  })
  expect(response.status()).toBe(200)
  const { responseId } = (await response.json()) as { responseId: string }

  await page.goto(`/e/${eventId}`)
  const editAda = page.getByRole("button", { name: "Edit Ada", exact: true })
  await expect(editAda).toBeVisible()
  await editAda.click()

  const deletion = page.waitForRequest(
    (request) =>
      request.method() === "DELETE" &&
      request.url().includes(`/api/events/${eventId}/response`),
  )
  await page.locator("#desktop-delete-availability-btn").click()
  await page
    .locator(".v-dialog")
    .getByRole("button", { name: "Delete", exact: true })
    .click()
  const sent = await deletion
  expect(new URL(sent.url()).searchParams.get("eventVisitorId")).toBe(
    eventVisitorId,
  )
  expect(sent.postDataJSON()).toEqual({ responseId })

  const final = await page.request.get(`/api/events/${eventId}`)
  expect(final.status()).toBe(200)
  const event = (await final.json()) as { responses: Record<string, unknown> }
  expect(Object.keys(event.responses)).toHaveLength(0)
  await expect(
    page.getByRole("button", { name: "Edit Ada", exact: true }),
  ).toHaveCount(0)
})

test("A signed-out visitor who owns one response can start adding another response", async ({
  page,
}) => {
  const created = await page.request.post("/api/events", {
    data: {
      name: "Visitor adds another response",
      type: "specific_dates",
      daysOnly: false,
      activeSlots: ["2026-10-05T14:00:00Z", "2026-10-05T14:15:00Z"],
      eventTimezone: "GMT",
      slotGeneration: {
        startTimeLocal: "14:00",
        endTimeLocal: "14:30",
        timeIncrementMinutes: 15,
      },
      timedRecurrence: {
        kind: "specific_dates",
        selectedDays: ["2026-10-05"],
        selectedDaysOfWeek: [],
        startOnMonday: false,
      },
    },
  })
  expect(created.status()).toBe(201)
  const { eventId } = (await created.json()) as { eventId: string }
  const response = await page.request.post(`/api/events/${eventId}/response`, {
    data: {
      createResponse: true,
      name: "Ada",
      availability: ["2026-10-05T14:00:00Z"],
    },
  })
  expect(response.status()).toBe(200)

  await page.goto(`/e/${eventId}`)
  await expect(
    page.getByRole("button", { name: "Edit Ada", exact: true }),
  ).toBeVisible()

  const addAvailability = page.getByRole("button", {
    name: "Add availability",
    exact: true,
  })
  await expect(addAvailability).toBeVisible()
  await addAvailability.click()

  const manual = page.getByRole("button", { name: "Manually", exact: true })
  await expect(manual).toBeVisible()
  await manual.click()

  await expect(
    page.getByRole("button", { name: "Cancel", exact: true }),
  ).toBeVisible()
  await expect(
    page.getByRole("button", { name: "Save", exact: true }),
  ).toBeVisible()
})

test("A blind-mode non-owner creates and manages a second response without seeing other responses", async ({
  page,
  request,
}) => {
  // The standalone request context owns the event; the page context becomes a
  // non-owner Event Guest that owns only the responses it creates.
  const created = await request.post("/api/events", {
    data: {
      name: "Blind visitor responses",
      type: "specific_dates",
      daysOnly: false,
      blindAvailabilityEnabled: true,
      activeSlots: ["2026-10-05T14:00:00Z", "2026-10-05T14:15:00Z"],
      eventTimezone: "GMT",
      slotGeneration: {
        startTimeLocal: "14:00",
        endTimeLocal: "14:30",
        timeIncrementMinutes: 15,
      },
      timedRecurrence: {
        kind: "specific_dates",
        selectedDays: ["2026-10-05"],
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
      availability: ["2026-10-05T14:00:00Z"],
    },
  })
  expect(otherResponse.status()).toBe(200)
  const adaResponse = await page.request.post(
    `/api/events/${eventId}/response`,
    {
      data: {
        createResponse: true,
        name: "Ada",
        availability: ["2026-10-05T14:00:00Z"],
      },
    },
  )
  expect(adaResponse.status()).toBe(200)

  await page.goto(`/e/${eventId}`)
  await expect(
    page.getByRole("button", { name: "Edit Ada", exact: true }),
  ).toBeVisible()
  await expect(
    page.getByRole("button", { name: "Edit Zoe", exact: true }),
  ).toHaveCount(0)
  await expect(page.getByText("Zoe", { exact: true })).toHaveCount(0)

  await page
    .getByRole("button", { name: "Add availability", exact: true })
    .click()
  await page.getByRole("button", { name: "Manually", exact: true }).click()
  await expect(
    page.getByTestId("schedule-overlap-time-grid-scroller"),
  ).toBeVisible()

  // The time grid indexes rows across the full day, so 14:00 is row 56.
  const firstSlot = page.locator(
    '#drag-section .timeslot[data-row="56"][data-col="0"]',
  )
  await firstSlot.scrollIntoViewIfNeeded()
  await firstSlot.click()
  await page.getByRole("button", { name: "Save", exact: true }).click()

  const nameField = page.getByLabel("Guest name (required)")
  await expect(nameField).toBeVisible()
  await nameField.fill("Ben")
  const createdSecond = page.waitForRequest(
    (mutation) =>
      mutation.method() === "POST" &&
      mutation.url().includes(`/api/events/${eventId}/response`),
  )
  await page.getByRole("button", { name: "Continue", exact: true }).click()
  const secondPayload = (await createdSecond).postDataJSON() as {
    createResponse?: boolean
    name?: string
  }
  expect(secondPayload).toMatchObject({ createResponse: true, name: "Ben" })

  const editBen = page.getByRole("button", { name: "Edit Ben", exact: true })
  await expect(editBen).toBeVisible()
  await expect(
    page.getByRole("button", { name: "Edit Ada", exact: true }),
  ).toBeVisible()

  await editBen.click()
  const secondSlot = page.locator(
    '#drag-section .timeslot[data-row="57"][data-col="0"]',
  )
  await secondSlot.scrollIntoViewIfNeeded()
  await secondSlot.click()
  const savedSecond = page.waitForRequest(
    (mutation) =>
      mutation.method() === "POST" &&
      mutation.url().includes(`/api/events/${eventId}/response`),
  )
  await page.getByRole("button", { name: "Save", exact: true }).click()
  const savedPayload = (await savedSecond).postDataJSON() as {
    createResponse?: boolean
    responseId?: string
    name?: string
  }
  expect(savedPayload).toMatchObject({ createResponse: false, name: "Ben" })
  expect(savedPayload.responseId).toBeTruthy()

  await editBen.click()
  await expect(page.locator("#desktop-delete-availability-btn")).toBeVisible()
  const deletion = page.waitForRequest(
    (mutation) =>
      mutation.method() === "DELETE" &&
      mutation.url().includes(`/api/events/${eventId}/response`),
  )
  await page.locator("#desktop-delete-availability-btn").click()
  await page
    .locator(".v-dialog")
    .getByRole("button", { name: "Delete", exact: true })
    .click()
  expect((await deletion).postDataJSON()).toMatchObject({
    responseId: savedPayload.responseId,
  })

  await expect(
    page.getByRole("button", { name: "Edit Ben", exact: true }),
  ).toHaveCount(0)
  await expect(
    page.getByRole("button", { name: "Edit Ada", exact: true }),
  ).toBeVisible()

  const final = await request.get(`/api/events/${eventId}`)
  const event = (await final.json()) as {
    responses: Record<string, { name: string }>
  }
  expect(
    Object.values(event.responses)
      .map((response) => response.name)
      .sort(),
  ).toEqual(["Ada", "Zoe"])
})
