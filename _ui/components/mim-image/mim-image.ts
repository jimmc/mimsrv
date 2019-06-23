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

  // imgitem is our input data that includes the caption
  @Polymer.decorators.property({type: Object})
  imgitem: NavItem;

  // preimgsrc is the API path for preloading the next image
  @Polymer.decorators.property({type: String})
  preimgsrc: string;

  // preimginfo is our input data for preloading the next image
  @Polymer.decorators.property({type: Object})
  preimginfo: any;

  // preimgitem is the data for the preload image that includes its caption
  @Polymer.decorators.property({type: Object})
  preimgitem: NavItem;

  lastResize = 0;       // Time of last resize
  maxResizeDelay = 500;    // We do at least one resize after this much time
  resizeTimeoutId = 0;  // Timer ID for dealing with resizes
  resizeTimeoutDelay = 100;  // Wait this long after event before resizing

  imageHeightNoCaption = 0;
  imageHeightWithCaption = 0;

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
    this.imgsrc = '';    // Clear it first so size calculations work correctly.
    this.imgsrc = this.imginfoToImgsrc(this.imginfo, this.imgitem, true);
  }

  // When preimginfo changes, we preload that image.
  @Polymer.decorators.observe('preimginfo')
  preimginfoChanged() {
    const img = new Image();
    img.src = this.imginfoToImgsrc(this.preimginfo, this.preimgitem, false);
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
  imginfoToImgsrc(imginfo: any, imgitem: NavItem, isDisplayed: boolean) {
    if (!imginfo) {
      return '';
    }
    let caption = ''
    if (imgitem && imgitem.text) {
      caption = imgitem.text;
    }
    const row = imginfo;
    let qParms = '';
    if (!row.zoom) {
      let height = this.offsetHeight;
      if (isDisplayed) {
        // If we are displaying this image, we have the real height,
        // so we save that so we can apply it to the preload image.
        if (!!caption) {
          // As a first approximation, we assume all captions take up the
          // same amount of vertial space, so we only distinguish between
          // caption or no-caption. This means if we get an image that has
          // a long caption that takes up multiple lines on the screen,
          // our preloading won't work as we move past that image.
          this.imageHeightWithCaption = height;
        } else {
          this.imageHeightNoCaption = height;
        }
      } else {
        // We pick our height based on whether we have a caption.
        if (!!caption) {
          height = this.imageHeightWithCaption;
        } else {
          height = this.imageHeightNoCaption;
        }
        if (height == 0) {
          height = this.offsetHeight;
        }
      }
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
