export type ApiError = {
  status: number;
  message: string;
};

export function encodeQueryString(params: Record<string, string | number | boolean> = {}): string {
  return params
    ? Object.keys(params)
        .sort()
        .map((key) => {
          const val = params[key];
          return encodeURIComponent(key) + '=' + encodeURIComponent(val);
        })
        .join('&')
    : '';
}

export default class ApiClient {
  server: string;
  token: string | null;
  csrf: string | null;
  onerror: (err: ApiError) => void;

  constructor(server: string, token: string | null, csrf: string | null) {
    this.server = server;
    this.token = token;
    this.csrf = csrf;
  }

  private _request(method: string, path: string, data: unknown): Promise<any> {
    var endpoint = `${this.server}${path}`;
    var xhr = new XMLHttpRequest();
    xhr.open(method, endpoint, true);

    if (this.token) {
      xhr.setRequestHeader('Authorization', 'Bearer ' + this.token);
    }

    if (method !== 'GET' && this.csrf) {
      xhr.setRequestHeader('X-CSRF-TOKEN', this.csrf);
    }

    return new Promise((resolve, reject) => {
      xhr.onload = () => {
        if (xhr.readyState === 4) {
          if (xhr.status >= 300) {
            const error: ApiError = {
              status: xhr.status,
              message: xhr.response,
            };
            if (this.onerror) {
              this.onerror(error);
            }
            reject(error);
            return;
          }
          const contentType = xhr.getResponseHeader('Content-Type');
          if (contentType && contentType.startsWith('application/json')) {
            resolve(JSON.parse(xhr.response));
          } else {
            resolve(xhr.response);
          }
        }
      };

      xhr.onerror = (e) => {
        reject(e);
      };

      if (data) {
        xhr.setRequestHeader('Content-Type', 'application/json');
        xhr.send(JSON.stringify(data));
      } else {
        xhr.send();
      }
    });
  }

  _get(path: string) {
    return this._request('GET', path, null);
  }

  _post(path: string, data?: unknown) {
    return this._request('POST', path, data);
  }

  _patch(path: string, data?: unknown) {
    return this._request('PATCH', path, data);
  }

  _delete(path: string) {
    return this._request('DELETE', path, null);
  }

  _subscribe(path: string, callback: (data: any) => void, opts = { reconnect: true }) {
    var query = encodeQueryString({
      access_token: this.token,
    });
    path = this.server ? this.server + path : path;
    path = this.token ? path + '?' + query : path;

    var events = new EventSource(path);
    events.onmessage = (event) => {
      var data = JSON.parse(event.data);
      callback(data);
    };

    if (!opts.reconnect) {
      events.onerror = (err) => {
        if (err.data === 'eof') {
          events.close();
        }
      };
    }
    return events;
  }
}
