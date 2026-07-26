<script lang="ts">
  import { onMount } from 'svelte';
  import { getToken, getPanelProject, getPanelProjectMetrics, getPanelProjectSummary } from '$lib/api';

  let { params } = $props();
  let slug = $derived(params.slug);

  let proj: any = $state(null);
  let metrics: any = $state(null);
  let summary: any = $state(null);
  let loading = $state(true);
  let days = $state('30');
  let error = $state('');

  onMount(() => {
    if (!getToken()) { window.location.href = '/login'; return; }
    load();
  });

  async function load() {
    try {
      const list = await getPanelProject(slug);
      if (list && list.length > 0) proj = list[0];
      const id = proj?.id;
      if (id) {
        const [m, s] = await Promise.all([
          getPanelProjectMetrics(id, parseInt(days)),
          getPanelProjectSummary(id, parseInt(days)),
        ]);
        metrics = m;
        summary = s;
      }
    } catch (err: any) { error = err.message; }
    finally { loading = false; }
  }
</script>

<div class="breadcrumb"><a href="/projects">Projects</a> <span class="sep">›</span> <span>{slug}</span></div>

{#if loading}
  <div class="card"><div class="skeleton"><div class="s-line w-1/4"></div><div class="s-line w-1/2"></div></div></div>
{:else if error}
  <div class="error">{error}</div>
{:else if proj}
  <div class="page-header">
    <div>
      <h1>{proj.name || slug}</h1>
      <p class="subtitle">{proj.provider} / {proj.bucket}</p>
    </div>
  </div>

  <div class="cards">
    <div class="card stat"><span class="stat-label">Requests Today</span><span class="stat-value">{summary?.requests_today ?? '-'}</span></div>
    <div class="card stat"><span class="stat-label">Cache Hit Ratio</span><span class="stat-value">{summary?.cache_hit_ratio?.toFixed(1) ?? '-'}%</span></div>
    <div class="card stat"><span class="stat-label">Bandwidth</span><span class="stat-value">{summary?.bandwidth_bytes_today ? (summary.bandwidth_bytes_today / 1048576).toFixed(1) + ' MB' : '-'}</span></div>
    <div class="card stat"><span class="stat-label">Errors</span><span class="stat-value" class:error={summary?.errors_today > 0}>{summary?.errors_today ?? 0}</span></div>
  </div>

  <div class="period">
    <span>Period:</span>
    <select bind:value={days} onchange={() => load()}>
      <option value="7">7 days</option>
      <option value="30">30 days</option>
      <option value="90">90 days</option>
    </select>
  </div>

  <div class="card">
    <h3>Daily Breakdown</h3>
    <table>
      <thead><tr><th>Date</th><th>Requests</th><th>Cache Hits</th><th>Misses</th><th>Bandwidth</th><th>Avg Time</th><th>Errors</th></tr></thead>
      <tbody>
        {#if metrics?.daily?.length}
          {#each metrics.daily as d}
            <tr>
              <td>{d.date}</td>
              <td>{d.requests ?? 0}</td>
              <td class="success">{d.cache_hits ?? 0} ({d.requests ? ((d.cache_hits/d.requests)*100).toFixed(1) : 0}%)</td>
              <td class="warn">{d.cache_misses ?? 0}</td>
              <td>{d.bandwidth_bytes ? (d.bandwidth_bytes/1048576).toFixed(1) + ' MB' : '-'}</td>
              <td>{d.avg_response_time_ms ?? '-'}ms</td>
              <td class:error={d.errors > 0}>{d.errors ?? 0}</td>
            </tr>
          {/each}
        {:else}
          <tr><td colspan="7" class="empty-cell">No data yet</td></tr>
        {/if}
      </tbody>
    </table>
  </div>
{/if}

<style>
  .breadcrumb { font-size: 0.875rem; color: var(--c-muted); margin-bottom: 0.5rem; }
  .breadcrumb a { color: var(--c-muted); } .sep { margin: 0 0.5rem; }
  .page-header { margin-bottom: 1.5rem; }
  h1 { font-size: 1.5rem; }
  .subtitle { color: var(--c-muted); font-size: 0.875rem; }
  .error { background: rgba(239,68,68,0.15); color: var(--c-danger); padding: 0.75rem; border-radius: var(--radius-sm); font-size: 0.875rem; }
  .cards { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 1rem; margin-bottom: 1.5rem; }
  .card { background: var(--c-surface); border: 1px solid var(--c-border); border-radius: var(--radius); padding: 1.25rem; }
  .stat { display: flex; flex-direction: column; gap: 0.25rem; }
  .stat-label { font-size: 0.8rem; color: var(--c-dim); text-transform: uppercase; letter-spacing: 0.05em; }
  .stat-value { font-size: 1.5rem; font-weight: 700; }
  .stat-value.error { color: var(--c-danger); }
  .success { color: var(--c-success); }
  .warn { color: var(--c-warning); }
  .period { display: flex; align-items: center; gap: 0.5rem; font-size: 0.875rem; color: var(--c-muted); margin-bottom: 1rem; }
  .period select { padding: 0.35rem 0.5rem; border: 1px solid var(--c-border); border-radius: var(--radius-sm); background: var(--c-surface); color: var(--c-text); }
  h3 { font-size: 0.9rem; margin-bottom: 0.75rem; color: var(--c-muted); }
  table { width: 100%; font-size: 0.8rem; }
  th { text-align: left; padding: 0.5rem; color: var(--c-dim); font-weight: 500; border-bottom: 1px solid var(--c-border); }
  td { padding: 0.5rem; border-bottom: 1px solid rgba(51,65,85,0.4); }
  .empty-cell { text-align: center; color: var(--c-dim); padding: 2rem !important; }
  .skeleton { padding: 1rem; }
  .s-line { height: 1rem; background: var(--c-border); border-radius: 0.25rem; margin-bottom: 0.75rem; }
  .w-1\/4 { width: 25%; } .w-1\/2 { width: 50%; }
</style>
