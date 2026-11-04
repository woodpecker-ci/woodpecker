import React from 'react';

export type InvaderState = 'ok' | 'failed';

const PIXELS: Record<InvaderState, string[]> = {
  ok: [
    '..X.....X..',
    '...X...X...',
    '..XXXXXXX..',
    '.XX.XXX.XX.',
    'XXXXXXXXXXX',
    'X.XXXXXXX.X',
    'X.X.....X.X',
    '...XX.XX...',
  ],
  failed: [
    '..X.....X..',
    '...X...X...',
    '..XX.X.XX..',
    '.XX.XXX.XX.',
    'XXXXX.XXXXX',
    'X.XXXXXXX.X',
    'X.X.....X.X',
    '...XX.XX...',
  ],
};

export function PixelInvader({
  state = 'ok',
  size = 22,
  className = '',
}: {
  state?: InvaderState;
  size?: number;
  className?: string;
}) {
  return (
    <svg
      className={`wp-pixel-invader ${className}`}
      width={size}
      height={(size * 8) / 11}
      viewBox="0 0 11 8"
      shapeRendering="crispEdges"
      aria-hidden="true"
      focusable="false"
    >
      {PIXELS[state].map((row, y) =>
        row
          .split('')
          .map((pixel, x) =>
            pixel === 'X' ? <rect key={`${x}-${y}`} x={x} y={y} width="1" height="1" fill="currentColor" /> : null,
          ),
      )}
    </svg>
  );
}
