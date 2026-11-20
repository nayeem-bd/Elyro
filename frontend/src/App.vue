<script setup lang="ts">
import Sidebar from './components/Sidebar.vue'
</script>

<template>
  <div class="flex h-full bg-white">
    <Sidebar />
    <main class="flex-1 overflow-y-auto">
      <div class="h-13" style="--wails-draggable: drag" />
      <RouterView />
    </main>
  </div>
</template>
