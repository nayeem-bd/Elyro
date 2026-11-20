<script setup lang="ts">
import { X } from '@lucide/vue'
import Sidebar from './components/Sidebar.vue'
import { lastError } from './errors'
</script>

<template>
  <div class="flex h-full bg-white">
    <Sidebar />
    <main class="flex-1 overflow-y-auto">
      <div class="h-13" style="--wails-draggable: drag" />
      <div
        v-if="lastError"
        role="alert"
        class="mx-8 mb-4 flex items-start gap-2 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800"
      >
        <span class="flex-1">{{ lastError }}</span>
        <button type="button" aria-label="Dismiss error" class="rounded p-0.5 hover:bg-red-100" @click="lastError = ''">
          <X class="size-4" aria-hidden="true" />
        </button>
      </div>
      <RouterView />
    </main>
  </div>
</template>
