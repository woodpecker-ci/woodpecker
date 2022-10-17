import { marked } from 'marked';
import { parse as YAMLParse } from 'yaml';

const regexHeader = new RegExp('^---([\\s|\\S]*?)---', 'm');
const regexContent = new RegExp('^ *?\\---[^]*?---*', 'm');

export function getHeader<T = any>(data: string): T {
  const header = regexHeader.exec(data);
  if (!header || header.length != 2) {
    throw new Error("Can't get the header");
  }

  return YAMLParse(header[1]) as T;
}

export function getContent(data: string): string {
  const content = data.replace(regexContent, '').replace(/<!--(.*?)-->/gm, '');
  if (!content) {
    throw new Error("Can't get the content");
  }
  return marked(content);
}
