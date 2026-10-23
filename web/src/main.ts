import { createApp } from 'vue'
import App from './App.vue'
import router from './router'

// Import Bootstrap CSS and JS
import 'bootstrap/dist/css/bootstrap.min.css'
import 'bootstrap/dist/js/bootstrap.bundle.min.js'

// Import BootstrapVue Next
import { createBootstrap } from 'bootstrap-vue-next'
import 'bootstrap-vue-next/dist/bootstrap-vue-next.css'

// Import custom styles
import './assets/styles.css'

import { registerServiceWorker } from './composables/usePwa'

// After a new deploy, an open tab may ask for a page chunk that no longer
// exists: reload once to get the new version.
router.onError((err, to) => {
  const msg = String((err as Error)?.message ?? err)
  if (/dynamically imported module|Importing a module script failed|Failed to fetch/i.test(msg) && !sessionStorage.getItem('chunk-reload')) {
    sessionStorage.setItem('chunk-reload', '1')
    window.location.assign(to.fullPath)
  }
})
router.afterEach(() => { try { sessionStorage.removeItem('chunk-reload') } catch { /* private mode */ } })

registerServiceWorker()

const app = createApp(App)
app.use(router)
app.use(createBootstrap())
app.mount('#app')

