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

  lastResize = 0;       // Time of last resize
  maxResizeDelay = 500;    // We do at least one resize after this much time
  resizeTimeoutId = 0;  // Timer ID for dealing with resizes
  resizeTimeoutDelay = 100;  // Wait this long after event before resizing

  ready() {
    super.ready();
    window.addEventListener('resize', () => this.compressResizeEvents());
    this.handleResize();
    this.$.imgContainer.addEventListener('keydown', this.keydown.bind(this));
  }

  // When the user resizes the window, we get a stream of resize events.
  // We don't want to have to process them all, so we discard most of them.
  compressResizeEvents() {
    const now = Date.now();
    if (now > this.lastResize + this.maxResizeDelay) {
      // It has been long enough since we last resized, do it now.
      this.clearResizeTimeout();
      this.handleResize();
      return;
    }
    this.setResizeTimeout();
  }

  setResizeTimeout() {
    this.clearResizeTimeout();
    this.resizeTimeoutId = window.setTimeout(() => {
        this.clearResizeTimeout();
        this.handleResize();
      },
      this.resizeTimeoutDelay);
  }

  clearResizeTimeout() {
    if (this.resizeTimeoutId > 0) {
      window.clearTimeout(this.resizeTimeoutId);
      this.resizeTimeoutId = 0;
    }
  }

  handleResize() {
    this.lastResize = Date.now();
    const width = this.$.imgContainer.clientWidth;
    const height = this.$.imgContainer.clientHeight;
    this.imgsize = {
      width,
      height,
    } as ImageSize;
  }

  keydown(e: any) {
    this.dispatchEvent(new CustomEvent('mimkey', {detail: e}));
  }
}
