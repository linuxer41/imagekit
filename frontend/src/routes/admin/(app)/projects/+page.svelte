<script lang="ts">
  import { onMount } from 'svelte';
  import { getToken, listAdminProjects } from '$lib/api';
  import type { Project } from '$lib/types';

  let projects = $state<Project[]>([]);
  let loading = $state(true);
  let error = $state('');

  onMount(() => {
    if (!getToken()) { window.location.href = '/admin/login'; return; }
    load();
  });

  async function load() {
    try { projects = await listAdminProjects(); }
    catch (err: any) { error = err.message; }
    finally { loading = false; }
  }
</script>

<div class="page-header">
  <div>
    <h1>All Projects</h1>
    <p class="subtitle">Manage all ImageKit projects</p>
  </div>
  <a href="/admin/projects/new" class="btn">+ New</a>
</div>

<div class="card">
  {#if loading}
    <div class="skeleton"><div class="s-line w-1/3"></div><div class="s-line w-1/2"></div></div>
  {:else if error}
    <div class="error">{error}</div>
  {:else if projects.length === 0}
    <div class="empty">No projects yet.</div>
  {:else}
    <table>
      <thead><tr><th>ID</th><th>Slug</th><th>Name</th><th>Tenant</th><th>Provider</th><th>Format</th></tr></thead>
      <tbody>
        {#each projects as p}
          <tr>
            <td>{p.id}</td>
            <td class="slug"><a href="/admin/projects/{p.slug}">{p.slug}</a></td>
            <td>{p.name || '-'}</td>
            <td>{p.tenant_id}</td>
            <td>{p.provider}</td>
            <td>{p.default_format}</td>
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
  .card { background: var(--c-surface); border: 1px solid var(--c-border); border-radius: var(--radius); overflow: hidden; }
  .error { padding: 1rem; color: var(--c-danger); }
  .empty { padding: 2rem; text-align: center; color: var(--c-muted); }
  .slug a { font-family: var(--font-mono); }
  table { width: 100%; font-size: 0.875rem; }
  th { text-align: left; padding: 0.75rem; color: var(--c-dim); font-weight: 500; border-bottom: 1px solid var(--c-border); }
  td { padding: 0.75rem; border-bottom: 1px solid rgba(51,65,85,0.4); }
  .skeleton { padding: 1rem; }
  .s-line { height: 1rem; background: var(--c-border); border-radius: 0.25rem; margin-bottom: 0.75rem; }
  .w-1\/3 { width: 33%; } .w-1\/2 { width: 50%; }
</style>
