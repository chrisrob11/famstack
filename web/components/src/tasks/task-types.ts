export interface Task {
  id: string;
  family_id: string;
  assigned_to?: string | null;
  title: string;
  description: string;
  task_type: 'todo' | 'chore' | 'appointment';
  status: 'pending' | 'completed';
  priority: number;
  due_date?: string | null;
  frequency?: string | null;
  created_by: string;
  created_at: string;
  completed_at?: string | null;
}
