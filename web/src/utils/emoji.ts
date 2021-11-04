import { emojify } from 'node-emoji';

export function convertEmojis(input: string): string {
  return emojify(input);
}
