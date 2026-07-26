<script lang="ts">
  import { onMount } from 'svelte';
  import { getToken, listAdminInvitations, createAdminInvitation } from '$lib/api';
  import type { Invitation } from '$lib/types';

  let invitations = $state<Invitation[]>([]);
  let loading = $state(true);
  let creating = $state(false);
  let error = $state('');

  onMount(() => {
    if (!getToken()) { window.location.href = '/admin/login'; return; }
    load();
  });

  async function load() {
    try { invitations = await listAdminInvitations(); }
    catch (err: any) { error = err.message; }
    finally { loading = false; }
  }

  async function generate() {
    creating = true;
    try {
      await createAdminInvitation();
      await load();
    } catch (err: any) { error = err.message; }
    finally { creating = false; }
  }
</script>

<div class="page-header">
  <div>
    <h1>Invitations</h1>
    <p class="subtitle">Manage invitation codes for tenant registration</p>
  </div>
  <button onclick={generate} disabled={creating} class="btn">{creating ? '...' : '+ Generate Code'}</button>
</div>

<div class="card">
  {#if loading}
    <div class="skeleton"><div class="s-line w-1/4"></div><div class="s-line w-1/2"></div></div>
  {:else if error}
    <div class="error">{error}</div>
  {:else if invitations.length === 0}
    <div class="empty">No invitations yet.</div>
  {:else}
    <table>
      <thead><tr><th>Code</th><th>Used</th><th>Created</th><th>Used By</th></tr></thead>
      <tbody>
        {#each invitations as inv}
          <tr>
            <td class="code">{inv.code}</td>
            <td class:success={inv.is_used}>{inv.is_used ? 'Yes' : 'No'}</td>
            <td>{inv.created_at?.slice(0, 10)}</td>
            <td>{inv.used_by ?? '-'}</td>
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
  .btn { padding: 0.5rem 1rem; background: var(--c-primary); color: #fff; border: none; border-radius: var(--radius-sm); font-size: 0.875rem; font-weight: 600; cursor: pointer; }
  .btn:hover { background: var(--c-primary-hover); }
  .btn:disabled { opacity: 0.5; cursor: not-allowed; }
  .card { background: var(--c-surface); border: 1px solid var(--c-border); border-radius: var(--radius); overflow: hidden; }
  .error { padding: 1rem; color: var(--c-danger); }
  .empty { padding: 2rem; text-align: center; color: var(--c-muted); }
  .success { color: var(--c-success); }
  .code { font-family: var(--font-mono); font-weight: 600; }
  table { width: 100%; font-size: 0.875rem; }
  th { text-align: left; padding: 0.75rem; color: var(--c-dim); font-weight: 500; border-bottom: 1px solid var(--c-border); }
  td { padding: 0.75rem; border-bottom: 1px solid rgba(51,65,85,0.4); }
  .skeleton { padding: 1rem; }
  .s-line { height: 1rem; background: var(--c-border); border-radius: 0.25rem; margin-bottom: 0.75rem; }
  .w-1\/4 { width: 25%; } .w-1\/2 { width: 50%; }
</style>
