import { TalksyVideoObserver } from "./talksy-video-observer.js";

const peerConnection = new RTCPeerConnection({
  iceServers: [
    {
      urls: "stun:stun.l.google.com:19302",
    },
  ],
});

let clientid = null;
let ws = null;
let internalChannel = null;
let videoObserver = null;
let red = true;
let negotiationNeeded = false;

peerConnection.ondatachannel = function (e) {
  console.log("ondatachannel: ", e.channel.label);
  if (e.channel.label == "internal") {
    internalChannel = e.channel;
    videoObserver = new TalksyVideoObserver(e.channel);

    internalChannel.addEventListener("message", (e) => {
      const msg = JSON.parse(e.data);
      if (msg.type == "vad_started" || msg.type == "vad_ended") {
        updateVoiceDetected(msg);
      }
    });
  }
};

const startWs = async () => {
  let debug = false;
  let dev = false;
  const urlParams = new URLSearchParams(window.location.search);
  if (urlParams.has("debug")) {
    debug = true;
  }
  if (urlParams.has("dev")) {
    dev = true;
  }

  if (urlParams.has("disablered")) {
    red = false;
  }

  var scheme = document.location.protocol == "https:" ? "wss" : "ws";

  ws = new WebSocket(
    `${scheme}://${window.location.host}/${scheme}?${
      debug ? "debug=1" : ""
    }&room_id=${urlParams.get("room_id")}`
  );

  const promise = new Promise((resolve, reject) => {
    ws.onopen = function () {
      resolve();
    };
    // ws.addEventListener("open", (event) => {
    //   console.log(event);
    //   resolve();
    // });

    ws.onerror = function (err) {
      console.log(err);
      reject(err);
    };
  });

  ws.onmessage = async function (e) {
    const msg = JSON.parse(e.data);
    try {
      if (msg.type == "clientid") {
        clientid = msg.data;
        document.getElementById("clientid").innerText = "ClientID: " + clientid;
      } else if (msg.type === "network_condition") {
        document.getElementById("network").innerText =
          msg.data == 0 ? "Unstable" : "Stable";
      } else if (msg.type == "offer") {
        console.log("client received offer: ", msg.data);
        await peerConnection.setRemoteDescription(msg.data);
        var answer = await peerConnection.createAnswer();
        await peerConnection.setLocalDescription(answer);
        ws.send(JSON.stringify({ type: "answer", data: answer.sdp }));
        negotiationNeeded = false;
        console.log("client send answer", answer);
      } else if (msg.type == "answer") {
        await peerConnection.setRemoteDescription(msg.data);
        console.log(
          "client received answer",
          peerConnection.currentRemoteDescription
        );
        negotiationNeeded = false;
      } else if (msg.type == "candidate") {
        await peerConnection.addIceCandidate(msg.data);
      } else if (msg.type == "tracks_added") {
        console.log("on tracks added", msg.data);
        const trackType = {};
        const tracksAdded = msg.data;
        Object.keys(tracksAdded).forEach((uid) => {
          // we suppose to check tracksAdded[key].id and compare it with current track id from  navigator.mediaDevices.getUserMedia() to know the source is media
          trackType[uid] = "media";
        });

        ws.send(JSON.stringify({ type: "tracks_added", data: trackType }));
      } else if (msg.type == "tracks_available") {
        console.log("on tracks available", msg.data);
        const subTracks = [];
        const availableTracks = msg.data;
        Object.keys(availableTracks).forEach((uid) => {
          // we suppose to check tracksAdded[key].id and compare it with current track id from  navigator.mediaDevices.getUserMedia() to know the source is media
          // ClientID string `json:"client_id"`
          // StreamID string `json:"stream_id"`
          // TrackID  string `json:"track_id"`
          // RID      string `json:"rid"`
          const track = availableTracks[uid];
          subTracks.push({
            client_id: track.client_id,
            track_id: track.id,
          });
        });

        console.log("subTracks: ", subTracks);

        ws.send(JSON.stringify({ type: "subscribe_tracks", data: subTracks }));
      } else if (msg.type == "allow_renegotiation") {
        const isAllowRenegotiation = msg.data;
        if (isAllowRenegotiation && negotiationNeeded) {
          negotiate();
        }
      } else if (msg.type == "track_stats") {
        updateTrackStats(msg.data);
      }
    } catch (error) {
      console.log(error);
    }
  };
  ws.onclose = function () {
    console.log("websocket close");
  };

  return promise;
};

function updateVoiceDetected(vad) {
  const streamid = vad.data.streamID;

  const videoEl = document.getElementById("video-" + streamid);
  if (!videoEl) {
    return;
  }

  if (vad.type === "vad_ended") {
    // voice ended
    videoEl.style.border = "3px solid gray";
  } else {
    // voice detected
    videoEl.style.border = "3px solid green";
  }
}

function startH264() {
  start("h264");
}

function startVP9() {
  start("vp9");
}

// if (!navigator.mediaDevices?.enumerateDevices) {
//     console.log("enumerateDevices() not supported.");
//   } else {
//     // List cameras and microphones.
//     navigator.mediaDevices
//       .enumerateDevices()
//       .then((devices) => {
//         devices.forEach((device) => {
//           console.log(`${device.kind}: ${device.label} id = ${device.deviceId}`);
//         });
//       })
//       .catch((err) => {
//         console.error(`${err.name}: ${err.message}`);
//       });
//   }

async function start(codec) {
  await startWs();
  document.getElementById("btnStart").disabled = true;
  document.getElementById("btnStartVP9").disabled = true;
  // const constraintList = document.querySelector("#constraintList");
  // const supportedConstraints = navigator.mediaDevices.getSupportedConstraints();

  // for (const constraint of Object.keys(supportedConstraints)) {
  //   const elem = document.createElement("li");
  //   elem.innerHTML = `<code>${constraint}</code>`;
  //   constraintList.appendChild(elem);
  // }

  const video = {
    width: { min: 1024, ideal: 1280, max: 1920 },
    height: { min: 776, ideal: 720, max: 1080 },
    frameRate: { ideal: 24, max: 60 },
    facingMode: "user",
    // width: { ideal: 1280 },
    // height: { ideal: 720 },
    // advanced: [
    //   { frameRate: { min: 30 } },
    //   { height: { min: 360 } },
    //   { width: { min: 720 } },
    //   { frameRate: { max: 30 } },
    //   { width: { max: 1280 } },
    //   { height: { max: 720 } },
    //   { aspectRatio: { exact: 1.77778 } },
    // ],
  };

  const constraints = {
    audio: true,
    video: video,
    // video: true,
  };

  if (navigator.mediaDevices === undefined) {
    navigator.mediaDevices = {};
  }

  if (navigator.mediaDevices.getUserMedia === undefined) {
    navigator.mediaDevices.getUserMedia = function (constraints) {
      var getUserMedia =
        navigator.webkitGetUserMedia || navigator.mozGetUserMedia;
      if (!getUserMedia) {
        alert(
            "244: getUserMedia() is not implemented in this browser. Chrome disables features such as getUserMedia when it comes from an unsecured origin. http://localhost is considered as a secure origin by default, however if you use an origin that does not have an SSL/TLS certificate then Chrome will consider the origin as unsecured and disable getUserMedia."
          );
        return Promise.reject(
          new Error("getUserMedia is not implemented in this browser")
        );
      }
      return new Promise(function (resolve, reject) {
        getUserMedia.call(navigator, constraints, resolve, reject);
      });
    };
  }

//   navigator.mediaDevices
//     .getUserMedia({ audio: true, video: true })
//     .then(function (stream) {
//       var video = document.querySelector("video");

//       if ("srcObject" in video) {
//         video.srcObject = stream;
//       } else {
//         // Не используем в новых браузерах
//         video.src = window.URL.createObjectURL(stream);
//       }
//       video.onloadedmetadata = function (e) {
//         video.play();
//       };
//     })
//     .catch(function (err) {
//       console.log(err.name + ": " + err.message);
//     });

  let stream;
  stream = await navigator.mediaDevices.getUserMedia(constraints);

  const streamid = stream.id.replace("{", "").replace("}", "");
  let container = document.getElementById("container-" + streamid);
  if (!container) {
    container = document.createElement("div");
    container.className = "container";
    container.id = "container-" + streamid;
    document.querySelector("main").appendChild(container);
  }

  let localVideo = document.getElementById("video-" + streamid);
  if (!localVideo) {
    localVideo = document.createElement("video");
    localVideo.id = "video-" + streamid;
    localVideo.autoplay = true;
    localVideo.muted = true;
    localVideo.controls = true;
    container.appendChild(localVideo);
  }

  if ("srcObject" in localVideo) {
    localVideo.srcObject = stream;
  } else {
    // dont use in modern browsers
    localVideo.src = window.URL.createObjectURL(stream);
  }

  peerConnection.ontrack = function (e) {
    e.receiver.playoutDelayHint = 0.1;

    e.streams.forEach((stream) => {
      console.log("ontrack", stream, e.track);
      const streamid = stream.id.replace("{", "").replace("}", "");
      let container = document.getElementById("container-" + streamid);
      if (!container) {
        container = document.createElement("div");
        container.className = "container";
        container.id = "container-" + streamid;
        document.querySelector("main").appendChild(container);
      }

      let remoteVideo = document.getElementById("video-" + streamid);
      if (!remoteVideo) {
        remoteVideo = document.createElement("video");
        remoteVideo.id = "video-" + streamid;
        remoteVideo.autoplay = true;
        remoteVideo.controls = true;
        container.appendChild(remoteVideo);
        if (videoObserver != null) {
          videoObserver.observe(remoteVideo);
        }
      }

      if (e.track.kind == "video") {
        const trackid = e.track.id.replace("{", "").replace("}", "");
        let stats = document.getElementById("stats-" + trackid);
        if (!stats) {
          const videoStats = document.createElement("div");
          videoStats.className = "video-stats";

          stats = document.createElement("div");
          stats.appendChild(videoStats);
          stats.className = "stats";
          stats.id = "stats-" + trackid;
          container.append(stats);
        }
      }

      if ("srcObject" in remoteVideo) {
        remoteVideo.srcObject = stream;
      } else {
        remoteVideo.src = window.URL.createObjectURL(stream);
      }

      stream.onremovetrack = function (e) {
        console.log("onremovetrack", stream, e.track);
        if ("srcObject" in remoteVideo) {
          remoteVideo.srcObject = null;
        } else {
          remoteVideo.src = null;
        }
        remoteVideo.remove();
        container.remove();
        if (videoObserver != null) {
          videoObserver.unobserve(remoteVideo);
        }
      };
    });
  };

  // send local video
  // peerConnection.addTrack(stream.getAudioTracks()[0], stream);
  // peerConnection.addTrack(stream.getVideoTracks()[0], stream);

  const audioTcvr = peerConnection.addTransceiver(stream.getAudioTracks()[0], {
    direction: "sendonly",
    streams: [stream],
    sendEncodings: [{ priority: "high" }],
  });

  if (
    audioTcvr.setCodecPreferences != undefined &&
    RTCRtpReceiver.getCapabilities != undefined
  ) {
    const audioCodecs = RTCRtpReceiver.getCapabilities("audio").codecs;

    let audioCodecsPref = [];
    if (red) {
      for (let i = 0; i < audioCodecs.length; i++) {
        // audio/red 48000 111/111
        if (audioCodecs[i].mimeType == "audio/red") {
          audioCodecsPref.push(audioCodecs[i]);
        }
      }
    }

    for (let i = 0; i < audioCodecs.length; i++) {
      if (audioCodecs[i].mimeType == "audio/opus") {
        audioCodecsPref.push(audioCodecs[i]);
      }
    }

    audioTcvr.setCodecPreferences(audioCodecsPref);
  }

  const isFirefox = navigator.userAgent.includes("Firefox");
  const isSimulcast = document.querySelector("#simulcast").checked;
  const isSvc = document.querySelector("#svc").checked;
  const maxBitrate = document.querySelector("#maxBitrate").value;

  //   console.log("isFirefox", isFirefox);
  if (codec === "vp9" && !isFirefox) {
    let videoTcvr = null;

    console.log("simulcast: ", isSimulcast);

    if (!isSimulcast) {
      videoTcvr = peerConnection.addTransceiver(stream.getVideoTracks()[0], {
        direction: "sendonly",
        streams: [stream],
        sendEncodings: [
          {
            maxBitrate: maxBitrate,
            scalabilityMode: isSvc ? "L3T3" : "L1T1",
          },
        ],
      });
    } else {
      videoTcvr = peerConnection.addTransceiver(stream.getVideoTracks()[0], {
        direction: "sendonly",
        streams: [stream],
        sendEncodings: [
          {
            rid: "high",
            maxBitrate: maxBitrate,
            maxFramerate: 30,
            scalabilityMode: isSvc ? "L3T3" : "L1T1",
          },
          {
            rid: "mid",
            scaleResolutionDownBy: 2.0,
            maxFramerate: 30,
            maxBitrate: maxBitrate / 2,
            scalabilityMode: isSvc ? "L3T3" : "L1T1",
          },
          {
            rid: "low",
            scaleResolutionDownBy: 4.0,
            maxBitrate: maxBitrate / 4,
            maxFramerate: 30,
            scalabilityMode: isSvc ? "L3T3" : "L1T1",
          },
        ],
      });
    }

    const codecs = RTCRtpReceiver.getCapabilities("video").codecs;
    let vp9_codecs = [];
    // iterate over supported codecs and pull out the codecs we want
    for (let i = 0; i < codecs.length; i++) {
      if (codecs[i].mimeType == "video/VP9") {
        vp9_codecs.push(codecs[i]);
      }
    }

    // push the rest of the codecs
    for (let i = 0; i < codecs.length; i++) {
      if (codecs[i].mimeType != "video/VP9") {
        vp9_codecs.push(codecs[i]);
      }
    }

    // currently not all browsers support setCodecPreferences
    if (videoTcvr.setCodecPreferences != undefined) {
      videoTcvr.setCodecPreferences(vp9_codecs);
    }
  } else {
    let videoTcvr = null;
    if (!isSimulcast) {
      videoTcvr = peerConnection.addTransceiver(stream.getVideoTracks()[0], {
        direction: "sendonly",
        streams: [stream],
        sendEncodings: [
          {
            maxBitrate: 1200 * 1000,
          },
        ],
      });
    } else {
      videoTcvr = peerConnection.addTransceiver(stream.getVideoTracks()[0], {
        direction: "sendonly",
        streams: [stream],
        sendEncodings: [
          {
            rid: "high",
            maxBitrate: 1200 * 1000,
            maxFramerate: 30,
          },
          {
            rid: "mid",
            scaleResolutionDownBy: 2.0,
            maxFramerate: 30,
            maxBitrate: 500 * 1000,
          },
          {
            rid: "low",
            scaleResolutionDownBy: 4.0,
            maxBitrate: 150 * 1000,
            maxFramerate: 30,
          },
        ],
      });
    }

    const codecs = RTCRtpReceiver.getCapabilities("video").codecs;
    let h264Codecs = [];
    // iterate over supported codecs and pull out the codecs we want
    for (let i = 0; i < codecs.length; i++) {
      if (codecs[i].mimeType == "video/H264") {
        h264Codecs.push(codecs[i]);
      }
    }

    // push the rest of the codecs
    for (let i = 0; i < codecs.length; i++) {
      if (codecs[i].mimeType != "video/H264") {
        h264Codecs.push(codecs[i]);
      }
    }

    // currently not all browsers support setCodecPreferences
    if (videoTcvr.setCodecPreferences != undefined) {
      videoTcvr.setCodecPreferences(h264Codecs);
    } else {
      console.log("setCodecPreferences not supported");
    }
  }

  const offer = await peerConnection.createOffer();

  await peerConnection.setLocalDescription(offer);

  ws.send(JSON.stringify({ type: "offer", data: offer.sdp }));
  console.log("browser send");

  peerConnection.onicecandidate = function (e) {
    if (e.candidate != null) {
      ws.send(
        JSON.stringify({ type: "candidate", data: e.candidate.candidate })
      );
    }
  };

  peerConnection.onconnectionstatechange = function (e) {
    console.log("onconnectionstatechange", peerConnection.connectionState);
    if (peerConnection.connectionState == "connected") {
      monitorStats();
      monitorBw();
    }
  };

  isAllowRenegotiation();
}

const sleep = (delay) => new Promise((resolve) => setTimeout(resolve, delay));

let prevHighBytesSent = 0;
let prevMidBytesSent = 0;
let prevLowBytesSent = 0;

const bwController = {
  low: 0,
  mid: 0,
  high: 0,
  available: 0,
};

const monitorBw = async () => {
  while (peerConnection.connectionState == "connected") {
    const totalBw = bwController.low + bwController.mid + bwController.high;

    if (
      bwController.available == 0 ||
      bwController.low == 0 ||
      bwController.mid == 0 ||
      bwController.high == 0
    ) {
      await sleep(5000);
      continue;
    }

    // const ratio = bwController.available / totalBw;

    await sleep(5000);
  }
};

const updateTrackStats = (trackStats) => {
  const sentStats = trackStats.sent_track_stats;
  sentStats.forEach((stat) => {
    const statsEl = document.getElementById("stats-" + stat.id);
    if (!statsEl) {
      return;
    }

    let trackStatsEl = statsEl.querySelector(".track-stats");
    if (!trackStatsEl) {
      trackStatsEl = document.createElement("div");
      trackStatsEl.className = "track-stats";
      statsEl.appendChild(trackStatsEl);
    }

    const statsText = `
                    <p>Packet Loss Ratio: ${
                      Math.round(stat.fraction_lost * 100) / 100
                    }</p>

                `;
    trackStatsEl.innerHTML = statsText;
  });
  //   const receivedStats = trackStats.received_track_stats;
};

const monitorStats = async () => {
  while (peerConnection.connectionState == "connected") {
    const stats = await peerConnection.getStats();

    const qualityLimitiationReason = {
      cpu: false,
      bandwidth: false,
    };

    stats.forEach((report) => {
      if (report.type === "inbound-rtp" && report.kind === "video") {
        const trackid = report.trackIdentifier
          .replace("{", "")
          .replace("}", "");
        const statsEl = document.querySelector(
          `#stats-${trackid} .video-stats`
        );
        if (!statsEl) {
          return;
        }

        if (typeof bwController[report.trackIdentifier] == "undefined") {
          bwController[report.trackIdentifier] = {
            prevBytesReceived: 0,
          };
        }

        if (
          bwController[report.trackIdentifier].prevBytesReceived == 0 ||
          report.bytesReceived == 0
        ) {
          bwController[report.trackIdentifier].prevBytesReceived =
            report.bytesReceived;
          return;
        }

        const deltaBytes =
          report.bytesReceived -
          bwController[report.trackIdentifier].prevBytesReceived;
        bwController[report.trackIdentifier].prevBytesReceived =
          report.bytesReceived;

        const statsText = `
                            <p>FrameRate: ${report.framesPerSecond}</p>
                            <p>Bitrate: ${(deltaBytes * 8) / 1000} kbps</p>
                            <p>Resolution: ${report.frameWidth}x${
          report.frameHeight
        }</p>
                            <p>Packet Lost: ${report.packetsLost}</p>
                            <p>Nack Count: ${report.nackCount}</p>
                        `;
        statsEl.innerHTML = statsText;
      }

      const statsEl = document.getElementById("stats-local");

      if (
        report.type === "candidate-pair" &&
        typeof report.availableOutgoingBitrate !== "undefined"
      ) {
        let bwStatsEl = document.getElementById("stats-local-bandwidth");
        if (!bwStatsEl) {
          bwStatsEl = document.createElement("div");
          bwStatsEl.id = "stats-local-bandwidth";
          statsEl.append(bwStatsEl);
        }

        bwStatsEl.innerHTML = `
                            <p>available bandwidth: ${
                              report.availableOutgoingBitrate / 1000
                            } kbps</p>
                            <p>current bitrate: ${
                              (bwController.low +
                                bwController.mid +
                                bwController.high) /
                              1000
                            } kbps</p>
                            `;

        bwController.available = report.availableOutgoingBitrate;
      }

      if (report.type === "outbound-rtp" && report.kind === "video") {
        if (report.rid === "high" || typeof report.rid === "undefined") {
          if (prevHighBytesSent === 0 || report.bytesSent == 0) {
            prevHighBytesSent = report.bytesSent;
            return;
          }

          let highStatsEl = document.getElementById("stats-local-high");
          if (!highStatsEl) {
            highStatsEl = document.createElement("div");
            highStatsEl.id = "stats-local-high";
            statsEl.append(highStatsEl);
          }

          const deltaBytes = Math.abs(report.bytesSent - prevHighBytesSent);
          prevHighBytesSent = report.bytesSent;
          const bitrate = deltaBytes * 8;
          bwController.high = bitrate;
          const qualityLimitation = `<p>Quality Limitation Reason: ${report.qualityLimitationReason}</p>`;
          highStatsEl.innerHTML = `
                                <h3>High</h3>
                                <p>FrameRate: ${report.framesPerSecond}</p>
                                <p>Bitrate: ${bitrate / 1000} kbps</p>
                                <p>Resolution: ${report.frameWidth}x${
            report.frameHeight
          }</p>
                                ${
                                  report.qualityLimitationReason
                                    ? qualityLimitation
                                    : ""
                                }
                            `;

          if (report.qualityLimitationReason == "cpu") {
            qualityLimitiationReason.cpu = true;
          }

          if (report.qualityLimitationReason == "bandwidth") {
            qualityLimitiationReason.bandwidth = true;
          }
        }

        if (report.rid === "mid") {
          if (prevMidBytesSent === 0 || report.bytesSent == 0) {
            prevMidBytesSent = report.bytesSent;
            return;
          }

          let midStatsEl = document.getElementById("stats-local-mid");
          if (!midStatsEl) {
            midStatsEl = document.createElement("div");
            midStatsEl.id = "stats-local-mid";
            statsEl.append(midStatsEl);
          }

          const deltaBytes = Math.abs(report.bytesSent - prevMidBytesSent);
          prevMidBytesSent = report.bytesSent;
          const bitrate = deltaBytes * 8;
          bwController.mid = bitrate;

          midStatsEl.innerHTML = `
                                <h3>Mid</h3>
                                <p>FrameRate: ${report.framesPerSecond}</p>
                                <p>Bitrate: ${bitrate / 1000} kbps</p>
                                <p>Resolution: ${report.frameWidth}x${
            report.frameHeight
          }</p>
                                <p>Quality Limitation Reason: ${
                                  report.qualityLimitationReason
                                }</p>
                            `;

          if (report.qualityLimitationReason == "cpu") {
            qualityLimitiationReason.cpu = true;
          }

          if (report.qualityLimitationReason == "bandwidth") {
            qualityLimitiationReason.bandwidth = true;
          }
        }

        if (report.rid === "low") {
          if (prevLowBytesSent === 0 || report.bytesSent == 0) {
            prevLowBytesSent = report.bytesSent;
            return;
          }

          let lowStatsEl = document.getElementById("stats-local-low");
          if (!lowStatsEl) {
            lowStatsEl = document.createElement("div");
            lowStatsEl.id = "stats-local-low";
            statsEl.append(lowStatsEl);
          }

          const deltaBytes = Math.abs(report.bytesSent - prevLowBytesSent);
          prevLowBytesSent = report.bytesSent;
          const bitrate = deltaBytes * 8;
          bwController.low = bitrate;

          lowStatsEl.innerHTML = `
                                <h3>Low</h3>
                                <p>FrameRate: ${report.framesPerSecond}</p>
                                <p>Bitrate: ${bitrate / 1000} kbps</p>
                                <p>Resolution: ${report.frameWidth}x${
            report.frameHeight
          }</p>
                                <p>Quality Limitation Reason: ${
                                  report.qualityLimitationReason
                                }</p>
                            `;

          if (report.qualityLimitationReason == "cpu") {
            qualityLimitiationReason.cpu = true;
          }

          if (report.qualityLimitationReason == "bandwidth") {
            qualityLimitiationReason.bandwidth = true;
          }
        }
      }
    });

    let qualityLimitiation = "none";

    if (qualityLimitiationReason.cpu && qualityLimitiationReason.bandwidth) {
      qualityLimitiation = "both";
    } else if (qualityLimitiationReason.cpu) {
      qualityLimitiation = "cpu";
    } else if (qualityLimitiationReason.bandwidth) {
      qualityLimitiation = "bandwidth";
    }

    if (internalChannel != null && internalChannel.readyState == "open") {
      const stats = {
        available_outgoing_bitrate: bwController.available,
        quality_limitation_reason: qualityLimitiation,
      };

      internalChannel.send(
        JSON.stringify({
          type: "stats",
          data: stats,
        })
      );
    }

    await sleep(1000);
  }
};

const toggleStats = () => {
  const statsEls = document.querySelectorAll(".stats");
  statsEls.forEach((el) => {
    if (el.style.display === "none") {
      el.style.display = "flex";
    } else {
      el.style.display = "none";
    }
  });
};

const switchQuality = () => {
  const quality = document.getElementById("selectQuality").value;
  ws.send(JSON.stringify({ type: "switch_quality", data: quality }));
};

const setBandwidthLimit = () => {
  const limit = document.getElementById("selectBandwidth").value;
  ws.send(JSON.stringify({ type: "set_bandwidth_limit", data: limit }));
};

const shareScreen = async () => {
  const stream = await navigator.mediaDevices.getDisplayMedia({
    video: true,
    audio: true,
  });

  const videoTrack = stream.getVideoTracks()[0];
  const audioTrack = stream.getAudioTracks()[0];

  let tscvAudio = null;
  let tscvVideo = null;

  if (typeof audioTrack != "undefined") {
    tscvAudio = peerConnection.addTransceiver(audioTrack, {
      direction: "sendonly",
      streams: [stream],
      sendEncodings: [{ priority: "high" }],
    });
  }

  if (!document.querySelector("#simulcast").checked) {
    tscvVideo = peerConnection.addTransceiver(videoTrack, {
      direction: "sendonly",
      streams: [stream],
      sendEncodings: [
        {
          maxBitrate: 1200 * 1000,
          maxFramerate: 30,
        },
      ],
    });
  } else {
    tscvVideo = peerConnection.addTransceiver(videoTrack, {
      direction: "sendonly",
      streams: [stream],
      sendEncodings: [
        {
          rid: "high",
          maxBitrate: 1200 * 1000,
          maxFramerate: 30,
          scalabilityMode: isSvc ? "L3T3" : "L1T1",
        },
        {
          rid: "mid",
          maxFramerate: 20,
          maxBitrate: 400 * 1000,
          scalabilityMode: isSvc ? "L3T3" : "L1T1",
        },
        {
          rid: "low",
          maxBitrate: 150 * 1000,
          maxFramerate: 15,
          scalabilityMode: isSvc ? "L3T3" : "L1T1",
        },
      ],
    });
  }

  const container = document.createElement("div");
  container.className = "container";
  container.id = "container-" + videoTrack.id;

  const video = document.createElement("video");
  video.id = "video-" + videoTrack.id;
  video.autoplay = true;
  if ("srcObject" in video) {
    video.srcObject = stream;
  } else {
    video.src = window.URL.createObjectURL(stream);
  }
  video.controls = true;
  container.appendChild(video);

  document.querySelector("main").appendChild(container);

  videoTrack.addEventListener("ended", (e) => {
    console.log("Video track ended, stopping screen sharing");

    document.querySelector("main").removeChild(container);

    peerConnection.removeTrack(tscvVideo.sender);

    peerConnection.removeTrack(tscvAudio.sender);

    isAllowRenegotiation();
  });

  isAllowRenegotiation();
};

const isAllowRenegotiation = () => {
  ws.send(JSON.stringify({ type: "is_allow_renegotiation" }));
  negotiationNeeded = true;
};

const negotiate = async () => {
  console.log("negotiate");
  const offer = await peerConnection.createOffer();

  await peerConnection.setLocalDescription(offer);
  ws.send(JSON.stringify({ type: "offer", data: offer.sdp }));
};

document.addEventListener("DOMContentLoaded", function (event) {
  document.getElementById("btnStart").onclick = startH264;
  document.getElementById("btnStartVP9").onclick = startVP9;
  document.getElementById("btnShareScreen").onclick = shareScreen;
  document.getElementById("btnStats").onclick = toggleStats;
  document.getElementById("selectQuality").onchange = switchQuality;
});
