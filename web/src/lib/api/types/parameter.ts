export type ParameterType = 'string' | 'number' | 'boolean' | 'choice';

export interface Parameter {
  id: number;
  repo_id: number;
  name: string;
  type: ParameterType;
  description: string;
  default: string;
  options: string[];
  required: boolean;
  order: number;
  // only 'repo_config' exists today; a future 'workflow' source (parameters defined
  // in the pipeline YAML) will be served through the same list endpoint
  source: string;
}
