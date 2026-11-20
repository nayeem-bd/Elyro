import { Events } from '@wailsio/runtime'
import { nextTick, onBeforeUnmount, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { Task } from '../bindings/elyro/internal/todo'
import { showError } from './errors'
import { useLists } from './stores/lists'
import { useTasks } from './stores/tasks'

// Keyboard control. ⌘N / ⇧⌘N / ⌘F come from the native menu (main.go) as "menu" events;
// the task keys are handled here because a native accelerator on Space or ⌘⌫ would fire while typing.
//   Esc close detail   ↑/↓ select   Space complete   Return edit title   ⌘⌫ delete   ⌘↑/⌘↓ reorder (in a list)
export function useShortcuts() {
  const tasks = useTasks()
  const lists = useLists()
  const route = useRoute()
  const router = useRouter()

  const focus = async (id: string) => {
    await nextTick()
    document.getElementById(id)?.focus()
  }
  function select(t: Task | undefined) {
    if (!t) return
    tasks.select(t)
    document.getElementById(`task-${t.id}`)?.scrollIntoView({ block: 'nearest' })
  }

  const commands: Record<string, () => unknown> = {
    async newTask() {
      if (!route.path.startsWith('/list/')) await router.push(`/list/${lists.inbox?.id}`)
      await focus('quick-add')
    },
    find: () => focus('search'),
  }
  // Listeners here run outside Vue's handlers, so failures are routed to the error banner by hand.
  const run = (fn?: () => unknown) => fn && Promise.resolve().then(fn).catch(showError)
  const off = Events.On('menu', (e: { data: string }) => run(commands[e.data]))

  function onKey(e: KeyboardEvent) {
    if (e.defaultPrevented || e.altKey || e.ctrlKey) return
    const act = (fn: () => unknown) => {
      e.preventDefault()
      run(fn)
    }
    if (e.key === 'Escape') return tasks.selected && act(() => tasks.select(null)) // works from inside fields too

    const el = e.target as HTMLElement
    const typing =
      el.isContentEditable ||
      ['TEXTAREA', 'SELECT'].includes(el.tagName) ||
      (el.tagName === 'INPUT' && (el as HTMLInputElement).type !== 'checkbox')
    if (typing) return

    const list = tasks.tasks
    const sel = tasks.selected
    const i = sel ? list.findIndex((t) => t.id === sel.id) : -1
    const key = (e.metaKey ? '⌘' : '') + e.key
    if (key === 'ArrowDown') return act(() => select(list[i + 1] ?? list[i]))
    if (key === 'ArrowUp') return act(() => select(i > 0 ? list[i - 1] : list[0]))
    if (!sel || i < 0) return

    const inList = route.path.startsWith('/list/')
    switch (key) {
      case ' ':
        if (el.tagName === 'INPUT' || el.tagName === 'BUTTON') return // native activation already toggles
        return act(() => tasks.setDone(sel.id, !sel.completedAt))
      case 'Enter':
        if (el.tagName === 'BUTTON' || el.tagName === 'A') return
        return act(() => focus('detail-title'))
      case '⌘Backspace':
        return act(() => tasks.remove(sel.id))
      case '⌘ArrowUp':
        if (inList && i > 0) act(() => tasks.move(sel.id, sel.listId, list[i - 1].id))
        return
      case '⌘ArrowDown':
        if (inList && i < list.length - 1) act(() => tasks.move(sel.id, sel.listId, list[i + 2]?.id ?? 0))
        return
    }
  }

  onMounted(() => window.addEventListener('keydown', onKey))
  onBeforeUnmount(() => {
    window.removeEventListener('keydown', onKey)
    off()
  })
}
