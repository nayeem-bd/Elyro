<script setup lang="ts">
import { useLists } from '../stores/lists'
import { useTasks } from '../stores/tasks'

const lists = useLists()
useTasks().show(async () => []) // no tasks here: clears the previous view's selection and keys
</script>

<template>
  <section class="mx-auto max-w-2xl px-8">
    <h1 class="px-1 text-2xl font-semibold text-gray-900">Settings</h1>

    <fieldset class="mt-6 px-1">
      <legend class="text-sm font-medium text-gray-900">Menu bar</legend>
      <p class="mt-1 text-sm text-gray-500">Lists to show in the ✓ menu bar menu, with their open task count.</p>
      <label v-for="l in lists.lists" :key="l.id" class="mt-2 flex items-center gap-2 text-sm text-gray-700">
        <input type="checkbox" :checked="l.inMenuBar" @change="lists.showInMenuBar(l.id, !l.inMenuBar)" />
        {{ l.name }}
      </label>
    </fieldset>
  </section>
</template>
