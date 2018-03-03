/* Mim-dialog component */

@Polymer.decorators.customElement('mim-dialog')
class MimDialog extends Polymer.Element {

  @Polymer.decorators.property({type: String})
  cancelLabel: string = "Cancel";

  @Polymer.decorators.property({type: String})
  confirmLabel: string = "OK";

  dialogResolve: (status: boolean) => void;
  dialogReject: () => void;

  open(): Promise<boolean> {
    return new Promise((resolve, reject) => {
      this.dialogResolve = resolve;
      this.dialogReject = reject;
      this.$.dialog.open();
    })
  }

  confirm() {
    this.$.dialog.close();
    if (this.dialogResolve) {
      this.dialogResolve(true);
    }
  }

  cancel() {
    this.$.dialog.close();
    if (this.dialogResolve) {
      this.dialogResolve(false);
    }
  }
}
