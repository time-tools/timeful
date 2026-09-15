import "vuetify/styles"
import { createVuetify } from "vuetify"
import * as directives from "vuetify/directives"
import { aliases, mdi } from "vuetify/iconsets/mdi-svg"

export default createVuetify({
  directives,
  icons: {
    defaultSet: "mdi",
    aliases,
    sets: { mdi },
  },
  theme: {
    defaultTheme: "light",
    themes: {
      light: {
        colors: {
          primary: "#00994C",
          error: "#DB1616",
        },
      },
    },
  },
  display: {
    thresholds: {
      sm: 640,
      md: 768,
      lg: 1024,
      xl: 1280,
    },
  },
})
