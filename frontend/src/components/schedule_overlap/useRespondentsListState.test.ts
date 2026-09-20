import { computed, nextTick, ref } from "vue"
import { describe, expect, it } from "vitest"
import { Temporal } from "temporal-polyfill"
import { UTC } from "@/constants"
import { ZdtSet } from "@/utils"
import type {
  ParsedResponse,
  ParsedResponses,
  TimedCellState,
} from "@/composables/schedule_overlap/types"
import {
  respondentStatusClass,
  useRespondentsListState,
} from "./useRespondentsListState"

const baseSlot = Temporal.Instant.from(
  "2026-01-01T09:00:00Z",
).toZonedDateTimeISO(UTC)
const otherSlot = Temporal.Instant.from(
  "2026-01-01T10:00:00Z",
).toZonedDateTimeISO(UTC)

function makeState(options: {
  active?: Temporal.ZonedDateTime
  inactive?: boolean
  cellState?: TimedCellState | null
  collapsed?: boolean
  availability?: Temporal.ZonedDateTime[]
  ifNeeded?: Temporal.ZonedDateTime[]
  hideIfNeeded?: boolean
}) {
  const parsedResponses = {
    "user-1": {
      user: { _id: "user-1", firstName: "Ada" } as never,
      availability: new ZdtSet(options.availability ?? []),
      ifNeeded: new ZdtSet(options.ifNeeded ?? []),
      guest: false,
    },
  }
  return useRespondentsListState({
    event: { blindAvailabilityEnabled: false },
    respondents: computed(() => [parsedResponses["user-1"].user]),
    curRespondents: computed(() => []),
    curTimeslotAvailability: computed(() => ({ "user-1": false })),
    curTimeslotInactive: computed(() => options.inactive ?? false),
    curTimeslotCellState: computed(() => options.cellState ?? null),
    curTimeslotCollapsed: computed(() => options.collapsed ?? false),
    parsedResponses: computed(() => parsedResponses),
    ownedGuestResponseLookupKeys: computed(() => new Set<string>()),
    curDate: computed(() => options.active ?? undefined),
    hideIfNeeded: computed(() => options.hideIfNeeded ?? false),
    isGroup: computed(() => false),
    attendees: computed(() => []),
    isOwner: computed(() => false),
    isPhone: computed(() => false),
  })
}

describe("useRespondentsListState respondentSlotStatus", () => {
  it("returns null when no slot is in context", () => {
    const state = makeState({})
    expect(state.respondentSlotStatus("user-1")).toBeNull()
  })

  it("returns available when the active slot is in the respondent availability", () => {
    const state = makeState({
      active: baseSlot,
      availability: [baseSlot],
      ifNeeded: [baseSlot],
    })
    expect(state.respondentSlotStatus("user-1")).toBe("available")
  })

  it("returns if-needed when the active slot only matches the if-needed set", () => {
    const state = makeState({ active: baseSlot, ifNeeded: [baseSlot] })
    expect(state.respondentSlotStatus("user-1")).toBe("if-needed")
  })

  it("returns unavailable when the active slot is in neither set", () => {
    const state = makeState({ active: otherSlot })
    expect(state.respondentSlotStatus("user-1")).toBe("unavailable")
  })

  it("returns disabled-inactive when hovering an enabled-inactive cell", () => {
    const state = makeState({
      inactive: true,
      cellState: "enabled_inactive",
      active: baseSlot,
    })
    expect(state.respondentSlotStatus("user-1")).toBe("disabled-inactive")
  })

  it("returns disabled-collapsed when hovering collapsed hours", () => {
    const state = makeState({
      collapsed: true,
      inactive: true,
      cellState: "enabled_inactive",
      active: baseSlot,
    })
    expect(state.respondentSlotStatus("user-1")).toBe("disabled-collapsed")
  })

  it("returns disabled-out-of-range for any other inactive cell state", () => {
    const outsideRange = makeState({
      inactive: true,
      cellState: "outside_range",
    })
    expect(outsideRange.respondentSlotStatus("user-1")).toBe(
      "disabled-out-of-range",
    )

    const padding = makeState({ inactive: true, cellState: "padding" })
    expect(padding.respondentSlotStatus("user-1")).toBe("disabled-out-of-range")

    const unknown = makeState({ inactive: true, cellState: null })
    expect(unknown.respondentSlotStatus("user-1")).toBe("disabled-out-of-range")
  })

  it("suppresses if-needed status when hideIfNeeded is on", () => {
    const state = makeState({
      active: baseSlot,
      ifNeeded: [baseSlot],
      hideIfNeeded: true,
    })
    expect(state.respondentSlotStatus("user-1")).toBe("unavailable")
  })

  it("maps statuses to their legend tailwind classes", () => {
    expect(respondentStatusClass("available")).toBe("tw:bg-[#00994C77]")
    expect(respondentStatusClass("if-needed")).toBe("tw:bg-yellow")
    expect(respondentStatusClass("unavailable")).toBe("tw:bg-[#F9CCCC]")
    expect(respondentStatusClass("disabled-inactive")).toBe(
      "tw:bg-light-gray-stroke",
    )
    expect(respondentStatusClass("disabled-collapsed")).toBe(
      "tw:bg-(--timeful-collapsed-hours-bg) respondent-status--collapsed",
    )
    expect(respondentStatusClass("disabled-out-of-range")).toBe("tw:bg-gray")
    expect(respondentStatusClass(null)).toBe("")
  })
})

function makeOrderingState(options: {
  respondents: {
    id: string
    firstName: string
    response?: Partial<ParsedResponse>
  }[]
  ownedGuestResponseLookupKeys?: string[]
}) {
  const parsedResponses: ParsedResponses = Object.fromEntries(
    options.respondents.map((respondent) => [
      respondent.id,
      {
        user: { _id: respondent.id, firstName: respondent.firstName },
        availability: new ZdtSet(),
        ifNeeded: new ZdtSet(),
        guest: false,
        ...respondent.response,
      },
    ]),
  )
  const selected = ref<string[]>([])
  const state = useRespondentsListState({
    event: { blindAvailabilityEnabled: false },
    respondents: computed(() =>
      options.respondents.map(
        (respondent) => parsedResponses[respondent.id].user,
      ),
    ),
    curRespondents: computed(() => selected.value),
    curTimeslotAvailability: computed(() =>
      Object.fromEntries(
        options.respondents.map((respondent) => [respondent.id, false]),
      ),
    ),
    curTimeslotInactive: computed(() => false),
    curTimeslotCellState: computed(() => null),
    curTimeslotCollapsed: computed(() => false),
    parsedResponses: computed(() => parsedResponses),
    ownedGuestResponseLookupKeys: computed(
      () => new Set(options.ownedGuestResponseLookupKeys ?? []),
    ),
    curDate: computed(() => undefined),
    hideIfNeeded: computed(() => false),
    isGroup: computed(() => false),
    attendees: computed(() => []),
    isOwner: computed(() => false),
    isPhone: computed(() => false),
  })
  return { state, selected }
}

describe("useRespondentsListState orderedRespondents", () => {
  function orderedIds(state: ReturnType<typeof makeOrderingState>["state"]) {
    return state.orderedRespondents.value.map((user) => user._id)
  }

  it("orders owned responses before other open responses and protected responses", () => {
    const { state } = makeOrderingState({
      respondents: [
        {
          id: "protected-ada",
          firstName: "Ada",
          response: { guest: true, guestEditPolicy: "protected" },
        },
        {
          id: "open-bea",
          firstName: "Bea",
          response: {
            guest: true,
            guestId: "token-open",
            guestEditPolicy: "open",
            guestOwnershipMode: "token",
          },
        },
        {
          id: "owned-cora",
          firstName: "Cora",
          response: {
            guest: true,
            guestId: "token-owned",
            guestEditPolicy: "protected",
            guestOwnershipMode: "token",
          },
        },
      ],
      ownedGuestResponseLookupKeys: ["token-owned"],
    })

    expect(orderedIds(state)).toEqual([
      "owned-cora",
      "open-bea",
      "protected-ada",
    ])
  })

  it("treats canonical canEdit responses as owned", () => {
    const { state } = makeOrderingState({
      respondents: [
        {
          id: "canonical-locked",
          firstName: "Ada",
          response: { publicId: "pub-locked", canEdit: false },
        },
        {
          id: "canonical-owned",
          firstName: "Zoe",
          response: { publicId: "pub-owned", canEdit: true },
        },
      ],
    })

    expect(orderedIds(state)).toEqual(["canonical-owned", "canonical-locked"])
  })

  it("treats legacy responses in the ownership keys as owned", () => {
    const { state } = makeOrderingState({
      respondents: [
        {
          id: "legacy-protected",
          firstName: "Ada",
          response: { guest: true, guestOwnershipMode: "legacy" },
        },
        {
          id: "legacy-owned",
          firstName: "Zoe",
          response: { guest: true, guestOwnershipMode: "legacy" },
        },
      ],
      ownedGuestResponseLookupKeys: ["legacy-owned"],
    })

    expect(orderedIds(state)).toEqual(["legacy-owned", "legacy-protected"])
  })

  it("keeps alphabetical order inside an ownership tier", () => {
    const { state } = makeOrderingState({
      respondents: [
        { id: "protected-bea", firstName: "Bea", response: { guest: true } },
        { id: "protected-ada", firstName: "Ada", response: { guest: true } },
      ],
    })

    expect(orderedIds(state)).toEqual(["protected-ada", "protected-bea"])
  })

  it("keeps selected responses first and orders the rest by ownership tier", async () => {
    const { state, selected } = makeOrderingState({
      respondents: [
        {
          id: "owned-ada",
          firstName: "Ada",
          response: {
            guest: true,
            guestId: "token-owned",
            guestOwnershipMode: "token",
          },
        },
        {
          id: "protected-zoe",
          firstName: "Zoe",
          response: { guest: true, guestEditPolicy: "protected" },
        },
        {
          id: "protected-bea",
          firstName: "Bea",
          response: { guest: true, guestEditPolicy: "protected" },
        },
      ],
      ownedGuestResponseLookupKeys: ["token-owned"],
    })

    selected.value = ["protected-zoe"]
    await nextTick()

    expect(orderedIds(state)).toEqual([
      "protected-zoe",
      "owned-ada",
      "protected-bea",
    ])
  })

  it("orders selected responses by click time", async () => {
    const { state, selected } = makeOrderingState({
      respondents: [
        { id: "protected-ada", firstName: "Ada", response: { guest: true } },
        { id: "protected-bea", firstName: "Bea", response: { guest: true } },
      ],
    })

    selected.value = ["protected-bea"]
    await nextTick()
    await new Promise((resolve) => setTimeout(resolve, 2))
    selected.value = ["protected-bea", "protected-ada"]
    await nextTick()

    expect(orderedIds(state)).toEqual(["protected-bea", "protected-ada"])
  })
})
