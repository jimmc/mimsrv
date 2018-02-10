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
    this.$.image.addEventListener('mimkey', this.onMimKey.bind(this));
    this.$.nav.addEventListener('mimdialog', this.onMimDialog.bind(this));
  }

  showDialogHtml(html: string) {
    if (html === this.dialogContent) {
      this.hideDialog();
    } else {
      this.dialogContent = html;
      this.$.dialogContent.innerHTML = html;
    }
  }

  showDialog(text: string) {
    if (text === this.dialogContent) {
      this.hideDialog();
    } else {
      this.dialogContent = text;
    }
  }

  hideDialog() {
    this.dialogContent = "";
  }

  onMimDialog(e: any) {
    if (e.detail.html) {
      this.showDialogHtml(e.detail.html);
    } else {
      this.showDialog(e.detail.message);
    }
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
    this.addKey('?', 'List key bindings',
        this.showKeyBindings.bind(this));
  }

  addKey(key: string, desc: string, f: () => void) {
    const keyFunc = new KeyFunc();
    keyFunc.desc = desc;
    keyFunc.f = f;
    this.keyMap[key] = keyFunc;
  }

  onMimKey(e: any) {
    console.log('nav mimkey', e);
    const key = e.detail.key;
    const keyFunc = this.keyMap[key];
    if (keyFunc) {
      keyFunc.f();
    }
  }
}
