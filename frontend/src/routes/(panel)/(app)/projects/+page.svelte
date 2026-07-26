<script lang="ts">
  import { onMount } from 'svelte';
  import { getToken, listPanelProjects } from '$lib/api';

  let projects = $state<any[]>([]);
  let loading = $state(true);
  let error = $state('');

  onMount(() => {
    if (!getToken()) { window.location.href = '/login'; return; }
    load();
  });

  async function load() {
    try { projects = await listPanelProjects(); }
    catch (err: any) { error = err.message; }
    finally { loading = false; }
  }
</script>

<div class="page-header">
  <div>
    <h1>My Projects</h1>
    <p class="subtitle">Manage your ImageKit projects</p>
  </div>
  <a href="/projects/new" class="btn">+ New</a>
</div>

<div class="card">
  {#if loading}
    <div class="skeleton"><div class="s-line w-1/4"></div><div class="s-line w-3/4"></div><div class="s-line w-1/2"></div></div>
  {:else if error}
    <div class="error-state">{error}</div>
  {:else if projects.length === 0}
    <div class="empty">No projects yet. <a href="/projects/new">Create one</a></div>
  {:else}
    <table>
      <thead><tr><th>Slug</th><th>Name</th><th>Format</th><th>RPS</th></tr></thead>
      <tbody>
        {#each projects as p}
          <tr>
            <td class="slug"><a href="/projects/{p.slug}">{p.slug}</a></td>
            <td>{p.name || '-'}</td>
            <td>{p.default_format}</td>
            <td>{p.rps}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>

<style>
  .page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 1.5rem; }
  h1 { font-size: 1.5rem; }
  .subtitle { color: var(--c-muted); font-size: 0.875rem; margin-top: 0.25rem; }
  .btn { display: inline-flex; align-items: center; gap: 0.35rem; padding: 0.5rem 1rem; background: var(--c-primary); color: #fff; border-radius: var(--radius-sm); font-size: 0.875rem; font-weight: 600; text-decoration: none !important; }
  .btn:hover { background: var(--c-primary-hover); }
  .card { background: var(--c-surface); border: 1px solid var(--c-border); border-radius: var(--radius); overflow: hidden; }
  table { width: 100%; font-size: 0.875rem; }
  th { text-align: left; padding: 0.75rem; color: var(--c-dim); font-weight: 500; border-bottom: 1px solid var(--c-border); }
  td { padding: 0.75rem; border-bottom: 1px solid rgba(51,65,85,0.4); }
  .slug a { font-family: var(--font-mono); }
  .empty { padding: 2rem; text-align: center; color: var(--c-muted); }
  .error-state { padding: 1rem; color: var(--c-danger); }
  .skeleton { padding: 1rem; }
  .s-line { height: 1rem; background: var(--c-border); border-radius: 0.25rem; margin-bottom: 0.75rem; }
  .w-1\/4 { width: 25%; } .w-3\/4 { width: 75%; } .w-1\/2 { width: 50%; }
</style>
