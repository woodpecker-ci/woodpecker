import { WebhookEvents } from './webhook';

export type Secret = {
  id: string;
  name: string;
  value: string;
  event: WebhookEvents[];
  image: string[];
};
