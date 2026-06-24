<script>
import { createEventDispatcher } from 'svelte'
import TaskItem from './TaskItem.svelte'

const STATUS = { pending: 0, processing: 1, paused: 2, done: 3, error: 4 }

export let tasks = []
export let taskProgress = {}
export let processing = false

const dispatch = createEventDispatcher()

$: doneCount = tasks.filter(t => t.status === STATUS.done).length
$: errorCount = tasks.filter(t => t.status === STATUS.error).length
$: pausedCount = tasks.filter(t => t.status === STATUS.paused).length
$: pendingCount = tasks.filter(t => t.status === STATUS.pending).length
$: processingCount = tasks.filter(t => t.status === STATUS.processing).length
$: hasErrors = errorCount + pausedCount > 0
$: hasPending = pendingCount > 0
</script>

<div class="mt-8">
  <div class="flex items-center justify-between mb-3">
    <span class="font-mono text-[11px] text-text-muted uppercase tracking-wider">
      Queue ({tasks.length})
      {#if doneCount > 0}
        <span class="text-success ml-1">{doneCount} done</span>
      {/if}
      {#if errorCount > 0}
        <span class="text-danger ml-1">{errorCount} failed</span>
      {/if}
    </span>
    <div class="flex gap-2">
      {#if hasPending}
        {#if !processing}
          <button
            class="font-mono text-xs px-4 py-1.5 border border-accent/40 text-accent hover:bg-accent/10 transition-colors"
            on:click={() => dispatch('start')}
          >Start</button>
        {:else}
          <button
            class="font-mono text-xs px-4 py-1.5 border border-app-border text-text-secondary hover:border-text-muted transition-colors"
            on:click={() => dispatch('pause')}
          >Pause</button>
        {/if}
      {/if}
      {#if processing || hasPending}
        <button
          class="font-mono text-xs px-4 py-1.5 border border-app-border text-text-muted hover:border-danger/40 hover:text-danger transition-colors"
          on:click={() => dispatch('cancel')}
        >Cancel</button>
      {/if}
      {#if !processing && hasErrors}
        <button
          class="font-mono text-xs px-4 py-1.5 border border-app-border text-text-muted hover:border-text-muted transition-colors"
          on:click={() => dispatch('clearErrors')}
        >Clear failed</button>
      {/if}
    </div>
  </div>

  <div class="flex flex-col gap-px">
    {#each tasks as task (task.id)}
      <TaskItem {task} progress={taskProgress[task.id] || {}}
        on:remove={e => dispatch('remove', e.detail)}
      />
    {/each}
  </div>
</div>
