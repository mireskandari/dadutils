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
  <h1 class="title">Dad PDF Stuff</h1>
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
    padding: 28px 24px;
    max-width: 640px;
    margin: 0 auto;
  }

  .title {
    font-size: 2.2rem;
    font-weight: 600;
    margin: 0 0 8px;
    color: var(--text-primary);
    letter-spacing: -0.02em;
  }

  .subtitle {
    color: var(--text-secondary);
    margin: 0 0 48px;
    font-size: 1.1rem;
  }

  .tools-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 20px;
  }

  .tool-card {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
    padding: 40px 24px;
    background: var(--bg-secondary);
    border: none;
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-sm);
    cursor: pointer;
    transition: all var(--transition-normal);
  }

  .tool-card.available:hover {
    transform: translateY(-4px);
    box-shadow: var(--shadow-hover);
  }

  .tool-card.coming-soon {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .tool-icon {
    font-size: 2.8rem;
    margin-bottom: 4px;
  }

  .tool-name {
    font-size: 1.25rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  .tool-description {
    font-size: 0.9rem;
    color: var(--text-secondary);
    text-align: center;
  }

  .coming-soon-badge {
    position: absolute;
    top: 12px;
    right: 12px;
    padding: 5px 10px;
    background: var(--bg-tertiary);
    border-radius: var(--radius-sm);
    font-size: 0.7rem;
    font-weight: 500;
    color: var(--text-muted);
  }

  @media (max-width: 480px) {
    .home {
      padding: 40px 20px;
    }

    .tools-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
