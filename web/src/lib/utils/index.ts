import { toRaw } from 'vue';

export function debounce<T extends unknown[]>(fn: (...args: T) => void, delay: number): (...args: T) => void {
  let timer: ReturnType<typeof setTimeout>;
  return (...args: T) => {
    clearTimeout(timer);
    timer = setTimeout(fn, delay, ...args);
  };
}

export function deepClone<T>(value: T): T {
  return JSON.parse(JSON.stringify(toRaw(value))) as T;
}

export function escapeHtml(text: string): string {
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#x27;');
}

// httpUrl returns url only if it is a http(s) URL, so other schemes like
// javascript: from forge or user data never end up in a link
export function httpUrl(url: string | undefined): string | undefined {
  if (url === undefined || url === '') {
    return undefined;
  }
  try {
    const { protocol } = new URL(url);
    return protocol === 'http:' || protocol === 'https:' ? url : undefined;
  } catch {
    return undefined;
  }
}
