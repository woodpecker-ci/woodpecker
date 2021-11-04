import marked from 'marked';

const tokens = ['---', '---'];
const regexHeader = new RegExp('^' + tokens[0] + '([\\s|\\S]*?)' + tokens[1]);
const regexContent = new RegExp(
  '^ *?\\' + tokens[0] + '[^]*?' + tokens[1] + '*'
);

export function getHeader<T = any>(data: string): T {
  const header = getRawHeader(data);

  const tmpObj = {};
  const lines = header.trim().split('\n');

  lines.forEach((line, i) => {
    var arr = line.trim().split(':');
    tmpObj[arr.shift()] = arr.join(':').trim();
  });

  return tmpObj as T;
}

export function getRawHeader(data: string): string {
  const header = regexHeader.exec(data);
  if (!header) {
    new Error("Can't get the header");
  }
  return header[1];
}

export function getContent(data): string {
  const content = data.replace(regexContent, '').replace(/<!--(.*?)-->/gm, '');
  if (!content) {
    throw new Error("Can't get the content");
  }
  return marked(content);
}
