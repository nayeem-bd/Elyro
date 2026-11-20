import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { ListService, type List } from '../../bindings/elyro/internal/todo'

export const useLists = defineStore('lists', () => {
  const lists = ref<List[]>([])
  const inbox = computed(() => lists.value.find((l) => l.isInbox))

  async function load() {
    lists.value = (await ListService.All()) ?? []
  }
  async function create(name: string) {
    const list = await ListService.Create(name)
    lists.value.push(list)
    return list
  }
  async function rename(id: number, name: string) {
    const list = await ListService.Rename(id, name)
    lists.value = lists.value.map((l) => (l.id === id ? list : l))
  }
  async function remove(id: number) {
    await ListService.Delete(id)
    lists.value = lists.value.filter((l) => l.id !== id)
  }

  return { lists, inbox, load, create, rename, remove }
})
