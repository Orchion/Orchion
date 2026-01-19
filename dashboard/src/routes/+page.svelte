<script lang="ts">
  import { onMount } from "svelte";
  import { getNodes } from "$lib/orchion";
  import type { Node } from "$lib/orchion";
  let nodes: Node[] = [];
  let error: string | null = null;

  onMount(async () => {
    try {
      nodes = await getNodes();
      error = null;
    } catch (err) {
      console.error("Failed to fetch nodes:", err);
      error = err instanceof Error ? err.message : "Failed to fetch nodes";
    }
  });
</script>

<h1>Orchion Nodes</h1>

{#if error}
  <p style="color: red;">Error: {error}</p>
{:else if nodes.length === 0}
  <p>No nodes registered yet.</p>
{:else}
  <ul>
    {#each nodes as node}
      <li>
        <strong>{node.hostname || node.id}</strong>
        <br />
        ID: {node.id}
        {#if node.capabilities}
          <br />
          CPU: {node.capabilities.cpu} | Memory: {node.capabilities.memory} | OS: {node.capabilities.os}
        {/if}
        {#if node.lastSeenUnix}
          <br />
          Last seen: {new Date(node.lastSeenUnix * 1000).toLocaleString()}
        {/if}
      </li>
    {/each}
  </ul>
{/if}