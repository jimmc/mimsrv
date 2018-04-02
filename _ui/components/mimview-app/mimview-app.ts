/* Mimview app */

class KeyFunc {
  desc: string;
  f: () => void;
}

@Polymer.decorators.customElement('mimview-app')
class MimviewApp extends Polymer.Element {

  @Polymer.decorators.property({type: Object})
  imgitem: NavItem;

  @Polymer.decorators.property({type: Boolean})
  showcaption: boolean;

  @Polymer.decorators.property({type: Boolean})
  loggedIn: boolean;

  @Polymer.decorators.property({type: Boolean})
  showPlayButton: boolean = false;

  @Polymer.decorators.property({type: Boolean})
  showVideoPlayer: boolean = false;

  @Polymer.decorators.property({type: String})
  videoSource: string;

  allowcaption = true;
  hascaption = false;

  keyMap: {[key: string]: KeyFunc};

  ready() {
    super.ready();
    this.initKeyMap();
    this.addEventListener('mimdialog', this.onMimDialog.bind(this));
    this.$.nav.addEventListener('mimdialog', this.onMimDialog.bind(this));
    this.$.main.addEventListener('keydown', this.keydown.bind(this));
    this.$.image.addEventListener('mimchecklogin', this.checkLogin.bind(this));
  }

  async showHtmlDialog(html: string) {
    const opts = {
      showTextarea: false,
      confirmLabel: '',
      cancelLabel: 'Dismiss',
      text: '',
      html: html,
    } as MimDialogOptions;
    return this.showDialog(opts);
  }

  async showTextDialog(text: string) {
    const opts = {
      showTextarea: false,
      confirmLabel: '',
      cancelLabel: 'Dismiss',
      text: text,
      html: '',
    } as MimDialogOptions;
    return this.showDialog(opts);
  }

  async showTextareaDialog(label: string, text: string) {
    const opts = {
      showTextarea: true,
      confirmLabel: 'OK',
      cancelLabel: 'Cancel',
      textarea: text,
      text: label,
      html: '',
    } as MimDialogOptions;
    return this.showDialog(opts);
  }

  async showDialog(opts: MimDialogOptions) {
    return this.$.dialog.open(opts);
  }

  hideDialog() {
    this.$.dialog.cancel();
  }

  async onMimDialog(e: any) {
    let ok: boolean;
    if (e.detail.html) {
      ok = await this.showHtmlDialog(e.detail.html);
    } else {
      ok = await this.showTextDialog(e.detail.message);
    }
  }

  resetMenu(item: number) {
    this.$.menuselection.select(item);
  }

  helpClicked() {
    this.resetMenu(0);
    this.showKeyBindings();
  }

  logoutClicked() {
    this.resetMenu(1);
    this.logout();
  }

  logout() {
    this.$.mimlogin.logout();
  }

  checkLogin() {
    this.$.mimlogin.checkStatus();
  }

  async putText(textFile: string, content: string) {
    try {
      const putTextUrl = "/api/text" + textFile;
      const formData = new FormData();
      formData.append("content", content);
      const options = {
        method: "PUT",
        params: formData,
      };
      const response = await ApiManager.xhrJson(putTextUrl, options);
      console.log("success, response:", response)
      return true;
    } catch (e) {
      console.error("putText failed:", e)
      return false;
    }
  }

  rotateCurrent(r: string) {
    if (!this.$.mimlogin.hasPermission('edit')) {
      this.showTextDialog('You do not have permission to edit');
      return
    }
    this.$.nav.rotateCurrent(r);
  }

  toggleAllowCaption() {
    this.allowcaption = !this.allowcaption;
    this.imgitemChanged();
  }

  async editImageDescription() {
    if (!this.$.mimlogin.hasPermission('edit')) {
      this.showTextDialog('You do not have permission to edit');
      return
    }
    this.editDescription(this.$.nav.currentImagePathAndText());
  }

  async editFolderDescription() {
    if (!this.$.mimlogin.hasPermission('edit')) {
      this.showTextDialog('You do not have permission to edit');
      return
    }
    this.editDescription(this.$.nav.currentFolderPathAndText());
  }

  async editDescription(pathAndText: [string]) {
    if (!pathAndText) {
      console.log("no current item")
      return;
    }
    const itemPath = pathAndText[0];
    const textPath = pathAndText[1];
    const currentText = pathAndText[2];
    console.log("textPath ", textPath);

    const ok = await this.showTextareaDialog("Description for " + itemPath, currentText);
    if (!ok) {
      console.log("editDescription canceled");
      return
    }
    const text = this.$.dialog.textareaValue();
    console.log("editDescription gets ", text);
    if (await this.putText(textPath, text)) {
      this.$.nav.updateText(itemPath, textPath, text);
    }
  }

  fullscreen() {
    const el = this.$.display;
    const f = el.requestFullscreen || el.webkitRequestFullscreen ||
      el.mozRequestFullscreen || el.msRequestFullscreen;
    f.call(el);
  }

  showKeyBindings() {
    var keys = [];
    for (var key in this.keyMap) {
      if (this.keyMap.hasOwnProperty(key)) {
        keys.push(key);
      }
    }
    keys.sort();
    const helpString = keys.map((key: any) => {
      const entry = this.keyMap[key];
      return key + ": " + entry.desc + "<br>";
    }).join('\n');
    this.showHtmlDialog(helpString);
  }

  initKeyMap() {
    this.keyMap = {};
    this.addKey('ArrowDown', 'Display the next image',
        () => this.$.nav.selectNext());
    this.addKey('ArrowUp', 'Display the previous image',
        () => this.$.nav.selectPrevious());
    this.addKey('Enter', 'Open or close the current folder',
        () => this.$.nav.toggleCurrent());
    this.addKey('?', 'List key bindings',
        this.showKeyBindings.bind(this));
    this.addKey('c', 'Toggle caption',
        () => this.toggleAllowCaption());
    this.addKey('e', 'Edit the image description',
        () => this.editImageDescription());
    this.addKey('E', 'Edit the folder description',
        () => this.editFolderDescription());
    this.addKey('f', 'Fullscreen mode',
        () => this.fullscreen());
    this.addKey('r', 'Rotate 90 degrees counterclockwise',
        () => this.rotateCurrent("+r"));
    this.addKey('R', 'Rotate 90 degrees clockwise',
        () => this.rotateCurrent("-r"));
    this.addKey('x', 'Logout',
        this.logout.bind(this));
    this.addKey('z', 'Zoom to unscaled image or back',
        () => this.$.nav.zoomCurrent());
  }

  addKey(key: string, desc: string, f: () => void) {
    const keyFunc = new KeyFunc();
    keyFunc.desc = desc;
    keyFunc.f = f;
    this.keyMap[key] = keyFunc;
  }

  keydown(e: any) {
    e.preventDefault(); // Prevent list from doing default scrolling on arrows
    const key = e.key;
    const modifierKeys = ['Shift', 'Control', 'Meta', 'Alt', 'CapsLock'];
    if (modifierKeys.indexOf(key) >= 0) {
      // Ignore presses of the modifier keys
      return;
    }
    const keyFunc = this.keyMap[key];
    this.hideDialog();
    if (keyFunc) {
      keyFunc.f();
    } else {
      console.log("Key", key, "not bound");
    }
  }

  playClicked() {
    this.showVideoPlayer = true;
    this.showPlayButton = false;
    this.videoSource = '/api/video' + this.imgitem.path;
    this.$.videoPlayer.load();
  }

  @Polymer.decorators.observe('imgitem')
  imgitemChanged() {
    this.showVideoPlayer = false;
    this.hascaption = !!(this.imgitem);
    this.showcaption = this.hascaption && this.allowcaption;
    this.$.image.handleResize();
    this.showPlayButton = this.imgitem && this.imgitem.type == 'video';
  }

  @Polymer.decorators.observe('loggedIn')
  loggedInChanged() {
    if (this.loggedIn) {
      this.$.mimlogin.checkStatus();    // Load our logged-in permissions
    }
  }

  imgOverflowClass(): string {
    if (this.imgitem && this.imgitem.zoom) {
      return 'overflowscroll';
    } else {
      return 'overflowhidden';
    }
  }
}
