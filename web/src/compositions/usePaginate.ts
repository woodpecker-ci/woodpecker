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

const lists: Record<number, PaginatedList> = {};
let currId = 0;

const scrollElement = document.querySelector('main > div');
if (!scrollElement) {
  throw new Error("Unexpected: Can't get scrollElement");
}
scrollElement.addEventListener('scroll', () => {
  if (scrollElement.scrollTop + scrollElement.clientHeight === scrollElement.scrollHeight) {
    (Object.keys(lists) as unknown as number[]).forEach((id) => {
      const list = lists[id];
      list.nextPage();
      lists[id] = list;
    });
  }
});

export class PaginatedList {
  private id = -1;

  private page = 1;

  private hasMore = true;

  private readonly load: (page: number) => Promise<boolean>;

  private readonly isActive: () => boolean;

  constructor(load: (page: number) => Promise<boolean>, isActive: () => boolean = () => true) {
    this.load = load;
    this.isActive = isActive;
  }

  public onMounted() {
    this.reset(true);
    this.id = currId;
    currId += 1;
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

  public nextPage() {
    this.page += 1;
    this.runLoad();
  }

  private async runLoad() {
    if (this.hasMore && this.isActive()) {
      // to prevent that load() is called multiple times, we set hasMore = false
      this.hasMore = false;
      this.hasMore = await this.load(this.page);
    }
  }
}
