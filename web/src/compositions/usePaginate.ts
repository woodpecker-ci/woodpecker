export async function usePaginate<T>(getSingle: (page: number) => Promise<T[]>): Promise<T[]> {
  let hasMore = true;
  let page = 1;
  const result: T[] = [];
  while (hasMore) {
    // eslint-disable-next-line no-await-in-loop
    const singleRes = await getSingle(page);
    result.push(...singleRes);
    hasMore = singleRes.length !== 0;
    page += 1;
  }
  return result;
}

const lists = {};
let currId = 0;

document.querySelector('main > div').addEventListener('scroll', (e: Event) => {
  if (e.target.scrollTop + e.target.clientHeight === e.target.scrollHeight) {
    for (const id in lists) {
      lists[id].page += 1;
      lists[id].runLoad();
    }
  }
});

export class PaginatedList {
  private id: number;
  private page = 1;
  private hasMore = true;
  private readonly load: (page: number) => Promise<boolean>;

  constructor(load: (page: number) => Promise<boolean>) {
    this.load = load;
  }

  public onMounted() {
    this.reset(true);
    this.id = currId++;
    lists[this.id] = this;
  }

  public reset(reload: boolean) {
    this.page = 1;
    this.hasMore = true;
    if (reload) {
      this.runLoad();
    }
  }

  public onUnmounted() {
    this.reset(false);
    delete lists[this.id];
  }

  private async runLoad() {
    if (this.hasMore) {
      // to prevent that load() is called multiple times, we set hasMore = false
      this.hasMore = false;
      this.hasMore = await this.load(this.page);
    }
  }
}
