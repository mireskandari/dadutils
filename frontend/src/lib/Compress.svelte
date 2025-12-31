<script>
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime.js';
  import { SelectPDFFile, CompressPDF, SaveFile, OpenFile } from '../../wailsjs/go/main/App.js';
  import FileDropZone from './components/FileDropZone.svelte';
  import ProgressBar from './components/ProgressBar.svelte';

  const dispatch = createEventDispatcher();

  // State
  let file = null;
  let preset = 'printer';
  let isProcessing = false;
  let progress = 0;
  let progressMessage = '';
  let logs = [];
  let result = null;
  let error = null;
  let savedPath = null;

  const presets = [
    { id: 'screen', name: 'Screen', description: 'Smallest file, lower quality' },
    { id: 'ebook', name: 'eBook', description: 'Good for digital reading' },
    { id: 'printer', name: 'Printer', description: 'Balanced quality and size' },
    { id: 'prepress', name: 'Prepress', description: 'High quality for printing' },
    { id: 'default', name: 'Default', description: 'Minimal compression' },
  ];

  // Event handlers
  function handleProgress(data) {
    progress = data.percent;
    progressMessage = data.message;
  }

  function handleLog(line) {
    logs = [...logs, line];
  }

  onMount(() => {
    EventsOn('compress:progress', handleProgress);
    EventsOn('compress:log', handleLog);
  });

  onDestroy(() => {
    EventsOff('compress:progress');
    EventsOff('compress:log');
  });

  // Actions
  async function handleBrowse() {
    try {
      const selected = await SelectPDFFile();
      if (selected) {
        file = selected;
        error = null;
        result = null;
        savedPath = null;
      }
    } catch (e) {
      error = e.message || 'Failed to select file';
    }
  }

  async function handleCompress() {
    if (!file) return;

    isProcessing = true;
    progress = 0;
    progressMessage = 'Starting...';
    logs = [];
    error = null;
    result = null;

    try {
      result = await CompressPDF(file.path, preset);

      if (result.success) {
        // Show save dialog
        const suggestedName = file.name.replace('.pdf', '_compressed.pdf');
        savedPath = await SaveFile(result.outputPath, suggestedName);
      } else {
        error = result.error || 'Compression failed';
      }
    } catch (e) {
      error = e.message || 'Compression failed';
    } finally {
      isProcessing = false;
    }
  }

  async function handleOpen() {
    if (savedPath) {
      try {
        await OpenFile(savedPath);
      } catch (e) {
        error = 'Failed to open file';
      }
    }
  }

  function handleReset() {
    file = null;
    result = null;
    savedPath = null;
    error = null;
    progress = 0;
    logs = [];
  }

  function goBack() {
    dispatch('navigate', { view: 'home' });
  }

  function formatSize(bytes) {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
  }
</script>

<div class="compress">
  <header class="header">
    <button class="back-btn" on:click={goBack}>
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M19 12H5M12 19l-7-7 7-7"/>
      </svg>
    </button>
    <h1>Compress PDF</h1>
  </header>

  <div class="content">
    {#if !file}
      <!-- File selection -->
      <FileDropZone on:browse={handleBrowse} on:files={(e) => console.log('dropped', e.detail)}>
        Drag & drop PDF here
      </FileDropZone>

    {:else if isProcessing}
      <!-- Processing -->
      <div class="file-info">
        <span class="file-name">{file.name}</span>
        <span class="file-size">{file.sizeText}</span>
      </div>

      <ProgressBar
        percent={progress}
        message={progressMessage}
        {logs}
        showLogs={true}
      />

    {:else if result && result.success}
      <!-- Success -->
      <div class="success-card">
        <div class="success-icon">✓</div>
        <h2>Compression Complete</h2>

        <div class="stats">
          <div class="stat">
            <span class="stat-label">Original</span>
            <span class="stat-value">{formatSize(result.originalSize)}</span>
          </div>
          <div class="stat-arrow">→</div>
          <div class="stat">
            <span class="stat-label">Compressed</span>
            <span class="stat-value">{formatSize(result.compressedSize)}</span>
          </div>
        </div>

        <div class="savings" class:negative={result.savingsPercent < 0}>
          {#if result.savingsPercent >= 0}
            Saved {result.savingsPercent}%
          {:else}
            File grew by {Math.abs(result.savingsPercent)}%
          {/if}
        </div>

        {#if savedPath}
          <div class="actions">
            <button class="btn primary" on:click={handleOpen}>Open</button>
            <button class="btn secondary" on:click={handleReset}>Compress Another</button>
          </div>
        {:else}
          <p class="save-cancelled">Save cancelled. <button class="link-btn" on:click={() => SaveFile(result.outputPath, file.name.replace('.pdf', '_compressed.pdf')).then(p => savedPath = p)}>Try again</button></p>
        {/if}
      </div>

    {:else}
      <!-- File loaded, ready to compress -->
      <div class="file-card">
        <div class="file-info">
          <span class="file-name">{file.name}</span>
          <span class="file-size">{file.sizeText}</span>
        </div>
        <button class="remove-btn" on:click={handleReset}>✕</button>
      </div>

      <div class="presets">
        <h3>Select Quality</h3>
        {#each presets as p}
          <label class="preset-option" class:selected={preset === p.id}>
            <input type="radio" bind:group={preset} value={p.id} />
            <span class="preset-name">{p.name}</span>
            <span class="preset-desc">{p.description}</span>
          </label>
        {/each}
      </div>

      <button class="btn primary compress-btn" on:click={handleCompress}>
        Compress
      </button>
    {/if}

    {#if error}
      <div class="error">{error}</div>
    {/if}
  </div>
</div>

<style>
  .compress {
    max-width: 500px;
    margin: 0 auto;
    padding: 20px;
  }

  .header {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 24px;
  }

  .header h1 {
    font-size: 1.5rem;
    margin: 0;
  }

  .back-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    background: var(--bg-secondary, #1a1a2e);
    border: 1px solid var(--border-color, #333);
    border-radius: 8px;
    color: var(--text-primary, #fff);
    cursor: pointer;
  }

  .back-btn:hover {
    background: var(--bg-hover, #232338);
  }

  .content {
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .file-card {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px;
    background: var(--bg-secondary, #1a1a2e);
    border-radius: 8px;
  }

  .file-info {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .file-name {
    font-weight: 500;
    color: var(--text-primary, #fff);
  }

  .file-size {
    font-size: 0.85rem;
    color: var(--text-secondary, #888);
  }

  .remove-btn {
    width: 28px;
    height: 28px;
    background: var(--bg-tertiary, #333);
    border: none;
    border-radius: 4px;
    color: var(--text-secondary, #888);
    cursor: pointer;
  }

  .remove-btn:hover {
    background: #c00;
    color: white;
  }

  .presets {
    background: var(--bg-secondary, #1a1a2e);
    border-radius: 8px;
    padding: 16px;
  }

  .presets h3 {
    margin: 0 0 12px;
    font-size: 0.9rem;
    color: var(--text-secondary, #888);
  }

  .preset-option {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px;
    margin: 4px 0;
    background: var(--bg-tertiary, #232338);
    border-radius: 6px;
    cursor: pointer;
    transition: background 0.2s;
  }

  .preset-option:hover {
    background: var(--bg-hover, #2a2a4e);
  }

  .preset-option.selected {
    background: var(--accent-color, #646cff);
  }

  .preset-option input {
    display: none;
  }

  .preset-name {
    font-weight: 500;
    min-width: 80px;
  }

  .preset-desc {
    font-size: 0.85rem;
    color: var(--text-secondary, #888);
  }

  .preset-option.selected .preset-desc {
    color: rgba(255,255,255,0.8);
  }

  .btn {
    padding: 12px 24px;
    border: none;
    border-radius: 8px;
    font-size: 1rem;
    cursor: pointer;
    transition: background 0.2s;
  }

  .btn.primary {
    background: var(--accent-color, #646cff);
    color: white;
  }

  .btn.primary:hover {
    background: var(--accent-hover, #535bf2);
  }

  .btn.secondary {
    background: var(--bg-secondary, #1a1a2e);
    color: var(--text-primary, #fff);
    border: 1px solid var(--border-color, #333);
  }

  .btn.secondary:hover {
    background: var(--bg-hover, #232338);
  }

  .compress-btn {
    width: 100%;
  }

  .success-card {
    text-align: center;
    padding: 32px;
    background: var(--bg-secondary, #1a1a2e);
    border-radius: 12px;
  }

  .success-icon {
    width: 48px;
    height: 48px;
    line-height: 48px;
    margin: 0 auto 16px;
    background: #2ecc71;
    border-radius: 50%;
    font-size: 1.5rem;
    color: white;
  }

  .success-card h2 {
    margin: 0 0 20px;
    font-size: 1.3rem;
  }

  .stats {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 16px;
    margin-bottom: 16px;
  }

  .stat {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .stat-label {
    font-size: 0.8rem;
    color: var(--text-secondary, #888);
  }

  .stat-value {
    font-size: 1.2rem;
    font-weight: 600;
  }

  .stat-arrow {
    color: var(--text-secondary, #888);
    font-size: 1.2rem;
  }

  .savings {
    font-size: 1.1rem;
    color: #2ecc71;
    margin-bottom: 20px;
  }

  .savings.negative {
    color: #e74c3c;
  }

  .actions {
    display: flex;
    gap: 12px;
    justify-content: center;
  }

  .save-cancelled {
    color: var(--text-secondary, #888);
  }

  .link-btn {
    background: none;
    border: none;
    color: var(--accent-color, #646cff);
    cursor: pointer;
    text-decoration: underline;
  }

  .error {
    padding: 12px;
    background: rgba(231, 76, 60, 0.2);
    border: 1px solid #e74c3c;
    border-radius: 8px;
    color: #e74c3c;
    text-align: center;
  }
</style>
