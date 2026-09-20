import { createApp } from "vue"
import { createPinia } from "pinia"
import { createHead } from "@unhead/vue/client"
import "@fontsource/chivo-mono/latin-400.css"
import "@fontsource/dm-sans/latin-400.css"
import "./index.css"
import "@/plugins/posthog"
import App from "./App.vue"
import router from "./router"
import vuetify from "./plugins/vuetify"

const app = createApp(App)

app.use(router)
app.use(createPinia())
app.use(vuetify)
app.use(createHead())

app.mount("#app")
