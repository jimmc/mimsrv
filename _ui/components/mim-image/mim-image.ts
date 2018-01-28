/* Mim-image component */

interface ImageSize {
  width: number;
  height: number;
}

@Polymer.decorators.customElement('mim-image')
class MimImage extends Polymer.Element {

  @Polymer.decorators.property({type: String})
  imgsrc: string;

  @Polymer.decorators.property({type: Object, notify: true})
  imgsize: ImageSize;

  ready() {
    super.ready();
    window.addEventListener('resize', () => this.handleResize());
    this.handleResize();
  }

  handleResize() {
    const width = this.$.img.parentElement.clientWidth;
    const height = this.$.img.parentElement.clientHeight;
    this.imgsize = {
      width,
      height,
    } as ImageSize;
    console.log('image size:', this.imgsize);
  }
}
