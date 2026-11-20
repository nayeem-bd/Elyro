import { defineStore } from 'pinia'
import { ref } from 'vue'
import { TaskService, type Task, type TaskInput } from '../../bindings/elyro/internal/todo'

export type Fetch = () => Promise<Task[] | null>

// Tasks of the view on screen (a list, Today, Search…). Go decides which tasks a view holds
// and in what order, so every change reloads the view rather than patching it locally.
export const useTasks = defineStore('tasks', () => {
  const tasks = ref<Task[]>([])
  const selected = ref<Task | null>(null) // kept apart from tasks: it stays open when an edit moves the task out of the view
  const changed = ref(0) // bumped when task counts may be stale (every write, a new day): the sidebar refetches
  let fetch: Fetch = async () => []
  let seq = 0

  async function reload() {
    const n = ++seq
    const fresh = (await fetch()) ?? []
    if (n === seq) tasks.value = fresh // drop a late reply from an older view or reload
  }
  async function show(f: Fetch) {
    fetch = f
    selected.value = null
    await reload()
  }
  // Runs a change, keeps the detail panel in sync with the returned task, then reloads the view.
  async function change(p: Promise<Task>) {
    const t = await p
    changed.value++
    if (selected.value?.id === t.id) selected.value = t
    await reload()
  }

  return {
    tasks,
    selected,
    changed,
    show,
    reload,
    select: (t: Task | null) => (selected.value = t),
    add: (listId: number, title: string) => change(TaskService.Create(listId, title)),
    update: (id: number, input: TaskInput) => change(TaskService.Update(id, input)),
    setDone: (id: number, done: boolean) => change(TaskService.SetDone(id, done)),
    move: (id: number, listId: number, beforeId = 0) => change(TaskService.Move(id, listId, beforeId)),
    async remove(id: number) {
      await TaskService.Delete(id)
      changed.value++
      if (selected.value?.id === id) selected.value = null
      await reload()
    },
  }
})
