import { ref } from 'vue'

// Last failed action, shown in a banner so a failed save is never silent.
export const lastError = ref('')

export function showError(e: unknown) {
  console.error(e)
  lastError.value = e instanceof Error ? e.message : String(e)
}
