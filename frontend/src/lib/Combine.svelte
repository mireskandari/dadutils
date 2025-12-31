<script>
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime.js';
  import {
    SelectPDFFiles,
    LoadPDFInfo,
    CombinePDFs,
    MergeTwoFiles,
    ReorderPages,
    SaveFile,
    OpenFile
  } from '../../wailsjs/go/main/App.js';
  import FileDropZone from './components/FileDropZone.svelte';
  import FileList from './components/FileList.svelte';
  import ProgressBar from './components/ProgressBar.svelte';

  const dispatch = createEventDispatcher();

  // State
  let documents = [];
  let selectedIds = [];
  let isProcessing = false;
  let progress = 0;
  let progressMessage = '';
  let logs = [];
  let result = null;
  let error = null;
  let savedPath = null;

  // Dialogs
  let showMergeDialog = false;
  let showEditPagesDialog = false;
  let editingDoc = null;
  let editPageOrder = [];

  // Event handlers
  function handleProgress(data) {
    progress = data.percent;
    progressMessage = data.message;
  }

  function handleLog(line) {
    logs = [...logs, line];
  }

  onMount(() => {
    EventsOn('combine:progress', handleProgress);
    EventsOn('combine:log', handleLog);
  });

  onDestroy(() => {
    EventsOff('combine:progress');
    EventsOff('combine:log');
  });

  // Derived state
  $: canMergePages = selectedIds.length === 2;
  $: canEditPages = selectedIds.length === 1;
  $: canCombine = documents.length >= 2;
  $: totalPages = documents.reduce((sum, d) => sum + d.pageCount, 0);

  // Actions
  async function handleBrowse() {
    try {
      const selected = await SelectPDFFiles();
      if (selected && selected.length > 0) {
        documents = [...documents, ...selected];
        error = null;
      }
    } catch (e) {
      error = e.message || 'Failed to select files';
    }
  }

  async function handleFileDrop(paths) {
    for (const path of paths) {
      try {
        const doc = await LoadPDFInfo(path);
        if (doc) {
          documents = [...documents, doc];
        }
      } catch (e) {
        error = `Failed to load ${path}: ${e.message}`;
      }
    }
  }

  function handleReorder(event) {
    documents = event.detail.files;
  }

  function handleRemove(event) {
    const id = event.detail.id;
    documents = documents.filter(d => d.id !== id);
    selectedIds = selectedIds.filter(i => i !== id);
  }

  function handleSelectionChange(event) {
    selectedIds = event.detail.selectedIds;
  }

  // Merge two files dialog
  function openMergeDialog() {
    if (canMergePages) {
      showMergeDialog = true;
    }
  }

  async function handleMerge(mode) {
    showMergeDialog = false;
    const [idA, idB] = selectedIds;
    const docA = documents.find(d => d.id === idA);
    const docB = documents.find(d => d.id === idB);

    if (!docA || !docB) return;

    try {
      isProcessing = true;
      progressMessage = `Merging ${docA.name} and ${docB.name}...`;
      progress = 50;

      const merged = await MergeTwoFiles(docA.path, docB.path, mode);

      if (merged) {
        // Replace the two files with the merged one
        const indexA = documents.findIndex(d => d.id === idA);
        documents = [
          ...documents.slice(0, indexA),
          merged,
          ...documents.slice(indexA + 1).filter(d => d.id !== idB)
        ];
        selectedIds = [];
      }
    } catch (e) {
      error = e.message || 'Merge failed';
    } finally {
      isProcessing = false;
    }
  }

  // Edit pages dialog
  function openEditPagesDialog() {
    if (canEditPages) {
      const doc = documents.find(d => d.id === selectedIds[0]);
      if (doc) {
        editingDoc = doc;
        // Initialize page order
        editPageOrder = doc.pageOrder || Array.from({ length: doc.pageCount }, (_, i) => i + 1);
        showEditPagesDialog = true;
      }
    }
  }

  function movePageUp(index) {
    if (index > 0) {
      const newOrder = [...editPageOrder];
      [newOrder[index - 1], newOrder[index]] = [newOrder[index], newOrder[index - 1]];
      editPageOrder = newOrder;
    }
  }

  function movePageDown(index) {
    if (index < editPageOrder.length - 1) {
      const newOrder = [...editPageOrder];
      [newOrder[index], newOrder[index + 1]] = [newOrder[index + 1], newOrder[index]];
      editPageOrder = newOrder;
    }
  }

  async function applyPageOrder() {
    if (!editingDoc) return;

    try {
      isProcessing = true;
      progressMessage = 'Reordering pages...';
      progress = 50;

      const reordered = await ReorderPages(editingDoc.path, editPageOrder);

      if (reordered) {
        // Replace the document
        const index = documents.findIndex(d => d.id === editingDoc.id);
        documents = [
          ...documents.slice(0, index),
          { ...reordered, pageOrder: editPageOrder },
          ...documents.slice(index + 1)
        ];
      }
    } catch (e) {
      error = e.message || 'Reorder failed';
    } finally {
      isProcessing = false;
      showEditPagesDialog = false;
      editingDoc = null;
      selectedIds = [];
    }
  }

  // Combine all
  async function handleCombine() {
    if (!canCombine) return;

    isProcessing = true;
    progress = 0;
    progressMessage = 'Starting...';
    logs = [];
    error = null;
    result = null;

    try {
      result = await CombinePDFs(documents);

      if (result.success) {
        const suggestedName = 'combined.pdf';
        savedPath = await SaveFile(result.outputPath, suggestedName);
      } else {
        error = result.error || 'Combine failed';
      }
    } catch (e) {
      error = e.message || 'Combine failed';
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
    documents = [];
    selectedIds = [];
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

<div class="combine">
  <header class="header">
    <button class="back-btn" on:click={goBack}>
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M19 12H5M12 19l-7-7 7-7"/>
      </svg>
    </button>
    <h1>Combine PDFs</h1>
  </header>

  <div class="content">
    {#if result && result.success}
      <!-- Success -->
      <div class="success-card">
        <div class="success-icon">✓</div>
        <h2>Combined Successfully</h2>

        <div class="stats-row">
          <span>{result.fileCount} files → 1 PDF</span>
          <span>{result.pageCount} pages total</span>
          <span>Size: {formatSize(result.outputSize)}</span>
        </div>

        {#if savedPath}
          <div class="actions">
            <button class="btn primary" on:click={handleOpen}>Open</button>
            <button class="btn secondary" on:click={handleReset}>Combine More</button>
          </div>
        {:else}
          <p class="save-cancelled">Save cancelled. <button class="link-btn" on:click={() => SaveFile(result.outputPath, 'combined.pdf').then(p => savedPath = p)}>Try again</button></p>
        {/if}
      </div>

    {:else if isProcessing}
      <!-- Processing -->
      <ProgressBar
        percent={progress}
        message={progressMessage}
        {logs}
        showLogs={true}
      />

    {:else}
      <!-- File list / empty state -->
      {#if documents.length === 0}
        <FileDropZone
          multiple={true}
          on:browse={handleBrowse}
          on:files={(e) => handleFileDrop(e.detail.paths)}
        >
          Drag & drop PDFs here
        </FileDropZone>
      {:else}
        <div class="toolbar">
          <button class="btn secondary" on:click={handleBrowse}>Add More</button>
        </div>

        <FileList
          files={documents}
          bind:selectedIds
          on:reorder={handleReorder}
          on:remove={handleRemove}
          on:selectionChange={handleSelectionChange}
        />

        {#if selectedIds.length > 0}
          <div class="selection-bar">
            <span>{selectedIds.length} file{selectedIds.length !== 1 ? 's' : ''} selected</span>
            <div class="selection-actions">
              {#if canMergePages}
                <button class="btn small" on:click={openMergeDialog}>Merge Pages</button>
              {/if}
              {#if canEditPages}
                <button class="btn small" on:click={openEditPagesDialog}>Edit Pages</button>
              {/if}
              <button class="btn small" on:click={() => selectedIds = []}>Clear</button>
            </div>
          </div>
        {/if}

        <button
          class="btn primary combine-btn"
          on:click={handleCombine}
          disabled={!canCombine}
        >
          Combine ({totalPages} pages)
        </button>
      {/if}

      <!-- Drop zone for adding more files -->
      {#if documents.length > 0}
        <FileDropZone
          multiple={true}
          on:browse={handleBrowse}
          on:files={(e) => handleFileDrop(e.detail.paths)}
        >
          <span slot="icon"></span>
          Drop more files here
        </FileDropZone>
      {/if}
    {/if}

    {#if error}
      <div class="error">{error}</div>
    {/if}
  </div>
</div>

<!-- Merge Dialog -->
{#if showMergeDialog}
  <div class="modal-overlay" on:click={() => showMergeDialog = false}>
    <div class="modal" on:click|stopPropagation>
      <h3>How should pages be merged?</h3>

      <button class="merge-option" on:click={() => handleMerge('interleave')}>
        <strong>Interleave</strong>
        <span>A1 → B1 → A2 → B2 → A3 → B3...</span>
        <span class="hint">Good for double-sided scans</span>
      </button>

      <button class="merge-option" on:click={() => handleMerge('append')}>
        <strong>Append</strong>
        <span>All of A → then all of B</span>
        <span class="hint">Standard merge</span>
      </button>

      <button class="btn secondary cancel-btn" on:click={() => showMergeDialog = false}>
        Cancel
      </button>
    </div>
  </div>
{/if}

<!-- Edit Pages Dialog -->
{#if showEditPagesDialog && editingDoc}
  <div class="modal-overlay" on:click={() => showEditPagesDialog = false}>
    <div class="modal wide" on:click|stopPropagation>
      <h3>Edit Pages: {editingDoc.name}</h3>
      <p class="modal-hint">Drag to reorder or use arrow buttons</p>

      <div class="page-grid">
        {#each editPageOrder as pageNum, index}
          <div class="page-item">
            <span class="page-number">Page {pageNum}</span>
            <div class="page-controls">
              <button
                class="arrow-btn"
                disabled={index === 0}
                on:click={() => movePageUp(index)}
              >↑</button>
              <button
                class="arrow-btn"
                disabled={index === editPageOrder.length - 1}
                on:click={() => movePageDown(index)}
              >↓</button>
            </div>
          </div>
        {/each}
      </div>

      <div class="modal-actions">
        <button class="btn secondary" on:click={() => showEditPagesDialog = false}>Cancel</button>
        <button class="btn primary" on:click={applyPageOrder}>Done</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .combine {
    max-width: 600px;
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
    gap: 16px;
  }

  .toolbar {
    display: flex;
    justify-content: flex-end;
  }

  .selection-bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 16px;
    background: var(--accent-color, #646cff);
    border-radius: 8px;
  }

  .selection-actions {
    display: flex;
    gap: 8px;
  }

  .btn {
    padding: 10px 20px;
    border: none;
    border-radius: 6px;
    font-size: 0.95rem;
    cursor: pointer;
    transition: background 0.2s;
  }

  .btn.small {
    padding: 6px 12px;
    font-size: 0.85rem;
  }

  .btn.primary {
    background: var(--accent-color, #646cff);
    color: white;
  }

  .btn.primary:hover:not(:disabled) {
    background: var(--accent-hover, #535bf2);
  }

  .btn.primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn.secondary {
    background: var(--bg-secondary, #1a1a2e);
    color: var(--text-primary, #fff);
    border: 1px solid var(--border-color, #333);
  }

  .btn.secondary:hover {
    background: var(--bg-hover, #232338);
  }

  .selection-bar .btn.small {
    background: rgba(255,255,255,0.2);
    color: white;
    border: none;
  }

  .selection-bar .btn.small:hover {
    background: rgba(255,255,255,0.3);
  }

  .combine-btn {
    width: 100%;
    padding: 14px;
    font-size: 1rem;
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
    margin: 0 0 16px;
    font-size: 1.3rem;
  }

  .stats-row {
    display: flex;
    justify-content: center;
    gap: 24px;
    margin-bottom: 20px;
    color: var(--text-secondary, #888);
    font-size: 0.9rem;
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

  /* Modal styles */
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .modal {
    background: var(--bg-primary, #0f0f1a);
    border: 1px solid var(--border-color, #333);
    border-radius: 12px;
    padding: 24px;
    max-width: 400px;
    width: 90%;
  }

  .modal.wide {
    max-width: 500px;
  }

  .modal h3 {
    margin: 0 0 16px;
  }

  .modal-hint {
    color: var(--text-secondary, #888);
    font-size: 0.9rem;
    margin: 0 0 16px;
  }

  .merge-option {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    width: 100%;
    padding: 16px;
    margin-bottom: 12px;
    background: var(--bg-secondary, #1a1a2e);
    border: 1px solid var(--border-color, #333);
    border-radius: 8px;
    color: var(--text-primary, #fff);
    text-align: left;
    cursor: pointer;
    transition: all 0.2s;
  }

  .merge-option:hover {
    background: var(--bg-hover, #232338);
    border-color: var(--accent-color, #646cff);
  }

  .merge-option strong {
    font-size: 1rem;
    margin-bottom: 4px;
  }

  .merge-option span {
    font-size: 0.85rem;
    color: var(--text-secondary, #888);
  }

  .merge-option .hint {
    font-size: 0.75rem;
    margin-top: 4px;
    font-style: italic;
  }

  .cancel-btn {
    width: 100%;
    margin-top: 8px;
  }

  .page-grid {
    max-height: 300px;
    overflow-y: auto;
    margin-bottom: 16px;
  }

  .page-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 10px 12px;
    margin-bottom: 6px;
    background: var(--bg-secondary, #1a1a2e);
    border-radius: 6px;
  }

  .page-number {
    font-size: 0.9rem;
  }

  .page-controls {
    display: flex;
    gap: 4px;
  }

  .arrow-btn {
    width: 28px;
    height: 28px;
    background: var(--bg-tertiary, #333);
    border: none;
    border-radius: 4px;
    color: var(--text-primary, #fff);
    cursor: pointer;
  }

  .arrow-btn:disabled {
    opacity: 0.3;
    cursor: not-allowed;
  }

  .arrow-btn:hover:not(:disabled) {
    background: var(--accent-color, #646cff);
  }

  .modal-actions {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
  }
</style>
