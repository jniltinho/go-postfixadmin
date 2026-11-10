export type Theme = { id: string; label: string; scheme: 'light' | 'dark' }

// Registry of the themes in this folder; ids match the [data-theme] blocks (checked by themes.test.ts).
export const themes: Theme[] = [
  { id: 'default', label: 'Default', scheme: 'light' },
  { id: 'admanager', label: 'ADManager', scheme: 'light' },
  { id: 'breeze', label: 'Breeze', scheme: 'light' },
  { id: 'breeze-dark', label: 'Breeze (dark)', scheme: 'dark' },
  { id: 'violet', label: 'Violet', scheme: 'light' },
  { id: 'violet-dark', label: 'Violet (dark)', scheme: 'dark' },
  { id: 'aurora', label: 'Aurora', scheme: 'light' },
  { id: 'aurora-dark', label: 'Aurora (dark)', scheme: 'dark' },
  { id: 'crystal', label: 'Crystal', scheme: 'light' },
  { id: 'crystal-dark', label: 'Crystal (dark)', scheme: 'dark' },
  { id: 'cone', label: 'Cone (dark)', scheme: 'dark' },
  { id: 'xp', label: 'Windows XP', scheme: 'light' },
  { id: 'classic-gray', label: 'Windows 98', scheme: 'light' },
  { id: 'graphite', label: 'Graphite (dark)', scheme: 'dark' },
  { id: 'grafana', label: 'Grafana (dark)', scheme: 'dark' },
  { id: 'rabbitmq', label: 'RabbitMQ', scheme: 'light' },
  { id: 'matrix', label: 'Matrix (dark)', scheme: 'dark' },
  { id: 'winamp', label: 'Winamp (dark)', scheme: 'dark' },
  { id: 'pink-panther', label: 'Pink Panther', scheme: 'light' },
  { id: 'high-contrast', label: 'High contrast (dark)', scheme: 'dark' },
]
