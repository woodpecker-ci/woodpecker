import { WebhookEvents } from './webhook';

export type Secret = {
  id: string;
  name: string;
  value: string;
  events: WebhookEvents[];
  images: string[];
};
