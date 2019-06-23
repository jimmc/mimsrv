/* Mim-image component */

interface ImageSize {
  width: number;
  height: number;
}

@Polymer.decorators.customElement('mim-image')
class MimImage extends Polymer.Element {

  // imgsrc is the API path to the current image to be displayed
  @Polymer.decorators.property({type: String})
  imgsrc: string;

  // imginfo is our input data from which we generate the API path for the current image
  @Polymer.decorators.property({type: Object})
  imginfo: any;

  // preimgsrc is the API path for preloading the next image
  @Polymer.decorators.property({type: String})
  preimgsrc: string;

  // preimginfo is our input data for preloading the next image
  @Polymer.decorators.property({type: Object})
  preimginfo: any;

  lastResize = 0;       // Time of last resize
  maxResizeDelay = 500;    // We do at least one resize after this much time
  resizeTimeoutId = 0;  // Timer ID for dealing with resizes
  resizeTimeoutDelay = 100;  // Wait this long after event before resizing

  ready() {
    super.ready();
    window.addEventListener('resize', () => this.compressResizeEvents());
  }

  connectedCallback() {
    super.connectedCallback();
    this.handleResize();
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
    const width = this.offsetWidth;
    const height = this.offsetHeight;
    this.imginfoChanged();
  }

  // When imginfo changes, we load a new current image.
  @Polymer.decorators.observe('imginfo')
  imginfoChanged() {
    this.imgsrc = ''    // Clear it first so size calculations work correctly.
    this.imgsrc = this.imginfoToImgsrc(this.imginfo)
  }

  // When preimginfo changes, we preload the next image.
  @Polymer.decorators.observe('preimginfo')
  preimginfoChanged() {
    this.preimgsrc = ''
    this.preimgsrc = this.imginfoToImgsrc(this.preimginfo)
  }

  // Given an imginfo, generate the API source string to load that image.
  // For unzoomed images, this string includes the image size, which is based
  // on the size of the img element in which it is being displayed.
  // The image size changes based on the size of the caption, which means
  // this may not be the right size when preloading the next image if the
  // caption on that image is not the same height. Perhaps some day that
  // will get fixed, but right now this code is taking the easy approach
  // on the assumption that much of the time the next image will have a
  // similar size caption.
  imginfoToImgsrc(imginfo: any) {
    if (!imginfo) {
      return '';
    }
    const row = imginfo;
    let qParms = '';
    if (!row.zoom) {
      const height = this.offsetHeight;
      const width = this.offsetWidth;
      qParms = '?w=' + width + '&h=' + height;
    }
    if (row.version) {
      if (qParms) {
        qParms = qParms + '&';
      } else {
        qParms = '?';
      }
      qParms = qParms + '_=' + row.version;
    }
    return "/api/image" + row.path + qParms;
  }

  errorloading() {
    // We failed to load our image, which might mean we got auto-logged out.
    this.dispatchEvent(new CustomEvent('mimchecklogin', {}));
  }
}
