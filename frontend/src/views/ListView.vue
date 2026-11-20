<script setup lang="ts">
import { Dialogs } from '@wailsio/runtime'
import { Trash2 } from '@lucide/vue'
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import InlineEdit from '../components/InlineEdit.vue'
import TaskRow from '../components/TaskRow.vue'
import { useLists } from '../stores/lists'
import { useTasks } from '../stores/tasks'

const route = useRoute()
const router = useRouter()
const lists = useLists()
const tasks = useTasks()

const id = computed(() => Number(route.params.id))
const list = computed(() => lists.lists.find((l) => l.id === id.value))
const draft = ref('')

watch(id, (v) => tasks.load(v), { immediate: true })

async function add() {
  const title = draft.value.trim()
  if (!title) return
  await tasks.add(title)
  draft.value = ''
}

async function remove() {
  const l = list.value
  if (!l) return
  const n = tasks.tasks.length
  const choice = await Dialogs.Question({
    Title: `Delete “${l.name}”?`,
    Message: n ? `Its ${n} task${n === 1 ? '' : 's'} will be deleted too. This can’t be undone.` : 'This can’t be undone.',
    Buttons: [{ Label: 'Delete' }, { Label: 'Cancel', IsCancel: true, IsDefault: true }],
  })
  if (choice !== 'Delete') return
  await lists.remove(l.id)
  await router.replace('/') // '/' redirects to the Inbox
}
</script>

<template>
  <section v-if="list" class="mx-auto max-w-2xl px-8 pb-8">
    <header class="flex items-center gap-2">
      <InlineEdit
        class="min-w-0 flex-1 text-2xl font-semibold text-gray-900"
        :value="list.name"
        label="List name"
        @save="(v) => lists.rename(list!.id, v)"
      />
      <button
        v-if="!list.isInbox"
        type="button"
        class="rounded p-1.5 text-gray-400 hover:bg-gray-100 hover:text-red-600"
        aria-label="Delete list"
        @click="remove"
      >
        <Trash2 class="size-4" aria-hidden="true" />
      </button>
    </header>

    <form class="mt-4" @submit.prevent="add">
      <input
        v-model="draft"
        maxlength="500"
        placeholder="Add a task, press Enter"
        aria-label="New task"
        class="w-full rounded-md border border-gray-200 px-3 py-2 text-sm outline-none focus:border-blue-400"
      />
    </form>

    <ul v-if="tasks.tasks.length" class="mt-3">
      <TaskRow v-for="t in tasks.tasks" :key="t.id" :task="t" />
    </ul>
    <p v-else class="mt-6 text-center text-sm text-gray-400">No tasks yet.</p>
  </section>
</template>
