<script lang="ts">
  import { onMount } from 'svelte';
  import { getToken, listAdminProjects, getAdminGlobalMetrics, getPanelProjectSummary } from '$lib/api';

  let { params } = $props();
  let slug = $derived(params.slug);
  let loading = $state(true);
  let error = $state('');
  let proj: any = $state(null);
  let summary: any = $state(null);
  let days = $state('30');

  onMount(() => {
    if (!getToken()) { window.location.href = '/admin/login'; return; }
    load();
  });

  async function load() {
    try {
      const projects = await listAdminProjects();
      proj = projects.find((p: any) => p.slug === slug) || null;
      if (proj) {
        summary = await getPanelProjectSummary(proj.id, parseInt(days));
      }
    } catch (err: any) { error = err.message; }
    finally { loading = false; }
  }
</script>

<div class="breadcrumb"><a href="/admin/projects">Projects</a> <span class="sep">›</span> <span>{slug}</span></div>

{#if loading}
  <div class="card"><div class="skeleton"><div class="s-line w-1/4"></div><div class="s-line w-1/2"></div></div></div>
{:else if error}
  <div class="error">{error}</div>
{:else if proj}
  <div class="page-header">
    <div>
      <h1>{proj.name || slug}</h1>
      <p class="subtitle">{proj.provider} / {proj.bucket} — Tenant #{proj.tenant_id}</p>
    </div>
  </div>

  <div class="cards">
    <div class="card stat"><span class="stat-label">Requests Today</span><span class="stat-value">{summary?.requests_today ?? '-'}</span></div>
    <div class="card stat"><span class="stat-label">Cache Hit Ratio</span><span class="stat-value">{summary?.cache_hit_ratio?.toFixed(1) ?? '-'}%</span></div>
    <div class="card stat"><span class="stat-label">Bandwidth</span><span class="stat-value">{summary?.bandwidth_bytes_today ? (summary.bandwidth_bytes_today / 1048576).toFixed(1) + ' MB' : '-'}</span></div>
    <div class="card stat"><span class="stat-label">Errors</span><span class="stat-value" class:error={summary?.errors_today > 0}>{summary?.errors_today ?? 0}</span></div>
  </div>

  <div class="card">
    <h3>Configuration</h3>
    <div class="props">
      <div class="prop"><span class="prop-key">Slug</span><span class="prop-val">{proj.slug}</span></div>
      <div class="prop"><span class="prop-key">Name</span><span class="prop-val">{proj.name || '-'}</span></div>
      <div class="prop"><span class="prop-key">Provider</span><span class="prop-val">{proj.provider}</span></div>
      <div class="prop"><span class="prop-key">Bucket</span><span class="prop-val">{proj.bucket}</span></div>
      <div class="prop"><span class="prop-key">Region</span><span class="prop-val">{proj.region || '-'}</span></div>
      <div class="prop"><span class="prop-key">Base Path</span><span class="prop-val">{proj.base_path}</span></div>
      <div class="prop"><span class="prop-key">Default Format</span><span class="prop-val">{proj.default_format}</span></div>
      <div class="prop"><span class="prop-key">Default Quality</span><span class="prop-val">{proj.default_quality}</span></div>
      <div class="prop"><span class="prop-key">Max Width</span><span class="prop-val">{proj.max_width}</span></div>
      <div class="prop"><span class="prop-key">Max Height</span><span class="prop-val">{proj.max_height}</span></div>
      <div class="prop"><span class="prop-key">RPS / Burst</span><span class="prop-val">{proj.rps} / {proj.burst}</span></div>
    </div>
  </div>
{:else}
  <div class="error">Project not found</div>
{/if}

<style>
  .breadcrumb { font-size: 0.875rem; color: var(--c-muted); margin-bottom: 0.5rem; }
  .sep { margin: 0 0.5rem; }
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
  h3 { font-size: 0.9rem; margin-bottom: 0.75rem; color: var(--c-muted); }
  .props { display: flex; flex-direction: column; gap: 0.5rem; }
  .prop { display: flex; gap: 0.75rem; font-size: 0.875rem; }
  .prop-key { color: var(--c-dim); min-width: 120px; }
  .prop-val { color: var(--c-text); font-family: var(--font-mono); }
  .skeleton { padding: 1rem; }
  .s-line { height: 1rem; background: var(--c-border); border-radius: 0.25rem; margin-bottom: 0.75rem; }
  .w-1\/4 { width: 25%; } .w-1\/2 { width: 50%; }
</style>
