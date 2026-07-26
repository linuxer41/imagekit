<script lang="ts">
  import { onMount } from 'svelte';
  import { getToken, getPanelAccount } from '$lib/api';

  let account = $state<any>(null);
  let loading = $state(true);
  let error = $state('');

  onMount(() => {
    if (!getToken()) { window.location.href = '/login'; return; }
    load();
  });

  async function load() {
    try { account = await getPanelAccount(); }
    catch (err: any) { error = err.message; }
    finally { loading = false; }
  }
</script>

<h1>Account Settings</h1>
<p class="subtitle">View your account information</p>

<div class="card">
  {#if loading}
    <div class="skeleton"><div class="s-line w-1/3"></div><div class="s-line w-1/2"></div></div>
  {:else if error}
    <div class="error-state">{error}</div>
  {:else if account}
    <div class="field"><span class="field-label">Email</span><span class="field-value">{account.email}</span></div>
    <div class="field"><span class="field-label">Name</span><span class="field-value">{account.name || '-'}</span></div>
  {/if}
</div>

<style>
  h1 { font-size: 1.5rem; }
  .subtitle { color: var(--c-muted); font-size: 0.875rem; margin-bottom: 1.5rem; }
  .card { background: var(--c-surface); border: 1px solid var(--c-border); border-radius: var(--radius); padding: 1.25rem; max-width: 500px; }
  .field { display: flex; justify-content: space-between; padding: 0.75rem 0; border-bottom: 1px solid rgba(51,65,85,0.4); }
  .field:last-child { border-bottom: none; }
  .field-label { color: var(--c-dim); font-size: 0.875rem; }
  .field-value { color: var(--c-text); font-size: 0.875rem; }
  .error-state { color: var(--c-danger); }
  .s-line { height: 1rem; background: var(--c-border); border-radius: 0.25rem; margin-bottom: 0.75rem; }
  .w-1\/3 { width: 33%; } .w-1\/2 { width: 50%; }
</style>
