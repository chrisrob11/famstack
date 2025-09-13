export interface Task {
  id: string;
  title: string;
  description: string;
  status: 'pending' | 'completed';
  task_type: 'todo' | 'chore' | 'appointment';
  assigned_to?: string | null;
  created_at: string;
  completed_at?: string | undefined;
  priority: number;
}
