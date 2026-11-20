import { createRouter, createWebHashHistory } from 'vue-router'
import { useLists } from './stores/lists'
import ListView from './views/ListView.vue'
import PlaceholderView from './views/PlaceholderView.vue'

// Hash history: the app is served from embedded files, so there is no server to rewrite deep links.
export const router = createRouter({
  history: createWebHashHistory(),
  linkActiveClass: 'bg-gray-200 text-gray-900',
  routes: [
    { path: '/', redirect: () => `/list/${useLists().inbox?.id}` },
    { path: '/today', component: PlaceholderView, props: { title: 'Today' } },
    { path: '/upcoming', component: PlaceholderView, props: { title: 'Upcoming' } },
    { path: '/completed', component: PlaceholderView, props: { title: 'Completed' } },
    { path: '/list/:id', component: ListView },
  ],
})
