import { createRouter, createWebHashHistory } from 'vue-router'
import PlaceholderView from './views/PlaceholderView.vue'

// Hash history: the app is served from embedded files, so there is no server to rewrite deep links.
export const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/', redirect: '/today' },
    { path: '/today', component: PlaceholderView, meta: { title: 'Today' } },
    { path: '/upcoming', component: PlaceholderView, meta: { title: 'Upcoming' } },
    { path: '/completed', component: PlaceholderView, meta: { title: 'Completed' } },
    { path: '/list/:id', component: PlaceholderView, meta: { title: 'List' } },
    { path: '/search', component: PlaceholderView, meta: { title: 'Search' } },
  ],
})
