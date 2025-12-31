<script>
  import { createEventDispatcher } from 'svelte';

  const dispatch = createEventDispatcher();

  const tools = [
    {
      id: 'compress',
      name: 'Compress',
      description: 'Reduce PDF file size',
      icon: '📦',
      available: true
    },
    {
      id: 'combine',
      name: 'Combine',
      description: 'Merge multiple PDFs',
      icon: '📑',
      available: true
    },
    {
      id: 'split',
      name: 'Split',
      description: 'Extract pages from PDF',
      icon: '✂️',
      available: false
    },
    {
      id: 'rotate',
      name: 'Rotate',
      description: 'Rotate PDF pages',
      icon: '🔄',
      available: false
    }
  ];

  function selectTool(tool) {
    if (tool.available) {
      dispatch('navigate', { view: tool.id });
    }
  }
</script>

<div class="home">
  <h1 class="title">Dad's PDF Tools</h1>
  <p class="subtitle">What would you like to do?</p>

  <div class="tools-grid">
    {#each tools as tool}
      <button
        class="tool-card"
        class:available={tool.available}
        class:coming-soon={!tool.available}
        on:click={() => selectTool(tool)}
        disabled={!tool.available}
      >
        <span class="tool-icon">{tool.icon}</span>
        <span class="tool-name">{tool.name}</span>
        <span class="tool-description">{tool.description}</span>
        {#if !tool.available}
          <span class="coming-soon-badge">Coming Soon</span>
        {/if}
      </button>
    {/each}
  </div>
</div>

<style>
  .home {
    padding: 40px 20px;
    max-width: 600px;
    margin: 0 auto;
  }

  .title {
    font-size: 2rem;
    font-weight: 600;
    margin: 0 0 8px;
    color: var(--text-primary, #fff);
  }

  .subtitle {
    color: var(--text-secondary, #888);
    margin: 0 0 40px;
    font-size: 1.1rem;
  }

  .tools-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 16px;
  }

  .tool-card {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    padding: 32px 20px;
    background: var(--bg-secondary, #1a1a2e);
    border: 1px solid var(--border-color, #333);
    border-radius: 12px;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .tool-card.available:hover {
    background: var(--bg-hover, #232338);
    border-color: var(--accent-color, #646cff);
    transform: translateY(-2px);
  }

  .tool-card.coming-soon {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .tool-icon {
    font-size: 2.5rem;
  }

  .tool-name {
    font-size: 1.2rem;
    font-weight: 600;
    color: var(--text-primary, #fff);
  }

  .tool-description {
    font-size: 0.85rem;
    color: var(--text-secondary, #888);
  }

  .coming-soon-badge {
    position: absolute;
    top: 8px;
    right: 8px;
    padding: 4px 8px;
    background: var(--bg-tertiary, #333);
    border-radius: 4px;
    font-size: 0.7rem;
    color: var(--text-secondary, #888);
  }

  @media (max-width: 480px) {
    .tools-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
