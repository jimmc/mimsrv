/* Mim-dialog component */

@Polymer.decorators.customElement('mim-dialog')
class MimDialog extends Polymer.Element {

  open() {
    this.$.dialog.open();
  }

  close() {
    this.$.dialog.close();
  }
}
