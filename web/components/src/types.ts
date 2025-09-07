export interface Task {
  id: string;
  familyId: string;
  assignedTo: string | null;
  title: string;
  description: string;
  taskType: 'todo' | 'chore' | 'appointment';
  status: 'pending' | 'completed';
  priority: number;
  dueDate: string | null;
  frequency: string | null;
  points: number;
  createdBy: string;
  createdAt: string;
  completedAt: string | null;
}

export interface User {
  id: string;
  familyId: string;
  name: string;
  email: string;
  role: string;
}

export interface Family {
  id: string;
  name: string;
  createdAt: string;
}

export interface ComponentConfig {
  apiBaseUrl: string;
  csrfToken: string;
}