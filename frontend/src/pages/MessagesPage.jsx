import { useEffect, useRef, useState } from "react";
import { MessageCircle, Mic, MicOff, PhoneOff, Plus, Send, Video, VideoOff, Loader2 } from "lucide-react";
import { useAuth } from "../context/AuthContext";
import { messageAPI } from "../services/message/messageApi";
import API, { getApiErrorMessage, getFileUrl } from "../utils/api";

const LIVEKIT_CDN = "https://cdn.jsdelivr.net/npm/livekit-client@2.22.3/dist/livekit-client.umd.min.js";
let sdkPromise;
function loadLiveKit() {
  if (window.LivekitClient) return Promise.resolve(window.LivekitClient);
  if (sdkPromise) return sdkPromise;
  sdkPromise = new Promise((resolve, reject) => {
    const script = document.createElement("script");
    script.src = LIVEKIT_CDN;
    script.async = true;
    script.onload = () => window.LivekitClient ? resolve(window.LivekitClient) : reject(new Error("LiveKit SDK failed to initialize"));
    script.onerror = () => reject(new Error("Unable to load the LiveKit client"));
    document.head.appendChild(script);
  });
  return sdkPromise;
}

function Avatar({ user }) {
  return <div className="h-11 w-11 shrink-0 overflow-hidden rounded-full bg-neutral-800">{user?.profilePicture ? <img src={getFileUrl(user.profilePicture)} alt="" className="h-full w-full object-cover"/> : <div className="flex h-full items-center justify-center font-bold">{user?.username?.[0]?.toUpperCase() || "?"}</div>}</div>;
}
function relationshipLabel(value) {
  return value === "mutual" ? "Mutual" : value === "following" ? "Following" : "Follower";
}

export default function MessagesPage() {
  const { user } = useAuth();
  const [contacts, setContacts] = useState([]);
  const [conversations, setConversations] = useState([]);
  const [selected, setSelected] = useState(null);
  const [messages, setMessages] = useState([]);
  const [hasMore, setHasMore] = useState(false);
  const [loadingMore, setLoadingMore] = useState(false);
  const [body, setBody] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [call, setCall] = useState(null);
  const [callState, setCallState] = useState("idle");
  const [callMuted, setCallMuted] = useState(false);
  const [cameraOff, setCameraOff] = useState(false);
  const [creating, setCreating] = useState(false);
  const wsRef = useRef(null);
  const callRef = useRef(null);
  const reconnectTimerRef = useRef(null);
  const reconnectAttemptRef = useRef(0);
  const bottomRef = useRef(null);
  const callRoomRef = useRef(null);
  const callSdkRef = useRef(null);
  const localTracksRef = useRef([]);
  const remoteTracksRef = useRef(new Set());
  const remoteContainerRef = useRef(null);
  const localVideoRef = useRef(null);

  const load = async () => {
    try {
      setLoading(true);
      const [contactsResponse, c] = await Promise.all([
        API.get("/users/message-contacts", { params: { limit: 50 } }),
        messageAPI.listConversations(),
      ]);
      setContacts(contactsResponse.data?.data || []);
      setConversations(c || []);
    } catch (e) {
      setError(getApiErrorMessage(e, "Unable to load messages."));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void load();
    return () => wsRef.current?.close();
  }, []);

  const cleanupCall = () => {
    remoteTracksRef.current.forEach((track) => {
      try { track.detach(); } catch {}
      try { track.stop?.(); } catch {}
    });
    remoteTracksRef.current.clear();
    localTracksRef.current.forEach((track) => {
      try { track.stop(); } catch {}
      try { track.detach(localVideoRef.current); } catch {}
    });
    localTracksRef.current = [];
    if (callRoomRef.current) {
      callRoomRef.current.disconnect();
      callRoomRef.current = null;
    }
    if (remoteContainerRef.current) remoteContainerRef.current.replaceChildren();
    if (localVideoRef.current) localVideoRef.current.srcObject = null;
    setCallState("idle");
    setCall(null);
  };

  useEffect(() => { callRef.current = call; }, [call]);\n\n  useEffect(() => () => cleanupCall(), []);

  useEffect(() => {
    if (!selected) return;
    let cancelled = false;
    messageAPI.history(selected.id)
      .then(async (r) => { if (cancelled) return; const history = Array.isArray(r) ? r : (r?.messages || []); setMessages(history); setHasMore(Boolean(r?.hasMore)); const lastIncoming = [...history].reverse().find((m) => String(m.senderId) !== String(user?.id)); if (lastIncoming) { try { await messageAPI.markRead(selected.id, lastIncoming.id); setConversations((current) => current.map((item) => item.id === selected.id ? { ...item, unreadCount: 0 } : item)); } catch {} } })
      .catch((e) => { if (!cancelled) setError(getApiErrorMessage(e, "Unable to load conversation.")); });

    const ws = new WebSocket(messageAPI.socketURL(selected.id));
    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.type === "message" && data.message) { setMessages((current) => current.some((m) => m.id === data.message.id) ? current : [...current, data.message]); if (String(data.message.senderId) !== String(user?.id)) setConversations((current) => current.map((item) => item.id === selected.id ? { ...item, updatedAt: data.message.createdAt, unreadCount: (item.unreadCount || 0) + 1 } : item)); }
        if (data.type === "read") { setMessages((current) => current.map((m) => m.id === data.messageId ? { ...m, readAt: data.readAt, readBy: data.userId } : m)); if (String(data.userId) === String(user?.id)) setConversations((current) => current.map((item) => item.id === selected.id ? { ...item, unreadCount: 0 } : item)); }
        if (data.type === "call_invite") setCall(data);
        if (data.type === "call_ended" && data.callId === call?.callId) cleanupCall();
      } catch {}
    };
    ws.onerror = () => setError("Message connection failed.");
    wsRef.current = ws;
    return () => {
      cancelled = true;
      ws.close();
      wsRef.current = null;
    };
  }, [selected, user?.id]);

  useEffect(() => bottomRef.current?.scrollIntoView({ behavior: "smooth" }), [messages]);

  const openContact = async (contact) => {
    setError("");
    try {
      const existing = conversations.find((c) => c.type === "DIRECT" && Array.isArray(c.memberIds) && c.memberIds.some((id) => String(id) === String(contact.id)));
      if (existing) {
        setSelected(existing);
        return;
      }
      const created = await messageAPI.createConversation({ type: "DIRECT", memberIds: [String(contact.id)] });
      setConversations((current) => [created, ...current]);
      setSelected(created);
    } catch (e) {
      setError(getApiErrorMessage(e, "Could not start conversation."));
    }
  };

  const loadOlder = async () => {
    if (!selected || loadingMore || !hasMore || !messages.length) return;
    setLoadingMore(true);
    try {
      const r = await messageAPI.history(selected.id, 50, messages[0].id);
      const older = r?.messages || [];
      setMessages((current) => [...older, ...current.filter((m) => !older.some((x) => x.id === m.id))]);
      setHasMore(Boolean(r?.hasMore));
    } catch (e) { setError(getApiErrorMessage(e, "Unable to load older messages.")); }
    finally { setLoadingMore(false); }
  };

  const send = () => {
    const text = body.trim();
    if (!text || !selected || wsRef.current?.readyState !== WebSocket.OPEN) return;
    wsRef.current.send(JSON.stringify({ type: "message", body: text, clientId: crypto.randomUUID() }));
    setBody("");
  };

  const createGroup = async () => {
    if (!contacts.length || creating) return;
    setCreating(true);
    try {
      const created = await messageAPI.createConversation({ type: "GROUP", title: "Relationship group", memberIds: contacts.map((f) => String(f.id)) });
      setConversations((current) => [created, ...current]);
      setSelected(created);
    } catch (e) {
      setError(getApiErrorMessage(e, "Could not create group."));
    } finally {
      setCreating(false);
    }
  };

  const connectCall = async (callId, initial = {}) => {
    try {
      setCallState("connecting");
      setError("");
      const sdk = await loadLiveKit();
      callSdkRef.current = sdk;
      const access = initial.token && initial.wsUrl
        ? initial
        : await messageAPI.callToken(callId);
      if (!access?.token || !access?.wsUrl) throw new Error("Video call transport is not configured");
      const room = new sdk.Room({ adaptiveStream: true, dynacast: true });
      callRoomRef.current = room;

      room.on(sdk.RoomEvent.TrackSubscribed, (track) => {
        if (!remoteContainerRef.current) return;
        const elements = track.attach();
        elements.forEach((element) => {
          element.className = "h-full w-full object-cover rounded-2xl";
          remoteContainerRef.current.appendChild(element);
        });
        remoteTracksRef.current.add(track);
      });
      room.on(sdk.RoomEvent.TrackUnsubscribed, (track) => {
        try { track.detach(); } catch {}
        remoteTracksRef.current.delete(track);
      });
      room.on(sdk.RoomEvent.Disconnected, () => {
        if (callRoomRef.current === room) callRoomRef.current = null;
        setCallState("idle");
        setCall(null);
      });

      await room.connect(access.wsUrl, access.token, { autoSubscribe: true });
      const localTracks = await sdk.createLocalTracks({ audio: true, video: true });
      for (const track of localTracks) {
        await room.localParticipant.publishTrack(track);
        localTracksRef.current.push(track);
        if (track.kind === sdk.Track.Kind.Video && localVideoRef.current) {
          track.attach(localVideoRef.current);
        }
      }
      setCall((current) => ({ ...(current || {}), callId, roomName: access.roomName || current?.roomName }));
      setCallState("live");
    } catch (e) {
      cleanupCall();
      setError(getApiErrorMessage(e, "Could not join video call."));
    }
  };

  const startCall = async () => {
    if (!selected) return;
    try {
      const result = await messageAPI.createCall(selected.id);
      setCall(result);
      await connectCall(result?.call?.id || result?.callId, result);
    } catch (e) {
      setError(getApiErrorMessage(e, "Could not start video call."));
    }
  };

  const joinIncomingCall = async () => {
    const callId = call?.callId || call?.call?.id || call?.id;
    if (!callId) return;
    await connectCall(callId, call);
  };

  const endCall = async () => {
    const callId = call?.callId || call?.call?.id || call?.id;
    try {
      if (callId) await messageAPI.endCall(callId);
    } catch (e) {
      setError(getApiErrorMessage(e, "Could not end video call."));
    } finally {
      cleanupCall();
    }
  };

  const toggleMute = () => {
    const sdk = callSdkRef.current;
    const room = callRoomRef.current;
    if (!sdk || !room) return;
    room.localParticipant.audioTrackPublications.forEach((publication) => publication.track?.enable(callMuted));
    setCallMuted((value) => !value);
  };

  const toggleCamera = () => {
    const sdk = callSdkRef.current;
    const room = callRoomRef.current;
    if (!sdk || !room) return;
    room.localParticipant.videoTrackPublications.forEach((publication) => publication.track?.enable(cameraOff));
    setCameraOff((value) => !value);
  };

  return <section className="flex min-h-screen w-full bg-neutral-950 pb-24 text-neutral-100 md:pb-6">
    <div className="mx-auto flex w-full max-w-6xl overflow-hidden md:my-5 md:h-[calc(100vh-2.5rem)] md:rounded-3xl md:border md:border-neutral-800">
      <aside className={`w-full border-neutral-800 md:w-80 md:border-r ${selected ? "hidden md:block" : "block"}`}>
        <div className="flex items-center justify-between border-b border-neutral-800 p-4">
          <div><h1 className="text-xl font-black">Messages</h1><p className="text-[11px] text-neutral-600">Followers & following</p></div>
          <button type="button" onClick={createGroup} disabled={creating || !contacts.length} className="rounded-xl border border-neutral-800 p-2 text-neutral-400 hover:text-white disabled:opacity-40" title="Create group"><Plus size={17}/></button>
        </div>
        <div className="border-b border-neutral-800 p-3">
          <p className="mb-2 px-1 text-[10px] font-black uppercase tracking-widest text-neutral-600">People you have a relationship with</p>
          {loading ? <Loader2 className="mx-auto my-8 animate-spin text-neutral-700"/> : contacts.length ? contacts.map((f) => <button type="button" key={f.id} onClick={() => void openContact(f)} className="flex w-full items-center gap-3 rounded-2xl p-2.5 text-left hover:bg-neutral-900"><Avatar user={f}/><div className="min-w-0"><p className="truncate text-sm font-bold">{f.username}</p><p className="text-[10px] text-neutral-600">{relationshipLabel(f.relationship)}</p></div></button>) : <p className="p-3 text-xs text-neutral-600">People you follow or who follow you will appear here.</p>}
        </div>
      </aside>

      <main className={`flex min-w-0 flex-1 flex-col ${selected ? "flex" : "hidden md:flex"}`}>
        {selected ? <><header className="flex items-center justify-between border-b border-neutral-800 p-4">
          <div className="flex items-center gap-3"><button type="button" onClick={() => setSelected(null)} className="text-neutral-500 md:hidden">←</button><div><h2 className="font-black">{selected.title || "Conversation"}</h2><p className="text-[10px] text-neutral-600">{selected.type === "GROUP" ? "Group chat" : "Direct message"}{selected.unreadCount > 0 ? ` · ${selected.unreadCount} unread` : ""}</p></div></div>
          <button type="button" onClick={() => void startCall()} className="rounded-xl border border-neutral-800 p-2 text-neutral-400 hover:text-white" title="Start video call"><Video size={18}/></button>
        </header>
        <div className="flex-1 overflow-y-auto p-4">{hasMore && <button type="button" onClick={() => void loadOlder()} disabled={loadingMore} className="mx-auto mb-3 block rounded-xl border border-neutral-800 px-3 py-2 text-[10px] font-bold text-neutral-500 disabled:opacity-40">{loadingMore ? "Loading…" : "Load older messages"}</button>}<div className="mx-auto max-w-2xl space-y-2">{messages.map((m) => <div key={m.id} className={`flex ${String(m.senderId) === String(user?.id) ? "justify-end" : "justify-start"}`}><div className={`max-w-[78%] rounded-2xl px-3.5 py-2.5 text-sm ${String(m.senderId) === String(user?.id) ? "bg-neutral-100 text-neutral-950" : "bg-neutral-900 text-neutral-200"}`}>{m.body}</div></div>)}<div ref={bottomRef}/></div></div>
        <form onSubmit={(e) => { e.preventDefault(); send(); }} className="border-t border-neutral-800 p-3"><div className="mx-auto flex max-w-2xl items-center gap-2"><input value={body} onChange={(e) => setBody(e.target.value)} placeholder="Message your connection…" className="min-w-0 flex-1 rounded-2xl border border-neutral-800 bg-neutral-900 px-4 py-3 text-sm outline-none placeholder:text-neutral-600 focus:border-neutral-600"/><button type="submit" className="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl bg-neutral-100 text-neutral-950 hover:bg-white"><Send size={17}/></button></div></form>
        </> : <div className="hidden flex-1 flex-col items-center justify-center text-center md:flex"><MessageCircle size={42} className="text-neutral-700"/><h2 className="mt-4 text-lg font-black">Your messages</h2><p className="mt-1 max-w-sm text-xs leading-5 text-neutral-600">Choose someone you follow or who follows you to chat, create a group, or start a video meeting.</p></div>}
      </main>
    </div>

    {error && <div className="fixed bottom-24 left-1/2 z-50 -translate-x-1/2 rounded-xl border border-red-900/50 bg-red-950/90 px-4 py-3 text-xs text-red-300">{error}</div>}

    {call && callState !== "live" && <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/80 p-4">
      <div className="w-full max-w-md rounded-3xl border border-neutral-800 bg-neutral-950 p-6">
        <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl bg-neutral-900"><Video/></div>
        <h3 className="mt-4 text-center text-lg font-black">{callState === "connecting" ? "Connecting…" : "Video call invitation"}</h3>
        <p className="mt-2 text-center text-xs leading-5 text-neutral-500">{callState === "connecting" ? "Joining the secure LiveKit room." : "You have been invited to a private video call."}</p>
        {callState !== "connecting" && <div className="mt-5 grid grid-cols-2 gap-2"><button type="button" onClick={() => cleanupCall()} className="rounded-xl border border-neutral-800 py-3 text-xs font-bold">Decline</button><button type="button" onClick={() => void joinIncomingCall()} className="rounded-xl bg-neutral-100 py-3 text-xs font-black text-neutral-950">Join call</button></div>}
      </div>
    </div>}

    {call && callState === "live" && <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/85 p-4">
      <div className="w-full max-w-4xl overflow-hidden rounded-3xl border border-neutral-800 bg-neutral-950">
        <div className="flex items-center justify-between border-b border-neutral-800 px-4 py-3"><div><p className="text-sm font-black">Video call</p><p className="text-[10px] text-neutral-600">{call.roomName || "LiveKit room"}</p></div><button type="button" onClick={() => void endCall()} className="rounded-xl bg-red-600 px-3 py-2 text-[10px] font-black text-white"><PhoneOff size={14} className="mr-1 inline"/> End</button></div>
        <div className="relative aspect-video bg-black">
          <div ref={remoteContainerRef} className="grid h-full w-full grid-cols-1 gap-2 p-2 sm:grid-cols-2"/>
          <video ref={localVideoRef} autoPlay muted playsInline className="absolute bottom-3 right-3 h-32 w-48 rounded-2xl border border-white/20 bg-neutral-900 object-cover shadow-2xl"/>
          <div className="absolute bottom-3 left-1/2 flex -translate-x-1/2 gap-2"><button type="button" onClick={toggleMute} className="rounded-full bg-black/75 p-3 text-white" aria-label={callMuted ? "Unmute microphone" : "Mute microphone"}>{callMuted?<MicOff size={17}/>:<Mic size={17}/>}</button><button type="button" onClick={toggleCamera} className="rounded-full bg-black/75 p-3 text-white" aria-label={cameraOff ? "Enable camera" : "Disable camera"}>{cameraOff?<VideoOff size={17}/>:<Video size={17}/>}</button><button type="button" onClick={() => void endCall()} className="rounded-full bg-red-600 p-3 text-white" aria-label="End call"><PhoneOff size={17}/></button></div>
        </div>
      </div>
    </div>}
  </section>;
}  useEffect(() => {
    if (!selected) return;
    let cancelled = false;
    let reconnectTimer;

    const mergeMessages = (incoming) => {
      setMessages((current) => {
        const byId = new Map(current.map((m) => [m.id, m]));
        incoming.forEach((m) => byId.set(m.id, m));
        return [...byId.values()].sort((a, b) => {
          const time = new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime();
          return time || String(a.id).localeCompare(String(b.id));
        });
      });
    };

    const syncHistory = async () => {
      const r = await messageAPI.history(selected.id);
      if (cancelled) return;
      const history = Array.isArray(r) ? r : (r?.messages || []);
      mergeMessages(history);
      setHasMore(Boolean(r?.hasMore));
      const lastIncoming = [...history].reverse().find((m) => String(m.senderId) !== String(user?.id));
      if (lastIncoming) {
        try {
          await messageAPI.markRead(selected.id, lastIncoming.id);
          setConversations((current) => current.map((item) => item.id === selected.id ? { ...item, unreadCount: 0 } : item));
        } catch {}
      }
    };

    const connect = () => {
      if (cancelled) return;
      const ws = new WebSocket(messageAPI.socketURL(selected.id));
      wsRef.current = ws;
      ws.onopen = () => { reconnectAttemptRef.current = 0; setError(""); };
      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          if (data.type === "message" && data.message) {
            mergeMessages([data.message]);
            if (String(data.message.senderId) !== String(user?.id)) {
              setConversations((current) => current.map((item) => item.id === selected.id ? { ...item, updatedAt: data.message.createdAt, unreadCount: (item.unreadCount || 0) + 1 } : item));
            }
          }
          if (data.type === "read") {
            setMessages((current) => current.map((m) => m.id === data.messageId ? { ...m, readAt: data.readAt, readBy: data.userId } : m));
            if (String(data.userId) === String(user?.id)) setConversations((current) => current.map((item) => item.id === selected.id ? { ...item, unreadCount: 0 } : item));
          }
          if (data.type === "call_invite") setCall(data);
          if (data.type === "call_ended" && data.callId === callRef.current?.callId) cleanupCall();
        } catch {}
      };
      ws.onerror = () => setError("Message connection failed. Reconnecting…");
      ws.onclose = () => {
        if (cancelled) return;
        if (wsRef.current === ws) wsRef.current = null;
        const attempt = reconnectAttemptRef.current++;
        const delay = Math.min(30000, 1000 * (2 ** Math.min(attempt, 5)));
        reconnectTimer = setTimeout(async () => {
          try { await syncHistory(); } catch {}
          connect();
        }, delay);
        reconnectTimerRef.current = reconnectTimer;
      };
    };

    syncHistory().catch((e) => { if (!cancelled) setError(getApiErrorMessage(e, "Unable to load conversation.")); });
    connect();

    return () => {
      cancelled = true;
      clearTimeout(reconnectTimer);
      if (reconnectTimerRef.current) clearTimeout(reconnectTimerRef.current);
      reconnectTimerRef.current = null;
      wsRef.current?.close();
      wsRef.current = null;
    };
  }, [selected, user?.id]);
