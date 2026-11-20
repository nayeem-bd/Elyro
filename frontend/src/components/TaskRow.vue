<script setup lang="ts">
import { Trash2 } from '@lucide/vue'
import type { Task } from '../../bindings/elyro/internal/todo'
import { useTasks } from '../stores/tasks'
import InlineEdit from './InlineEdit.vue'

defineProps<{ task: Task }>()
const tasks = useTasks()
</script>

<template>
  <li class="group flex items-center gap-3 rounded-md px-2 py-1.5 hover:bg-gray-50">
    <input
      type="checkbox"
      class="size-4 shrink-0 accent-blue-500"
      :checked="!!task.completedAt"
      :aria-label="`Mark “${task.title}” ${task.completedAt ? 'not done' : 'done'}`"
      @change="tasks.setDone(task.id, !task.completedAt)"
    />
    <InlineEdit
      class="min-w-0 flex-1 text-sm"
      :class="task.completedAt ? 'text-gray-400 line-through' : 'text-gray-900'"
      :value="task.title"
      label="Task title"
      @save="(v) => tasks.rename(task.id, v)"
    />
    <button
      type="button"
      class="rounded p-1 text-gray-400 opacity-0 group-hover:opacity-100 hover:text-red-600 focus:opacity-100"
      :aria-label="`Delete “${task.title}”`"
      @click="tasks.remove(task.id)"
    >
      <Trash2 class="size-4" aria-hidden="true" />
    </button>
  </li>
</template>
