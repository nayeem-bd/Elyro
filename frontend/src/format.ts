const pad = (n: number) => String(n).padStart(2, '0')
export const isoDay = (d: Date) => `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`

// Built once: creating an Intl formatter per row is the costly part of formatting.
const sameYear = new Intl.DateTimeFormat(undefined, { weekday: 'short', month: 'short', day: 'numeric' })
const otherYear = new Intl.DateTimeFormat(undefined, { weekday: 'short', month: 'short', day: 'numeric', year: 'numeric' })

// Label for a 'YYYY-MM-DD' due date (the form Go stores), relative to the local today.
export function dueLabel(due: string) {
  const now = new Date()
  const today = isoDay(now)
  const overdue = due < today // ISO days compare correctly as strings
  if (due === today) return { text: 'Today', overdue }
  if (due === isoDay(new Date(now.getFullYear(), now.getMonth(), now.getDate() + 1))) return { text: 'Tomorrow', overdue }
  const [y, m, d] = due.split('-').map(Number)
  return { text: (y === now.getFullYear() ? sameYear : otherYear).format(new Date(y, m - 1, d)), overdue }
}

export const priorities = ['None', 'Low', 'Medium', 'High']
export const priorityColor = ['', 'text-blue-500', 'text-amber-500', 'text-red-500']
