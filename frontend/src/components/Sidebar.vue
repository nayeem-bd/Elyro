<script setup lang="ts">
import { Sun, CalendarDays, CircleCheck, Inbox, List, Plus, Search, Settings } from '@lucide/vue'
import { Events } from '@wailsio/runtime'
import { nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { List as TodoList } from '../../bindings/elyro/internal/todo'
import { carries, droppedId, leftTarget, LIST, startDrag, TASK } from '../dnd'
import { useLists } from '../stores/lists'
import { useTasks } from '../stores/tasks'

const views = [
  { to: '/today', label: 'Today', icon: Sun },
  { to: '/upcoming', label: 'Upcoming', icon: CalendarDays },
  { to: '/completed', label: 'Completed', icon: CircleCheck },
]
const linkClass = 'flex items-center gap-2 rounded-md px-2 py-1.5 text-sm text-gray-700 hover:bg-gray-200/60'

const lists = useLists()
const tasks = useTasks()
const route = useRoute()
const router = useRouter()

// Search: the query lives in the URL (/search?q=); typing pauses 200 ms before it runs.
const q = ref('')
watch(() => route.path, (p) => p !== '/search' && (q.value = ''))
let timer: ReturnType<typeof setTimeout>
function search() {
  clearTimeout(timer)
  timer = setTimeout(() => router.replace({ path: '/search', query: { q: q.value } }), 200)
}

// Drag a task onto a list to move it there; drag a list onto another to reorder (Inbox stays first).
const over = ref<{ id: number; kind: 'task' | 'list' } | null>(null) // id 0 = after the last list
function dragOver(e: DragEvent, l?: TodoList) {
  const kind = carries(e, TASK) && l ? 'task' : carries(e, LIST) && !l?.isInbox ? 'list' : null
  if (!kind) return
  e.preventDefault()
  over.value = { id: l?.id ?? 0, kind }
}
async function drop(e: DragEvent, l?: TodoList) {
  over.value = null
  const taskId = droppedId(e, TASK)
  const listId = droppedId(e, LIST)
  if (taskId && l) await tasks.move(taskId, l.id)
  else if (listId && listId !== l?.id) await lists.move(listId, l?.id ?? 0)
}
const adding = ref(false)
const draft = ref('')
const input = ref<HTMLInputElement>()

async function startAdd() {
  draft.value = ''
  adding.value = true
  await nextTick()
  input.value?.focus()
}
onBeforeUnmount(Events.On('menu', (e: { data: string }) => e.data === 'newList' && startAdd()))
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
    <label class="relative mx-2 mb-3 block">
      <Search class="pointer-events-none absolute top-1/2 left-2 size-3.5 -translate-y-1/2 text-gray-400" aria-hidden="true" />
      <input
        id="search"
        v-model="q"
        type="search"
        placeholder="Search"
        aria-label="Search tasks"
        class="w-full rounded-md border border-gray-200 bg-white py-1 pr-2 pl-7 text-sm outline-none focus:border-blue-400"
        @input="search"
      />
    </label>
    <ul class="space-y-0.5 px-2">
      <li v-for="v in views" :key="v.to">
        <RouterLink :to="v.to" :class="linkClass">
          <component :is="v.icon" class="size-4" aria-hidden="true" />
          {{ v.label }}
        </RouterLink>
      </li>
    </ul>

    <h2 class="mt-6 mb-1 px-4 text-xs font-medium tracking-wide text-gray-400 uppercase">Lists</h2>
    <ul
      class="min-h-0 flex-1 space-y-0.5 overflow-y-auto px-2 pb-4"
      @dragover="dragOver($event)"
      @dragleave="leftTarget($event) && (over = null)"
      @drop.prevent="drop($event)"
    >
      <li
        v-for="l in lists.lists"
        :key="l.id"
        :draggable="!l.isInbox"
        class="rounded-md border-t-2"
        :class="[
          over?.id === l.id && over.kind === 'list' ? 'border-blue-400' : 'border-transparent',
          over?.id === l.id && over.kind === 'task' && 'ring-2 ring-blue-400',
        ]"
        @dragstart="startDrag($event, LIST, l.id)"
        @dragover.stop="dragOver($event, l)"
        @drop.stop.prevent="drop($event, l)"
      >
        <RouterLink :to="`/list/${l.id}`" :class="linkClass" draggable="false">
          <component :is="l.isInbox ? Inbox : List" class="size-4 shrink-0" aria-hidden="true" />
          <span class="truncate">{{ l.name }}</span>
        </RouterLink>
      </li>
      <li v-if="over?.id === 0" class="h-0.5 rounded bg-blue-400" aria-hidden="true" />
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

    <div class="m-2 flex items-center">
      <button
        type="button"
        class="flex flex-1 items-center gap-2 rounded-md px-2 py-1.5 text-sm text-gray-500 hover:bg-gray-200/60 hover:text-gray-900"
        @click="startAdd"
      >
        <Plus class="size-4" aria-hidden="true" /> New list
      </button>
      <RouterLink
        to="/settings"
        class="rounded-md p-1.5 text-gray-500 hover:bg-gray-200/60 hover:text-gray-900"
        aria-label="Settings"
        title="Settings"
      >
        <Settings class="size-4" aria-hidden="true" />
      </RouterLink>
    </div>
  </nav>
</template>
