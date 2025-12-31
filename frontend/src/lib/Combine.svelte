<script>
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import { flip } from 'svelte/animate';
  import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime.js';
  import {
    SelectPDFFiles,
    LoadPDFInfo,
    CombinePDFs,
    MergeTwoFiles,
    ReorderPages,
    SaveFile,
    OpenFile,
    GenerateAllThumbnails
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

  // Thumbnails state: { [docId]: firstPageThumbnailDataUrl }
  let thumbnails = {};
  // Edit pages thumbnails: [{ pageIndex, imageData }]
  let editPageThumbnails = [];

  // Dialogs
  let showMergeDialog = false;
  let showEditPagesDialog = false;
  let editingDoc = null;
  let editPageOrder = [];

  // Drag-and-drop state for Edit Pages
  let draggedPageIndex = null;
  let dragOverPageIndex = null;

  // Generate thumbnail for a document (first page only for list view)
  async function generateThumbnail(doc) {
    try {
      const results = await GenerateAllThumbnails(doc.path, 80, 110);
      if (results && results.length > 0) {
        thumbnails = { ...thumbnails, [doc.id]: results[0].imageData };
      }
    } catch (e) {
      console.error('Failed to generate thumbnail:', e);
    }
  }

  // Generate all page thumbnails for edit dialog
  async function generateEditThumbnails(doc) {
    try {
      editPageThumbnails = [];
      const results = await GenerateAllThumbnails(doc.path, 120, 165);
      if (results) {
        editPageThumbnails = results;
      }
    } catch (e) {
      console.error('Failed to generate page thumbnails:', e);
    }
  }

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
  $: canSaveSingle = documents.length === 1;
  $: totalPages = documents.reduce((sum, d) => sum + d.pageCount, 0);

  // Actions
  async function handleBrowse() {
    try {
      const selected = await SelectPDFFiles();
      if (selected && selected.length > 0) {
        documents = [...documents, ...selected];
        error = null;
        // Generate thumbnails for new files
        for (const doc of selected) {
          generateThumbnail(doc);
        }
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
          // Generate thumbnail for new file
          generateThumbnail(doc);
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
    // Clean up thumbnail
    const { [id]: _, ...rest } = thumbnails;
    thumbnails = rest;
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
  async function openEditPagesDialog() {
    if (canEditPages) {
      const doc = documents.find(d => d.id === selectedIds[0]);
      if (doc) {
        editingDoc = doc;
        // Initialize page order
        editPageOrder = doc.pageOrder || Array.from({ length: doc.pageCount }, (_, i) => i + 1);
        showEditPagesDialog = true;
        // Generate page thumbnails
        generateEditThumbnails(doc);
      }
    }
  }

  // Drag-and-drop handlers for page reordering
  function handlePageDragStart(e, index) {
    draggedPageIndex = index;
    e.dataTransfer.effectAllowed = 'move';
  }

  function handlePageDragOver(e, index) {
    e.preventDefault();
    if (draggedPageIndex === null || draggedPageIndex === index) {
      dragOverPageIndex = null;
      return;
    }
    dragOverPageIndex = index;
  }

  function handlePageDrop(e, targetIndex) {
    e.preventDefault();
    if (draggedPageIndex === null || draggedPageIndex === targetIndex) return;

    const newOrder = [...editPageOrder];
    const [draggedPage] = newOrder.splice(draggedPageIndex, 1);
    newOrder.splice(targetIndex, 0, draggedPage);
    editPageOrder = newOrder;

    draggedPageIndex = null;
    dragOverPageIndex = null;
  }

  function handlePageDragEnd() {
    draggedPageIndex = null;
    dragOverPageIndex = null;
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
        const oldId = editingDoc.id;
        documents = [
          ...documents.slice(0, index),
          { ...reordered, pageOrder: editPageOrder },
          ...documents.slice(index + 1)
        ];

        // Clean up old thumbnail and regenerate for new document
        const { [oldId]: _, ...rest } = thumbnails;
        thumbnails = rest;
        generateThumbnail(reordered);
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

  // Save single document (for reordered PDFs)
  async function handleSaveSingle() {
    if (!canSaveSingle) return;

    const doc = documents[0];
    try {
      const suggestedName = doc.name.replace('.pdf', '-reordered.pdf');
      savedPath = await SaveFile(doc.path, suggestedName);
      if (savedPath) {
        result = {
          success: true,
          fileCount: 1,
          pageCount: doc.pageCount,
          outputSize: doc.size,
          outputPath: doc.path
        };
      }
    } catch (e) {
      error = e.message || 'Save failed';
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
    thumbnails = {};
    editPageThumbnails = [];
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
          {thumbnails}
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

        {#if canCombine}
          <button
            class="btn primary combine-btn"
            on:click={handleCombine}
          >
            Combine ({totalPages} pages)
          </button>
        {:else if canSaveSingle}
          <button
            class="btn primary combine-btn"
            on:click={handleSaveSingle}
          >
            Save PDF ({totalPages} pages)
          </button>
        {/if}
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
      <p class="modal-hint">Drag pages to reorder</p>

      <div class="page-grid-drag">
        {#each editPageOrder as pageNum, index (pageNum)}
          <div
            class="page-card-drag"
            class:dragging={draggedPageIndex === index}
            class:drag-over={dragOverPageIndex === index}
            draggable="true"
            animate:flip={{ duration: 200 }}
            on:dragstart={(e) => handlePageDragStart(e, index)}
            on:dragover={(e) => handlePageDragOver(e, index)}
            on:drop={(e) => handlePageDrop(e, index)}
            on:dragend={handlePageDragEnd}
          >
            <div class="drag-grip">
              <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
                <circle cx="5" cy="3" r="1.5"/>
                <circle cx="11" cy="3" r="1.5"/>
                <circle cx="5" cy="8" r="1.5"/>
                <circle cx="11" cy="8" r="1.5"/>
                <circle cx="5" cy="13" r="1.5"/>
                <circle cx="11" cy="13" r="1.5"/>
              </svg>
            </div>
            <div class="page-thumbnail">
              {#if editPageThumbnails[pageNum - 1]}
                <img src={editPageThumbnails[pageNum - 1].imageData} alt="Page {pageNum}" />
              {:else}
                <div class="thumb-placeholder">{pageNum}</div>
              {/if}
            </div>
            <div class="page-label">Page {pageNum}</div>
          </div>
        {/each}
      </div>

      <div class="modal-actions">
        <button class="btn secondary" on:click={() => showEditPagesDialog = false}>Cancel</button>
        <button class="btn primary" on:click={applyPageOrder}>Apply</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .combine {
    max-width: 620px;
    margin: 0 auto;
    padding: 24px;
  }

  .header {
    display: flex;
    align-items: center;
    gap: 14px;
    margin-bottom: 28px;
  }

  .header h1 {
    font-size: 1.5rem;
    font-weight: 600;
    margin: 0;
    color: var(--text-primary);
  }

  .back-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 40px;
    height: 40px;
    background: var(--bg-secondary);
    border: none;
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-sm);
    color: var(--text-secondary);
    cursor: pointer;
    transition: all var(--transition-fast);
  }

  .back-btn:hover {
    background: var(--bg-hover);
    color: var(--accent-color);
    box-shadow: var(--shadow-md);
  }

  .content {
    display: flex;
    flex-direction: column;
    gap: 18px;
  }

  .toolbar {
    display: flex;
    justify-content: flex-end;
  }

  .selection-bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 14px 18px;
    background: var(--accent-light);
    border: 1px solid var(--accent-color);
    border-radius: var(--radius-md);
    color: var(--accent-color);
    font-weight: 500;
  }

  .selection-actions {
    display: flex;
    gap: 8px;
  }

  .btn {
    padding: 12px 22px;
    border: none;
    border-radius: var(--radius-sm);
    font-size: 0.95rem;
    font-weight: 500;
    cursor: pointer;
    transition: all var(--transition-fast);
  }

  .btn.small {
    padding: 8px 14px;
    font-size: 0.85rem;
  }

  .btn.primary {
    background: var(--accent-color);
    color: white;
  }

  .btn.primary:hover:not(:disabled) {
    background: var(--accent-hover);
    transform: translateY(-1px);
    box-shadow: var(--shadow-md);
  }

  .btn.primary:disabled {
    background: var(--border-color);
    color: var(--text-muted);
    cursor: not-allowed;
  }

  .btn.secondary {
    background: var(--bg-secondary);
    color: var(--text-primary);
    border: 1px solid var(--border-color);
  }

  .btn.secondary:hover {
    background: var(--bg-hover);
  }

  .selection-bar .btn.small {
    background: var(--bg-secondary);
    color: var(--accent-color);
    border: 1px solid var(--accent-color);
  }

  .selection-bar .btn.small:hover {
    background: var(--accent-color);
    color: white;
  }

  .combine-btn {
    width: 100%;
    padding: 16px;
    font-size: 1rem;
    margin-top: 4px;
  }

  .success-card {
    text-align: center;
    padding: 36px;
    background: var(--bg-secondary);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-sm);
  }

  .success-icon {
    width: 56px;
    height: 56px;
    line-height: 56px;
    margin: 0 auto 20px;
    background: var(--success-light);
    border-radius: 50%;
    font-size: 1.6rem;
    color: var(--success-color);
  }

  .success-card h2 {
    margin: 0 0 20px;
    font-size: 1.4rem;
    color: var(--text-primary);
  }

  .stats-row {
    display: flex;
    justify-content: center;
    gap: 20px;
    margin-bottom: 24px;
    padding: 14px;
    background: var(--bg-tertiary);
    border-radius: var(--radius-sm);
    color: var(--text-secondary);
    font-size: 0.9rem;
  }

  .actions {
    display: flex;
    gap: 12px;
    justify-content: center;
  }

  .save-cancelled {
    color: var(--text-secondary);
    font-size: 0.9rem;
  }

  .link-btn {
    background: none;
    border: none;
    color: var(--accent-color);
    cursor: pointer;
    text-decoration: underline;
    font-size: inherit;
  }

  .link-btn:hover {
    color: var(--accent-hover);
  }

  .error {
    padding: 14px 18px;
    background: var(--error-light);
    border: 1px solid var(--error-color);
    border-radius: var(--radius-sm);
    color: var(--error-color);
    text-align: center;
    font-size: 0.95rem;
  }

  /* Modal styles */
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.4);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    animation: fadeIn 200ms ease;
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  .modal {
    background: var(--bg-secondary);
    border: none;
    border-radius: var(--radius-xl);
    box-shadow: var(--shadow-lg);
    padding: 28px;
    max-width: 420px;
    width: 90%;
    animation: slideUp 200ms cubic-bezier(0, 0, 0.2, 1);
  }

  @keyframes slideUp {
    from {
      opacity: 0;
      transform: translateY(10px) scale(0.98);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }

  .modal.wide {
    max-width: 540px;
  }

  .modal h3 {
    margin: 0 0 8px;
    font-size: 1.2rem;
    color: var(--text-primary);
  }

  .modal-hint {
    color: var(--text-secondary);
    font-size: 0.9rem;
    margin: 0 0 20px;
  }

  .merge-option {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    width: 100%;
    padding: 18px 20px;
    margin-bottom: 12px;
    background: var(--bg-tertiary);
    border: 2px solid transparent;
    border-radius: var(--radius-md);
    color: var(--text-primary);
    text-align: left;
    cursor: pointer;
    transition: all var(--transition-fast);
  }

  .merge-option:hover {
    background: var(--accent-light);
    border-color: var(--accent-color);
  }

  .merge-option strong {
    font-size: 1rem;
    margin-bottom: 4px;
    color: var(--text-primary);
  }

  .merge-option span {
    font-size: 0.85rem;
    color: var(--text-secondary);
  }

  .merge-option .hint {
    font-size: 0.75rem;
    margin-top: 6px;
    font-style: italic;
    color: var(--text-muted);
  }

  .cancel-btn {
    width: 100%;
    margin-top: 8px;
  }

  /* Drag-and-drop page grid */
  .page-grid-drag {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
    gap: 14px;
    max-height: 400px;
    overflow-y: auto;
    margin-bottom: 20px;
    padding: 8px;
  }

  .page-card-drag {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 12px;
    background: var(--bg-tertiary);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-sm);
    cursor: grab;
    transition: all var(--transition-fast);
  }

  .page-card-drag:hover {
    box-shadow: var(--shadow-md);
    transform: translateY(-2px);
  }

  .page-card-drag:active {
    cursor: grabbing;
  }

  .page-card-drag.dragging {
    opacity: 0.4;
    transform: scale(0.95);
  }

  .page-card-drag.drag-over {
    box-shadow: var(--shadow-md), inset 0 0 0 2px var(--accent-color);
    background: var(--accent-light);
  }

  .drag-grip {
    position: absolute;
    top: 6px;
    right: 6px;
    color: var(--text-muted);
    opacity: 0;
    transition: opacity var(--transition-fast);
  }

  .page-card-drag:hover .drag-grip {
    opacity: 1;
  }

  .page-thumbnail {
    width: 80px;
    height: 110px;
    margin-bottom: 10px;
    border-radius: var(--radius-sm);
    overflow: hidden;
    background: var(--bg-secondary);
    box-shadow: var(--shadow-sm);
  }

  .page-thumbnail img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .thumb-placeholder {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.4rem;
    font-weight: 600;
    color: var(--text-muted);
    background: var(--bg-secondary);
  }

  .page-label {
    font-size: 0.8rem;
    font-weight: 500;
    color: var(--text-secondary);
  }

  .modal-actions {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
  }
</style>
