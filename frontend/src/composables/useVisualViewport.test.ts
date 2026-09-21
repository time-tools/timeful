// @vitest-environment happy-dom

import { effectScope } from "vue"
import { afterEach, describe, expect, it } from "vitest"
import { useVisualViewport } from "./useVisualViewport"

type FakeVisualViewport = EventTarget & {
  height: number
  offsetTop: number
  width: number
}

const installVisualViewport = (rect: {
  height: number
  offsetTop: number
}): FakeVisualViewport => {
  const viewport = Object.assign(new EventTarget(), {
    height: rect.height,
    offsetTop: rect.offsetTop,
    width: 390,
  })
  Object.defineProperty(window, "visualViewport", {
    configurable: true,
    value: viewport,
  })
  return viewport
}

afterEach(() => {
  Reflect.deleteProperty(window, "visualViewport")
})

describe("useVisualViewport", () => {
  it("reads the visible viewport rect on creation", () => {
    installVisualViewport({ height: 640, offsetTop: 0 })

    const scope = effectScope()
    const viewport = scope.run(() => useVisualViewport())

    expect(viewport?.value).toEqual({ top: 0, height: 640 })
    scope.stop()
  })

  it("tracks visible viewport resize and scroll events", () => {
    const visibleViewport = installVisualViewport({ height: 640, offsetTop: 0 })
    const scope = effectScope()
    const viewport = scope.run(() => useVisualViewport())

    visibleViewport.height = 340
    visibleViewport.offsetTop = 120
    visibleViewport.dispatchEvent(new Event("resize"))
    expect(viewport?.value).toEqual({ top: 120, height: 340 })

    visibleViewport.offsetTop = 40
    visibleViewport.dispatchEvent(new Event("scroll"))
    expect(viewport?.value).toEqual({ top: 40, height: 340 })

    scope.stop()
  })

  it("stops tracking once its scope is disposed", () => {
    const visibleViewport = installVisualViewport({ height: 640, offsetTop: 0 })
    const scope = effectScope()
    const viewport = scope.run(() => useVisualViewport())
    expect(viewport?.value).toEqual({ top: 0, height: 640 })

    scope.stop()

    visibleViewport.height = 340
    visibleViewport.dispatchEvent(new Event("resize"))
    expect(viewport?.value).toEqual({ top: 0, height: 640 })
  })

  it("reports no rect when the visible viewport API is unavailable", () => {
    const scope = effectScope()
    const viewport = scope.run(() => useVisualViewport())

    expect(viewport?.value).toBeNull()
    scope.stop()
  })
})
