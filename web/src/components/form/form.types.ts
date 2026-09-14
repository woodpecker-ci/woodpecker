export interface SelectOption {
  value: string;
  text: string;
  description?: string;
  badge?: string;
}

export type RadioOption = SelectOption;

export type CheckboxOption = SelectOption;
