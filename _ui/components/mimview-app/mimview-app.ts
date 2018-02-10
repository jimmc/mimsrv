/* Mimview app */

@Polymer.decorators.customElement('mimview-app')
class MimviewApp extends Polymer.Element {

  @Polymer.decorators.property({type: String})
  dialogContent: string = "";

  ready() {
    super.ready();
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

  onMimKey(e: any) {
    this.$.nav.onMimKey(e);
  }

  onMimDialog(e: any) {
    if (e.detail.html) {
      this.showDialogHtml(e.detail.html);
    } else {
      this.showDialog(e.detail.message);
    }
  }
}
