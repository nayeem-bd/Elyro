import { createRouter, createWebHashHistory } from 'vue-router'
import { TaskService } from '../bindings/elyro/internal/todo'
import { useLists } from './stores/lists'
import ListView from './views/ListView.vue'
import SmartView from './views/SmartView.vue'

// Hash history: the app is served from embedded files, so there is no server to rewrite deep links.
export const router = createRouter({
  history: createWebHashHistory(),
  linkActiveClass: 'bg-gray-200 text-gray-900',
  routes: [
    { path: '/', redirect: () => `/list/${useLists().inbox?.id}` },
    {
      path: '/today',
      component: SmartView,
      props: { title: 'Today', fetch: TaskService.Today, empty: 'Nothing due today.' },
    },
    {
      path: '/upcoming',
      component: SmartView,
      props: { title: 'Upcoming', fetch: TaskService.Upcoming, empty: 'Nothing scheduled.' },
    },
    {
      path: '/completed',
      component: SmartView,
      props: { title: 'Completed', fetch: TaskService.Completed, empty: 'No completed tasks.' },
    },
    {
      path: '/search',
      component: SmartView,
      props: (r) => {
        const q = String(r.query.q ?? '')
        return { title: q ? `Results for “${q}”` : 'Search', fetch: () => TaskService.Search(q), empty: 'No matches.' }
      },
    },
    { path: '/list/:id', component: ListView },
  ],
})
