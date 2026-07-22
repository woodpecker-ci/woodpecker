import BrowserOnly from '@docusaurus/BrowserOnly';
import type { WrapperProps } from '@docusaurus/types';
import Redoc from '@theme-original/Redoc';
import type RedocType from '@theme/Redoc';
import React from 'react';

type Props = WrapperProps<typeof RedocType>;

// Redoc reads Docusaurus' color-mode context (useColorMode) while rendering.
// That context is unavailable during static site generation, so rendering it
// server-side throws a ReactContextError and fails the `/api` build. Restrict
// Redoc to the browser; the surrounding Layout still renders during SSG.
export default function RedocWrapper(props: Props) {
  return (
    <BrowserOnly fallback={<div className="redocusaurus">Loading API reference…</div>}>
      {() => <Redoc {...props} />}
    </BrowserOnly>
  );
}
