<script>
  import Home from './lib/Home.svelte';
  import Compress from './lib/Compress.svelte';
  import Combine from './lib/Combine.svelte';

  let currentView = 'home';

  function handleNavigate(event) {
    currentView = event.detail.view;
  }
</script>

<main>
  {#if currentView === 'home'}
    <Home on:navigate={handleNavigate} />
  {:else if currentView === 'compress'}
    <Compress on:navigate={handleNavigate} />
  {:else if currentView === 'combine'}
    <Combine on:navigate={handleNavigate} />
  {:else}
    <Home on:navigate={handleNavigate} />
  {/if}
</main>

<style>
  main {
    min-height: 100vh;
    padding: 20px;
  }
</style>
