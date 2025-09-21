/**
 * Form Data Type Definitions
 * Provides type safety for form data extraction
 */

// Integration creation form data
export interface IntegrationFormData {
  integration_type: string;
  provider: string;
  display_name: string;
  description?: string;
  auth_method?: string;
}

// OAuth configuration form data
export interface OAuthConfigFormData {
  client_id: string;
  client_secret: string;
  redirect_url?: string;
  scopes?: string[];
}

// Google OAuth specific form data
export interface GoogleOAuthFormData extends OAuthConfigFormData {
  configured: boolean;
}

// Form validation result
export interface FormValidationResult {
  isValid: boolean;
  errors: Record<string, string>;
}

// Generic form state
export interface FormState<T = Record<string, any>> {
  data: Partial<T>;
  errors: Record<keyof T, string>;
  isValid: boolean;
  isSubmitting: boolean;
}

// Form field configuration
export interface FormFieldConfig {
  type: 'text' | 'email' | 'password' | 'select' | 'textarea' | 'checkbox';
  label: string;
  placeholder?: string;
  required?: boolean;
  options?: Array<{ value: string; label: string }>;
  validation?: {
    pattern?: RegExp;
    minLength?: number;
    maxLength?: number;
    custom?: (value: any) => string | null;
  };
}

// Integration form configuration
export interface IntegrationFormConfig {
  fields: Record<keyof IntegrationFormData, FormFieldConfig>;
}

// Form event types
export interface FormEvents {
  'form-submit': CustomEvent<{ data: any; isValid: boolean }>;
  'form-change': CustomEvent<{ field: string; value: any }>;
  'form-error': CustomEvent<{ field: string; error: string }>;
  'form-reset': CustomEvent<void>;
}

// Form validation rules
export type ValidationRule<T = any> = (value: T) => string | null;

export type ValidationRules<T = Record<string, any>> = {
  [K in keyof T]?: ValidationRule<T[K]>[];
};

// Common validation rules
export const ValidationRules = {
  required: (message = 'This field is required'): ValidationRule =>
    (value) => (!value || (typeof value === 'string' && value.trim() === '')) ? message : null,

  email: (message = 'Please enter a valid email address'): ValidationRule<string> =>
    (value) => {
      if (!value) return null;
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      return emailRegex.test(value) ? null : message;
    },

  minLength: (min: number, message?: string): ValidationRule<string> =>
    (value) => {
      if (!value) return null;
      const msg = message || `Must be at least ${min} characters`;
      return value.length >= min ? null : msg;
    },

  maxLength: (max: number, message?: string): ValidationRule<string | undefined> =>
    (value) => {
      if (!value) return null;
      const msg = message || `Must be no more than ${max} characters`;
      return value.length <= max ? null : msg;
    },

  pattern: (regex: RegExp, message = 'Invalid format'): ValidationRule<string> =>
    (value) => {
      if (!value) return null;
      return regex.test(value) ? null : message;
    },

  url: (message = 'Please enter a valid URL'): ValidationRule<string | undefined> =>
    (value) => {
      if (!value) return null;
      try {
        new URL(value);
        return null;
      } catch {
        return message;
      }
    },

  oneOf: <T>(options: T[], message = 'Please select a valid option'): ValidationRule<T> =>
    (value) => options.includes(value) ? null : message,
};

// Form validation utility
export function validateFormData<T extends Record<string, any>>(
  data: Partial<T>,
  rules: ValidationRules<T>
): FormValidationResult {
  const errors: Record<string, string> = {};

  for (const [field, fieldRules] of Object.entries(rules)) {
    const value = data[field as keyof T];

    if (fieldRules && Array.isArray(fieldRules)) {
      for (const rule of fieldRules) {
        const error = rule(value);
        if (error) {
          errors[field] = error;
          break; // Stop at first error for this field
        }
      }
    }
  }

  return {
    isValid: Object.keys(errors).length === 0,
    errors
  };
}

// Integration form validation rules
export const integrationFormRules: ValidationRules<IntegrationFormData> = {
  integration_type: [ValidationRules.required('Please select an integration type')],
  provider: [ValidationRules.required('Please select a provider')],
  display_name: [
    ValidationRules.required('Please enter a display name'),
    ValidationRules.minLength(3, 'Display name must be at least 3 characters'),
    ValidationRules.maxLength(50, 'Display name must be no more than 50 characters')
  ],
  description: [
    ValidationRules.maxLength(200, 'Description must be no more than 200 characters')
  ]
};

// OAuth form validation rules
export const oauthFormRules: ValidationRules<GoogleOAuthFormData> = {
  client_id: [ValidationRules.required('Please enter the Client ID')],
  client_secret: [ValidationRules.required('Please enter the Client Secret')],
  redirect_url: [ValidationRules.url('Please enter a valid redirect URL')]
};