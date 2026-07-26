<script lang="ts">
  import { panelLogin, setPanelToken } from '$lib/api';

  let email = $state('');
  let password = $state('');
  let error = $state('');
  let loading = $state(false);

  async function handleSubmit(e: Event) {
    e.preventDefault();
    error = '';
    loading = true;
    try {
      const res = await panelLogin(email, password);
      setPanelToken(res.token);
      window.location.href = '/projects';
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  }
</script>

<div class="auth-page">
  <div class="auth-card">
    <h1>Sign In</h1>
    <p class="subtitle">Access your ImageKit dashboard</p>
    {#if error}<div class="error">{error}</div>{/if}
    <form onsubmit={handleSubmit}>
      <label>Email <input type="email" bind:value={email} required placeholder="you@example.com" /></label>
      <label>Password <input type="password" bind:value={password} required placeholder="••••••••" /></label>
      <button type="submit" disabled={loading}>{loading ? 'Signing in...' : 'Sign In'}</button>
    </form>
    <p class="footer-link">Don't have an account? <a href="/register">Register</a></p>
  </div>
</div>

<style>
  .auth-page { display: flex; align-items: center; justify-content: center; min-height: 100vh; }
  .auth-card { background: var(--c-surface); border: 1px solid var(--c-border); border-radius: var(--radius); padding: 2rem; width: 100%; max-width: 400px; box-shadow: var(--shadow); }
  h1 { font-size: 1.5rem; margin-bottom: 0.25rem; }
  .subtitle { color: var(--c-muted); font-size: 0.875rem; margin-bottom: 1.5rem; }
  .error { background: rgba(239,68,68,0.15); color: var(--c-danger); padding: 0.75rem; border-radius: var(--radius-sm); font-size: 0.875rem; margin-bottom: 1rem; }
  form { display: flex; flex-direction: column; gap: 1rem; }
  label { display: flex; flex-direction: column; gap: 0.35rem; font-size: 0.875rem; color: var(--c-muted); }
  input { padding: 0.625rem 0.75rem; border: 1px solid var(--c-border); border-radius: var(--radius-sm); background: var(--c-bg); color: var(--c-text); outline: none; }
  input:focus { border-color: var(--c-primary); }
  button { padding: 0.625rem; background: var(--c-primary); color: #fff; border: none; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
  button:hover { background: var(--c-primary-hover); }
  button:disabled { opacity: 0.5; cursor: not-allowed; }
  .footer-link { text-align: center; margin-top: 1rem; font-size: 0.875rem; color: var(--c-muted); }
</style>
