import { ComponentConfig } from '../common/types.js';

export interface FamilyMember {
  id: string;
  family_id: string;
  name: string;
  nickname?: string;
  member_type: 'adult' | 'child' | 'pet';
  age?: number;
  avatar_url?: string;
  user_id?: string;
  display_order: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  user?: {
    id: string;
    email: string;
    first_name: string;
    last_name: string;
    role: string;
  };
}

export interface CreateFamilyMemberRequest {
  name: string;
  nickname?: string;
  member_type: 'adult' | 'child' | 'pet';
  age?: number;
  avatar_url?: string;
  display_order?: number;
}

export interface Family {
  id: string;
  name: string;
  created_at: string;
}

export interface CreateFamilyRequest {
  name: string;
}

export class FamilyService {
  private config: ComponentConfig;

  constructor(config: ComponentConfig) {
    this.config = config;
  }

  // Family Management
  async listFamilies(): Promise<Family[]> {
    const response = await fetch(`${this.config.apiBaseUrl}/families`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'X-CSRF-Token': this.config.csrfToken,
      },
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch families: ${response.statusText}`);
    }

    return response.json();
  }

  async createFamily(familyData: CreateFamilyRequest): Promise<Family> {
    const response = await fetch(`${this.config.apiBaseUrl}/families`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-CSRF-Token': this.config.csrfToken,
      },
      body: JSON.stringify(familyData),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(errorText || `Failed to create family: ${response.statusText}`);
    }

    return response.json();
  }

  async getFamily(familyId: string): Promise<Family> {
    const response = await fetch(`${this.config.apiBaseUrl}/families/${familyId}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'X-CSRF-Token': this.config.csrfToken,
      },
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch family: ${response.statusText}`);
    }

    return response.json();
  }

  async updateFamily(familyId: string, updates: Partial<CreateFamilyRequest>): Promise<Family> {
    const response = await fetch(`${this.config.apiBaseUrl}/families/${familyId}`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
        'X-CSRF-Token': this.config.csrfToken,
      },
      body: JSON.stringify(updates),
    });

    if (!response.ok) {
      throw new Error(`Failed to update family: ${response.statusText}`);
    }

    return response.json();
  }

  async deleteFamily(familyId: string): Promise<void> {
    const response = await fetch(`${this.config.apiBaseUrl}/families/${familyId}`, {
      method: 'DELETE',
      headers: {
        'X-CSRF-Token': this.config.csrfToken,
      },
    });

    if (!response.ok) {
      throw new Error(`Failed to delete family: ${response.statusText}`);
    }
  }

  // Family Member Management
  async listFamilyMembers(familyId: string): Promise<FamilyMember[]> {
    const response = await fetch(`${this.config.apiBaseUrl}/families/${familyId}/members`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'X-CSRF-Token': this.config.csrfToken,
      },
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch family members: ${response.statusText}`);
    }

    return response.json();
  }

  async createFamilyMember(familyId: string, memberData: CreateFamilyMemberRequest): Promise<FamilyMember> {
    const response = await fetch(`${this.config.apiBaseUrl}/families/${familyId}/members`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-CSRF-Token': this.config.csrfToken,
      },
      body: JSON.stringify(memberData),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(errorText || `Failed to create family member: ${response.statusText}`);
    }

    return response.json();
  }

  async getFamilyMember(familyId: string, memberId: string): Promise<FamilyMember> {
    const response = await fetch(`${this.config.apiBaseUrl}/families/${familyId}/members/${memberId}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'X-CSRF-Token': this.config.csrfToken,
      },
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch family member: ${response.statusText}`);
    }

    return response.json();
  }

  async updateFamilyMember(
    familyId: string,
    memberId: string,
    updates: Partial<CreateFamilyMemberRequest>
  ): Promise<FamilyMember> {
    const response = await fetch(`${this.config.apiBaseUrl}/families/${familyId}/members/${memberId}`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
        'X-CSRF-Token': this.config.csrfToken,
      },
      body: JSON.stringify(updates),
    });

    if (!response.ok) {
      throw new Error(`Failed to update family member: ${response.statusText}`);
    }

    return response.json();
  }

  async deleteFamilyMember(familyId: string, memberId: string): Promise<void> {
    const response = await fetch(`${this.config.apiBaseUrl}/families/${familyId}/members/${memberId}`, {
      method: 'DELETE',
      headers: {
        'X-CSRF-Token': this.config.csrfToken,
      },
    });

    if (!response.ok) {
      throw new Error(`Failed to delete family member: ${response.statusText}`);
    }
  }
}
