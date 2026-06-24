<script>
import { createEventDispatcher } from 'svelte'

export let disabled = false

const dispatch = createEventDispatcher()

function handleDragOver(e) {
  e.preventDefault()
}

function handleDrop(e) {
  e.preventDefault()
  const paths = []
  for (const item of e.dataTransfer.items) {
    if (item.kind === 'file') {
      const f = item.getAsFile()
      if (f.path) paths.push(f.path)
      else if (f.name) paths.push(f.name)
    }
  }
  if (paths.length) dispatch('files', { paths })
}

function handleClick() {
  dispatch('browse')
}
</script>

{#if !disabled}
  <div
    class="border border-dashed border-app-border hover:border-text-muted transition-colors p-12 text-center cursor-pointer"
    on:dragover={handleDragOver}
    on:drop={handleDrop}
    on:click={handleClick}
    on:keydown={(e) => e.key === 'Enter' && handleClick()}
    role="button"
    tabindex="0"
  >
    <p class="text-text-secondary text-sm font-ui">Drop video files here</p>
    <p class="text-text-muted text-xs mt-2 font-ui">or click to browse</p>
  </div>
{:else}
  <div class="border border-app-border p-12 text-center opacity-40">
    <p class="text-text-muted text-sm font-ui">Processing — adding files blocked</p>
  </div>
{/if}
