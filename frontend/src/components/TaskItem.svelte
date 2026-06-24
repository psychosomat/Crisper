<script>
import { createEventDispatcher } from 'svelte'

const STATUS = { pending: 0, processing: 1, paused: 2, done: 3, error: 4 }

export let task = {}
export let progress = {}

const dispatch = createEventDispatcher()

$: pct = progress.progress ? Math.round(progress.progress * 100) : (task.status === STATUS.done ? 100 : 0)
$: phase = progress.phase || ''
$: isProcessing = task.status === STATUS.processing
$: isDone = task.status === STATUS.done
$: isError = task.status === STATUS.error
$: isPaused = task.status === STATUS.paused
$: isPending = task.status === STATUS.pending
$: canRemove = isPending || isError || isDone || isPaused

const statusClasses = ['border-app-border', 'border-accent/40', 'border-app-border', 'border-success/30', 'border-danger/30']
$: borderClass = statusClasses[task.status] || 'border-app-border'

$: barClass = isDone ? 'bg-success' : isError ? 'bg-danger' : 'bg-accent'

const phaseLabels = {
  extracting: 'Extracting audio',
  reading: 'Reading WAV',
  'detecting speech': 'Detecting speech',
  'labeling speakers': 'Labeling speakers',
  transcribing: 'Transcribing',
  filtering: 'Filtering silence',
  stabilizing: 'Stabilizing output',
  assembling: 'Assembling text',
  saving: 'Saving file',
  done: 'Done',
}

$: phaseText = phaseLabels[phase] || phase

const statusIcons = ['', '', '\u23F8', '\u2713', '\u2717']
$: statusIcon = statusIcons[task.status] || ''

$: iconColor = isDone ? 'text-success' : isError ? 'text-danger' : 'text-text-muted'
</script>

<div class="border {borderClass} bg-app-surface px-4 py-3 flex items-center gap-3">
  <span class="font-mono text-xs {iconColor} w-5 text-center flex-shrink-0">{statusIcon}</span>

  <div class="flex-1 min-w-0">
    <div class="font-mono text-xs text-text-primary truncate">{task.file_name}</div>

    {#if isProcessing || isDone || isError}
      <div class="mt-1.5 h-1 bg-app-bg overflow-hidden">
        <div class="h-full {barClass} transition-all duration-300" style="width: {pct}%"></div>
      </div>
      <div class="mt-1 flex justify-between text-[10px]">
        <span class="font-mono text-text-muted">{phaseText}</span>
        {#if isProcessing}
          <span class="font-mono text-text-secondary tabular-nums">{pct}%</span>
        {/if}
      </div>
    {:else if isPaused}
      <div class="mt-1 text-[11px] font-mono text-text-muted">Paused</div>
    {:else}
      <div class="mt-1 text-[11px] font-mono text-text-muted">Pending</div>
    {/if}

    {#if isError && task.error_msg}
      <div class="mt-1 text-[11px] font-mono text-danger truncate">{task.error_msg}</div>
    {/if}
  </div>

  {#if canRemove}
    <button
      class="font-mono text-text-muted hover:text-danger transition-colors px-1 text-sm flex-shrink-0"
      on:click={() => dispatch('remove', { id: task.id })}
    >&times;</button>
  {/if}
</div>
