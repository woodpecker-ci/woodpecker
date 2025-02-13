export enum ParameterType {
  Boolean = 'boolean',
  SingleChoice = 'single_choice',
  MultipleChoice = 'multiple_choice',
  String = 'string',
  Text = 'text',
  Password = 'password'
}

export interface Parameter {
  id: string;
  repo_id: number;
  name: string;
  branch: string;
  type: ParameterType;
  description: string;
  default_value: string;
  trim_string: boolean;
}
