const API_URL = process.env.API_URL ?? "http://localhost:3000";

export function apiUrl(path: string): string {
  return `${API_URL}${path}`;
}
