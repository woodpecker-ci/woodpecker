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
