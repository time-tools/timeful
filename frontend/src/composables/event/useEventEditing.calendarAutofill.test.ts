// @vitest-environment happy-dom

import { computed, nextTick, ref } from "vue"
import { beforeEach, describe, expect, it, vi } from "vitest"
import { calendarTypes } from "@/constants"
import type * as UtilsModule from "@/utils"
import type { Event, User } from "@/types"
import type { ScheduleOverlapInstance } from "./types"
import { useEventEditing } from "./useEventEditing"

const { showInfoMock, showErrorMock, signInGoogleMock, signInOutlookMock } =
  vi.hoisted(() => ({
    showInfoMock: vi.fn(),
    showErrorMock: vi.fn(),
    signInGoogleMock: vi.fn(),
    signInOutlookMock: vi.fn(),
  }))

vi.mock("@/stores/main", () => ({
  useMainStore: () => ({
    showInfo: showInfoMock,
    showError: showErrorMock,
  }),
}))

vi.mock("@/utils", async (importOriginal) => {
  const actual = await importOriginal<typeof UtilsModule>()
  return {
    ...actual,
    signInGoogle: signInGoogleMock,
    signInOutlook: signInOutlookMock,
  }
})

vi.mock("@/utils/calendarAutofillAvailability", () => ({
  calendarAutofillEnabled: false,
}))

function createScheduleOverlap() {
  return ref({
    startEditing: vi.fn(),
    setAvailabilityAutomatically: vi.fn(),
    clearSelectedGuestOwnership: vi.fn(),
  } as unknown as ScheduleOverlapInstance)
}

function createEditing(opts: {
  scheduleOverlapRef: ReturnType<typeof createScheduleOverlap>
  authUser: User | null
  calendarPermissionGranted: boolean
  userHasResponded: boolean
}) {
  return useEventEditing({
    event: ref({
      _id: "evt-1",
      eventVisitorId: "visitor-1",
      responses: {},
    } as Event),
    eventId: ref("evt-1"),
    authUser: computed(() => opts.authUser),
    scheduleOverlapRef: opts.scheduleOverlapRef,
    isSignUp: computed(() => false),
    isGroup: computed(() => false),
    userHasResponded: computed(() => opts.userHasResponded),
    curGuestId: ref(""),
    addingAvailabilityAsGuest: ref(false),
    calendarPermissionGranted: ref(opts.calendarPermissionGranted),
    refreshEvent: vi.fn().mockResolvedValue(undefined),
  })
}

describe("useEventEditing with calendar autofill disabled", () => {
  beforeEach(() => {
    showInfoMock.mockReset()
    showErrorMock.mockReset()
    signInGoogleMock.mockReset()
    signInOutlookMock.mockReset()
  })

  it("starts manual editing instead of opening the choice dialog for a signed-out viewer", async () => {
    const scheduleOverlapRef = createScheduleOverlap()
    const editing = createEditing({
      scheduleOverlapRef,
      authUser: null,
      calendarPermissionGranted: false,
      userHasResponded: false,
    })

    editing.addAvailability()
    await nextTick()

    expect(scheduleOverlapRef.value.startEditing).toHaveBeenCalledTimes(1)
    expect(
      scheduleOverlapRef.value.setAvailabilityAutomatically,
    ).not.toHaveBeenCalled()
    expect(editing.choiceDialog.value).toBe(false)
  })

  it("skips automatic calendar filling for an unresponded viewer with calendar access", async () => {
    const scheduleOverlapRef = createScheduleOverlap()
    const editing = createEditing({
      scheduleOverlapRef,
      authUser: { _id: "user-1" },
      calendarPermissionGranted: true,
      userHasResponded: false,
    })

    editing.addAvailability()
    await nextTick()

    expect(scheduleOverlapRef.value.startEditing).toHaveBeenCalledTimes(1)
    expect(
      scheduleOverlapRef.value.setAvailabilityAutomatically,
    ).not.toHaveBeenCalled()
    expect(editing.choiceDialog.value).toBe(false)
  })

  it("keeps automatic calendar filling inert", () => {
    const scheduleOverlapRef = createScheduleOverlap()
    const editing = createEditing({
      scheduleOverlapRef,
      authUser: { _id: "user-1" },
      calendarPermissionGranted: true,
      userHasResponded: false,
    })

    editing.setAvailabilityAutomatically(calendarTypes.GOOGLE)
    editing.setAvailabilityAutomatically(calendarTypes.OUTLOOK)

    expect(signInGoogleMock).not.toHaveBeenCalled()
    expect(signInOutlookMock).not.toHaveBeenCalled()
    expect(editing.choiceDialog.value).toBe(false)
  })
})
