<script lang="ts">
  import { onMount } from 'svelte';
  import { getToken, listAdminTenants } from '$lib/api';
  import type { Tenant } from '$lib/types';

  let tenants = $state<Tenant[]>([]);
  let loading = $state(true);
  let error = $state('');

  onMount(() => {
    if (!getToken()) { window.location.href = '/admin/login'; return; }
    load();
  });

  async function load() {
    try { tenants = await listAdminTenants(); }
    catch (err: any) { error = err.message; }
    finally { loading = false; }
  }
</script>

<h1>Tenants</h1>
<p class="subtitle">All registered tenants</p>

<div class="card">
  {#if loading}
    <div class="skeleton"><div class="s-line w-1/4"></div><div class="s-line w-1/2"></div><div class="s-line w-2/3"></div></div>
  {:else if error}
    <div class="error">{error}</div>
  {:else if tenants.length === 0}
    <div class="empty">No tenants yet.</div>
  {:else}
    <table>
      <thead><tr><th>ID</th><th>Name</th><th>Email</th><th>Active</th><th>Created</th></tr></thead>
      <tbody>
        {#each tenants as t}
          <tr>
            <td>{t.id}</td>
            <td>{t.name}</td>
            <td>{t.email}</td>
            <td class:success={t.is_active} class:danger={!t.is_active}>{t.is_active ? 'Yes' : 'No'}</td>
            <td>{t.created_at?.slice(0, 10)}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>

<style>
  h1 { font-size: 1.5rem; }
  .subtitle { color: var(--c-muted); font-size: 0.875rem; margin-bottom: 1rem; }
  .card { background: var(--c-surface); border: 1px solid var(--c-border); border-radius: var(--radius); overflow: hidden; }
  .error { padding: 1rem; color: var(--c-danger); }
  .empty { padding: 2rem; text-align: center; color: var(--c-muted); }
  .success { color: var(--c-success); }
  .danger { color: var(--c-danger); }
  table { width: 100%; font-size: 0.875rem; }
  th { text-align: left; padding: 0.75rem; color: var(--c-dim); font-weight: 500; border-bottom: 1px solid var(--c-border); }
  td { padding: 0.75rem; border-bottom: 1px solid rgba(51,65,85,0.4); }
  .skeleton { padding: 1rem; }
  .s-line { height: 1rem; background: var(--c-border); border-radius: 0.25rem; margin-bottom: 0.75rem; }
  .w-1\/4 { width: 25%; } .w-1\/2 { width: 50%; } .w-2\/3 { width: 66%; }
</style>
