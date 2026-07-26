<script lang="ts">
  import { onMount } from 'svelte';
  import { getToken, getPanelProject, getPanelProjectMetrics, getPanelProjectSummary, updatePanelProject, deletePanelProject } from '$lib/api';

  let { params } = $props();
  let slug = $derived(params.slug);

  let proj: any = $state(null);
  let metrics: any = $state(null);
  let summary: any = $state(null);
  let loading = $state(true);
  let days = $state('30');
  let error = $state('');

  let editing = $state(false);
  let saving = $state(false);
  let saveError = $state('');

  let eName = $state('');
  let eBasePath = $state('/');
  let eProvider = $state('gcs');
  let eBucket = $state('');
  let eRegion = $state('');
  let eEndpoint = $state('');
  let eAccessKey = $state('');
  let eSecretKey = $state('');
  let eDefaultQuality = $state(80);
  let eDefaultFormat = $state('webp');
  let eMaxWidth = $state(4096);
  let eMaxHeight = $state(4096);

  let showDeleteConfirm = $state(false);
  let deleting = $state(false);

  const providers = [
    { value: 'gcs', label: 'Google Cloud Storage' },
    { value: 's3', label: 'Amazon S3' },
    { value: 'rustfs', label: 'RustFS (S3-compatible)' },
  ];

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

  function startEdit() {
    if (!proj) return;
    eName = proj.name || '';
    eBasePath = proj.base_path || '/';
    eProvider = proj.provider || 'gcs';
    eBucket = proj.bucket || '';
    eRegion = proj.region || '';
    eEndpoint = proj.endpoint || '';
    eAccessKey = proj.access_key_id || '';
    eSecretKey = '';
    eDefaultQuality = proj.default_quality || 80;
    eDefaultFormat = proj.default_format || 'webp';
    eMaxWidth = proj.max_width || 4096;
    eMaxHeight = proj.max_height || 4096;
    saveError = '';
    editing = true;
  }

  function cancelEdit() {
    editing = false;
    saveError = '';
  }

  async function handleSave() {
    if (!proj) return;
    saveError = '';
    saving = true;
    try {
      const body: Record<string, any> = {
        name: eName || '',
        base_path: eBasePath || '/',
        provider: eProvider,
        bucket: eBucket,
        region: eRegion,
        default_quality: eDefaultQuality,
        default_format: eDefaultFormat,
        max_width: eMaxWidth,
        max_height: eMaxHeight,
      };
      if (eProvider === 's3' || eProvider === 'rustfs') {
        body.access_key_id = eAccessKey;
        if (eSecretKey) body.secret_access_key = eSecretKey;
      }
      if (eProvider === 'rustfs') body.endpoint = eEndpoint;
      await updatePanelProject(proj.id, body);
      editing = false;
      await load();
    } catch (err: any) { saveError = err.message; }
    finally { saving = false; }
  }

  async function handleDelete() {
    if (!proj) return;
    deleting = true;
    try {
      await deletePanelProject(proj.id);
      window.location.href = '/projects';
    } catch (err: any) {
      error = err.message;
      showDeleteConfirm = false;
      deleting = false;
    }
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
    <div class="actions">
      <button class="btn btn-outline" onclick={startEdit}>Edit</button>
      <button class="btn btn-danger" onclick={() => showDeleteConfirm = true}>Delete</button>
    </div>
  </div>

  {#if editing}
    <div class="form-wrap">
      {#if saveError}<div class="error">{saveError}</div>{/if}
      <form onsubmit={(e) => { e.preventDefault(); handleSave(); }}>
        <label>Slug <input value={proj.slug} disabled /></label>
        <label>Name <input bind:value={eName} placeholder="My Project" /></label>
        <label>Base Path <input bind:value={eBasePath} /></label>
        <div class="grid-2">
          <label>Default Quality <input type="number" bind:value={eDefaultQuality} /></label>
          <label>Default Format
            <select bind:value={eDefaultFormat}>
              <option value="webp">WebP</option><option value="jpeg">JPEG</option><option value="png">PNG</option>
            </select>
          </label>
        </div>
        <div class="grid-2">
          <label>Max Width <input type="number" bind:value={eMaxWidth} /></label>
          <label>Max Height <input type="number" bind:value={eMaxHeight} /></label>
        </div>
        <label>Storage Provider
          <select bind:value={eProvider}>
            {#each providers as p}<option value={p.value}>{p.label}</option>{/each}
          </select>
        </label>
        <label>Bucket <input bind:value={eBucket} placeholder="my-bucket" /></label>
        {#if eProvider !== 'rustfs'}<label>Region <input bind:value={eRegion} placeholder="us-east-1" /></label>{/if}
        {#if eProvider === 'rustfs'}<label>Endpoint URL <input bind:value={eEndpoint} /></label>{/if}
        {#if eProvider === 's3' || eProvider === 'rustfs'}
          <label>Access Key ID <input bind:value={eAccessKey} /></label>
          <label>Secret Access Key <input type="password" bind:value={eSecretKey} placeholder="Leave empty to keep current" /></label>
        {/if}
        <div class="form-actions">
          <button type="submit" disabled={saving} class="btn">{saving ? 'Saving...' : 'Save'}</button>
          <button type="button" class="btn btn-ghost" onclick={cancelEdit}>Cancel</button>
        </div>
      </form>
    </div>
  {/if}

  {#if showDeleteConfirm}
    <div class="modal-overlay" onclick={() => showDeleteConfirm = false}>
      <div class="modal" onclick={(e) => e.stopPropagation()}>
        <h3>Delete Project</h3>
        <p>Are you sure you want to delete <strong>{proj.name || proj.slug}</strong>? This action cannot be undone.</p>
        <div class="form-actions">
          <button class="btn btn-danger" onclick={handleDelete} disabled={deleting}>{deleting ? 'Deleting...' : 'Confirm Delete'}</button>
          <button class="btn btn-ghost" onclick={() => showDeleteConfirm = false}>Cancel</button>
        </div>
      </div>
    </div>
  {/if}

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
  .page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 1.5rem; }
  h1 { font-size: 1.5rem; }
  .subtitle { color: var(--c-muted); font-size: 0.875rem; }
  .actions { display: flex; gap: 0.5rem; }
  .btn { padding: 0.5rem 1rem; border-radius: var(--radius-sm); font-size: 0.875rem; font-weight: 600; cursor: pointer; border: none; }
  .btn-outline { background: transparent; border: 1px solid var(--c-border); color: var(--c-text); }
  .btn-outline:hover { background: var(--c-surface); }
  .btn-danger { background: var(--c-danger); color: #fff; }
  .btn-danger:hover { opacity: 0.9; }
  .btn-ghost { background: transparent; color: var(--c-muted); border: 1px solid var(--c-border); }
  .btn-ghost:hover { background: var(--c-surface); }
  .error { background: rgba(239,68,68,0.15); color: var(--c-danger); padding: 0.75rem; border-radius: var(--radius-sm); font-size: 0.875rem; margin-bottom: 1rem; }
  .form-wrap { max-width: 600px; margin-bottom: 1.5rem; }
  form { background: var(--c-surface); border: 1px solid var(--c-border); border-radius: var(--radius); padding: 1.5rem; display: flex; flex-direction: column; gap: 1rem; }
  .grid-2 { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
  label { display: flex; flex-direction: column; gap: 0.35rem; font-size: 0.875rem; color: var(--c-muted); }
  input, select { padding: 0.5rem 0.75rem; border: 1px solid var(--c-border); border-radius: var(--radius-sm); background: var(--c-bg); color: var(--c-text); outline: none; }
  input:focus, select:focus { border-color: var(--c-primary); }
  input:disabled { opacity: 0.5; cursor: not-allowed; }
  .form-actions { display: flex; gap: 0.5rem; margin-top: 0.5rem; }
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
  .modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 100; }
  .modal { background: var(--c-surface); border: 1px solid var(--c-border); border-radius: var(--radius); padding: 1.5rem; max-width: 420px; width: 90%; }
  .modal h3 { font-size: 1.1rem; margin-bottom: 0.75rem; color: var(--c-text); }
  .modal p { font-size: 0.875rem; color: var(--c-muted); margin-bottom: 1.25rem; line-height: 1.5; }
</style>
