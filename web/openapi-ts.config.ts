import { defineConfig } from '@hey-api/openapi-ts';

export default defineConfig({
  client: '@hey-api/client-fetch',
  input: '../docs/swagger.json',
  output: 'src/lib/api/',
  schemas: false,
  // services: {},
});
