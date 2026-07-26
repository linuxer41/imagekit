<script lang="ts">
  import { onMount } from 'svelte';
  import { getToken, createAdminProject } from '$lib/api';

  let tenantId = $state('');
  let slug = $state('');
  let name = $state('');
  let provider = $state('gcs');
  let bucket = $state('');
  let region = $state('us-east-1');
  let endpoint = $state('');
  let accessKey = $state('');
  let secretKey = $state('');
  let error = $state('');
  let loading = $state(false);

  onMount(() => { if (!getToken()) window.location.href = '/admin/login'; });

  function onProviderChange() {
    region = provider === 'rustfs' ? '' : 'us-east-1';
    endpoint = provider === 'rustfs' ? 'https://storage.iathings.com' : '';
  }

  async function handleSubmit(e: Event) {
    e.preventDefault();
    error = '';
    loading = true;
    try {
      const body: Record<string, any> = {
        tenant_id: parseInt(tenantId), slug, name, provider, bucket, region,
        default_quality: 80, default_format: 'webp', max_width: 4096, max_height: 4096,
      };
      if (provider === 's3' || provider === 'rustfs') { body.access_key_id = accessKey; body.secret_access_key = secretKey; }
      if (provider === 'rustfs') body.endpoint = endpoint;
      await createAdminProject(body);
      window.location.href = '/admin/projects/' + slug;
    } catch (err: any) { error = err.message; }
    finally { loading = false; }
  }
</script>

<div class="breadcrumb"><a href="/admin/projects">Projects</a> <span class="sep">/</span> <span>New</span></div>
<h1>New Project</h1>

<div class="form-wrap">
  {#if error}<div class="error">{error}</div>{/if}
  <form onsubmit={handleSubmit}>
    <label>Tenant ID * <input type="number" bind:value={tenantId} required placeholder="123" /></label>
    <div class="grid-2">
      <label>Slug * <input bind:value={slug} required placeholder="my-project" /></label>
      <label>Name <input bind:value={name} placeholder="My Project" /></label>
    </div>
    <label>Storage Provider *
      <select bind:value={provider} onchange={onProviderChange}>
        <option value="gcs">Google Cloud Storage</option>
        <option value="s3">Amazon S3</option>
        <option value="rustfs">RustFS (S3-compatible)</option>
      </select>
    </label>
    <label>Bucket * <input bind:value={bucket} required placeholder="my-bucket" /></label>
    {#if provider !== 'rustfs'}<label>Region <input bind:value={region} /></label>{/if}
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
  .sep { margin: 0 0.5rem; }
  h1 { font-size: 1.5rem; margin-bottom: 1rem; }
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
