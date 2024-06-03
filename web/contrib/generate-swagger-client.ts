import { generateApi } from 'swagger-typescript-api';
import path from 'node:path';

async function run() {
  console.log(`Generating swagger client ...`);

  await generateApi({
    name: 'api.ts',
    output: path.resolve(__dirname, '..', 'src', 'lib', 'api'),
    input: path.resolve(__dirname, '..', '..', 'docs', 'swagger.json'),
  });

  console.log(`Done.`);
}

run();
