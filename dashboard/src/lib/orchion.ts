export async function getNodes() {
  const res = await fetch("http://localhost:8080/api/nodes");
  return res.json();
}