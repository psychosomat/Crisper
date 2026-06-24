<script>
import { createEventDispatcher, onMount } from 'svelte'
import { DownloadModel, IsModelDownloaded, SelectOutputDirectory, IsWhisperInstalled, DownloadWhisperCLI } from '../../wailsjs/go/main/App.js'
import { EventsOn } from '../../wailsjs/runtime/runtime.js'

export let models = []
export let recommended = null

const dispatch = createEventDispatcher()

let selectedModel = recommended?.name || ''
let outputDir = ''
let downloading = ''
let downloadProgress = 0
let downloadErr = ''
let modelDownloaded = {}
let browseError = ''

let whisperInstalled = false
let whisperDownloading = false
let whisperProgress = 0
let whisperErr = ''
let starting = false

onMount(async () => {
  whisperInstalled = await IsWhisperInstalled()

  EventsOn('whisper-download-progress', (data) => {
    whisperProgress = data.progress
  })
})

async function checkDownloaded(name) {
  if (!name) return
  if (modelDownloaded[name] !== undefined) return modelDownloaded[name]
  modelDownloaded[name] = await IsModelDownloaded(name)
  modelDownloaded = modelDownloaded
  return modelDownloaded[name]
}

async function downloadModel(name) {
  downloading = name
  downloadProgress = 0
  downloadErr = ''

  EventsOn('download-progress', (data) => {
    if (data.model === name) downloadProgress = data.progress
  })

  try {
    await DownloadModel(name)
    modelDownloaded[name] = true
    modelDownloaded = modelDownloaded
  } catch (e) {
    downloadErr = String(e)
  } finally {
    downloading = ''
  }
}

async function downloadWhisper() {
  whisperDownloading = true
  whisperProgress = 0
  whisperErr = ''
  try {
    await DownloadWhisperCLI()
    whisperInstalled = true
  } catch (e) {
    whisperErr = String(e)
  } finally {
    whisperDownloading = false
  }
}

async function handleBrowse() {
  browseError = ''
  const dir = await SelectOutputDirectory()
  if (dir) outputDir = dir
}

async function handleStart() {
  if (!selectedModel) return
  starting = true
  whisperErr = ''

  if (!whisperInstalled) {
    await downloadWhisper()
    if (!whisperInstalled) {
      starting = false
      return
    }
  }

  dispatch('confirm', { model: selectedModel, outputDir })
}

function handleCancel() {
  dispatch('close')
}

function formatSize(gb) {
  if (gb < 1) return Math.round(gb * 1000) + ' MB'
  return gb.toFixed(1) + ' GB'
}
</script>

<div class="fixed inset-0 z-50 flex items-center justify-center" style="background: rgba(0,0,0,0.7)">
<div class="bg-app-elevated border border-app-border w-full max-w-lg mx-4 p-6">

  <h2 class="font-mono text-xs uppercase tracking-wider text-text-secondary mb-6">Start Transcription</h2>

  {#if whisperDownloading}
    <div class="mb-5 p-3 border border-accent/40 bg-accent/5">
      <div class="h-1 bg-app-bg overflow-hidden mb-1.5">
        <div class="h-full bg-accent transition-all duration-200" style="width: {whisperProgress * 100}%"></div>
      </div>
      <span class="font-mono text-[10px] text-text-muted">Installing whisper-cli... {Math.round(whisperProgress * 100)}%</span>
    </div>
  {:else if !whisperInstalled}
    <div class="mb-5 p-3 border border-accent/40 bg-accent/5 font-mono text-[11px] text-text-secondary">
      whisper-cli not found &mdash; will be auto-installed on start
    </div>
  {/if}

  <div class="mb-5">
    <span class="block font-mono text-[11px] uppercase tracking-wider text-text-muted mb-2">Model</span>
    <div class="flex flex-col gap-px max-h-48 overflow-y-auto">
      {#each models as m (m.name)}
        {@const downloaded = modelDownloaded[m.name]}
        <button
          class="flex items-center justify-between px-3 py-2 text-left border transition-colors
                 {selectedModel === m.name ? 'border-accent/40 bg-accent/5' : 'border-app-border hover:border-app-border-hover'}"
          on:click={async () => {
            selectedModel = m.name
            await checkDownloaded(m.name)
          }}
        >
          <div class="min-w-0">
            <div class="font-mono text-xs text-text-primary">{m.display_name}</div>
            <div class="font-mono text-[10px] text-text-muted mt-0.5">
              {formatSize(m.size_gb)} &middot; {formatSize(m.min_ram_gb)}+ RAM
              {#if recommended?.name === m.name}
                <span class="text-accent ml-1">Recommended</span>
              {/if}
            </div>
          </div>
          <span class="font-mono text-[10px] flex-shrink-0 ml-3
            {downloaded === true ? 'text-success' : downloaded === false ? 'text-text-muted' : ''}">
            {downloaded === true ? 'ready' : downloaded === false ? 'needs download' : ''}
          </span>
        </button>
      {/each}
    </div>
  </div>

  {#if downloadErr}
    <div class="mb-4 px-3 py-2 border border-danger/30 bg-danger/5 font-mono text-[11px] text-danger">{downloadErr}</div>
  {/if}

  {#if selectedModel && modelDownloaded[selectedModel] === false && downloading !== selectedModel}
    <div class="mb-4">
      <button
        class="font-mono text-xs px-4 py-1.5 border border-accent/40 text-accent hover:bg-accent/10 transition-colors w-full"
        on:click={() => downloadModel(selectedModel)}
      >Download {selectedModel}</button>
    </div>
  {/if}

  {#if downloading}
    <div class="mb-4">
      <div class="h-1 bg-app-bg overflow-hidden mb-1">
        <div class="h-full bg-accent transition-all duration-200" style="width: {downloadProgress * 100}%"></div>
      </div>
      <span class="font-mono text-[10px] text-text-muted">{Math.round(downloadProgress * 100)}%</span>
    </div>
  {/if}

  <div class="mb-6">
    <label for="sd-outdir" class="block font-mono text-[11px] uppercase tracking-wider text-text-muted mb-2">Output Directory</label>
    <div class="flex gap-2">
      <input
        id="sd-outdir"
        type="text"
        class="flex-1 bg-app-bg border border-app-border px-3 py-1.5 font-mono text-xs text-text-primary placeholder-text-muted focus:border-accent/40 focus:outline-none"
        placeholder="Same as video file"
        bind:value={outputDir}
        readonly
      />
      <button
        class="font-mono text-xs px-3 py-1.5 border border-app-border text-text-secondary hover:border-text-muted transition-colors flex-shrink-0"
        on:click={handleBrowse}
      >Browse</button>
    </div>
  </div>

  {#if whisperErr}
    <div class="mb-4 px-3 py-2 border border-danger/30 bg-danger/5 font-mono text-[11px] text-danger">{whisperErr}</div>
  {/if}

  <div class="flex justify-end gap-2">
    <button
      class="font-mono text-xs px-4 py-1.5 border border-app-border text-text-muted hover:border-text-muted transition-colors"
      on:click={handleCancel}
      disabled={starting}
    >Cancel</button>
    <button
      class="font-mono text-xs px-4 py-1.5 border border-accent/40 text-accent hover:bg-accent/10 transition-colors disabled:opacity-30 disabled:cursor-not-allowed"
      disabled={!selectedModel || (modelDownloaded[selectedModel] === false) || starting || whisperDownloading}
      on:click={handleStart}
    >{starting ? 'Starting...' : 'Start'}</button>
  </div>

</div>
</div>
