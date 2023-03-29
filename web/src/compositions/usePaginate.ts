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

export class PaginatedList {
  private page: number = 1;
  private hasMore: boolean = true;
  private readonly load: (page: number) => Promise<boolean>;
  private readonly scrollComponent: Element;

  constructor(load: (page: number) => Promise<boolean>, elem: String = 'main > div') {
    this.load = load;
    this.scrollComponent = document.querySelector(elem);
    if (!this.scrollComponent) {
      throw new Error('Unexpected: "scrollComponent" should be provided at this place');
    }
  }

  public onMounted() {
    this.reset(true);
    this.scrollComponent.addEventListener('scroll', this.handleScroll());
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
    this.scrollComponent.removeEventListener('scroll', this.handleScroll());
  }

  private async runLoad() {
    if (this.hasMore) {
      // to prevent that load() is called multiple times, we set hasMore = false
      this.hasMore = false;
      this.hasMore = await this.load(this.page);
    }
  }

  private handleScroll() {
    const list = this;
    return (e: Event) => {
      if (e.target.scrollTop + e.target.clientHeight === e.target.scrollHeight) {
        list.page += 1;
        list.runLoad();
      }
    };
  }
}
