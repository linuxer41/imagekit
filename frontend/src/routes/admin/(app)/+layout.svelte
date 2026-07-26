<script lang="ts">
  import { onMount } from 'svelte';
  import { getToken, clearTokens } from '$lib/api';

  let { children } = $props();
  let loggedIn = $state(false);
  let mobileOpen = $state(false);

  onMount(() => { loggedIn = !!getToken(); });

  function logout() {
    clearTokens();
    window.location.href = '/admin/login';
  }

  const navItems = [
    { href: '/admin/tenants', label: 'Tenants' },
    { href: '/admin/projects', label: 'Projects' },
    { href: '/admin/invitations', label: 'Invitations' },
  ];
</script>

<div class="layout">
  <aside class="sidebar" class:open={mobileOpen}>
    <div class="sidebar-header">
      <a href="/admin" class="logo">Admin Panel</a>
      <button class="close-btn" onclick={() => mobileOpen = false}>✕</button>
    </div>
    <nav class="sidebar-nav">
      {#each navItems as item}
        <a href={item.href} class="nav-link">{item.label}</a>
      {/each}
      {#if loggedIn}
        <button onclick={logout} class="nav-link logout">Logout</button>
      {/if}
    </nav>
  </aside>
  <div class="overlay" class:visible={mobileOpen} onclick={() => mobileOpen = false}></div>
  <main class="main">
    <button class="mobile-toggle" onclick={() => mobileOpen = true}>☰</button>
    {@render children()}
  </main>
</div>

<style>
  .layout { display: flex; min-height: 100vh; }
  .sidebar { width: 240px; background: var(--c-surface); border-right: 1px solid var(--c-border); display: flex; flex-direction: column; flex-shrink: 0; }
  .sidebar-header { display: flex; align-items: center; justify-content: space-between; padding: 1rem; border-bottom: 1px solid var(--c-border); }
  .logo { font-weight: 700; font-size: 1.1rem; color: var(--c-text) !important; text-decoration: none !important; }
  .sidebar-nav { padding: 0.5rem; display: flex; flex-direction: column; gap: 0.25rem; flex: 1; }
  .nav-link { display: flex; align-items: center; gap: 0.5rem; padding: 0.625rem 0.75rem; border-radius: var(--radius-sm); color: var(--c-muted) !important; font-size: 0.9rem; text-decoration: none !important; transition: background 0.15s; }
  .nav-link:hover { background: var(--c-border); color: var(--c-text) !important; }
  .logout { margin-top: auto; color: var(--c-danger) !important; background: none; border: none; cursor: pointer; width: 100%; }
  .main { flex: 1; padding: 1.5rem; min-width: 0; }
  .close-btn { display: none; background: none; border: none; color: var(--c-muted); font-size: 1.25rem; cursor: pointer; }
  .mobile-toggle { display: none; background: none; border: none; color: var(--c-text); font-size: 1.5rem; cursor: pointer; margin-bottom: 1rem; }
  .overlay { display: none; }
  @media (max-width: 768px) {
    .sidebar { position: fixed; top: 0; left: 0; bottom: 0; z-index: 200; transform: translateX(-100%); transition: transform 0.2s; }
    .sidebar.open { transform: translateX(0); }
    .close-btn { display: block; }
    .mobile-toggle { display: inline-block; }
    .overlay.visible { display: block; position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 150; }
  }
</style>
