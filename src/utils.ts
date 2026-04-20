
export function getEnvVar(name: string): string {
  const envVar = process.env[name];
  if (!envVar || !envVar.trim()) throw new Error(`${name} is required env var`);
  return envVar;
}

