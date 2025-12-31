<script>
  export let percent = 0;
  export let message = '';
  export let logs = [];
  export let showLogs = false;

  let logsExpanded = false;

  $: displayPercent = Math.min(100, Math.max(0, percent));
</script>

<div class="progress-container">
  <div class="progress-header">
    <span class="progress-message">{message}</span>
    <span class="progress-percent">{displayPercent}%</span>
  </div>

  <div class="progress-bar">
    <div class="progress-fill" style="width: {displayPercent}%"></div>
  </div>

  {#if showLogs && logs.length > 0}
    <button class="logs-toggle" on:click={() => logsExpanded = !logsExpanded}>
      <span class="toggle-icon">{logsExpanded ? '▼' : '▶'}</span>
      Details ({logs.length})
    </button>

    {#if logsExpanded}
      <div class="logs-container">
        {#each logs as log}
          <div class="log-line">{log}</div>
        {/each}
      </div>
    {/if}
  {/if}
</div>

<style>
  .progress-container {
    width: 100%;
    padding: 16px;
    background: var(--bg-secondary, #1a1a2e);
    border-radius: 8px;
  }

  .progress-header {
    display: flex;
    justify-content: space-between;
    margin-bottom: 8px;
    font-size: 0.9rem;
  }

  .progress-message {
    color: var(--text-primary, #fff);
  }

  .progress-percent {
    color: var(--text-secondary, #888);
    font-family: monospace;
  }

  .progress-bar {
    height: 8px;
    background: var(--bg-tertiary, #2a2a3e);
    border-radius: 4px;
    overflow: hidden;
  }

  .progress-fill {
    height: 100%;
    background: var(--accent-color, #646cff);
    border-radius: 4px;
    transition: width 0.3s ease;
  }

  .logs-toggle {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-top: 12px;
    padding: 4px 8px;
    background: none;
    border: none;
    color: var(--text-secondary, #888);
    font-size: 0.85rem;
    cursor: pointer;
  }

  .logs-toggle:hover {
    color: var(--text-primary, #fff);
  }

  .toggle-icon {
    font-size: 0.7rem;
  }

  .logs-container {
    margin-top: 8px;
    padding: 12px;
    background: var(--bg-tertiary, #0a0a14);
    border-radius: 6px;
    max-height: 150px;
    overflow-y: auto;
    font-family: monospace;
    font-size: 0.8rem;
    text-align: left;
  }

  .log-line {
    color: var(--text-secondary, #888);
    padding: 2px 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .log-line:last-child {
    color: var(--text-primary, #fff);
  }
</style>
