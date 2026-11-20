<script setup lang="ts">
// shortcut: one placeholder for the smart views, replaced by real views in Phase 3.
defineProps<{ title: string }>()
</script>

<template>
  <section class="p-8">
    <h1 class="text-2xl font-semibold text-gray-900">{{ title }}</h1>
    <p class="mt-2 text-sm text-gray-500">Nothing here yet.</p>
  </section>
</template>
