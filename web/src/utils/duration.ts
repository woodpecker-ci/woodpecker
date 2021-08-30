import humanizeDuration from 'humanize-duration';

const en_short = {
  w: (count?: number) => 'w',
  d: (count?: number) => 'd',
  h: (count?: number) => 'h',
  m: (count?: number) => 'min',
  s: (count?: number) => 'sec',
};
const durationOptions: humanizeDuration.HumanizerOptions = {
  round: true,
  languages: { en_short },
  language: 'en_short',
};

export function prettyDuration(durationMs: number) {
  return humanizeDuration(durationMs, durationOptions);
}

function leadingZeros(n: number, length: number): string {
  let res = n.toString();
  while (res.length < length) res = '0' + res;
  return res;
}

export function durationAsNumber(durationMs: number): string {
  const durationSeconds = durationMs / 1000;
  const seconds = leadingZeros(Math.floor(durationSeconds % 60), 2);
  const minutes = leadingZeros(Math.floor(durationSeconds / 60) % 60, 2);
  const hours = Math.floor(durationSeconds / 3600);

  if (hours !== 0) {
    return `${hours}:${minutes}:${seconds}`;
  }

  return `${minutes}:${seconds}`;
}
