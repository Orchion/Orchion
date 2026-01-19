// Get the orchestrator base URL from environment or use default
const getBaseUrl = (): string => {
  if (typeof window !== 'undefined') {
    // Browser environment - check for env var or use default
    return import.meta.env.VITE_ORCHESTRATOR_URL || 'http://localhost:8080';
  }
  return 'http://localhost:8080';
};

export interface Node {
  id: string;
  hostname: string;
  capabilities?: {
    cpu: string;
    memory: string;
    os: string;
  };
  lastSeenUnix?: number;
}

export async function getNodes(): Promise<Node[]> {
  const baseUrl = getBaseUrl();
  const url = `${baseUrl}/api/nodes`;
  
  try {
    const res = await fetch(url);
    
    if (!res.ok) {
      throw new Error(`Failed to fetch nodes: ${res.status} ${res.statusText}`);
    }
    
    const nodes = await res.json();
    
    // Validate response is an array
    if (!Array.isArray(nodes)) {
      throw new Error('Invalid response: expected array of nodes');
    }
    
    return nodes;
  } catch (error) {
    console.error('Error fetching nodes:', error);
    throw error;
  }
}