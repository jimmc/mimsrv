/* Mimview app */

@Polymer.decorators.customElement('mimview-app')
class MimviewApp extends Polymer.Element {

  ready() {
    super.ready();
    this.$.image.addEventListener('mimkey', this.onMimKey.bind(this));
  }

  onMimKey(e: any) {
    this.$.nav.onMimKey(e);
  }
}
