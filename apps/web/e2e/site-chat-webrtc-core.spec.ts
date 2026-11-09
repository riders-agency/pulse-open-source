import { expect, test, type Page } from "@playwright/test";
import { readFile } from "node:fs/promises";
import path from "node:path";

test.use({
  launchOptions: {
    args: [
      "--autoplay-policy=no-user-gesture-required",
      "--disable-features=WebRtcHideLocalIpsWithMdns",
      "--use-fake-ui-for-media-stream",
    ],
  },
});

type SignalMessage = {
  type: string;
  call_session_id?: string;
  payload?: Record<string, unknown>;
};

const call = {
  call_session_id: "66666666-6666-6666-6666-666666666666",
  conversation_id: "55555555-5555-5555-5555-555555555555",
  visitor_id: "11111111-1111-1111-1111-111111111111",
  visit_session_id: "22222222-2222-2222-2222-222222222222",
  media_kind: "audio",
  direction: "visitor_to_operator",
  requested_by: "visitor",
  status: "ringing",
  recording_enabled: false,
};

const voice = {
  ice_servers: [
    { urls: ["stun:stun.test:19302"] },
    { urls: ["turn:turn.test:3478?transport=tcp"], username: "turn-user", credential: "turn-secret" },
  ],
  ice_transport_policy: "all",
  audio: {
    echo_cancellation: true,
    noise_suppression: true,
    auto_gain_control: true,
    channel_count: 1,
  },
};

async function installToneMicrophone(page: Page, frequency: number) {
  await page.addInitScript((toneFrequency) => {
    const existing = navigator.mediaDevices ?? {};
    const contexts: AudioContext[] = [];
    Object.defineProperty(navigator, "mediaDevices", {
      configurable: true,
      value: {
        ...existing,
        getUserMedia: async () => {
          const AudioContextCtor = window.AudioContext || (window as typeof window & { webkitAudioContext?: typeof AudioContext }).webkitAudioContext;
          const ctx = new AudioContextCtor();
          await ctx.resume().catch(() => undefined);
          const destination = ctx.createMediaStreamDestination();
          const oscillator = ctx.createOscillator();
          const gain = ctx.createGain();
          oscillator.type = "sine";
          oscillator.frequency.value = toneFrequency as number;
          gain.gain.value = 0.22;
          oscillator.connect(gain);
          gain.connect(destination);
          oscillator.start();
          contexts.push(ctx);
          (window as typeof window & { __rpToneAudioContexts?: AudioContext[] }).__rpToneAudioContexts = contexts;
          return destination.stream;
        },
      },
    });
  }, frequency);
}

async function rmsFromAudioElement(page: Page, expression: string) {
  return page.evaluate(async (sourceExpression) => {
    async function measure(stream: MediaStream) {
      const AudioContextCtor = window.AudioContext || (window as typeof window & { webkitAudioContext?: typeof AudioContext }).webkitAudioContext;
      const ctx = new AudioContextCtor();
      await ctx.resume().catch(() => undefined);
      const source = ctx.createMediaStreamSource(stream);
      const analyser = ctx.createAnalyser();
      analyser.fftSize = 2048;
      source.connect(analyser);
      const data = new Float32Array(analyser.fftSize);
      let max = 0;
      for (let frame = 0; frame < 24; frame += 1) {
        await new Promise((resolve) => setTimeout(resolve, 50));
        analyser.getFloatTimeDomainData(data);
        let sum = 0;
        for (let i = 0; i < data.length; i += 1) sum += data[i] * data[i];
        max = Math.max(max, Math.sqrt(sum / data.length));
      }
      await ctx.close().catch(() => undefined);
      return max;
    }
    const audio = Function(`return (${sourceExpression})`)() as HTMLAudioElement | null;
    if (!audio) return 0;
    audio.muted = false;
    audio.volume = 1;
    await audio.play().catch(() => undefined);
    const capture = audio.captureStream?.() ?? (audio as HTMLAudioElement & { mozCaptureStream?: () => MediaStream }).mozCaptureStream?.();
    if (!capture) return 0;
    return measure(capture);
  }, expression);
}

async function inboundAudioPackets(page: Page, name: string) {
  return page.evaluate(async (pcName) => {
    const pc = (window as unknown as Record<string, RTCPeerConnection | undefined>)[pcName];
    if (!pc) return 0;
    const stats = await pc.getStats();
    let packets = 0;
    stats.forEach((report) => {
      if (report.type === "inbound-rtp" && report.kind === "audio") {
        packets += Number(report.packetsReceived || 0);
      }
    });
    return packets;
  }, name);
}

test("site chat voice uses WebRTC media tracks with configured ICE servers", async ({ browser }) => {
  const widgetPath = path.resolve(process.cwd(), "../api/internal/modules/sitechat/transport/http/widget.js");
  const widget = await readFile(widgetPath, "utf8");
  expect(widget).toContain("RTCPeerConnection");
  expect(widget).toContain("call.audio_chunk");

  const visitor = await browser.newPage();
  const operator = await browser.newPage();

  await installToneMicrophone(visitor, 440);
  await installToneMicrophone(operator, 660);
  await visitor.context().grantPermissions(["microphone"], { origin: "https://visitor.test" });
  await operator.context().grantPermissions(["microphone"], { origin: "https://operator.test" });

  const visitorMessages: SignalMessage[] = [];
  const waiters = new Map<string, ((message: SignalMessage) => void)[]>();
  const waitForVisitorMessage = (type: string, after = -1) => new Promise<SignalMessage>((resolve) => {
    const existing = visitorMessages.find((message, index) => index > after && message.type === type);
    if (existing) {
      resolve(existing);
      return;
    }
    waiters.set(type, [...(waiters.get(type) ?? []), resolve]);
  });

  async function sendToVisitor(message: SignalMessage) {
    await visitor.evaluate((msg) => {
      const win = window as typeof window & { __rpMockWS?: { onmessage?: (event: MessageEvent) => void } };
      win.__rpMockWS?.onmessage?.({ data: JSON.stringify(msg) } as MessageEvent);
    }, message);
  }

  await visitor.exposeBinding("__rpVisitorWS", async (_source, raw: string) => {
    const message = JSON.parse(raw) as SignalMessage;
    visitorMessages.push(message);
    const queue = waiters.get(message.type) ?? [];
    waiters.delete(message.type);
    queue.forEach((resolve) => resolve(message));
    if (message.type === "webrtc.ice" && message.payload?.candidate) {
      await operator.evaluate(async (candidate) => {
        const win = window as typeof window & { __operatorPC?: RTCPeerConnection; __operatorIce?: RTCIceCandidateInit[] };
        if (!win.__operatorPC) return;
        if (!win.__operatorPC.remoteDescription) {
          win.__operatorIce = [...(win.__operatorIce ?? []), candidate as RTCIceCandidateInit];
          return;
        }
        await win.__operatorPC.addIceCandidate(candidate as RTCIceCandidateInit).catch(() => undefined);
      }, message.payload.candidate);
    }
  });

  await operator.exposeBinding("__rpOperatorSignal", async (_source, message: SignalMessage) => {
    await sendToVisitor(message);
  });

  await visitor.addInitScript(() => {
    const NativeRTCPeerConnection = window.RTCPeerConnection;
    function PatchedRTCPeerConnection(
      this: RTCPeerConnection,
      ...args: ConstructorParameters<typeof RTCPeerConnection>
    ) {
      const pc = new NativeRTCPeerConnection(...args);
      (window as typeof window & { __debugPC?: RTCPeerConnection; __debugPCConfig?: RTCConfiguration }).__debugPC = pc;
      (window as typeof window & { __debugPCConfig?: RTCConfiguration }).__debugPCConfig = args[0] ?? {};
      return pc;
    }
    PatchedRTCPeerConnection.prototype = NativeRTCPeerConnection.prototype;
    Object.assign(PatchedRTCPeerConnection, NativeRTCPeerConnection);
    Object.defineProperty(window, "RTCPeerConnection", { configurable: true, value: PatchedRTCPeerConnection });

    class MockWebSocket {
      static CONNECTING = 0;
      static OPEN = 1;
      static CLOSING = 2;
      static CLOSED = 3;
      readyState = MockWebSocket.CONNECTING;
      bufferedAmount = 0;
      onopen?: () => void;
      onmessage?: (event: MessageEvent) => void;
      onclose?: () => void;
      onerror?: () => void;
      constructor(public url: string) {
        (window as typeof window & { __rpMockWS?: MockWebSocket }).__rpMockWS = this;
        setTimeout(() => {
          this.readyState = MockWebSocket.OPEN;
          this.onopen?.();
          this.onmessage?.({ data: JSON.stringify({ type: "ack", payload: {} }) } as MessageEvent);
        }, 80);
      }
      send(data: string) {
        void (window as typeof window & { __rpVisitorWS: (raw: string) => Promise<void> }).__rpVisitorWS(data);
      }
      close() {
        this.readyState = MockWebSocket.CLOSED;
        this.onclose?.();
      }
    }
    Object.defineProperty(window, "WebSocket", { configurable: true, value: MockWebSocket });
  });

  await visitor.route("https://visitor.test/**", async (route) => {
    await route.fulfill({
      contentType: "text/html",
      body: `<!doctype html><html><head><meta charset="utf-8"><title>Visitor audio test</title></head><body>
        <script async src="https://pulse.test/site-chat/widget.js" data-site-chat-key="k"></script>
      </body></html>`,
    });
  });
  await visitor.route("https://pulse.test/site-chat/widget.js", async (route) => {
    await route.fulfill({ contentType: "application/javascript", body: widget });
  });
  await visitor.route("https://pulse.test/api/v1/public/site-chat/widgets/k/config", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify({ title: "Test", subtitle: "", accent_color: "#00B2FF", position: "right", settings_json: {}, voice }),
    });
  });
  await visitor.route("https://pulse.test/api/v1/public/site-chat/presence", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify({
        visitor_id: call.visitor_id,
        visit_session_id: call.visit_session_id,
        chat_widget_id: "33333333-3333-3333-3333-333333333333",
        project_id: "44444444-4444-4444-4444-444444444444",
        conversation_id: call.conversation_id,
        monitoring_status: "stopped",
        visitor_state_snapshot: {},
        voice,
      }),
    });
  });
  await visitor.route("https://pulse.test/api/v1/public/site-chat/conversations/**/messages?**", async (route) => {
    await route.fulfill({ contentType: "application/json", body: JSON.stringify([]) });
  });
  await visitor.route("https://pulse.test/api/v1/public/site-chat/conversations", async (route) => {
    await route.fulfill({ contentType: "application/json", body: JSON.stringify({ conversation_id: call.conversation_id }) });
  });
  await visitor.route("https://pulse.test/api/v1/public/site-chat/calls/request", async (route) => {
    await route.fulfill({ contentType: "application/json", body: JSON.stringify(call) });
  });
  await visitor.route("https://pulse.test/api/v1/public/site-chat/events", async (route) => {
    await route.fulfill({ status: 202, contentType: "application/json", body: JSON.stringify({ status: "accepted" }) });
  });
  await operator.route("https://operator.test/**", async (route) => {
    await route.fulfill({ contentType: "text/html", body: "<!doctype html><html><body><button>operator</button></body></html>" });
  });

  await operator.goto("https://operator.test/", { waitUntil: "domcontentloaded" });
  await operator.locator("button").click();
  await visitor.goto("https://visitor.test/?rp_audio=1", { waitUntil: "domcontentloaded" });
  await visitor.waitForFunction(() => document.querySelector(".rp-site-chat")?.shadowRoot?.querySelector("[data-call-open]"));
  await visitor.locator("[data-open]").click();
  await visitor.locator("[data-privacy]").check();
  await visitor.locator("[data-call-open]").click();

  await waitForVisitorMessage("call.request");
  await sendToVisitor({ type: "call.accept", call_session_id: call.call_session_id, payload: { call_session: { ...call, status: "accepted" } } });

  const offer = await operator.evaluate(async ({ callSessionId, voiceConfig }) => {
    const win = window as typeof window & {
      __operatorPC?: RTCPeerConnection;
      __operatorRemoteStream?: MediaStream;
      __operatorIce?: RTCIceCandidateInit[];
      __rpOperatorSignal: (message: SignalMessage) => Promise<void>;
    };
    const pc = new RTCPeerConnection({ iceServers: voiceConfig.ice_servers as RTCIceServer[] });
    win.__operatorPC = pc;
    pc.onicecandidate = (event) => {
      if (!event.candidate) return;
      void win.__rpOperatorSignal({
        type: "webrtc.ice",
        call_session_id: callSessionId,
        payload: { candidate: event.candidate.toJSON(), sender: "operator" },
      });
    };
    pc.ontrack = (event) => {
      win.__operatorRemoteStream = event.streams[0] ?? new MediaStream([event.track]);
    };
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
    stream.getAudioTracks().forEach((track) => pc.addTrack(track, stream));
    await pc.setLocalDescription(await pc.createOffer());
    return { description: pc.localDescription?.toJSON() };
  }, { callSessionId: call.call_session_id, voiceConfig: voice });

  await sendToVisitor({ type: "webrtc.offer", call_session_id: call.call_session_id, payload: { description: offer.description, sender: "operator" } });
  const answer = await waitForVisitorMessage("webrtc.answer");
  await operator.evaluate(async ({ description }) => {
    const win = window as typeof window & { __operatorPC?: RTCPeerConnection; __operatorIce?: RTCIceCandidateInit[] };
    const pc = win.__operatorPC;
    if (!pc) throw new Error("missing operator peer");
    await pc.setRemoteDescription(description as RTCSessionDescriptionInit);
    const pending = win.__operatorIce ?? [];
    win.__operatorIce = [];
    await Promise.all(pending.map((candidate) => pc.addIceCandidate(candidate).catch(() => undefined)));
  }, { description: answer.payload?.description });

  await expect.poll(() => visitor.evaluate(() => {
    const cfg = (window as typeof window & { __debugPCConfig?: RTCConfiguration }).__debugPCConfig;
    const urls = cfg?.iceServers?.flatMap((server) => Array.isArray(server.urls) ? server.urls : [server.urls]) ?? [];
    return urls.some((url) => String(url).startsWith("turn:turn.test"));
  })).toBe(true);
  await expect.poll(() => visitorMessages.filter((message) => message.type === "call.audio_chunk").length).toBeGreaterThan(0);
  await expect.poll(() => operator.evaluate(() => {
    const win = window as typeof window & { __operatorRemoteStream?: MediaStream };
    return Boolean(win.__operatorRemoteStream?.getAudioTracks().length);
  })).toBe(true);
  await expect.poll(() => visitor.evaluate(() => {
    const root = document.querySelector(".rp-site-chat") as HTMLElement | null;
    const audio = root?.shadowRoot?.querySelector("[data-remote-audio]") as HTMLAudioElement | null;
    return Boolean((audio?.srcObject as MediaStream | null)?.getAudioTracks().length);
  })).toBe(true);

  await expect.poll(() => inboundAudioPackets(operator, "__operatorPC")).toBeGreaterThan(0);
  await expect.poll(() => rmsFromAudioElement(visitor, "document.querySelector('.rp-site-chat').shadowRoot.querySelector('[data-remote-audio]')")).toBeGreaterThan(0.02);

  await operator.close();
  await visitor.close();
});
