// HTML5 drag-and-drop helpers. Custom types keep a dragged task and a dragged list apart;
// during dragover only the types are readable, the ids arrive on drop.
export const TASK = 'application/x-elyro-task'
export const LIST = 'application/x-elyro-list'

export function startDrag(e: DragEvent, type: string, id: number) {
  e.dataTransfer?.setData(type, String(id))
  if (e.dataTransfer) e.dataTransfer.effectAllowed = 'move'
}
export const carries = (e: DragEvent, type: string) => !!e.dataTransfer?.types.includes(type)
export const droppedId = (e: DragEvent, type: string) => Number(e.dataTransfer?.getData(type)) || 0
// dragleave also fires when moving onto a child; true only when the pointer really left the element.
export const leftTarget = (e: DragEvent) => !(e.currentTarget as Element).contains(e.relatedTarget as Node)
