import type { Event } from "@/types"

// Canonical public identifier used to open an event. Canonical events carry
// their public identifier in _id; shortId covers payloads that only expose the
// compatibility field.
export function eventPublicId(event: Pick<Event, "_id" | "shortId">): string {
  return event._id ?? event.shortId ?? ""
}
