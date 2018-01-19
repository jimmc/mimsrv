class ApiManager {
  public static async xhrJson(url: string) {
    const response = await this.xhrText(url);
    return JSON.parse(response || 'null');
  }

  public static async xhrText(url: string) {
    const request = await this.xhr(url);
    return (request as any).responseText;
  }

  public static xhr(url: string) {
    const request = new XMLHttpRequest();
    return new Promise((resolve, reject) => {
      request.onreadystatechange = () => {
        if (request.readyState === 4) {
          if (request.status === 200) {
            try {
              resolve(request);
            } catch (e) {
              reject(e);
            }
          } else {
            reject(request);
          }
        }
      };
      const method = "GET";
      request.open(method, url);
      const params = {};
      request.send(params);
    })
  }
}
