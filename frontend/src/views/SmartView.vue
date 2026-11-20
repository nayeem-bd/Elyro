<script setup lang="ts">
import TaskList from '../components/TaskList.vue'
import { useTasks, type Fetch } from '../stores/tasks'

// Today, Upcoming, Completed and Search: tasks from every list, ordered by Go.
// Remounted per route (App keys RouterView by path), so it loads once.
const props = defineProps<{ title: string; fetch: Fetch; empty: string }>()
useTasks().show(props.fetch)
</script>

<template>
  <section class="mx-auto max-w-2xl px-8">
    <h1 class="px-1 text-2xl font-semibold text-gray-900">{{ title }}</h1>
    <div class="mt-4"><TaskList show-list :empty="empty" /></div>
  </section>
</template>
