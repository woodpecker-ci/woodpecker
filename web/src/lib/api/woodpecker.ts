type Pet = {
  id: number;
  name: string;
};

type Endpoint = {
  request: {
    query: Record<string, any>;
    body: Record<string, any>;
    headers: Record<string, any>;
  };
  response: {
    body: Record<string, any>;
    headers: Record<string, any>;
  };
};

type Paths = {
  'GET /pets': {
    response: {
      body: {
        agents: Pet[];
      };
    };
    request: {
      query: {
        page: number;
        per_page: number;
      };
    };
  };
};

class Client<Paths> {
  baseUrl = 'https://api.hey.com';

  request<Url = keyof Paths>(methodWithPath: Url, options: Paths[Url]['request']) {
    const method = methodWithPath.split(' ')[0].toLowerCase();

    return fetch({
      method: methodWithPath.method,
      url: `${this.baseUrl}${methodWithPath.path}`,
    });
  }
}

const client = new Client<Paths>();
const pets = client.request('GET /pets', {});
