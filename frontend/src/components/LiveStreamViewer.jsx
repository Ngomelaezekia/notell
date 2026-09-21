import { useEffect, useRef, useState } from "react";
import { Loader2, Radio, Volume2, VolumeX } from "lucide-react";
import { liveAPI } from "../services/liveApi";
import { getApiErrorMessage } from "../utils/api";

const LIVEKIT_CDN = "https://cdn.jsdelivr.net/npm/livekit-client@2.22.3/dist/livekit-client.umd.min.js";

let sdkPromise;
function loadLiveKit() {
  if (window.LivekitClient) return Promise.resolve(window.LivekitClient);
  if (sdkPromise) return sdkPromise;
  sdkPromise = new Promise((resolve, reject) => {
    const script = document.createElement("script");
    script.src = LIVEKIT_CDN;
    script.async = true;
    script.onload = () => (window.LivekitClient ? resolve(window.LivekitClient) : reject(new Error("LiveKit SDK failed to initialize")));
    script.onerror = () => reject(new Error("Unable to load the LiveKit client"));
    document.head.appendChild(script);
  });
  return sdkPromise;
}

export default function LiveStreamViewer({ channelId }) {
  const [stream, setStream] = useState(null);
  const [status, setStatus] = useState("loading");
  const [error, setError] = useState("");
  const [muted, setMuted] = useState(false);
  const videoRef = useRef(null);
  const roomRef = useRef(null);
  const tracksRef = useRef(new Set());

  useEffect(() => {
    if (videoRef.current) videoRef.current.muted = muted;
  }, [muted]);

  useEffect(() => {
    let cancelled = false;
    let reconnectTimer;

    const cleanupTracks = () => {
      tracksRef.current.forEach((track) => {
        try { track.detach(videoRef.current); } catch { /* track already detached */ }
      });
      tracksRef.current.clear();
    };

    const disconnect = () => {
      cleanupTracks();
      if (roomRef.current) {
        roomRef.current.disconnect();
        roomRef.current = null;
      }
    };

    const connect = async (current) => {
      try {
        const [sdk, playback] = await Promise.all([loadLiveKit(), liveAPI.playbackToken(current.id)]);
        if (cancelled) return;
        if (!playback?.token || !playback?.livekitUrl) throw new Error("Live playback is not configured");

        const room = new sdk.Room({
          adaptiveStream: true,
          dynacast: true,
          audioCaptureDefaults: { autoGainControl: false },
        });
        roomRef.current = room;

        room.on(sdk.RoomEvent.TrackSubscribed, (track) => {
          if (!videoRef.current || track.kind !== sdk.Track.Kind.Video && track.kind !== sdk.Track.Kind.Audio) return;
          track.attach(videoRef.current);
          tracksRef.current.add(track);
          if (track.kind === sdk.Track.Kind.Audio) videoRef.current.muted = muted;
        });
        room.on(sdk.RoomEvent.TrackUnsubscribed, (track) => {
          try { track.detach(videoRef.current); } catch { /* already detached */ }
          tracksRef.current.delete(track);
        });
        room.on(sdk.RoomEvent.Disconnected, () => {
          if (!cancelled) {
            setStatus("offline");
            reconnectTimer = window.setTimeout(() => void refresh(), 5000);
          }
        });

        await room.connect(playback.livekitUrl, playback.token, { autoSubscribe: true });
        if (!cancelled) setStatus("live");
      } catch (err) {
        if (cancelled) return;
        disconnect();
        if (err?.response?.status === 403 || err?.response?.status === 402) {
          setError(getApiErrorMessage(err, "You need an active channel subscription to watch this private live stream."));
          setStatus("denied");
          return;
        }
        setError(getApiErrorMessage(err, "Unable to connect to the live stream."));
        setStatus("error");
      }
    };

    const refresh = async () => {
      try {
        const result = await liveAPI.current(channelId);
        const current = result?.stream || result?.data?.stream;
        if (cancelled) return;
        disconnect();
        setError("");
        setStream(current || null);
        if (!current) {
          setStatus("offline");
          reconnectTimer = window.setTimeout(() => void refresh(), 5000);
          return;
        }
        setStatus("connecting");
        await connect(current);
      } catch (err) {
        if (cancelled) return;
        if (err?.response?.status === 404) {
          setStream(null);
          setStatus("offline");
          reconnectTimer = window.setTimeout(() => void refresh(), 5000);
          return;
        }
        setError(getApiErrorMessage(err, "Unable to load the live channel."));
        setStatus("error");
      }
    };

    void refresh();
    return () => {
      cancelled = true;
      window.clearTimeout(reconnectTimer);
      disconnect();
    };
  }, [channelId]);

  return (
    <section className="mt-8 overflow-hidden rounded-3xl border border-neutral-800 bg-black shadow-2xl">
      <div className="relative aspect-video bg-neutral-950">
        <video ref={videoRef} autoPlay playsInline controls className="h-full w-full object-contain" muted={muted} />
        {status !== "live" && (
          <div className="absolute inset-0 flex flex-col items-center justify-center gap-3 bg-neutral-950/90 px-6 text-center">
            {status === "loading" || status === "connecting" ? <Loader2 className="animate-spin text-neutral-500" /> : <Radio className="text-neutral-700" />}
            <div>
              <p className="text-sm font-black text-neutral-200">{status === "denied" ? "Plan required" : status === "offline" ? "Channel is offline" : status === "error" ? "Live stream unavailable" : "Connecting to live…"}</p>
              {error && <p className="mt-1 max-w-md text-xs text-neutral-500">{error}</p>}
            </div>
          </div>
        )}
        {status === "live" && (
          <div className="absolute left-3 top-3 inline-flex items-center gap-1.5 rounded-full bg-red-500 px-2.5 py-1 text-[10px] font-black text-white shadow-lg">
            <span className="h-1.5 w-1.5 rounded-full bg-white" /> LIVE
          </div>
        )}
        <button type="button" onClick={() => setMuted((value) => !value)} className="absolute bottom-3 right-3 rounded-full bg-black/60 p-2 text-white backdrop-blur" aria-label={muted ? "Unmute live stream" : "Mute live stream"}>
          {muted ? <VolumeX size={17} /> : <Volume2 size={17} />}
        </button>
      </div>
      {stream && <div className="border-t border-neutral-800 bg-neutral-900/80 px-4 py-3"><h2 className="text-sm font-black">{stream.title}</h2>{stream.description && <p className="mt-1 text-xs text-neutral-500">{stream.description}</p>}</div>}
    </section>
  );
}
