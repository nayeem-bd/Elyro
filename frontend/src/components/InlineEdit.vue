<script setup lang="ts">
import { nextTick, ref } from 'vue'

// Text that turns into an input on double-click (or Enter). Saves on Enter/blur, cancels on Escape.
const props = defineProps<{ value: string; label: string }>()
const emit = defineEmits<{ save: [value: string] }>()

const editing = ref(false)
const draft = ref('')
const input = ref<HTMLInputElement>()

async function start() {
  draft.value = props.value
  editing.value = true
  await nextTick()
  input.value?.select()
}
function commit() {
  if (!editing.value) return
  editing.value = false
  const v = draft.value.trim()
  if (v && v !== props.value) emit('save', v)
}
</script>

<template>
  <input
    v-if="editing"
    ref="input"
    v-model="draft"
    :aria-label="label"
    maxlength="500"
    class="w-full rounded border border-blue-400 bg-white px-1 outline-none"
    @keydown.enter.prevent="commit"
    @keydown.esc.prevent="editing = false"
    @blur="commit"
  />
  <span
    v-else
    tabindex="0"
    title="Double-click to edit"
    class="block truncate rounded px-1 focus:outline-2 focus:outline-blue-400"
    @dblclick="start"
    @keydown.enter.prevent="start"
  >{{ value }}</span>
</template>
