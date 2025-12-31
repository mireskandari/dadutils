<script>
  import { createEventDispatcher } from 'svelte';

  export let files = [];
  export let selectedIds = [];
  export let thumbnails = {}; // Map of file.id -> thumbnail data URL

  const dispatch = createEventDispatcher();

  let draggedIndex = null;

  function toggleSelection(id) {
    if (selectedIds.includes(id)) {
      selectedIds = selectedIds.filter(i => i !== id);
    } else {
      selectedIds = [...selectedIds, id];
    }
    dispatch('selectionChange', { selectedIds });
  }

  function removeFile(id) {
    dispatch('remove', { id });
  }

  function handleDragStart(e, index) {
    draggedIndex = index;
    e.dataTransfer.effectAllowed = 'move';
  }

  function handleDragOver(e, index) {
    e.preventDefault();
    if (draggedIndex === null || draggedIndex === index) return;

    // Reorder files
    const newFiles = [...files];
    const [draggedItem] = newFiles.splice(draggedIndex, 1);
    newFiles.splice(index, 0, draggedItem);

    dispatch('reorder', { files: newFiles });
    draggedIndex = index;
  }

  function handleDragEnd() {
    draggedIndex = null;
  }

  function getTotalPages() {
    return files.reduce((sum, f) => sum + (f.pageCount || 0), 0);
  }

  function getTotalSize() {
    return files.reduce((sum, f) => sum + (f.size || 0), 0);
  }

  function formatSize(bytes) {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
  }
</script>

<div class="file-list">
  {#each files as file, index (file.id)}
    <div
      class="file-item"
      class:selected={selectedIds.includes(file.id)}
      class:dragging={draggedIndex === index}
      draggable="true"
      on:dragstart={(e) => handleDragStart(e, index)}
      on:dragover={(e) => handleDragOver(e, index)}
      on:dragend={handleDragEnd}
    >
      <div class="drag-handle" title="Drag to reorder">
        <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
          <circle cx="5" cy="3" r="1.5"/>
          <circle cx="11" cy="3" r="1.5"/>
          <circle cx="5" cy="8" r="1.5"/>
          <circle cx="11" cy="8" r="1.5"/>
          <circle cx="5" cy="13" r="1.5"/>
          <circle cx="11" cy="13" r="1.5"/>
        </svg>
      </div>

      <label class="checkbox-wrapper">
        <input
          type="checkbox"
          checked={selectedIds.includes(file.id)}
          on:change={() => toggleSelection(file.id)}
        />
        <span class="checkmark"></span>
      </label>

      <div class="thumbnail-wrapper">
        {#if thumbnails[file.id]}
          <img src={thumbnails[file.id]} alt="Preview" class="thumbnail" />
        {:else}
          <div class="thumbnail-placeholder">
            <span>PDF</span>
          </div>
        {/if}
      </div>

      <div class="file-info">
        <span class="file-name">{file.name}</span>
        <span class="file-meta">
          {file.pageCount} page{file.pageCount !== 1 ? 's' : ''} • {file.sizeText}
        </span>
      </div>

      <button class="remove-btn" on:click={() => removeFile(file.id)} title="Remove">
        ✕
      </button>
    </div>
  {/each}

  {#if files.length > 0}
    <div class="total-row">
      <span>Total: {getTotalPages()} pages • {formatSize(getTotalSize())}</span>
    </div>
  {/if}
</div>

<style>
  .file-list {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .file-item {
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 14px 16px;
    background: var(--bg-secondary);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-sm);
    transition: all var(--transition-fast);
  }

  .file-item:hover {
    box-shadow: var(--shadow-md);
  }

  .file-item.selected {
    background: var(--accent-light);
    box-shadow: var(--shadow-md), inset 0 0 0 2px var(--accent-color);
  }

  .file-item.dragging {
    opacity: 0.4;
    transform: scale(0.98);
  }

  .drag-handle {
    cursor: grab;
    color: var(--text-muted);
    user-select: none;
    padding: 4px;
    border-radius: 4px;
    transition: all var(--transition-fast);
  }

  .drag-handle:hover {
    color: var(--accent-color);
    background: var(--accent-light);
  }

  .drag-handle:active {
    cursor: grabbing;
  }

  .checkbox-wrapper {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 22px;
    height: 22px;
    cursor: pointer;
    flex-shrink: 0;
  }

  .checkbox-wrapper input {
    position: absolute;
    opacity: 0;
    width: 100%;
    height: 100%;
    cursor: pointer;
  }

  .checkmark {
    width: 20px;
    height: 20px;
    border: 2px solid var(--border-color);
    border-radius: 6px;
    background: var(--bg-secondary);
    transition: all var(--transition-fast);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .checkbox-wrapper:hover .checkmark {
    border-color: var(--accent-color);
  }

  .checkbox-wrapper input:checked ~ .checkmark {
    background: var(--accent-color);
    border-color: var(--accent-color);
  }

  .checkmark:after {
    content: '';
    display: none;
    width: 5px;
    height: 10px;
    border: solid white;
    border-width: 0 2px 2px 0;
    transform: rotate(45deg);
    margin-bottom: 2px;
  }

  .checkbox-wrapper input:checked ~ .checkmark:after {
    display: block;
  }

  .thumbnail-wrapper {
    width: 40px;
    height: 56px;
    flex-shrink: 0;
  }

  .thumbnail {
    width: 100%;
    height: 100%;
    object-fit: cover;
    border-radius: 6px;
    background: var(--bg-tertiary);
    box-shadow: var(--shadow-sm);
  }

  .thumbnail-placeholder {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--bg-tertiary);
    border-radius: 6px;
    border: 1px solid var(--border-color);
  }

  .thumbnail-placeholder span {
    font-size: 0.65rem;
    color: var(--text-muted);
    font-weight: 600;
  }

  .file-info {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 3px;
    min-width: 0;
  }

  .file-name {
    font-weight: 500;
    color: var(--text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .file-meta {
    font-size: 0.8rem;
    color: var(--text-secondary);
  }

  .remove-btn {
    width: 28px;
    height: 28px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: transparent;
    border: none;
    border-radius: var(--radius-sm);
    color: var(--text-muted);
    cursor: pointer;
    opacity: 0;
    transition: all var(--transition-fast);
  }

  .file-item:hover .remove-btn {
    opacity: 1;
  }

  .remove-btn:hover {
    background: var(--error-light);
    color: var(--error-color);
  }

  .total-row {
    padding: 14px 16px;
    text-align: right;
    font-size: 0.9rem;
    color: var(--text-secondary);
    background: var(--bg-tertiary);
    border-radius: var(--radius-sm);
    margin-top: 4px;
  }
</style>
