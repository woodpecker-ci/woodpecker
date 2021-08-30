import { WebhookEvents } from './webhook_events';

export type Secret = {
  id: string;
  name: string;
  value: string;
  event: WebhookEvents[];
  image: string[];
};
