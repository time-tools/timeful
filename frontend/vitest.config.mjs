import { defineConfig } from "vitest/config"
import vue from "@vitejs/plugin-vue"
import Icons from "unplugin-icons/vite"
import path from "path"
import { fileURLToPath } from "url"

const rootDirectory = path.dirname(fileURLToPath(import.meta.url))

export default defineConfig({
  plugins: [vue(), Icons({ compiler: "vue3", scale: 1 })],
  resolve: {
    alias: {
      "@": path.resolve(rootDirectory, "./src"),
    },
  },
  test: {
    environment: "node",
    include: ["src/**/*.test.{js,ts}", "eslint/**/*.test.{js,ts}"],
    setupFiles: ["./src/test/silenceNodeLocalStorageWarning.ts"],
    testTimeout: 15000,
    slowTestThreshold: 100,
  },
})
