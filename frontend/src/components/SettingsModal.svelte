<script>
import { createEventDispatcher } from 'svelte'
import { SaveSettings, SetWindowFrame, IsModelDownloaded, DownloadModel } from '../../wailsjs/go/main/App.js'
import { EventsOn } from '../../wailsjs/runtime/runtime.js'

export let settings = {}
export let models = []

const dispatch = createEventDispatcher()

let local = { ...settings }
let modelDownloaded = {}
let downloadErr = ''
let downloadProgress = 0
let downloading = false

const frameOptions = [
  { value: 'system', label: 'System default' },
  { value: 'custom', label: 'Custom titlebar' },
  { value: 'none', label: 'None (for Hyprland/tiling WM)' },
]

async function save() {
  if (local.window_frame !== settings.window_frame) {
    await SetWindowFrame(local.window_frame)
  }
  await SaveSettings(local)
  dispatch('close')
}

async function checkDownloaded(name) {
  if (!name) return
  if (modelDownloaded[name] !== undefined) return modelDownloaded[name]
  modelDownloaded[name] = await IsModelDownloaded(name)
  modelDownloaded = modelDownloaded
  return modelDownloaded[name]
}

async function downloadModel() {
  const name = local.model_name
  if (!name) return
  downloading = true
  downloadErr = ''
  downloadProgress = 0

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
    downloading = false
  }
}

async function handleModelChange() {
  const name = local.model_name
  if (name) await checkDownloaded(name)
}
</script>

<div class="fixed inset-0 z-50 flex items-center justify-center" style="background: rgba(0,0,0,0.7)">
<div class="bg-app-elevated border border-app-border w-full max-w-md mx-4 p-6">

  <h2 class="font-mono text-xs uppercase tracking-wider text-text-secondary mb-6">Settings</h2>

  <div class="mb-4">
    <label for="s-model" class="block font-mono text-[11px] uppercase tracking-wider text-text-muted mb-2">Default model</label>
    <select
      id="s-model"
      class="w-full bg-app-bg border border-app-border px-3 py-1.5 font-mono text-xs text-text-primary focus:border-accent/40 focus:outline-none appearance-none"
      bind:value={local.model_name}
      on:change={handleModelChange}
    >
      <option value="">— Select —</option>
      {#each models as m}
        <option value={m.name}>{m.display_name}</option>
      {/each}
    </select>

    {#if local.model_name && modelDownloaded[local.model_name] === false}
      <div class="mt-2">
        {#if downloading}
          <div class="h-1 bg-app-bg overflow-hidden mb-1">
            <div class="h-full bg-accent transition-all duration-200" style="width: {downloadProgress * 100}%"></div>
          </div>
          <span class="font-mono text-[10px] text-text-muted">{Math.round(downloadProgress * 100)}%</span>
        {:else}
          <button
            class="font-mono text-[11px] px-3 py-1 border border-accent/40 text-accent hover:bg-accent/10 transition-colors"
            on:click={downloadModel}
          >Download</button>
        {/if}
      </div>
    {:else if local.model_name && modelDownloaded[local.model_name] === true}
      <div class="mt-1 font-mono text-[10px] text-success">Ready</div>
    {/if}

    {#if downloadErr}
      <div class="mt-2 font-mono text-[10px] text-danger">{downloadErr}</div>
    {/if}
  </div>

  <div class="mb-4">
    <label for="s-lang" class="block font-mono text-[11px] uppercase tracking-wider text-text-muted mb-2">Language</label>
    <select
      id="s-lang"
      class="w-full bg-app-bg border border-app-border px-3 py-1.5 font-mono text-xs text-text-primary focus:border-accent/40 focus:outline-none appearance-none"
      bind:value={local.language}
    >
      <option value="auto">Auto-detect (recommended)</option>
      <option value="en">English</option>
      <option value="ru">Russian</option>
      <option value="de">German</option>
      <option value="fr">French</option>
      <option value="es">Spanish</option>
      <option value="it">Italian</option>
      <option value="ja">Japanese</option>
      <option value="ko">Korean</option>
      <option value="zh">Chinese</option>
      <option value="pt">Portuguese</option>
      <option value="ar">Arabic</option>
      <option value="hi">Hindi</option>
      <option value="tr">Turkish</option>
      <option value="nl">Dutch</option>
      <option value="pl">Polish</option>
      <option value="uk">Ukrainian</option>
    </select>
  </div>

  <div class="mb-4">
    <label for="s-threads" class="block font-mono text-[11px] uppercase tracking-wider text-text-muted mb-2">Threads</label>
    <input
      id="s-threads"
      type="number"
      class="w-full bg-app-bg border border-app-border px-3 py-1.5 font-mono text-xs text-text-primary focus:border-accent/40 focus:outline-none"
      bind:value={local.threads}
      min="1"
      max="32"
    />
  </div>

  <div class="mb-4 flex items-center gap-2">
    <input
      type="checkbox"
      id="stamps"
      class="bg-app-bg border-app-border text-accent focus:ring-0 focus:ring-offset-0"
      bind:checked={local.show_timestamps}
    />
    <label for="stamps" class="font-mono text-[11px] uppercase tracking-wider text-text-muted">Show timestamps</label>
  </div>

  <div class="mb-6">
    <label for="s-frame" class="block font-mono text-[11px] uppercase tracking-wider text-text-muted mb-2">Window frame</label>
    <select
      id="s-frame"
      class="w-full bg-app-bg border border-app-border px-3 py-1.5 font-mono text-xs text-text-primary focus:border-accent/40 focus:outline-none appearance-none"
      bind:value={local.window_frame}
    >
      {#each frameOptions as opt}
        <option value={opt.value}>{opt.label}</option>
      {/each}
    </select>
  </div>

  <div class="flex justify-end gap-2">
    <button
      class="font-mono text-xs px-4 py-1.5 border border-app-border text-text-muted hover:border-text-muted transition-colors"
      on:click={() => dispatch('close')}
    >Cancel</button>
    <button
      class="font-mono text-xs px-4 py-1.5 border border-accent/40 text-accent hover:bg-accent/10 transition-colors"
      on:click={save}
    >Save</button>
  </div>

</div>
</div>
