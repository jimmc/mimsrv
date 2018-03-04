/* Mim-dialog component */

interface MimDialogOptions {
  cancelLabel: string;
  confirmLabel: string;
  showTextarea: boolean;
  text: string;
  html: string;
  textarea: string,
}

@Polymer.decorators.customElement('mim-dialog')
class MimDialog extends Polymer.Element {

  @Polymer.decorators.property({type: String})
  cancelLabel: string = "Cancel";

  @Polymer.decorators.property({type: String})
  confirmLabel: string = "OK";

  @Polymer.decorators.property({type: String})
  text: string = "";

  @Polymer.decorators.property({type: String})
  html: string = "";

  @Polymer.decorators.property({type: String})
  textarea: string = "";

  @Polymer.decorators.property({type: Boolean})
  showTextarea: boolean = false;

  dialogResolve: (status: boolean) => void;
  dialogReject: () => void;

  open(opts: MimDialogOptions): Promise<boolean> {
    this.confirmLabel = opts.confirmLabel;
    this.cancelLabel = opts.cancelLabel;
    this.showTextarea = opts.showTextarea;
    this.text = opts.text;
    this.$.dialogHtml.innerHTML = opts.html;
    this.$.dialogTextarea.value = opts.textarea;
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

  textareaValue(): string {
    return this.$.dialogTextarea.value;
  }
}
