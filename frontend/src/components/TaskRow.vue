<script setup lang="ts">
import { Flag, Trash2 } from '@lucide/vue'
import { computed } from 'vue'
import type { Task } from '../../bindings/elyro/internal/todo'
import { startDrag, TASK } from '../dnd'
import { dueLabel, priorities, priorityColor } from '../format'
import { useLists } from '../stores/lists'
import { useTasks } from '../stores/tasks'
import InlineEdit from './InlineEdit.vue'

const props = defineProps<{ task: Task; showList?: boolean }>()
const tasks = useTasks()
const lists = useLists()
const isSelected = computed(() => tasks.selected?.id === props.task.id)
const due = computed(() => props.task.dueDate && dueLabel(props.task.dueDate))
const listName = computed(() => props.showList && lists.lists.find((l) => l.id === props.task.listId)?.name)
</script>

<template>
  <li
    :id="`task-${task.id}`"
    :aria-current="isSelected || undefined"
    class="group flex items-start gap-3 rounded-md px-2 py-1.5"
    :class="isSelected ? 'bg-blue-50' : 'hover:bg-gray-50'"
    draggable="true"
    @dragstart="startDrag($event, TASK, task.id)"
    @click="tasks.select(task)"
  >
    <input
      type="checkbox"
      class="mt-0.5 size-4 shrink-0 accent-blue-500"
      :checked="!!task.completedAt"
      :aria-label="`Mark “${task.title}” ${task.completedAt ? 'not done' : 'done'}`"
      @click.stop
      @change="tasks.setDone(task.id, !task.completedAt)"
    />
    <div class="min-w-0 flex-1">
      <InlineEdit
        class="text-sm"
        :class="task.completedAt ? 'text-gray-400 line-through' : 'text-gray-900'"
        :value="task.title"
        label="Task title"
        @save="(v) => tasks.update(task.id, { ...task, title: v })"
      />
      <div v-if="due || listName || task.priority" class="mt-0.5 flex items-center gap-2 px-1 text-xs text-gray-500">
        <span v-if="due" :class="due.overdue && !task.completedAt && 'text-red-600'">{{ due.text }}</span>
        <span v-if="listName" class="truncate">{{ listName }}</span>
        <Flag
          v-if="task.priority"
          class="size-3"
          :class="priorityColor[task.priority]"
          :aria-label="`${priorities[task.priority]} priority`"
        />
      </div>
    </div>
    <button
      type="button"
      class="rounded p-1 text-gray-400 opacity-0 group-hover:opacity-100 hover:text-red-600 focus:opacity-100"
      :aria-label="`Delete “${task.title}”`"
      @click.stop="tasks.remove(task.id)"
    >
      <Trash2 class="size-4" aria-hidden="true" />
    </button>
  </li>
</template>
