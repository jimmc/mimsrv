/* Nav component */

// ListItem is what we get back from the API list call.
interface ListItem {
  Name: string;
  Path: string;
  IndexPath: string;
  IndexEntry: string;
  IsDir: boolean;
  Size: number;
  Type: string;
  ModTime: number;      // seconds since the epoch
  ModTimeStr: string;   // ModTime converted to string in the server
  ExifDateTime: Date;
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
  indexPath: string;    // API path to index containing this item, if not default index
  indexEntry: string;   // Path to file relative to index, if not default index
  level: number;
  expanded: boolean;
  pending: boolean;
  isDir: boolean;
  size: number;
  type: string;
  modTime: number;
  modTimeOutOfOrder: boolean;
  modTimeStr: string;
  exifDateTime: Date;
  exifDateTimeOutOfOrder: boolean;
  text: string;
  textWithoutFlags: string;
  textError: string;
  index: string;
  filtered: boolean;
  version: number;
  zoom: boolean;        // When true, request unscaled image
}

// This is the info we pass to mim-image so it can display the image
interface ImageInfo {
  path: string;
  version: number,
  zoom: boolean;
}

@Polymer.decorators.customElement('mim-nav')
class MimNav extends Polymer.Element {

  @Polymer.decorators.property({type: Boolean, notify: true})
  loggedIn: boolean;

  @Polymer.decorators.property({type: Object, notify: true})
  imginfo: ImageInfo | undefined;

  @Polymer.decorators.property({type: Object, notify: true})
  imgitem: NavItem | undefined;

  @Polymer.decorators.property({type: Object, notify: true})
  preimginfo: ImageInfo | undefined;

  @Polymer.decorators.property({type: Object, notify: true})
  preimgitem: NavItem | undefined;

  @Polymer.decorators.property({type: Array})
  rows: NavItem[] = [];

  @Polymer.decorators.property({type: Number})
  selectedIndex: number;

  @Polymer.decorators.property({type: Object})
  route: any;

  @Polymer.decorators.property({type: Boolean})
  showExifDateTimes: boolean;

  publishChannel: BroadcastChannel;
  subscribeChannel: BroadcastChannel;
  requestedLocation: string;
  preloadDirectCount = 1;       // Preload this many images when we direct-access an image.
  preloadMovingCount = 3;       // Preload this many images when we are moving through a list.

  ready() {
    super.ready();
    this.setupRequestedLocation();
    this.queryApiList('');
    this.setupChannels();
    this.showExifDateTimes = (this.getQueryParm('showExifDateTimes') != null);
  }

  setupRequestedLocation() {
    let loc = this.route.__queryParams.loc;
    if (typeof loc === 'undefined') {
      return;
    }
    // Make sure requestedLocation starts with a slash and doesn't end with one
    if (loc.endsWith('/')) {
      loc = loc.substr(0, loc.length - 1);
    }
    if (!loc.startsWith('/')) {
      loc = '/' + loc;
    }
    this.requestedLocation = loc;
    this.set(['route', '__queryParams', 'loc'], loc);
  }

  setupChannels() {
    // We want to publish if we are not subscribing, or if
    // we are subscribing and are publishing to a different channel.
    const pubChan = this.getQueryParm('publish');
    const subChan = this.getQueryParm('subscribe');
    const shouldPublish = subChan == null || (pubChan !== null && subChan != pubChan);
    if (shouldPublish) {
      this.publishChannel = this.createNamedChannel('publish')
    }
    if (this.getQueryParm('subscribe') !== null) {
      this.subscribeChannel = this.createNamedChannel('subscribe')
      this.subscribeChannel.onmessage = (b) => this.receiveBroadcast(b)
    }
  }

  createNamedChannel(parmName: string) {
    const parmValue = this.getQueryParm(parmName);
    const channelName = !!parmValue ? ('mimview-' + parmValue) : 'mimview';
    return new BroadcastChannel(channelName);
  }

  getQueryParm(parmName: string) {
    const query = location.search.substring(1);
    const parms = query.split('&');
    for (let i = 0; i < parms.length; i++) {
      const kv = parms[i].split('=');
      if (kv[0] === parmName) {
        if (kv.length >= 2) {
          return kv[1];
        } else {
          return '';
        }
      }
    }
    return null;
  }

  publishLocation(loc: string) {
    if (this.publishChannel) {
      this.publishChannel.postMessage(loc);
    }
  }

  receiveBroadcast(b: MessageEvent) {
    const data: string = b.data;
    this.selectLocation(data);
  }

  // Returns the delta count for this.rows
  async queryApiList(dir: string) {
    // console.log("queryApiList:", dir);
    try {
      const listUrl = "/api/list/" + dir;
      const response = await ApiManager.xhrJson(listUrl);
      this.loggedIn = true;
      this.handleListResponse(dir, response);
    } catch (e) {
      if (e.status == 401 /*Unauthorized*/ || e.status == 403 /*Forbidden*/) {
        this.loggedIn = false;
      } else {
        alert('Error loading ' + dir + ': ' + e.responseText)
      }
      console.log("Query failed:", e);
    }
  }

  handleListResponse(dir: string, list: ListResponse) {
    const navItems = list.Items.map(
        (listItem) => this.listToNav(listItem, dir));
    for (let i = 1; i < navItems.length; i++) {
        if (!navItems[i].isDir && !navItems[i-1].isDir &&
                navItems[i].modTime < navItems[i-1].modTime) {
            navItems[i].modTimeOutOfOrder = true;
        }
        if (!navItems[i].isDir && !navItems[i-1].isDir &&
                navItems[i].exifDateTime < navItems[i-1].exifDateTime) {
            navItems[i].exifDateTimeOutOfOrder = true;
        }
    }
    this.updateDirRows(dir, navItems, list);
  }

  listToNav(listItem: ListItem, dir: string): NavItem {
    const level = dir.split('/').length;
    const path = listItem.Path || dir + '/' + listItem.Name;
    return {
      path,
      name: listItem.Name,
      indexPath: listItem.IndexPath,
      indexEntry: listItem.IndexEntry,
      level,
      expanded: false,
      isDir: listItem.IsDir || listItem.Type == 'index',
      size: listItem.Size,
      type: listItem.Type,
      modTime: listItem.ModTime,
      modTimeStr: this.dtFromFlags(listItem.Text) || listItem.ModTimeStr,
      exifDateTime: listItem.ExifDateTime,
      text: listItem.Text,
      textWithoutFlags: this.stripFlags(listItem.Text),
      textError: listItem.TextError,
      version: 0,
    } as NavItem;
  }

  stripFlags(textWithFlags: string): string {
    const lines = textWithFlags.split('\n');
    while (lines && lines[0] && lines[0].startsWith('!')) {
      lines.shift()
    }
    return lines.join('\n')
  }

  dtFromFlags(textWithFlags: string): string {
    if (!textWithFlags) {
      return ''
    }
    const lines = textWithFlags.split('\n');
    for (const l in lines) {
      const line = lines[l]
      if (line && line.startsWith('!dt=')) {
        return line.substring(4);
      }
    }
    return ''
  }

  updateDirRows(dir: string, rows: NavItem[], list: ListResponse) {
    if (!dir) {
      this.rows = rows;
      this.handleRequestedLocation(dir);
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
    this.handleRequestedLocation(dir);
  }

  // Looks to see if we have a requested location, and if so, whether we
  // still need to descend down that path.
  handleRequestedLocation(dir: string) {
    if (!this.requestedLocation) {
      // console.log("No requestedLocation");
      return;           // No requested location, so no need to check anything here.
    }
    const index = this.rows.findIndex((row) => this.rowMatchesLocation(row, this.requestedLocation));
    if (index >= 0) {
      // console.log("Found requestedLocation in current rows");
      this.openPath(this.requestedLocation);
      this.requestedLocation = '';
      return;
    }
    if (dir == this.requestedLocation) {
      // console.log("Made it to requestedLocation");
      this.requestedLocation = '';
      return;           // We made it to the requested location
    }
    // console.log("At ", dir, ", still need to get to requestedLocation ", this.requestedLocation);
    const dirParts = dir.split('/');
    const locParts = this.requestedLocation.split('/');
    if (dirParts.length >= locParts.length) {
      console.log("Length mismatch in requestedLocation?");
      this.requestedLocation = '';
      return;
    }
    const nextDir = dir + '/' + locParts[dirParts.length];
    // console.log("nextDir:", nextDir);
    this.queryApiList(nextDir);
  }

  // Returns true of the given row matches the requested location.
  rowMatchesLocation(row: NavItem, location: string) {
    if (row.indexPath) {
      return ('/' + row.indexPath + '/' + row.indexEntry) == location;
    }
    return row.path == location;
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
    if (row.type == 'index') {
      classList.push('indexfile');
    }
    if (row.modTimeOutOfOrder) {
      classList.push('outoforder');
    }
    if (row.exifDateTimeOutOfOrder) {
      classList.push('exifoutoforder');
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
    this.preloadN(e.model.index, this.preloadDirectCount, 1);
    this.preloadN(e.model.index, this.preloadDirectCount, -1);
  }

  rowToggled(e: any) {
    this.selectAt(e.model.index);
    this.preloadN(e.model.index, this.preloadDirectCount, 1);
    this.preloadN(e.model.index, this.preloadDirectCount, -1);
    this.toggleCurrent();
  }

  selectAt(index: number) {
    this.selectedIndex = index;
    if (index < 0) {
      this.set(['route', '__queryParams', 'loc'], '');
      // this.setImageRow(undefined);
      return;
    }
    this.scrollRowIntoView(index);
    const row = this.rows[index];
    let path = row.path;
    if (row.indexPath) {
      path = '/' + row.indexPath + '/' + row.indexEntry;
    }
    this.set(['route', '__queryParams', 'loc'], path);
    if (row.isDir) {
      this.setImageRow(undefined);
    } else {
      this.setImageRow(row);
    }
  }

  // Preload multiple images. Delta should be 1 or -1 for preloading
  // when moving forwards or backwards. Preloads are relative to index.
  preloadN(index: number, count: number, delta: number) {
    while (count > 0) {
        index += delta;
        this.preloadAt(index)
        count--;
    }
  }

  // If the given index is a valid image, use the preload
  // mechanism to preload it.
  preloadAt(index: number) {
    if (index < 0 || index >= this.rows.length) {
      this.setPreImageRow(undefined);
      return;
    }
    const row = this.rows[index];
    if (row.isDir) {
      this.setPreImageRow(undefined);
    } else {
      this.setPreImageRow(row);
    }
  }

  async selectLocation(loc: string) {
    if (this.selectedIndex >= 0) {
      const row = this.rows[this.selectedIndex];
      if (this.rowMatchesLocation(row, loc)) {
        // Already selected, do nothing
        return
      }
    }
    if (!loc) {
      this.selectAt(-1);        // Deselect
      return;
    }
    await this.openPath(loc);
  }

  async openPath(path: string) {
    const x = path.lastIndexOf('/');
    if (x < 0) {
      return;
    }
    const parentPath = path.substr(0, x);
    await this.openPath(parentPath);
    const index = this.rows.findIndex((row) => this.rowMatchesLocation(row, path));
    if (index < 0) {
      console.log("path not found:", path);
      return;
    }
    const row = this.rows[index];
    if (row.isDir) {
      if (!row.expanded) {
        await this.toggleAt(index);     // expand the folder
      }
    } else {
      const oldIndex = this.selectedIndex;
      this.selectAt(index);     // Select the image
      // If we notice we are moving, preload a few more images.
      if (index == oldIndex + 1) {
        this.preloadN(index, this.preloadMovingCount, 1);
      } else if (index == oldIndex - 1) {
        this.preloadN(index, this.preloadMovingCount, -1);
      } else {
        this.preloadN(index, this.preloadDirectCount, 1);
        this.preloadN(index, this.preloadDirectCount, -1);
      }
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
      return await this.toggleAt(this.selectedIndex);
    }
    return 0
  }

  async toggleAt(rowIndex: number) {
    const row = this.rows[rowIndex];
    if (row.isDir) {
      const preRowCount = this.rows.length;
      if (row.expanded) {
        this.collapseRowAt(rowIndex);
      } else {
        this.set(['rows', rowIndex, 'pending'], true);
        await this.queryApiList(row.path);
        this.set(['rows', rowIndex, 'pending'], false);
      }
      const postRowCount = this.rows.length;
      return postRowCount - preRowCount;
    } else {
      return 0
    }
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
        // selectAt updates selectedIndex
        this.preloadN(this.selectedIndex, this.preloadMovingCount, 1);
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
        this.preloadN(this.selectedIndex, this.preloadMovingCount, 1);
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
        // selectAt updates this.selectedIndex
        this.preloadN(this.selectedIndex, this.preloadMovingCount, -1);
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
        this.selectAt(this.selectedIndex - 1);
        // selectAt updates this.selectedIndex
        this.preloadN(this.selectedIndex, this.preloadMovingCount, -1);
        return;
      }
      this.setImageRow(undefined);
      const prevIndexUnexpanded = this.findPreviousUnexpandedOrFile(
          this.selectedIndex);
      if (prevIndexUnexpanded >= 0) {
        const row = this.rows[prevIndexUnexpanded];
        if (!row.isDir) {
          this.scrollRowIntoView(prevIndexUnexpanded + 1);
          this.scrollRowIntoView(prevIndexUnexpanded - 1);
          this.selectAt(prevIndexUnexpanded);
          this.preloadN(prevIndexUnexpanded, this.preloadMovingCount, -1);
          return;
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
    // Text updates are done directly on the image location and its text file,
    // so we compare row.path directly instead of using this.rowMatchesLocation.
    const index = this.rows.findIndex((row) => row.path == itemPath);
    if (index < 0) {
      console.error("Can't find entry for item", itemPath);
      return;
    }
    this.set(["rows", index, "text"], text);
    this.set(["rows", index, "textWithoutFlags"], this.stripFlags(text));
    if (index == this.selectedIndex) {
      // Propagate to imgitem to update the caption right away.
      if (this.imgitem && this.imgitem.text == text) {
        this.notifyPath("imgitem.text", text);
      } else {
        this.set(["imgitem", "text"], text);
      }
    }
    // If we overrode the modTimeStr from the server with a dt string
    // from the image text file, we lost that server-provided info.
    // If we don't have the dt string any more, we have nothing else,
    // so we just leave it there.
    this.set(["rows", index, "modTimeStr"],
        this.dtFromFlags(text) || this.rows[index].modTimeStr);
    this.updateImageAliases(index);
  }

  async dropCurrent() {
    if (this.selectedIndex < 0) {
      console.log("No image selected")
      return
    }
    const row = this.rows[this.selectedIndex];
    if (row.isDir) {
      console.log("Can't drop a directory")
      return
    }
    let indexFile = ''
    if (row.indexPath) {
      // Custom index file
      indexFile = "/" + row.indexPath
    } else {
      // Default index file
      const lastSlash = row.path.lastIndexOf('/')
      const dir = row.path.substr(0, lastSlash)
      indexFile = dir + "/index.mpr"
    }
    try {
      const indexUrl = "/api/index" + indexFile;
      const formData = new FormData();
      formData.append("item", row.indexEntry || row.name);
      formData.append("action", "drop");
      formData.append("autocreate", "true");
      const options = {
        method: "POST",
        params: formData,
      };
      const response = await ApiManager.xhrJson(indexUrl, options);
      // After success from the server, drop the row from our index also.
      this.splice('rows', this.selectedIndex, 1);
      // We don't change this.selectedIndex, so we will display the next row,
      // unless we are at the end, in which case we back up by one.
      if (this.selectedIndex >= this.rows.length) {
        this.selectedIndex = this.selectedIndex - 1;
      }
      this.redisplayCurrent();
    } catch (e) {
      console.error("drop failed:", e)
    }
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
    // We always do the rotations in the default index, because the
    // server always looks in the default index.
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
      // If there are any custom indexes, there may be other places in
      // the currently displayed list that reference the same image, so
      // make sure we update the version number there also.
      this.updateImageAliases(this.selectedIndex);
      this.redisplayCurrent();
    } catch (e) {
      console.error("rotation failed:", e)
    }
  }

  // We want to keep the version numbers and texts of aliases images in sync,
  // so that when we change one (in particular, rotation) and then
  // look at the other, we will see that change there as well.
  updateImageAliases(index: number) {
    const row = this.rows[index];
    // Look for the same row.path in other entries.
    for (let i = 0; i < this.rows.length; i++) {
      if (i == index) {
        continue;
      }
      // We specifically match on path here rather than calling
      // this.rowMatchesLocation, because we are looking for aliases.
      if (this.rows[i].path == row.path) {
        this.rows[i].version = row.version;
        // When updating text, we want to propagate the change to the ui.
        this.set(["rows", i, "text"], row.text)
        this.set(["rows", i, "textWithoutFlags"], row.textWithoutFlags);
        if (i == this.selectedIndex) {
          // Propagate to imgitem to update the caption right away.
          if (this.imgitem && this.imgitem.text == row.text) {
            this.notifyPath("imgitem.text", row.text);
          } else {
            this.set(["imgitem", "text"], row.text);
          }
        }
      }
    }
  }

  zoomCurrent() {
    if (this.selectedIndex < 0) {
      return
    }
    const row = this.rows[this.selectedIndex];
    row.zoom = !row.zoom;
    row.version = row.version + 1;
    this.updateImageAliases(this.selectedIndex);  // Keep version numbers of aliases in sync.
    this.redisplayCurrent();
  }

  redisplayCurrent() {
    this.selectAt(this.selectedIndex);  // redisplay
    this.preloadN(this.selectedIndex, this.preloadDirectCount, 1);
    this.preloadN(this.selectedIndex, this.preloadDirectCount, -1);
  }

  // Set the row item so that the corresponding image gets displayed.
  setImageRow(row?: NavItem) {
    this.imgitem = row;
    if (!row || !row.path) {
      this.imginfo = undefined;
      this.publishLocation('');
      return
    }
    this.imginfo = {
      path: row.path,
      version: row.version,
      zoom: row.zoom,
    } as ImageInfo;
    let loc = row.path;
    if (row.indexPath) {
      loc = '/' + row.indexPath + '/' + row.indexEntry;
    }
    this.publishLocation(loc);
  }

  // Preload the image on the specified row.
  setPreImageRow(row?: NavItem) {
    this.preimgitem = row;
    if (!row || !row.path) {
      this.preimginfo = undefined;
      return
    }
    this.preimginfo = {
      path: row.path,
      version: row.version,
      zoom: row.zoom,
    } as ImageInfo;
  }

  // Preload the image on the specified row.
  setPre2ImageRow(row?: NavItem) {
    this.setPreImageRow(row);
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
