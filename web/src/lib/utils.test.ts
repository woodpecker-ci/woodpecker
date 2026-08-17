import { describe, expect, it } from 'vitest';

import { escapeHtml } from './utils';

describe('escapeHtml', () => {
  it('should return plain text unchanged', () => {
    expect(escapeHtml('hello world')).toBe('hello world');
  });

  it('should return empty string unchanged', () => {
    expect(escapeHtml('')).toBe('');
  });

  it('should escape HTML tags', () => {
    expect(escapeHtml('<b>bold</b>')).toBe('&lt;b&gt;bold&lt;/b&gt;');
    expect(escapeHtml('<script>alert("xss")</script>')).toBe('&lt;script&gt;alert(&quot;xss&quot;)&lt;/script&gt;');
  });

  it('should escape ampersands', () => {
    expect(escapeHtml('foo & bar')).toBe('foo &amp; bar');
    expect(escapeHtml('a&&b')).toBe('a&amp;&amp;b');
  });

  it('should escape double quotes', () => {
    expect(escapeHtml('say "hello"')).toBe('say &quot;hello&quot;');
  });

  it('should escape single quotes', () => {
    expect(escapeHtml("it's")).toBe('it&#x27;s');
  });

  it('should escape greater-than signs', () => {
    expect(escapeHtml('a > b')).toBe('a &gt; b');
  });

  it('should escape mixed content', () => {
    expect(escapeHtml(`<a href="foo">it's & that's <b>all</b>`)).toBe(
      '&lt;a href=&quot;foo&quot;&gt;it&#x27;s &amp; that&#x27;s &lt;b&gt;all&lt;/b&gt;',
    );
  });

  it('should escape already-escaped ampersands', () => {
    expect(escapeHtml('&amp;')).toBe('&amp;amp;');
  });
});
