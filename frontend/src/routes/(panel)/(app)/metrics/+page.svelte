<script lang="ts">
  import { onMount } from 'svelte';
  import { getToken, getPanelTenantMetrics } from '$lib/api';

  let metrics = $state<any>(null);
  let days = $state('7');
  let loading = $state(true);
  let error = $state('');

  onMount(() => {
    if (!getToken()) { window.location.href = '/login'; return; }
    load();
  });

  async function load() {
    loading = true;
    try { metrics = await getPanelTenantMetrics(parseInt(days)); }
    catch (err: any) { error = err.message; }
    finally { loading = false; }
  }

  function fmt(n: number): string {
    if (!n) return '0';
    if (n >= 1e9) return (n/1e9).toFixed(1)+'B';
    if (n >= 1e6) return (n/1e6).toFixed(1)+'M';
    if (n >= 1e3) return (n/1e3).toFixed(1)+'K';
    return n.toString();
  }
</script>

<h1>Global Metrics</h1>
<p class="subtitle">Aggregated metrics across all your projects</p>

<div class="period">
  <span>Period:</span>
  <select bind:value={days} onchange={() => load()}>
    <option value="7">7 days</option>
    <option value="30">30 days</option>
    <option value="90">90 days</option>
  </select>
</div>

{#if loading}
  <div class="cards"><div class="card skeleton-card"></div><div class="card skeleton-card"></div><div class="card skeleton-card"></div></div>
{:else if error}
  <div class="error">{error}</div>
{:else if metrics}
  <div class="cards">
    <div class="card stat"><span class="stat-label">Total Requests</span><span class="stat-value">{fmt(metrics.total_requests)}</span></div>
    <div class="card stat"><span class="stat-label">Cache Hit Ratio</span><span class="stat-value">{metrics.cache_hit_ratio?.toFixed(1) ?? '-'}%</span></div>
    <div class="card stat"><span class="stat-label">Total Bandwidth</span><span class="stat-value">{metrics.total_bandwidth_bytes ? (metrics.total_bandwidth_bytes / 1073741824).toFixed(2) + ' GB' : '-'}</span></div>
  </div>

  <div class="card">
    <table>
      <thead><tr><th>Project</th><th>Requests</th><th>Cache Hits</th><th>Bandwidth</th><th>Errors</th></tr></thead>
      <tbody>
        {#if metrics.per_project?.length}
          {#each metrics.per_project as p}
            <tr>
              <td><a href="/projects/{p.slug}">{p.slug}</a></td>
              <td>{fmt(p.requests)}</td>
              <td class="success">{fmt(p.cache_hits)} ({p.requests ? ((p.cache_hits/p.requests)*100).toFixed(1) : 0}%)</td>
              <td>{p.bandwidth_bytes ? (p.bandwidth_bytes/1048576).toFixed(1) + ' MB' : '-'}</td>
              <td class:error={p.errors > 0}>{p.errors ?? 0}</td>
            </tr>
          {/each}
        {:else}
          <tr><td colspan="5" class="empty-cell">No data yet</td></tr>
        {/if}
      </tbody>
    </table>
  </div>
{/if}

<style>
  h1 { font-size: 1.5rem; }
  .subtitle { color: var(--c-muted); font-size: 0.875rem; margin-bottom: 1rem; }
  .period { display: flex; align-items: center; gap: 0.5rem; font-size: 0.875rem; color: var(--c-muted); margin-bottom: 1rem; }
  .period select { padding: 0.35rem 0.5rem; border: 1px solid var(--c-border); border-radius: var(--radius-sm); background: var(--c-surface); color: var(--c-text); }
  .error { background: rgba(239,68,68,0.15); color: var(--c-danger); padding: 0.75rem; border-radius: var(--radius-sm); font-size: 0.875rem; }
  .cards { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 1rem; margin-bottom: 1.5rem; }
  .card { background: var(--c-surface); border: 1px solid var(--c-border); border-radius: var(--radius); padding: 1.25rem; }
  .skeleton-card { height: 80px; animation: pulse 2s infinite; }
  @keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.4; } }
  .stat { display: flex; flex-direction: column; gap: 0.25rem; }
  .stat-label { font-size: 0.8rem; color: var(--c-dim); text-transform: uppercase; letter-spacing: 0.05em; }
  .stat-value { font-size: 1.5rem; font-weight: 700; }
  .stat-value.error { color: var(--c-danger); }
  .success { color: var(--c-success); }
  table { width: 100%; font-size: 0.8rem; }
  th { text-align: left; padding: 0.5rem; color: var(--c-dim); font-weight: 500; border-bottom: 1px solid var(--c-border); }
  td { padding: 0.5rem; border-bottom: 1px solid rgba(51,65,85,0.4); }
  .empty-cell { text-align: center; color: var(--c-dim); padding: 2rem !important; }
</style>
