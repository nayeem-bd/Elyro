import { defineStore } from 'pinia'
import { ref } from 'vue'
import { TaskService, type Task } from '../../bindings/elyro/internal/todo'

// Tasks of the list currently on screen. Go owns ordering, so changes that move a task reload.
export const useTasks = defineStore('tasks', () => {
  const tasks = ref<Task[]>([])
  let listId = 0

  async function load(id: number) {
    listId = id
    const fresh = (await TaskService.ByList(id)) ?? []
    if (listId === id) tasks.value = fresh // drop a late reply after switching lists
  }
  async function add(title: string) {
    await TaskService.Create(listId, title)
    await load(listId)
  }
  async function setDone(id: number, done: boolean) {
    await TaskService.SetDone(id, done)
    await load(listId)
  }
  async function rename(id: number, title: string) {
    const task = await TaskService.Update(id, { title })
    tasks.value = tasks.value.map((t) => (t.id === id ? task : t))
  }
  async function remove(id: number) {
    await TaskService.Delete(id)
    tasks.value = tasks.value.filter((t) => t.id !== id)
  }

  return { tasks, load, add, setDone, rename, remove }
})
