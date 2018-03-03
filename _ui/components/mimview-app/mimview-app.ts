/* Mimview app */

class KeyFunc {
  desc: string;
  f: () => void;
}

@Polymer.decorators.customElement('mimview-app')
class MimviewApp extends Polymer.Element {

  @Polymer.decorators.property({type: String})
  dialogContent: string = "";

  @Polymer.decorators.property({type: Object})
  imgitem: NavItem;

  @Polymer.decorators.property({type: Boolean})
  hascaption: boolean;

  keyMap: {[key: string]: KeyFunc};

  ready() {
    super.ready();
    this.initKeyMap();
    this.addEventListener('mimdialog', this.onMimDialog.bind(this));
    this.$.main.addEventListener('keydown', this.keydown.bind(this));
    this.$.image.addEventListener('mimchecklogin', this.checkLogin.bind(this));
  }

  async showDialogHtml(html: string) {
    this.dialogContent = html;
    this.$.dialogContent.innerHTML = html;
    const ok = await this.$.dialog.open();
    console.log("dialog done, ok =", ok);
    return ok;
  }

  async showDialog(text: string) {
    this.dialogContent = text;
    const ok = await this.$.dialog.open();
    console.log("dialog done, ok =", ok);
    return ok;
  }

  hideDialog() {
    this.dialogContent = "";
    this.$.dialog.cancel();
  }

  async onMimDialog(e: any) {
    let ok: boolean;
    if (e.detail.html) {
      ok = await this.showDialogHtml(e.detail.html);
    } else {
      ok = await this.showDialog(e.detail.message);
    }
    if (e.callback) {
      e.callback(ok);
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

  toggleCurrent() {
    this.$.nav.toggleCurrent();
  }

  selectNext() {
    this.$.nav.selectNext();
  }

  selectPrevious() {
    this.$.nav.selectPrevious();
  }

  rotate(value: string) {
    this.$.nav.rotateCurrent(value);
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
    this.showDialogHtml(helpString);
  }

  initKeyMap() {
    this.keyMap = {};
    this.addKey('ArrowDown', 'Display the next image',
        this.selectNext.bind(this));
    this.addKey('ArrowUp', 'Display the previous image',
        this.selectPrevious.bind(this));
    this.addKey('Enter', 'Open or close the current folder',
        this.toggleCurrent.bind(this));
    this.addKey('?', 'List key bindings',
        this.showKeyBindings.bind(this));
    this.addKey('x', 'Logout',
        this.logout.bind(this));
    this.addKey('r', 'Rotate 90 degrees counterclockwise',
        () => this.rotate("+r"))
    this.addKey('R', 'Rotate 90 degrees clockwise',
        () => this.rotate("-r"))
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
    // console.log("Key: ", key);
    const keyFunc = this.keyMap[key];
    this.hideDialog();
    if (keyFunc) {
      keyFunc.f();
    }
  }

  @Polymer.decorators.observe('imgitem')
  imgitemChanged() {
    this.hascaption = !!(this.imgitem);
    this.$.image.handleResize();
  }
}
