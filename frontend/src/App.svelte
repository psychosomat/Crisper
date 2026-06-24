<script>
import { onMount } from 'svelte'
import { EventsOn } from '../wailsjs/runtime/runtime.js'
import { GetSettings, GetAvailableModels, GetRecommendedModel, AddFiles, RemoveFile, GetQueue, StartQueue, PauseQueue, CancelQueue, SelectFiles, IsProcessing, SetWindowFrame } from '../wailsjs/go/main/App.js'

import Header from './components/Header.svelte'
import Titlebar from './components/Titlebar.svelte'
import DropZone from './components/DropZone.svelte'
import QueuePanel from './components/QueuePanel.svelte'
import StartDialog from './components/StartDialog.svelte'
import SettingsModal from './components/SettingsModal.svelte'

let settings = {}
let models = []
let recommended = null
let queueTasks = []
let processing = false
let taskProgress = {}
let statusMsg = ''
let showStartDialog = false
let showSettings = false
let batchResult = null

onMount(async () => {
  settings = await GetSettings()
  models = await GetAvailableModels()
  recommended = await GetRecommendedModel()
  queueTasks = await GetQueue()
  processing = await IsProcessing()

  if (settings.window_frame === 'none') {
    await SetWindowFrame('none')
  }

  EventsOn('queue-update', async () => {
    queueTasks = await GetQueue()
    processing = await IsProcessing()
  })

  EventsOn('download-progress', (data) => {
    taskProgress[data.model] = data
    taskProgress = taskProgress
  })

  EventsOn('task-progress', (data) => {
    taskProgress[data.id] = data
    taskProgress = taskProgress
  })

  EventsOn('batch-complete', (data) => {
    batchResult = data
  })
})

async function handleSelectFiles() {
  const files = await SelectFiles()
  if (files?.length) {
    await AddFiles(files)
    queueTasks = await GetQueue()
  }
}

function handleStartClick() {
  if (queueTasks.length === 0) return
  showStartDialog = true
}

async function handleStartConfirm(e) {
  showStartDialog = false
  statusMsg = ''
  try {
    await StartQueue(e.detail.model, e.detail.outputDir || '')
    processing = true
  } catch (err) {
    statusMsg = String(err)
    processing = false
  }
}

async function handlePause() {
  await PauseQueue()
  processing = false
}

async function handleCancel() {
  await CancelQueue()
  processing = false
  queueTasks = await GetQueue()
}

async function handleRemove(e) {
  await RemoveFile(e.detail.id)
  queueTasks = await GetQueue()
}

async function handleClearErrorTasks() {
  for (const t of queueTasks) {
    if (t.status === 4 || t.status === 2) {
      await RemoveFile(t.id)
    }
  }
  queueTasks = await GetQueue()
  batchResult = null
}

async function handleDismissBatch() {
  batchResult = null
  await handleClearErrorTasks()
}

function handleDropFiles(e) {
  AddFiles(e.detail.paths).then(async () => {
    queueTasks = await GetQueue()
  })
}
</script>

{#if settings.window_frame === 'custom'}
  <Titlebar />
{/if}

<main class="max-w-app mx-auto px-5 pb-16">

  <Header {processing} />

  {#if showStartDialog}
    <StartDialog
      {models}
      {recommended}
      on:close={() => showStartDialog = false}
      on:confirm={handleStartConfirm}
    />
  {/if}

  {#if showSettings}
    <SettingsModal
      {settings}
      {models}
      on:close={async () => {
        showSettings = false
        settings = await GetSettings()
        await SetWindowFrame(settings.window_frame)
      }}
    />
  {/if}

  <DropZone
    disabled={processing}
    on:browse={handleSelectFiles}
    on:files={handleDropFiles}
  />

  {#if statusMsg}
    <div class="mt-4 px-4 py-3 border border-danger/30 bg-danger/5">
      <span class="font-mono text-xs text-danger">{statusMsg}</span>
    </div>
  {/if}

  {#if batchResult}
    <div class="mt-4 px-4 py-3 border border-success/30 bg-success/5 flex items-center justify-between">
      <span class="font-mono text-xs text-success">
        {batchResult.processed} file{batchResult.processed !== 1 ? 's' : ''} processed
        {#if batchResult.errors > 0}
          &middot; {batchResult.errors} error{batchResult.errors !== 1 ? 's' : ''}
        {/if}
      </span>
      <button
        class="font-mono text-[10px] text-text-muted hover:text-text-secondary transition-colors"
        on:click={handleDismissBatch}
      >Dismiss</button>
    </div>
  {/if}

  {#if queueTasks.length > 0}
    <QueuePanel
      tasks={queueTasks}
      {taskProgress}
      {processing}
      on:start={handleStartClick}
      on:pause={handlePause}
      on:cancel={handleCancel}
      on:remove={handleRemove}
      on:clearErrors={handleClearErrorTasks}
    />
  {/if}

  <div class="mt-12 flex justify-center">
    <button
      class="font-mono text-[10px] uppercase tracking-wider text-text-muted hover:text-text-secondary transition-colors"
      on:click={() => showSettings = true}
    >Settings</button>
  </div>

</main>
