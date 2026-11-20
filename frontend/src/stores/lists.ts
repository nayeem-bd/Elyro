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
  // Swaps in the list a change returned.
  async function replace(p: Promise<List>) {
    const list = await p
    lists.value = lists.value.map((l) => (l.id === list.id ? list : l))
  }
  const rename = (id: number, name: string) => replace(ListService.Rename(id, name))
  const showInMenuBar = (id: number, on: boolean) => replace(ListService.ShowInMenuBar(id, on))
  async function move(id: number, beforeId = 0) {
    await ListService.Move(id, beforeId)
    await load()
  }
  async function remove(id: number) {
    await ListService.Delete(id)
    lists.value = lists.value.filter((l) => l.id !== id)
  }

  return { lists, inbox, load, create, rename, showInMenuBar, move, remove }
})
