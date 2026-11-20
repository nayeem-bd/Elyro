<script setup lang="ts">
import { useRoute } from 'vue-router'

// shortcut: one placeholder for every route, replaced by real views in Phases 2–3.
const route = useRoute()
</script>

<template>
  <section class="p-8">
    <h1 class="text-2xl font-semibold text-gray-900">{{ route.meta.title }}</h1>
    <p class="mt-2 text-sm text-gray-500">Nothing here yet.</p>
  </section>
</template>
