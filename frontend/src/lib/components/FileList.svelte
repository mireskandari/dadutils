<script>
  import { createEventDispatcher } from 'svelte';

  export let files = [];
  export let selectedIds = [];

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
      <div class="drag-handle" title="Drag to reorder">≡</div>

      <label class="checkbox-wrapper">
        <input
          type="checkbox"
          checked={selectedIds.includes(file.id)}
          on:change={() => toggleSelection(file.id)}
        />
        <span class="checkmark"></span>
      </label>

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
    gap: 8px;
  }

  .file-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px;
    background: var(--bg-secondary, #1a1a2e);
    border-radius: 8px;
    border: 1px solid transparent;
    transition: all 0.2s;
  }

  .file-item.selected {
    border-color: var(--accent-color, #646cff);
    background: var(--bg-hover, #232338);
  }

  .file-item.dragging {
    opacity: 0.5;
  }

  .drag-handle {
    cursor: grab;
    color: var(--text-secondary, #888);
    font-size: 1.2rem;
    user-select: none;
  }

  .drag-handle:active {
    cursor: grabbing;
  }

  .checkbox-wrapper {
    position: relative;
    width: 20px;
    height: 20px;
    cursor: pointer;
  }

  .checkbox-wrapper input {
    position: absolute;
    opacity: 0;
    width: 100%;
    height: 100%;
    cursor: pointer;
  }

  .checkmark {
    position: absolute;
    top: 0;
    left: 0;
    width: 18px;
    height: 18px;
    border: 2px solid var(--border-color, #444);
    border-radius: 4px;
    background: var(--bg-tertiary, #232338);
    transition: all 0.2s;
  }

  .checkbox-wrapper input:checked ~ .checkmark {
    background: var(--accent-color, #646cff);
    border-color: var(--accent-color, #646cff);
  }

  .checkmark:after {
    content: '';
    position: absolute;
    display: none;
    left: 5px;
    top: 2px;
    width: 5px;
    height: 9px;
    border: solid white;
    border-width: 0 2px 2px 0;
    transform: rotate(45deg);
  }

  .checkbox-wrapper input:checked ~ .checkmark:after {
    display: block;
  }

  .file-info {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }

  .file-name {
    font-weight: 500;
    color: var(--text-primary, #fff);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .file-meta {
    font-size: 0.8rem;
    color: var(--text-secondary, #888);
  }

  .remove-btn {
    width: 24px;
    height: 24px;
    background: transparent;
    border: none;
    border-radius: 4px;
    color: var(--text-secondary, #888);
    cursor: pointer;
    opacity: 0.5;
    transition: all 0.2s;
  }

  .file-item:hover .remove-btn {
    opacity: 1;
  }

  .remove-btn:hover {
    background: #c00;
    color: white;
  }

  .total-row {
    padding: 12px;
    text-align: right;
    font-size: 0.9rem;
    color: var(--text-secondary, #888);
    border-top: 1px solid var(--border-color, #333);
    margin-top: 4px;
  }
</style>
