import { onScopeDispose, readonly, ref, type Ref } from "vue"

export interface VisualViewportRect {
  top: number
  height: number
}

const readVisualViewportRect = (): VisualViewportRect | null => {
  if (typeof window === "undefined") return null
  const visibleViewport = window.visualViewport
  if (!visibleViewport) return null
  return { top: visibleViewport.offsetTop, height: visibleViewport.height }
}

export function useVisualViewport(): Readonly<Ref<VisualViewportRect | null>> {
  const rect = ref<VisualViewportRect | null>(readVisualViewportRect())

  const update = () => {
    rect.value = readVisualViewportRect()
  }

  if (typeof window !== "undefined") {
    window.visualViewport?.addEventListener("resize", update)
    window.visualViewport?.addEventListener("scroll", update)
    onScopeDispose(() => {
      window.visualViewport?.removeEventListener("resize", update)
      window.visualViewport?.removeEventListener("scroll", update)
    })
  }

  return readonly(rect)
}
