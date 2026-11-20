import '@wailsio/runtime' // native glue: window dragging, bindings transport
import { createPinia } from 'pinia'
import { createApp } from 'vue'
import App from './App.vue'
import { showError } from './errors'
import { router } from './router'
import { useLists } from './stores/lists'
import './style.css'

const app = createApp(App).use(createPinia())
app.config.errorHandler = showError // also catches rejected promises from event handlers
router.onError(showError)

// Lists load before the router starts, so '/' can redirect straight to the Inbox.
useLists()
  .load()
  .catch(showError)
  .finally(() => app.use(router).mount('#app'))
