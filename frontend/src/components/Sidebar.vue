<script setup lang="ts">
import { Sun, CalendarDays, CircleCheck, Inbox, List, Plus } from '@lucide/vue'
import { nextTick, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useLists } from '../stores/lists'

const views = [
  { to: '/today', label: 'Today', icon: Sun },
  { to: '/upcoming', label: 'Upcoming', icon: CalendarDays },
  { to: '/completed', label: 'Completed', icon: CircleCheck },
]
const linkClass = 'flex items-center gap-2 rounded-md px-2 py-1.5 text-sm text-gray-700 hover:bg-gray-200/60'

const lists = useLists()
const router = useRouter()
const adding = ref(false)
const draft = ref('')
const input = ref<HTMLInputElement>()

async function startAdd() {
  draft.value = ''
  adding.value = true
  await nextTick()
  input.value?.focus()
}
async function add() {
  if (!adding.value) return // Enter hides the input, whose blur would call add() again
  const name = draft.value.trim()
  adding.value = false
  if (!name) return
  const list = await lists.create(name)
  await router.push(`/list/${list.id}`)
}
</script>

<template>
  <nav class="flex h-full w-56 shrink-0 flex-col border-r border-gray-200 bg-gray-50" aria-label="Sidebar">
    <!-- Title-bar strip: drag the window here; the traffic lights sit on the left. -->
    <div class="h-13 shrink-0" style="--wails-draggable: drag" />
    <ul class="space-y-0.5 px-2">
      <li v-for="v in views" :key="v.to">
        <RouterLink :to="v.to" :class="linkClass">
          <component :is="v.icon" class="size-4" aria-hidden="true" />
          {{ v.label }}
        </RouterLink>
      </li>
    </ul>

    <h2 class="mt-6 mb-1 px-4 text-xs font-medium tracking-wide text-gray-400 uppercase">Lists</h2>
    <ul class="min-h-0 flex-1 space-y-0.5 overflow-y-auto px-2">
      <li v-for="l in lists.lists" :key="l.id">
        <RouterLink :to="`/list/${l.id}`" :class="linkClass">
          <component :is="l.isInbox ? Inbox : List" class="size-4 shrink-0" aria-hidden="true" />
          <span class="truncate">{{ l.name }}</span>
        </RouterLink>
      </li>
      <li v-if="adding">
        <input
          ref="input"
          v-model="draft"
          maxlength="500"
          placeholder="List name"
          aria-label="New list name"
          class="w-full rounded-md border border-blue-400 bg-white px-2 py-1 text-sm outline-none"
          @keydown.enter.prevent="add"
          @keydown.esc.prevent="adding = false"
          @blur="add"
        />
      </li>
    </ul>

    <button
      type="button"
      class="m-2 flex items-center gap-2 rounded-md px-2 py-1.5 text-sm text-gray-500 hover:bg-gray-200/60 hover:text-gray-900"
      @click="startAdd"
    >
      <Plus class="size-4" aria-hidden="true" /> New list
    </button>
  </nav>
</template>
