<script setup lang="ts">
import { ref } from 'vue'
import { carries, droppedId, leftTarget, TASK } from '../dnd'
import { useTasks } from '../stores/tasks'
import TaskRow from './TaskRow.vue'

// listId set = a real list, so rows can be reordered here. Smart views pass showList instead;
// their rows can still be dragged onto a sidebar list.
const props = defineProps<{ listId?: number; showList?: boolean; empty: string }>()
const tasks = useTasks()
const over = ref<number | null>(null) // row the drag hovers (inserts before it); 0 = end of list

function dragOver(e: DragEvent, beforeId: number) {
  if (!props.listId || !carries(e, TASK)) return
  e.preventDefault() // allows the drop
  over.value = beforeId
}
async function drop(e: DragEvent, beforeId: number) {
  over.value = null
  const id = droppedId(e, TASK)
  if (props.listId && id && id !== beforeId) await tasks.move(id, props.listId, beforeId)
}
</script>

<template>
  <ul
    v-if="tasks.tasks.length"
    class="border-b-2 pb-6"
    :class="over === 0 ? 'border-blue-400' : 'border-transparent'"
    @dragover="dragOver($event, 0)"
    @dragleave="leftTarget($event) && (over = null)"
    @drop.prevent="drop($event, 0)"
  >
    <TaskRow
      v-for="t in tasks.tasks"
      :key="t.id"
      :task="t"
      :show-list="showList"
      class="border-t-2"
      :class="over === t.id ? 'border-blue-400' : 'border-transparent'"
      @dragover.stop="dragOver($event, t.id)"
      @drop.stop.prevent="drop($event, t.id)"
    />
  </ul>
  <p v-else class="mt-6 text-center text-sm text-gray-400">{{ empty }}</p>
</template>
