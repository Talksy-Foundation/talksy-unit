export class TalksyVideoObserver {
  /**
   * Constructor.
   * @param {RTCDataChannel} dataChannel - Data channel to use for reporting video size
   * @param {number} intervalGap - interval time gap between report
   * @returns {void}
   */
  constructor(dataChannel, intervalGap) {
    this._lastReportTime = [];
    this._intervalGap = typeof intervalGap !== "number" ? 1000 : intervalGap;
    this._dataChannel = dataChannel;
    this._videoElements = [];
    this._resizeObserver = new ResizeObserver(this._onResize.bind(this));
    this._intersectionObserver = new IntersectionObserver(
        this._onIntersection.bind(this)
    );
  }

  get lastReportTime() {
    return this._lastReportTime;
  }

  set lastReportTime(value) {
    this._lastReportTime = value;
  }

  get intervalGap() {
    return this._intervalGap;
  }

  set intervalGap(value) {
    this._intervalGap = value;
  }

  get dataChannel() {
    return this._dataChannel;
  }

  set dataChannel(value) {
    this._dataChannel = value;
  }

  get videoElements() {
    return this._videoElements;
  }

  set videoElements(value) {
    this._videoElements = value;
  }

  get resizeObserver() {
    return this._resizeObserver;
  }

  set resizeObserver(value) {
    this._resizeObserver = value;
  }

  get intersectionObserver() {
    return this._intersectionObserver;
  }

  set intersectionObserver(value) {
    this._intersectionObserver = value;
  }

  /**
   * Callback when video element is resized.
   * @param {ResizeObserverEntry[]} entries - Resize observer entries
   * @returns {void}
   */
  _onResize(entries) {
    entries.forEach((entry) => {
      if (entry.contentBoxSize) {
        const videoTracks = entry.target.srcObject.getVideoTracks();
        if (videoTracks.length > 0) {
          const trackid = videoTracks[0].id;
          const contentBoxSize = entry.contentBoxSize[0];
          const width = contentBoxSize.inlineSize;
          const height = contentBoxSize.blockSize;
          this._onVideoSizeChanged(trackid, width, height);
        }
      }
    });
  }

  /**
   * Callback when video element is intersected.
   * @param {IntersectionObserverEntry[]} entries - Intersection observer entries
   * @returns {void}
   */
  _onIntersection(entries) {
    entries.forEach((entry) => {
      const videoTracks = entry.target.srcObject.getVideoTracks();
      if (videoTracks.length > 0) {
        const trackid = videoTracks[0].id;
        const width = entry.isIntersecting ? entry.target.width : 0;
        const height = entry.isIntersecting ? entry.target.height : 0;
        this._onVideoSizeChanged(trackid, width, height);
      }
    });
  }

  /**
   * Observe video element for any visibility or resize changes.
   * @param {HTMLVideoElement} videoElement - Video element to watch
   * @returns {void}
   */
  observe(videoElement) {
    this._watchVideoElement(videoElement);
    this._videoElements.push(videoElement);
  }

  /**
   * Unobserve video element for any visibility or resize changes.
   * @param {HTMLVideoElement} videoElement - Video element to unwatch
   * @returns {void}
   */
  unobserve(videoElement) {
    this._intersectionObserver.unobserve(videoElement);
    this._resizeObserver.unobserve(videoElement);
  }

  /**
   * Watch video element events.
   * @param {HTMLVideoElement} videoElement - Video element to watch
   * @returns {void}
   */
  _watchVideoElement(videoElement) {
    this._intersectionObserver.observe(videoElement);
    this._resizeObserver.observe(videoElement);
  }

  /**
   * Report video size to peer connection.
   * @param {string} id - MediaStreamTrack id
   * @param {number} width - Video width
   * @param {number} height - Video height
   * @returns {void}
   */
  _onVideoSizeChanged(id, width, height) {
    if (
      id in this._lastReportTime &&
      Date.now() - this._lastReportTime[id] < this._intervalGap
    ) {
      return;
    }

    this._lastReportTime[id] = Date.now();

    if (this._dataChannel.readyState == "open") {
      const data = JSON.stringify({
        type: "video_size",
        data: {
          track_id: id,
          width: Math.floor(width),
          height: Math.floor(height),
        },
      });

      console.log("Sending video size data: ", data);
      this._dataChannel.send(data);
    } else {
      const listener = () => {
        this._dataChannel.send(
          JSON.stringify({
            type: "video_size",
            data: {
              track_id: id,
              width: Math.floor(width),
              height: Math.floor(height),
            },
          })
        );

        this._dataChannel.removeEventListener("open", listener);
      };

      this._dataChannel.addEventListener("open", listener);
    }
  }
}
