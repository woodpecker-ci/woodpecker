export default async function <T>(getSingle: (page) => Promise<T[]>): Promise<T[]> {
  let hasMore = true;
  let page = 1;
  const result: T[] = [];
  while (hasMore) {
    const singleRes = await getSingle(page);
    result.push(...singleRes);
    hasMore = singleRes.length !== 0;
    page++;
  }
  return result;
}
