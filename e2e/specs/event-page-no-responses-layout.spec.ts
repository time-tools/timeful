import { expect, test } from "@playwright/test"
import {
  buildSpecificDateSeed,
  dismissConsent,
  openEventPage,
  seedCanonicalTimedEvent,
  waitForEventShell,
  waitForScheduleOverlapMounted,
} from "../helpers/timed-event-helpers"
import { Temporal } from "temporal-polyfill"

test("event page without responses aligns desktop header details with their control column", async ({
  page,
}, testInfo) => {
  test.skip(
    testInfo.project.name === "chromium-mobile",
    "Desktop-only header layout",
  )

  const now = Temporal.Now.instant()
  const today = now.toZonedDateTimeISO("UTC").toPlainDate().toString()

  const seed = await seedCanonicalTimedEvent(
    page.request,
    buildSpecificDateSeed({
      name: `Layout test ${String(now.epochMilliseconds)}`,
      selectedDays: [today],
      activeSlots: [`${today}T09:00:00.000Z`, `${today}T10:00:00.000Z`],
      eventTimezone: "UTC",
      startTimeLocal: "09:00",
      endTimeLocal: "17:00",
      timeIncrementMinutes: 60,
    }),
  )

  await openEventPage(page, seed.shortId)

  // Verify "Add availability" button exists
  const addAvailabilityBtn = page.locator("#desktop-primary-availability-btn")
  await expect(addAvailabilityBtn).toBeVisible()
  await expect(addAvailabilityBtn).toHaveText(/Add availability/i)

  const collapseDisabledTimesToggle = page.locator(
    ".desktop-event-header-options__collapse-disabled-times-switch",
  )
  const scheduleEventButton = page.getByRole("button", {
    name: /^Schedule event$/i,
  })
  await expect(collapseDisabledTimesToggle).toBeVisible()
  await expect(scheduleEventButton).toBeVisible()

  if (testInfo.project.name === "chromium-desktop") {
    const timeFormatToggle = page.locator(".time-format-toggle")
    const firstTimeGridRow = page
      .locator(".schedule-overlap-time-grid__body-row")
      .first()
    const nextPageButton = page.locator(
      ".schedule-overlap-sidebar__pager button.v-btn",
    )
    const title = page.locator("#event-header-title")
    const editEventButton = page.locator("#edit-event-btn")
    const ownerActionRow = page.locator("#event-header-button-row")
    const [
      titleBox,
      addAvailabilityBox,
      editEventBox,
      ownerActionRowBox,
      collapseDisabledTimesBox,
      scheduleEventBox,
      timeFormatToggleBox,
      firstTimeGridRowBox,
      nextPageButtonBox,
      sidebarBox,
      githubLinkBox,
    ] = await Promise.all([
      title.boundingBox(),
      addAvailabilityBtn.boundingBox(),
      editEventButton.boundingBox(),
      ownerActionRow.boundingBox(),
      collapseDisabledTimesToggle.boundingBox(),
      scheduleEventButton.boundingBox(),
      timeFormatToggle.boundingBox(),
      firstTimeGridRow.boundingBox(),
      nextPageButton.boundingBox(),
      page.locator(".schedule-overlap-sidebar").boundingBox(),
      page.getByRole("link", { name: "GitHub" }).boundingBox(),
    ])
    if (
      titleBox === null ||
      addAvailabilityBox === null ||
      editEventBox === null ||
      ownerActionRowBox === null ||
      collapseDisabledTimesBox === null ||
      scheduleEventBox === null ||
      timeFormatToggleBox === null ||
      firstTimeGridRowBox === null ||
      nextPageButtonBox === null ||
      sidebarBox === null ||
      githubLinkBox === null
    ) {
      throw new Error(
        "Expected each desktop header detail and control to have boxes",
      )
    }
    // An [Event Owner](../../../docs/terminology/glossary.md#event-owner)
    // additionally renders the access-transfer action, so the owner action
    // cluster may wrap to a second line; the collapse toggle stays centered on
    // the cluster rather than on any single button.
    expect(editEventBox.y).toBeGreaterThanOrEqual(ownerActionRowBox.y)
    expect(editEventBox.y + editEventBox.height).toBeLessThanOrEqual(
      ownerActionRowBox.y + ownerActionRowBox.height,
    )
    for (const [detailBox, actionBox] of [
      [titleBox, addAvailabilityBox],
      [ownerActionRowBox, collapseDisabledTimesBox],
    ]) {
      expect(
        Math.abs(
          detailBox.y +
            detailBox.height / 2 -
            (actionBox.y + actionBox.height / 2),
        ),
      ).toBeLessThanOrEqual(1)
      expect(
        Math.abs(actionBox.width - addAvailabilityBox.width),
      ).toBeLessThanOrEqual(1)
      expect(Math.abs(actionBox.x - addAvailabilityBox.x)).toBeLessThanOrEqual(
        1,
      )
    }
    expect(
      Math.abs(scheduleEventBox.width - addAvailabilityBox.width),
    ).toBeLessThanOrEqual(1)
    expect(
      Math.abs(scheduleEventBox.x - addAvailabilityBox.x),
    ).toBeLessThanOrEqual(1)
    expect(
      Math.abs(timeFormatToggleBox.y - firstTimeGridRowBox.y),
    ).toBeLessThanOrEqual(1)
    const gridRightToTimeFormatToggleLeft =
      timeFormatToggleBox.x -
      (firstTimeGridRowBox.x + firstTimeGridRowBox.width)
    expect(gridRightToTimeFormatToggleLeft).toBeGreaterThanOrEqual(16)
    expect(gridRightToTimeFormatToggleLeft).toBeLessThanOrEqual(20)
    expect(
      Math.abs(nextPageButtonBox.x - timeFormatToggleBox.x),
    ).toBeLessThanOrEqual(2)
    expect(
      Math.abs(
        addAvailabilityBox.x +
          addAvailabilityBox.width -
          (sidebarBox.x + sidebarBox.width),
      ),
    ).toBeLessThanOrEqual(1)
    expect(
      Math.abs(
        githubLinkBox.x +
          githubLinkBox.width -
          (sidebarBox.x + sidebarBox.width),
      ),
    ).toBeLessThanOrEqual(1)

    const viewport = page.viewportSize()
    if (viewport === null) {
      throw new Error("Expected a viewport size")
    }
    const columnSideInset = (viewport.width - 1024) / 2
    expect(
      Math.abs(
        sidebarBox.x +
          sidebarBox.width -
          (viewport.width - columnSideInset - 16),
      ),
    ).toBeLessThanOrEqual(1)
    expect(Math.abs(titleBox.x - (columnSideInset + 16))).toBeLessThanOrEqual(1)

    const collapseContentBounds = await page.evaluate<{
      left: number
      right: number
      center: number
    } | null>(() => {
      const toggle = document.querySelector<HTMLElement>(
        "#collapse-disabled-times-toggle",
      )
      const control = toggle?.closest<HTMLElement>(".v-selection-control")
      const input = control?.querySelector<HTMLElement>(
        ".v-selection-control__wrapper",
      )
      const label = control?.querySelector<HTMLElement>(".v-label")

      if (!toggle || !input || !label) return null

      const inputRect = input.getBoundingClientRect()
      const labelRect = label.getBoundingClientRect()
      const left = Math.min(inputRect.left, labelRect.left)
      const right = Math.max(inputRect.right, labelRect.right)
      return { left, right, center: (left + right) / 2 }
    })
    expect(collapseContentBounds).not.toBeNull()
    if (collapseContentBounds === null) {
      throw new Error("Expected the Collapse disabled times switch content")
    }
    expect(collapseContentBounds.left).toBeGreaterThanOrEqual(
      addAvailabilityBox.x - 1,
    )
    expect(collapseContentBounds.right).toBeLessThanOrEqual(
      addAvailabilityBox.x + addAvailabilityBox.width + 1,
    )
    expect(
      Math.abs(
        collapseContentBounds.center -
          (collapseDisabledTimesBox.x + collapseDisabledTimesBox.width / 2),
      ),
    ).toBeLessThanOrEqual(2)
  }

  // Verify the parent wrapper does NOT have tw:col-span-2 (which makes it very wide)
  const parentWrapper = page.locator(
    "#event-header-actions .desktop-primary-availability-anchor",
  )
  const parentClass = await parentWrapper.getAttribute("class")
  expect(parentClass).not.toContain("tw:col-span-2")

  // Verify "More options" is NOT present
  const moreOptions = page.locator("#desktop-header-more-options")
  await expect(moreOptions).not.toBeVisible()
})

test("timed event timezone stays close to Responses", async ({
  page,
  request,
}, testInfo) => {
  test.skip(
    testInfo.project.name === "chromium-mobile",
    "Desktop-only sidebar layout",
  )

  const now = Temporal.Now.instant()
  const today = now.toZonedDateTimeISO("UTC").toPlainDate().toString()
  const seed = await seedCanonicalTimedEvent(
    request,
    buildSpecificDateSeed({
      name: `Timed responses spacing ${String(now.epochMilliseconds)}`,
      selectedDays: [today],
      activeSlots: [`${today}T09:00:00.000Z`, `${today}T10:00:00.000Z`],
      eventTimezone: "UTC",
      startTimeLocal: "09:00",
      endTimeLocal: "17:00",
      timeIncrementMinutes: 60,
    }),
  )

  await openEventPage(page, seed.shortId)

  const timezone = page.getByTestId("timezone-select-trigger")
  const responsesHeading = page.getByText("Responses", { exact: true })
  await expect(timezone).toBeVisible()
  await expect(responsesHeading).toBeVisible()

  const [timezoneBox, responsesHeadingBox] = await Promise.all([
    timezone.boundingBox(),
    responsesHeading.boundingBox(),
  ])

  if (timezoneBox === null || responsesHeadingBox === null) {
    throw new Error(
      "Expected the timezone selector and Responses heading to have boxes",
    )
  }

  expect(
    responsesHeadingBox.y - (timezoneBox.y + timezoneBox.height),
  ).toBeLessThanOrEqual(10)
})

test("timed add availability controls stay close to the Legend", async ({
  page,
  request,
}, testInfo) => {
  test.skip(
    testInfo.project.name === "chromium-mobile",
    "Desktop-only sidebar layout",
  )

  const now = Temporal.Now.instant()
  const today = now.toZonedDateTimeISO("UTC").toPlainDate().toString()
  const seed = await seedCanonicalTimedEvent(
    request,
    buildSpecificDateSeed({
      name: `Timed add availability spacing ${String(now.epochMilliseconds)}`,
      selectedDays: [today],
      activeSlots: [`${today}T09:00:00.000Z`, `${today}T10:00:00.000Z`],
      eventTimezone: "UTC",
      startTimeLocal: "09:00",
      endTimeLocal: "17:00",
      timeIncrementMinutes: 60,
    }),
  )

  await openEventPage(page, seed.shortId)
  await waitForScheduleOverlapMounted(page)
  await page.locator("#desktop-primary-availability-btn").click()
  await page.getByRole("button", { name: "Manually", exact: true }).click()

  const lastEditControl = page.locator(".calendar-options-button")
  const legend = page.getByText("Legend", { exact: true })
  await expect(lastEditControl).toBeVisible()
  await expect(legend).toBeVisible()

  const [lastEditControlBox, legendBox] = await Promise.all([
    lastEditControl.boundingBox(),
    legend.boundingBox(),
  ])
  if (lastEditControlBox === null || legendBox === null) {
    throw new Error(
      "Expected Add availability controls and Legend to have boxes",
    )
  }

  expect(
    legendBox.y - (lastEditControlBox.y + lastEditControlBox.height),
  ).toBeLessThanOrEqual(10)
})

test("desktop header details and controls stack independently", async ({
  page,
}, testInfo) => {
  test.skip(
    testInfo.project.name !== "chromium-desktop",
    "Desktop-only header layout geometry",
  )

  const now = Temporal.Now.instant()
  const today = now.toZonedDateTimeISO("UTC").toPlainDate().toString()
  const seedFor = (name: string) =>
    seedCanonicalTimedEvent(page.request, {
      ...buildSpecificDateSeed({
        name,
        selectedDays: [today],
        activeSlots: [`${today}T09:00:00.000Z`, `${today}T10:00:00.000Z`],
        eventTimezone: "UTC",
        startTimeLocal: "09:00",
        endTimeLocal: "17:00",
        timeIncrementMinutes: 60,
      }),
      description: "A saved description",
    })

  const measureHeader = async () => {
    const detailsColumn = page.locator("#event-header-details-column")
    const controlsColumn = page.locator("#event-header-controls-column")
    const toggle = page.locator(
      ".desktop-event-header-options__collapse-disabled-times-switch",
    )
    const descriptionRow = page.locator("#event-header-description-row")
    const scheduleButton = page.locator("#desktop-schedule-event-btn")
    await expect(detailsColumn).toBeVisible()
    await expect(controlsColumn).toBeVisible()
    await expect(toggle).toBeVisible()
    await expect(descriptionRow).toBeVisible()
    await expect(scheduleButton).toBeVisible()

    const [
      detailsBox,
      controlsBox,
      titleBox,
      metaRowBox,
      descriptionBox,
      toggleBox,
      scheduleBox,
    ] = await Promise.all([
      detailsColumn.boundingBox(),
      controlsColumn.boundingBox(),
      page.locator("#event-header-title").boundingBox(),
      page.locator("#event-header-meta-row").boundingBox(),
      descriptionRow.boundingBox(),
      toggle.boundingBox(),
      scheduleButton.boundingBox(),
    ])
    if (
      detailsBox === null ||
      controlsBox === null ||
      titleBox === null ||
      metaRowBox === null ||
      descriptionBox === null ||
      toggleBox === null ||
      scheduleBox === null
    ) {
      throw new Error("Expected the desktop header columns to have boxes")
    }

    return {
      detailsBox,
      controlsBox,
      titleBox,
      metaRowBox,
      descriptionBox,
      toggleBox,
      scheduleBox,
    }
  }

  type HeaderBoxes = Awaited<ReturnType<typeof measureHeader>>
  const offsetsFor = (header: HeaderBoxes) => ({
    title: header.titleBox.y - header.detailsBox.y,
    meta: header.metaRowBox.y - header.detailsBox.y,
    description: header.descriptionBox.y - header.detailsBox.y,
    toggle: header.toggleBox.y - header.controlsBox.y,
    schedule: header.scheduleBox.y - header.controlsBox.y,
  })

  const shortSeed = await seedFor(`Short name ${String(now.epochMilliseconds)}`)
  await openEventPage(page, shortSeed.shortId)
  const shortHeader = await measureHeader()

  const longName =
    (
      "A long desktop event name that has to wrap " +
      "onto several lines in the header "
    ).slice(0, 80) + String(now.epochMilliseconds)
  const longSeed = await seedFor(longName)
  await openEventPage(page, longSeed.shortId)
  const longHeader = await measureHeader()

  await test.step("the paired groups share the same gaps", () => {
    const offsets = offsetsFor(shortHeader)
    expect(
      Math.abs(shortHeader.detailsBox.y - shortHeader.controlsBox.y),
    ).toBeLessThanOrEqual(1)
    expect(Math.abs(offsets.title)).toBeLessThanOrEqual(1)
    expect(Math.abs(offsets.meta - offsets.toggle)).toBeLessThanOrEqual(1)
    expect(
      Math.abs(offsets.description - offsets.schedule),
    ).toBeLessThanOrEqual(1)
  })

  await test.step("a wrapped title moves only the details column", () => {
    const shortOffsets = offsetsFor(shortHeader)
    const longOffsets = offsetsFor(longHeader)
    expect(longHeader.titleBox.height).toBeGreaterThan(
      shortHeader.titleBox.height,
    )
    expect(longHeader.detailsBox.height).toBeGreaterThan(
      shortHeader.detailsBox.height,
    )
    expect(
      Math.abs(longHeader.controlsBox.height - shortHeader.controlsBox.height),
    ).toBeLessThanOrEqual(1)
    expect(longOffsets.meta).toBeGreaterThan(shortOffsets.meta)
    expect(longOffsets.description).toBeGreaterThan(shortOffsets.description)
    expect(
      Math.abs(longOffsets.toggle - shortOffsets.toggle),
    ).toBeLessThanOrEqual(1)
    expect(
      Math.abs(longOffsets.schedule - shortOffsets.schedule),
    ).toBeLessThanOrEqual(1)
  })

  await test.step("a taller control group moves only the controls column", async () => {
    await page
      .locator("#event-header-controls-column")
      .evaluate((controlsColumn) => {
        const firstGroup = controlsColumn.firstElementChild
        if (!(firstGroup instanceof HTMLElement)) {
          throw new Error("Expected a first controls group")
        }
        firstGroup.style.minHeight = "160px"
      })

    const grownHeader = await measureHeader()
    const longOffsets = offsetsFor(longHeader)
    const grownOffsets = offsetsFor(grownHeader)

    expect(
      Math.abs(grownHeader.detailsBox.height - longHeader.detailsBox.height),
    ).toBeLessThanOrEqual(1)
    expect(
      Math.abs(grownOffsets.title - longOffsets.title),
    ).toBeLessThanOrEqual(1)
    expect(Math.abs(grownOffsets.meta - longOffsets.meta)).toBeLessThanOrEqual(
      1,
    )
    expect(
      Math.abs(grownOffsets.description - longOffsets.description),
    ).toBeLessThanOrEqual(1)
    expect(grownOffsets.toggle - longOffsets.toggle).toBeGreaterThanOrEqual(100)
    expect(grownOffsets.schedule - longOffsets.schedule).toBeGreaterThanOrEqual(
      100,
    )
  })
})

test("mobile group header keeps availability actions between the title and the metadata actions", async ({
  page,
}, testInfo) => {
  test.skip(
    testInfo.project.name !== "chromium-mobile",
    "Mobile-only header order assertions",
  )

  const now = Temporal.Now.instant()
  const today = now.toZonedDateTimeISO("UTC").toPlainDate().toString()
  const seed = await seedCanonicalTimedEvent(page.request, {
    name: `Mobile header order ${String(now.epochMilliseconds)}`,
    type: "group",
    description: "A saved description",
    activeSlots: [`${today}T09:00:00.000Z`, `${today}T10:00:00.000Z`],
    eventTimezone: "UTC",
    slotGeneration: {
      startTimeLocal: "09:00",
      endTimeLocal: "17:00",
      timeIncrementMinutes: 60,
    },
    timedRecurrence: {
      kind: "weekly",
      selectedDays: [today],
      selectedDaysOfWeek: [Temporal.PlainDate.from(today).dayOfWeek],
      startOnMonday: true,
    },
  })

  await page.goto(`/g/${seed.shortId}`, { waitUntil: "domcontentloaded" })
  await dismissConsent(page)
  await waitForEventShell(page)

  const title = page.locator("#event-header-title")
  const groupActions = page.locator("#event-header-mobile-group-actions")
  const metaRow = page.locator("#event-header-meta-row")
  const descriptionRow = page.locator("#event-header-description-row")
  await expect(title).toBeVisible()
  await expect(groupActions).toHaveCount(1)
  await expect(groupActions).toBeVisible()
  await expect(metaRow).toBeVisible()
  await expect(descriptionRow).toBeVisible()

  const [titleBox, groupActionsBox, metaRowBox, descriptionBox] =
    await Promise.all([
      title.boundingBox(),
      groupActions.boundingBox(),
      metaRow.boundingBox(),
      descriptionRow.boundingBox(),
    ])
  if (
    titleBox === null ||
    groupActionsBox === null ||
    metaRowBox === null ||
    descriptionBox === null
  ) {
    throw new Error("Expected the mobile header groups to have boxes")
  }

  expect(titleBox.y + titleBox.height).toBeLessThanOrEqual(
    groupActionsBox.y + 1,
  )
  expect(groupActionsBox.y + groupActionsBox.height).toBeLessThanOrEqual(
    metaRowBox.y + 1,
  )
  expect(metaRowBox.y + metaRowBox.height).toBeLessThanOrEqual(
    descriptionBox.y + 1,
  )

  // The desktop controls column contributes no duplicate availability actions
  // on a phone viewport.
  await expect(page.locator("#event-header-controls-column > *")).toHaveCount(0)
})
