<script setup lang="ts">
import { Trash2, X } from '@lucide/vue'
import { ref, watch } from 'vue'
import type { Task, TaskInput } from '../../bindings/elyro/internal/todo'
import { priorities } from '../format'
import { useLists } from '../stores/lists'
import { useTasks } from '../stores/tasks'

const props = defineProps<{ task: Task }>()
const tasks = useTasks()
const lists = useLists()

const draft = ref<TaskInput>({ title: '', notes: '', dueDate: null, priority: 0 })
// Reset the form whenever the saved task changes (opened, saved, renamed from its row…).
// Saves fire on blur/change, so the field being typed in is never the one being reset.
watch(
  () => props.task,
  () => {
    const { title, notes, dueDate, priority } = props.task
    draft.value = { title, notes, dueDate, priority }
  },
  { immediate: true },
)

async function save() {
  const d = { ...draft.value, title: draft.value.title.trim(), dueDate: draft.value.dueDate || null }
  const t = props.task
  if (!d.title) {
    draft.value.title = t.title // a task can't lose its title
    return
  }
  if (d.title === t.title && d.notes === t.notes && d.dueDate === t.dueDate && d.priority === t.priority) return
  await tasks.update(t.id, d)
}

const field = 'mt-1 w-full rounded-md border border-gray-200 bg-white px-2 py-1.5 text-sm outline-none focus:border-blue-400'
const label = 'block text-xs font-medium text-gray-500'
</script>

<template>
  <aside class="flex w-80 shrink-0 flex-col border-l border-gray-200 bg-gray-50" aria-label="Task details">
    <div class="flex h-13 shrink-0 items-center justify-end px-3" style="--wails-draggable: drag">
      <button type="button" class="rounded p-1 text-gray-500 hover:bg-gray-200" aria-label="Close details" @click="tasks.select(null)">
        <X class="size-4" aria-hidden="true" />
      </button>
    </div>

    <form class="flex-1 space-y-4 overflow-y-auto px-4 pb-4" @submit.prevent="save">
      <label :class="label">
        Title
        <input id="detail-title" v-model="draft.title" maxlength="500" :class="field" @blur="save" />
      </label>

      <label :class="label">
        List
        <select
          :value="task.listId"
          :class="field"
          @change="tasks.move(task.id, Number(($event.target as HTMLSelectElement).value))"
        >
          <option v-for="l in lists.lists" :key="l.id" :value="l.id">{{ l.name }}</option>
        </select>
      </label>

      <div class="flex gap-2">
        <label :class="[label, 'flex-1']">
          Due date
          <input v-model="draft.dueDate" type="date" :class="field" @change="save" />
        </label>
        <label :class="[label, 'w-28']">
          Priority
          <select v-model.number="draft.priority" :class="field" @change="save">
            <option v-for="(p, i) in priorities" :key="p" :value="i">{{ p }}</option>
          </select>
        </label>
      </div>
      <button
        v-if="draft.dueDate"
        type="button"
        class="-mt-2 text-xs text-blue-600 hover:underline"
        @click="(draft.dueDate = null), save()"
      >
        Clear due date
      </button>

      <label :class="label">
        Notes
        <textarea v-model="draft.notes" rows="8" maxlength="10000" :class="[field, 'resize-none']" @blur="save" />
      </label>
    </form>

    <div class="border-t border-gray-200 p-3">
      <button
        type="button"
        class="flex items-center gap-2 rounded-md px-2 py-1.5 text-sm text-red-600 hover:bg-red-50"
        @click="tasks.remove(task.id)"
      >
        <Trash2 class="size-4" aria-hidden="true" /> Delete task
      </button>
    </div>
  </aside>
</template>
