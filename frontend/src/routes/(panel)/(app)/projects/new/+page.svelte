<script lang="ts">
  import { onMount } from 'svelte';
  import { getToken, createPanelProject } from '$lib/api';

  let slug = $state('');
  let name = $state('');
  let basePath = $state('/');
  let provider = $state('gcs');
  let bucket = $state('');
  let region = $state('us-east-1');
  let endpoint = $state('');
  let accessKey = $state('');
  let secretKey = $state('');
  let defaultQuality = $state(80);
  let defaultFormat = $state('webp');
  let maxWidth = $state(4096);
  let maxHeight = $state(4096);
  let error = $state('');
  let loading = $state(false);

  onMount(() => { if (!getToken()) window.location.href = '/login'; });

  const providers = [
    { value: 'gcs', label: 'Google Cloud Storage' },
    { value: 's3', label: 'Amazon S3' },
    { value: 'rustfs', label: 'RustFS (S3-compatible)' },
  ];

  function onProviderChange() {
    region = provider === 'rustfs' ? '' : 'us-east-1';
    endpoint = provider === 'rustfs' ? 'https://storage.iathings.com' : '';
  }

  async function handleSubmit(e: Event) {
    e.preventDefault();
    error = '';
    loading = true;
    try {
      const body: Record<string, any> = { slug, name, base_path: basePath, provider, bucket, region, default_quality: defaultQuality, default_format: defaultFormat, max_width: maxWidth, max_height: maxHeight };
      if (provider === 's3' || provider === 'rustfs') { body.access_key_id = accessKey; body.secret_access_key = secretKey; }
      if (provider === 'rustfs') body.endpoint = endpoint;
      if (provider === 'gcs') body.credentials_json = '';
      await createPanelProject(body);
      window.location.href = '/projects/' + slug;
    } catch (err: any) { error = err.message; }
    finally { loading = false; }
  }
</script>

<div class="breadcrumb"><a href="/projects">Projects</a> <span class="sep">/</span> <span>New Project</span></div>
<h1>New Project</h1>
<p class="subtitle">Configure a new ImageKit project</p>

<div class="form-wrap">
  {#if error}<div class="error">{error}</div>{/if}
  <form onsubmit={handleSubmit}>
    <div class="grid-2">
      <label>Slug * <input bind:value={slug} required placeholder="my-project" /></label>
      <label>Name <input bind:value={name} placeholder="My Project" /></label>
    </div>
    <label>Base Path <input bind:value={basePath} /></label>
    <div class="grid-2">
      <label>Default Quality <input type="number" bind:value={defaultQuality} /></label>
      <label>Default Format
        <select bind:value={defaultFormat}>
          <option value="webp">WebP</option><option value="jpeg">JPEG</option><option value="png">PNG</option>
        </select>
      </label>
    </div>
    <div class="grid-2">
      <label>Max Width <input type="number" bind:value={maxWidth} /></label>
      <label>Max Height <input type="number" bind:value={maxHeight} /></label>
    </div>
    <label>Storage Provider *
      <select bind:value={provider} onchange={onProviderChange}>
        <option value="">Select</option>
        {#each providers as p}<option value={p.value}>{p.label}</option>{/each}
      </select>
    </label>
    <label>Bucket * <input bind:value={bucket} required placeholder="my-bucket" /></label>
    {#if provider !== 'rustfs'}<label>Region <input bind:value={region} placeholder="us-east-1" /></label>{/if}
    {#if provider === 'rustfs'}<label>Endpoint URL <input bind:value={endpoint} /></label>{/if}
    {#if provider === 's3' || provider === 'rustfs'}
      <label>Access Key ID <input bind:value={accessKey} /></label>
      <label>Secret Access Key <input type="password" bind:value={secretKey} /></label>
    {/if}
    <button type="submit" disabled={loading} class="btn">{loading ? 'Creating...' : 'Create Project'}</button>
  </form>
</div>

<style>
  .breadcrumb { font-size: 0.875rem; color: var(--c-muted); margin-bottom: 0.5rem; }
  .breadcrumb a { color: var(--c-muted); } .sep { margin: 0 0.5rem; }
  h1 { font-size: 1.5rem; }
  .subtitle { color: var(--c-muted); font-size: 0.875rem; margin-bottom: 1.5rem; }
  .form-wrap { max-width: 600px; }
  .error { background: rgba(239,68,68,0.15); color: var(--c-danger); padding: 0.75rem; border-radius: var(--radius-sm); font-size: 0.875rem; margin-bottom: 1rem; }
  form { background: var(--c-surface); border: 1px solid var(--c-border); border-radius: var(--radius); padding: 1.5rem; display: flex; flex-direction: column; gap: 1rem; }
  .grid-2 { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
  label { display: flex; flex-direction: column; gap: 0.35rem; font-size: 0.875rem; color: var(--c-muted); }
  input, select { padding: 0.5rem 0.75rem; border: 1px solid var(--c-border); border-radius: var(--radius-sm); background: var(--c-bg); color: var(--c-text); outline: none; }
  input:focus, select:focus { border-color: var(--c-primary); }
  .btn { padding: 0.625rem; background: var(--c-primary); color: #fff; border: none; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
  .btn:hover { background: var(--c-primary-hover); }
  .btn:disabled { opacity: 0.5; cursor: not-allowed; }
</style>
