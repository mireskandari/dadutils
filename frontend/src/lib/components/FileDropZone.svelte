<script>
  import { createEventDispatcher } from 'svelte';

  export let multiple = false;
  export let accept = ".pdf";
  export let disabled = false;

  const dispatch = createEventDispatcher();

  let isDragging = false;
  let fileInput;

  function handleDragOver(e) {
    if (disabled) return;
    e.preventDefault();
    isDragging = true;
  }

  function handleDragLeave() {
    isDragging = false;
  }

  function handleDrop(e) {
    if (disabled) return;
    e.preventDefault();
    isDragging = false;

    const files = Array.from(e.dataTransfer.files);
    const pdfFiles = files.filter(f => f.name.toLowerCase().endsWith('.pdf'));

    if (pdfFiles.length > 0) {
      // Get file paths - in Wails, dropped files have a path property
      const paths = pdfFiles.map(f => f.path).filter(Boolean);
      if (paths.length > 0) {
        dispatch('files', { paths });
      }
    }
  }

  function handleClick() {
    if (!disabled) {
      dispatch('browse');
    }
  }

  function handleFileSelect(e) {
    const files = Array.from(e.target.files);
    const paths = files.map(f => f.path).filter(Boolean);
    if (paths.length > 0) {
      dispatch('files', { paths });
    }
    // Reset input so same file can be selected again
    e.target.value = '';
  }
</script>

<div
  class="dropzone"
  class:dragging={isDragging}
  class:disabled={disabled}
  on:dragover={handleDragOver}
  on:dragleave={handleDragLeave}
  on:drop={handleDrop}
  on:click={handleClick}
  role="button"
  tabindex="0"
  on:keypress={(e) => e.key === 'Enter' && handleClick()}
>
  <input
    type="file"
    bind:this={fileInput}
    {accept}
    {multiple}
    on:change={handleFileSelect}
    style="display: none"
  />

  <div class="dropzone-content">
    <div class="dropzone-icon">
      <slot name="icon">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M9 12h6m-3-3v6m-7 4h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
        </svg>
      </slot>
    </div>
    <div class="dropzone-text">
      <slot>
        Drag & drop PDF{multiple ? 's' : ''} here<br/>
        <span class="dropzone-or">or</span>
      </slot>
    </div>
    <button class="dropzone-btn" on:click|stopPropagation={() => dispatch('browse')} {disabled}>
      {multiple ? 'Select Files' : 'Select PDF'}
    </button>
  </div>
</div>

<style>
  .dropzone {
    border: 2px dashed var(--border-color, #444);
    border-radius: 12px;
    padding: 40px 20px;
    text-align: center;
    cursor: pointer;
    transition: all 0.2s ease;
    background: var(--bg-secondary, #1a1a2e);
  }

  .dropzone:hover:not(.disabled) {
    border-color: var(--accent-color, #646cff);
    background: var(--bg-hover, #232338);
  }

  .dropzone.dragging {
    border-color: var(--accent-color, #646cff);
    background: var(--bg-hover, #232338);
    transform: scale(1.01);
  }

  .dropzone.disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .dropzone-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
  }

  .dropzone-icon {
    color: var(--text-secondary, #888);
  }

  .dropzone-text {
    color: var(--text-secondary, #888);
    line-height: 1.6;
  }

  .dropzone-or {
    font-size: 0.85em;
    opacity: 0.7;
  }

  .dropzone-btn {
    margin-top: 8px;
    padding: 10px 24px;
    border: none;
    border-radius: 6px;
    background: var(--accent-color, #646cff);
    color: white;
    font-size: 0.95rem;
    cursor: pointer;
    transition: background 0.2s;
  }

  .dropzone-btn:hover:not(:disabled) {
    background: var(--accent-hover, #535bf2);
  }

  .dropzone-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
