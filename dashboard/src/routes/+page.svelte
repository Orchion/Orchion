<script>
  import { getNodes } from "$lib/orchion";
  let nodes = [];

  onMount(async () => {
    nodes = await getNodes();
  });
</script>

<h1>Orchion Nodes</h1>

<ul>
  {#each nodes as node}
    <li>{node.id} — {node.status}</li>
  {/each}
</ul>