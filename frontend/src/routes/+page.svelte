<script lang="ts">
  import { onMount } from 'svelte';
  import { getToken } from '$lib/api';
  onMount(() => {
    if (getToken()) window.location.href = '/projects';
    else window.location.href = '/login';
  });
</script>
