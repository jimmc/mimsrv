/* Nav component */

// ListItem is what we get back from the API list call.
interface ListItem {
  Name: string;
  IsDir: boolean;
  Size: number;
  ModTime: number;      // seconds since the epoch
  ModTimeStr: string;   // ModTime converted to string in the server
  Text: string;
  TextError: string;
}

interface ListResponse {
  IndexName: string;
  UnfilteredFileCount: number;
  Items: ListItem[];
}

// NavItem is what we maintain locally for our nav list.
interface NavItem {
  path: string;         // Full path to the this item
  name: string;         // Final component of the path
  level: number;
  expanded: boolean;
  pending: boolean;
  isDir: boolean;
  size: number;
  modTime: number;
  modTimeStr: string;
  text: string;
  textError: string;
  index: string;
  filtered: boolean;
  version: number;
}

@Polymer.decorators.customElement('mim-nav')
class MimNav extends Polymer.Element {

  @Polymer.decorators.property({type: Boolean, notify: true})
  loggedIn: boolean;

  @Polymer.decorators.property({type: String, notify: true})
  imgsrc: string;

  @Polymer.decorators.property({type: Object, notify: true})
  imgitem: NavItem | undefined;

  @Polymer.decorators.property({type: Object})
  imgsize: any;

  @Polymer.decorators.property({type: Array})
  rows: NavItem[] = [];

  @Polymer.decorators.property({type: Number})
  selectedIndex: number;

  ready() {
    super.ready();
    this.queryApiList('');
  }

  // Returns the delta count for this.rows
  async queryApiList(dir: string) {
    try {
      const listUrl = "/api/list/" + dir;
      const response = await ApiManager.xhrJson(listUrl);
      this.loggedIn = true;
      this.handleListResponse(dir, response);
    } catch (e) {
      this.loggedIn = false;
      console.log("Query failed:", e);
    }
  }

  handleListResponse(dir: string, list: ListResponse) {
    const navItems = list.Items.map(
        (listItem) => this.listToNav(listItem, dir));
    this.updateDirRows(dir, navItems, list);
  }

  listToNav(listItem: ListItem, dir: string): NavItem {
    const level = dir.split('/').length;
    const path = dir + '/' + listItem.Name;
    return {
      path,
      name: listItem.Name,
      level,
      expanded: false,
      isDir: listItem.IsDir,
      size: listItem.Size,
      modTime: listItem.ModTime,
      modTimeStr: listItem.ModTimeStr,
      text: listItem.Text,
      textError: listItem.TextError,
      version: 0,
    } as NavItem;
  }

  updateDirRows(dir: string, rows: NavItem[], list: ListResponse) {
    if (!dir) {
      this.rows = rows;
      return;
    }
    // We are updating in the middle somewhere, look for our dir,
    // replace its children with the new items, and expand it.
    const index = this.rows.findIndex((row) => row.path == dir);
    if (index < 0) {
      console.error("Can't find entry for dir", dir);
      return;
    }
    this.set(["rows", index, "index"], list.IndexName)
    this.set(["rows", index, "filtered"],
        list.UnfilteredFileCount != rows.length);
    const nextIndex = this.nextIndex(index);
    const updatedRows = this.rows.slice(0, index + 1)
      .concat(rows)
      .concat(this.rows.slice(nextIndex, this.rows.length));
    this.rows = updatedRows;
    this.set(['rows', index, 'expanded'], true);
  }

  // Looks at the level of the row at the specified index and returns the
  // index of the first following row with the same or lower level,
  // otherwise the length of the rows list.
  nextIndex(index: number) {
    const rowLevel = this.rows[index].level;
    let nextIndex = this.rows.findIndex((row, i) => {
      if (i <= index) {
        return false;
      }
      if (row.level <= rowLevel) {
        return true;
      }
      return false;
    });
    if (nextIndex < 0) {
      nextIndex = this.rows.length;
    }
    return nextIndex;
  }

  collapseRowAt(index: number) {
    const nextIndex = this.nextIndex(index);
    const updatedRows = this.rows.slice(0, index + 1)
      .concat(this.rows.slice(nextIndex, this.rows.length));
    this.rows = updatedRows;
    this.set(['rows', index, 'expanded'], false);
  }

  getRowClass(row: NavItem, selectedIndex: number) {
    let classList = ['nav-item'];
    if (row.isDir) {
      classList.push('dir');
    }
    const rowIndex = this.rows.indexOf(row);
    if (rowIndex >= 0) {
      if (rowIndex === selectedIndex) {
        classList.push('selected');
      }
    }
    return classList.join(' ');
  }

  indentsForRow(row: NavItem) {
    return new Array(row.level);
  }

  sizeAsString(row: NavItem) {
    if (row.size <= 999) {
      return row.size + "B";
    }
    if (row.size <= 9999) {
      return Math.round(row.size/10)/100 + "K";
    }
    if (row.size <= 99999) {
      return Math.round(row.size/100)/10 + "K";
    }
    if (row.size <= 999999) {
      return Math.round(row.size/1000) + "K";
    }
    if (row.size <= 9999999) {
      return Math.round(row.size/10000)/100 + "M";
    }
    if (row.size <= 99999999) {
      return Math.round(row.size/100000)/10 + "M";
    }
    return Math.round(row.size/1000000) + "M";
  }

  rowClicked(e: any) {
    if (e.clientX == 0 && e.clientY == 0) {
      // We get an on-click for an Enter key as well as a mouse click.
      // We want to handle them separately, so we check here to see if this
      // is a real mouse-click. If not, we ignore it here, and process it
      // separately elsewhere.
      return;
    }
    this.selectAt(e.model.index);
  }

  rowToggled(e: any) {
    this.selectAt(e.model.index);
    this.toggleCurrent();
  }

  selectAt(index: number) {
    this.selectedIndex = index;
    this.scrollRowIntoView(index);
    const row = this.rows[index];
    if (row.isDir) {
      this.setImageRow(undefined);
    } else {
      this.setImageRow(row);
    }
  }

  scrollRowIntoView(index: number) {
    if (index < 0 || index >= this.rows.length) {
      return;
    }
    const rowElements = this.$.listContainer.querySelectorAll('.nav-item');
    const rowElement = rowElements[index];
    if (rowElement.offsetTop < this.scrollTop) {
      rowElement.scrollIntoView(true);
    } else if (rowElement.offsetTop + rowElement.offsetHeight > this.scrollTop + this.offsetHeight) {
      rowElement.scrollIntoView(false);
    }
  }

  async toggleCurrent() {
    if (this.selectedIndex >= 0) {
      const row = this.rows[this.selectedIndex];
      const rowIndex = this.selectedIndex;
      if (row.isDir) {
        const preRowCount = this.rows.length;
        if (row.expanded) {
          this.collapseRowAt(this.selectedIndex);
        } else {
          this.set(['rows', rowIndex, 'pending'], true);
          await this.queryApiList(row.path);
          this.set(['rows', rowIndex, 'pending'], false);
        }
        const postRowCount = this.rows.length;
        return postRowCount - preRowCount;
      }
    }
    return 0
  }

  selectNext() {
    if (this.selectedIndex >= 0 && this.selectedIndex < this.rows.length - 1) {
      if (!this.rows[this.selectedIndex].isDir) {
        // If we are currently viewing a file, then we want to move to the
        // next file, even if it is in another folder.
        this.selectNextFile();
      } else {
        this.scrollRowIntoView(this.selectedIndex + 2);
        this.selectAt(this.selectedIndex + 1);
      }
    }
  }

  async selectNextFile() {
    if (this.selectedIndex >= 0 && this.selectedIndex < this.rows.length - 1) {
      this.selectedIndex = this.selectedIndex + 1;
      const row = this.rows[this.selectedIndex];
      if (row.isDir) {
        this.setImageRow(undefined);
        if (!row.expanded) {
          await this.toggleCurrent();
        }
        this.selectNextFile();
      } else {
        this.scrollRowIntoView(this.selectedIndex + 1);
        this.selectAt(this.selectedIndex);
      }
    }
  }

  selectPrevious() {
    if (this.selectedIndex > 0 && this.selectedIndex < this.rows.length) {
      if (!this.rows[this.selectedIndex].isDir) {
        // If we are currently viewing a file, then we want to move to the
        // previous file, even if it is in another folder.
        this.selectPreviousFile();
      } else {
        this.scrollRowIntoView(this.selectedIndex - 2);
        this.selectAt(this.selectedIndex - 1);
      }
    }
  }

  async selectPreviousFile() {
    if (this.selectedIndex > 0 && this.selectedIndex < this.rows.length) {
      const row = this.rows[this.selectedIndex - 1];
      const rowIndex = this.selectedIndex - 1;
      if (!row.isDir) {
        this.scrollRowIntoView(this.selectedIndex);
        this.scrollRowIntoView(this.selectedIndex - 2);
        return this.selectAt(this.selectedIndex - 1);
      }
      this.setImageRow(undefined);
      const prevIndexUnexpanded = this.findPreviousUnexpandedOrFile(
          this.selectedIndex);
      if (prevIndexUnexpanded >= 0) {
        const row = this.rows[prevIndexUnexpanded];
        if (!row.isDir) {
          this.scrollRowIntoView(prevIndexUnexpanded + 1);
          this.scrollRowIntoView(prevIndexUnexpanded - 1);
          return this.selectAt(prevIndexUnexpanded);
        }
        const preRowCount = this.rows.length;
        this.set(['rows', rowIndex, 'pending'], true);
        await this.queryApiList(row.path);
        this.set(['rows', rowIndex, 'pending'], false);
        const postRowCount = this.rows.length;
        const deltaRowCount = postRowCount - preRowCount;
        this.selectedIndex = this.selectedIndex + deltaRowCount;
        this.selectPreviousFile();
      }
    }
  }

  findPreviousUnexpandedOrFile(index: number) {
    index = index - 1;
    while (index >= 0) {
      const row = this.rows[index];
      if (!row.isDir || !row.expanded) {
        return index;
      }
      index = index - 1;
    }
    return index;
  }

  currentImagePathAndText() {
    if (this.selectedIndex < 0) {
      return null;
    }
    const row = this.rows[this.selectedIndex];
    if (row.isDir) {
      return null;
    }
    const itemPath = row.path;
    const lastDot = itemPath.lastIndexOf('.');
    const textPath = itemPath.substr(0, lastDot) + ".txt";
    return [itemPath, textPath, row.text];
  }

  currentFolderPathAndText() {
    let index = this.selectedIndex;
    if (index < 0) {
      return null;
    }
    let row = this.rows[index];
    // Walk backwards until we get to a folder
    while (!row.isDir && index > 0) {
      index = index - 1;
      row = this.rows[index];
    }
    const textPath = row.path + "/summary.txt";
    return [row.path, textPath, row.text];
  }

  updateText(itemPath: string, textPath: string, text: string) {
    const index = this.rows.findIndex((row) => row.path == itemPath);
    if (index < 0) {
      console.error("Can't find entry for item", itemPath);
      return;
    }
    this.set(["rows", index, "text"], text);
  }

  async rotateCurrent(value: string) {
    if (this.selectedIndex < 0) {
      console.log("No image selected")
      return
    }
    const row = this.rows[this.selectedIndex];
    if (row.isDir) {
      console.log("Can't rotate a directory")
      return
    }
    const lastSlash = row.path.lastIndexOf('/')
    const dir = row.path.substr(0, lastSlash)
    const indexFile = dir + "/index.mpr"
    try {
      const indexUrl = "/api/index" + indexFile;
      const formData = new FormData();
      formData.append("item", row.name);
      formData.append("action", "deltarotation");
      formData.append("value", value)
      formData.append("autocreate", "true");
      const options = {
        method: "POST",
        params: formData,
      };
      const response = await ApiManager.xhrJson(indexUrl, options);
      row.version = row.version + 1;
      this.imgsizeChanged();    // reload the image at the new version
    } catch (e) {
      console.error("rotation failed:", e)
    }
  }

  @Polymer.decorators.observe('imgsize')
  imgsizeChanged() {
    if (this.selectedIndex >= 0) {
      const row = this.rows[this.selectedIndex];
      if (!row.isDir) {
        this.selectAt(this.selectedIndex);
      }
    }
  }

  setImageRow(row?: NavItem) {
    let qParms = '';
    this.imgitem = row;
    if (this.imgsize) {
      const height = this.imgsize.height;
      const width = this.imgsize.width;
      qParms = '?w=' + width + '&h=' + height;
      if (row && row.version) {
        qParms = qParms + '&_=' + row.version;
      }
    }
    if (row && row.path) {
      this.imgsrc = "/api/image" + row.path + qParms;
    } else {
      this.imgsrc = '';
    }
  }

  showDialogHtml(html: string) {
    const detail = {html: html};
    this.dispatchEvent(new CustomEvent('mimdialog', {detail: detail}));
  }

  showDialog(msg: string) {
    const detail = {message: msg};
    this.dispatchEvent(new CustomEvent('mimdialog', {detail: detail}));
  }
}
