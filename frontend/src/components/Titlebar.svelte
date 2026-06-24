<script>
import { onMount } from 'svelte'
import { WindowMinimize, WindowToggleMaximize, WindowClose } from '../../wailsjs/go/main/App.js'
import { WindowIsMaximised } from '../../wailsjs/runtime/runtime.js'

let maximized = false

onMount(async () => {
  maximized = await WindowIsMaximised()
})

function handleMaximize() {
  WindowToggleMaximize()
  maximized = !maximized
}
</script>

<div class="titlebar" on:dblclick={handleMaximize}>
  <div class="window-controls">
    <button class="control-btn" on:click={WindowMinimize} title="Minimize">
      <svg width="12" height="12" viewBox="0 0 12 12">
        <line x1="2" y1="6" x2="10" y2="6" stroke="currentColor" stroke-width="1"/>
      </svg>
    </button>

    <button class="control-btn" on:click={handleMaximize} title={maximized ? 'Restore' : 'Maximize'}>
      {#if maximized}
        <svg width="12" height="12" viewBox="0 0 12 12">
          <rect x="3" y="1" width="7" height="7" fill="none" stroke="currentColor" stroke-width="1"/>
          <rect x="1" y="3" width="7" height="7" fill="var(--app-bg)" stroke="currentColor" stroke-width="1"/>
        </svg>
      {:else}
        <svg width="12" height="12" viewBox="0 0 12 12">
          <rect x="2" y="2" width="8" height="8" fill="none" stroke="currentColor" stroke-width="1"/>
        </svg>
      {/if}
    </button>

    <button class="control-btn control-btn-close" on:click={WindowClose} title="Close">
      <svg width="12" height="12" viewBox="0 0 12 12">
        <line x1="2" y1="2" x2="10" y2="10" stroke="currentColor" stroke-width="1"/>
        <line x1="10" y1="2" x2="2" y2="10" stroke="currentColor" stroke-width="1"/>
      </svg>
    </button>
  </div>
</div>

<style>
  .titlebar {
    position: relative;
	--wails-draggable: drag;
    display: flex;
    align-items: center;
    justify-content: flex-end;
    height: 32px;
    background: var(--app-bg);
    border-bottom: 1px solid var(--app-border);
    -webkit-app-region: drag;
  }

  .window-controls {
    display: flex;
    height: 100%;
    -webkit-app-region: no-drag;
  }

  .control-btn {
    display: flex;
	--wails-draggable: no-drag; 
    align-items: center;
    justify-content: center;
    width: 46px;
    height: 100%;
    background: transparent;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    transition: background 0.15s, color 0.15s;
  }

  .control-btn:hover {
    background: var(--app-border);
    color: var(--text-primary);
  }

  .control-btn-close:hover {
    background: var(--danger);
    color: white;
  }
</style>
