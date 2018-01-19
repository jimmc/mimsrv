/* Nav component */

interface ListResponse {
  Names: string[];
}

@Polymer.decorators.customElement('mim-nav')
class MimNav extends Polymer.Element {

  @Polymer.decorators.property({type: String, notify: true})
  imgsrc: string;

  @Polymer.decorators.property({type: Array})
  rows: string[] = [];

  ready() {
    super.ready();
    this.queryApiList('');
  }

  queryApiList(dir: string) {
    const listUrl = "/api/list/" + dir;
    ApiManager.xhrJson(listUrl)
      .then((response) => this.handleListResponse(dir, response));
  }

  handleListResponse(dir: string, list: ListResponse) {
    console.log("list:", list);
    const fullPaths = list.Names.map((entry) => dir + "/" + entry);
    this.setRows(fullPaths);
  }

  setRows(rr: string[]) {
    this.rows = rr;
  }

  rowClicked(e: MouseEvent) {
    const target = e.target as HTMLSpanElement;
    const text = target.innerText;
    console.log('Clicked:', text);
    if (text.endsWith('.jpg')) {
      this.setImageSource(text);
    } else {
      this.queryApiList(text);
    }
  }

  setImageSource(src: string) {
    /*
    const height = this.$.img.clientHeight;
    const width = this.$.img.clientWidth;
    console.log('Height:', height, ' Width:', width);
    */
    this.imgsrc = "/api/image/" + src;
  }
}
