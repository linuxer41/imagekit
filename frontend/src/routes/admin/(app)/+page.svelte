<script lang="ts">
  import { onMount } from 'svelte';
  import { getToken } from '$lib/api';
  onMount(() => {
    if (getToken()) window.location.href = '/admin/tenants';
    else window.location.href = '/admin/login';
  });
</script>
