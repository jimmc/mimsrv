/* Mimview app */

class KeyFunc {
  desc: string;
  f: () => void;
}

@Polymer.decorators.customElement('mimview-app')
class MimviewApp extends Polymer.Element {

  @Polymer.decorators.property({type: String})
  dialogContent: string = "";

  keyMap: {[key: string]: KeyFunc};

  ready() {
    super.ready();
    this.initKeyMap();
    this.addEventListener('mimdialog', this.onMimDialog.bind(this));
    this.$.main.addEventListener('keydown', this.keydown.bind(this));
  }

  showDialogHtml(html: string) {
    this.dialogContent = html;
    this.$.dialogContent.innerHTML = html;
    this.$.dialog.open();
  }

  showDialog(text: string) {
    this.dialogContent = text;
    this.$.dialog.open();
  }

  hideDialog() {
    this.dialogContent = "";
    this.$.dialog.close();
  }

  onMimDialog(e: any) {
    if (e.detail.html) {
      this.showDialogHtml(e.detail.html);
    } else {
      this.showDialog(e.detail.message);
    }
  }

  logout() {
    this.$.mimlogin.logout();
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
}
