<script lang="ts">
  import { onMount } from 'svelte';
  import { getToken, clearTokens } from '$lib/api';

  let { children } = $props();
  let loggedIn = $state(false);
  let mobileOpen = $state(false);

  onMount(() => { loggedIn = !!getToken(); });

  function logout() {
    clearTokens();
    loggedIn = false;
    window.location.href = '/login';
  }
</script>

<div class="layout">
  <nav class="nav">
    <div class="nav-inner">
      <a href="/" class="logo">ImageKit</a>
      <button class="mobile-toggle" onclick={() => mobileOpen = !mobileOpen}>
        <span class="bar"></span><span class="bar"></span><span class="bar"></span>
      </button>
      <div class="nav-links" class:open={mobileOpen}>
        {#if loggedIn}
          <a href="/projects" class="nav-link">Projects</a>
          <a href="/metrics" class="nav-link">Metrics</a>
          <a href="/account" class="nav-link">Account</a>
          <button onclick={logout} class="nav-link logout">Logout</button>
        {:else}
          <a href="/login" class="nav-link">Login</a>
          <a href="/register" class="nav-link">Register</a>
        {/if}
      </div>
    </div>
  </nav>
  <main class="main">
    {@render children()}
  </main>
</div>

<style>
  .layout { min-height: 100vh; display: flex; flex-direction: column; }
  .nav { background: var(--c-surface); border-bottom: 1px solid var(--c-border); position: sticky; top: 0; z-index: 100; }
  .nav-inner { max-width: 1200px; margin: 0 auto; display: flex; align-items: center; justify-content: space-between; padding: 0 1rem; height: 56px; }
  .logo { font-weight: 700; font-size: 1.1rem; color: var(--c-text) !important; text-decoration: none !important; }
  .nav-links { display: flex; align-items: center; gap: 0.25rem; }
  .nav-link { padding: 0.5rem 0.75rem; border-radius: var(--radius-sm); color: var(--c-muted) !important; text-decoration: none !important; font-size: 0.9rem; transition: background 0.15s; }
  .nav-link:hover { background: var(--c-border); color: var(--c-text) !important; }
  .logout { color: var(--c-danger) !important; background: none; border: none; cursor: pointer; }
  .main { flex: 1; max-width: 1200px; width: 100%; margin: 0 auto; padding: 1.5rem 1rem; }
  .mobile-toggle { display: none; background: none; border: none; cursor: pointer; padding: 0.5rem; }
  .bar { display: block; width: 20px; height: 2px; background: var(--c-text); margin: 4px 0; border-radius: 1px; }
  @media (max-width: 640px) {
    .mobile-toggle { display: block; }
    .nav-links { display: none; position: absolute; top: 56px; left: 0; right: 0; background: var(--c-surface); border-bottom: 1px solid var(--c-border); flex-direction: column; padding: 0.5rem; }
    .nav-links.open { display: flex; }
  }
</style>
