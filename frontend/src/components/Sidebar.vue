<script setup lang="ts">
import { Sun, CalendarDays, CircleCheck } from '@lucide/vue'

const views = [
  { to: '/today', label: 'Today', icon: Sun },
  { to: '/upcoming', label: 'Upcoming', icon: CalendarDays },
  { to: '/completed', label: 'Completed', icon: CircleCheck },
]
</script>

<template>
  <nav class="flex h-full w-56 shrink-0 flex-col border-r border-gray-200 bg-gray-50" aria-label="Sidebar">
    <!-- Title-bar strip: drag the window here; the traffic lights sit on the left. -->
    <div class="h-13 shrink-0" style="--wails-draggable: drag" />
    <ul class="space-y-0.5 px-2">
      <li v-for="v in views" :key="v.to">
        <RouterLink
          :to="v.to"
          class="flex items-center gap-2 rounded-md px-2 py-1.5 text-sm text-gray-700 hover:bg-gray-200/60"
          active-class="bg-gray-200 text-gray-900"
        >
          <component :is="v.icon" class="size-4" aria-hidden="true" />
          {{ v.label }}
        </RouterLink>
      </li>
    </ul>
    <h2 class="mt-6 px-4 text-xs font-medium tracking-wide text-gray-400 uppercase">Lists</h2>
  </nav>
</template>
