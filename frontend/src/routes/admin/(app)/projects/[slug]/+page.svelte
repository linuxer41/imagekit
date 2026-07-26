<script lang="ts">
  import { onMount } from 'svelte';
  import { getToken } from '$lib/api';

  let { params } = $props();
  let slug = $derived(params.slug);
  let loading = $state(true);

  onMount(() => {
    if (!getToken()) { window.location.href = '/admin/login'; return; }
    loading = false;
  });
</script>

<div class="breadcrumb"><a href="/admin/projects">Projects</a> <span class="sep">›</span> <span>{slug}</span></div>
<h1>{slug}</h1>
<p class="subtitle">Project details</p>

<div class="card">
  {#if loading}<div class="skeleton"><div class="s-line w-1/3"></div><div class="s-line w-1/2"></div></div>
  {:else}<p class="text-muted">Select a period to view metrics.</p>{/if}
</div>

<style>
  .breadcrumb { font-size: 0.875rem; color: var(--c-muted); margin-bottom: 0.5rem; }
  .sep { margin: 0 0.5rem; }
  h1 { font-size: 1.5rem; }
  .subtitle { color: var(--c-muted); font-size: 0.875rem; margin-bottom: 1rem; }
  .card { background: var(--c-surface); border: 1px solid var(--c-border); border-radius: var(--radius); padding: 1.25rem; }
  .text-muted { color: var(--c-dim); }
  .s-line { height: 1rem; background: var(--c-border); border-radius: 0.25rem; margin-bottom: 0.75rem; }
  .w-1\/3 { width: 33%; } .w-1\/2 { width: 50%; }
</style>
