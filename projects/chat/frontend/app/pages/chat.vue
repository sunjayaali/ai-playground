<script setup lang="ts">
import { useWebSocket } from "@vueuse/core";

definePageMeta({
  layout: false,
});

// --- socket state -----------------------------------------------------------

interface RawMessage {
  type: "echo" | "system" | "pong" | "connected_clients" | "server_time";
  text: string;
  ts: number;
  from?: string;
  count?: number;
  time?: string;
}

type SocketStatus = "connecting" | "open" | "closed" | "error";

const MAX_MESSAGES = 200;
const HEARTBEAT_MS = 10_000;

const config = useRuntimeConfig();
const status = ref<SocketStatus>("closed");
const rawMessages = ref<RawMessage[]>([]);
const latency = ref<number | null>(null);
const connectedClients = ref(0);
const serverTime = ref("");

// A per-browser fingerprint so we can identify
// which echoed frames are our own.
const selfId = crypto.randomUUID();

let heartbeat: ReturnType<typeof setInterval> | null = null;
let pingSentAt = 0;

function push(msg: RawMessage) {
  rawMessages.value.push(msg);
  if (rawMessages.value.length > MAX_MESSAGES) {
    rawMessages.value.splice(0, rawMessages.value.length - MAX_MESSAGES);
  }
}

const socket = useWebSocket(config.public.wsUrl, {
  autoReconnect: {
    retries: Infinity,
    delay: (retries) => Math.min(500 * 2 ** retries, 10_000),
  },
  onMessage(_ws, event) {
    const raw = typeof event.data === "string" ? event.data : null;
    if (raw === null) return;
    try {
      const msg = JSON.parse(raw) as RawMessage;
      if (msg.type === "pong") {
        latency.value = Math.round(performance.now() - pingSentAt);
        return;
      }
      if (msg.type === "connected_clients") {
        connectedClients.value = msg.count ?? 0;
        return;
      }
      if (msg.type === "server_time") {
        serverTime.value = msg.time ?? "";
        return;
      }
      push(msg);
    } catch {
      push({ type: "system", text: raw, ts: Date.now() });
    }
  },
});

watch(
  () => socket.status.value,
  (next) => {
    status.value =
      next === "OPEN"
        ? "open"
        : next === "CONNECTING"
          ? "connecting"
          : "closed";
    if (next === "OPEN") {
      if (heartbeat) return;
      heartbeat = setInterval(() => {
        if (socket.status.value !== "OPEN") return;
        pingSentAt = performance.now();
        socket.send(JSON.stringify({ type: "ping", text: "" }));
      }, HEARTBEAT_MS);
    } else if (heartbeat) {
      clearInterval(heartbeat);
      heartbeat = null;
    }
  },
  { immediate: true },
);

const isOpen = computed(() => status.value === "open");

function send(text: string) {
  if (!isOpen.value) return false;
  socket.send(JSON.stringify({ type: "echo", text, from: selfId }));
  return true;
}

onUnmounted(() => {
  if (heartbeat) clearInterval(heartbeat);
  heartbeat = null;
});

// --- map raw messages to UIMessage shape for UChatMessage -----------------

function isMine(msg: RawMessage) {
  return msg.type === "echo" && msg.from === selfId;
}

// UChatMessage renders each message directly (as
// UChatMessages would internally). Own echoes are "user"
// messages on the right; everything else is an "assistant"
// message on the left, both as bubbles with distinct colors.
interface ChatMessage {
  id: string;
  role: "user" | "assistant";
  side: "left" | "right";
  variant: "solid" | "soft";
  parts: { type: "text"; text: string }[];
}

const chatMessages = computed<ChatMessage[]>(() =>
  rawMessages.value.map((msg, i): ChatMessage => ({
    id: `${msg.ts}-${i}`,
    role: isMine(msg) ? "user" : "assistant",
    side: isMine(msg) ? "right" : "left",
    variant: isMine(msg) ? "solid" : "soft",
    parts: [{ type: "text", text: msg.text }],
  })),
);

// Pin the feed to the newest message — scroll to bottom after each push
const scrollAreaRef = useTemplateRef("scrollAreaRef");

watch(
  () => rawMessages.value.length,
  async () => {
    await nextTick();
    const el = scrollAreaRef.value?.$el;
    el?.scrollTo({ top: el.scrollHeight, behavior: "smooth" });
  },
);

// --- composer -----------------------------------------------------------------

const draft = ref("");

function submit() {
  const text = draft.value.trim();
  if (!text || !send(text)) return;
  draft.value = "";
}
</script>

<template>
  <UApp>
    <UMain :ui="{ base: 'flex flex-col h-screen bg-background' }">
      <!-- header -->
      <UHeader
        title="Chat"
        :toggle="false"
        :ui="{
          root: 'h-auto py-2 border-b border-default bg-default/75 backdrop-blur-sm',
          left: 'gap-3',
          right: 'gap-4 text-xs text-muted',
          title: 'font-semibold text-base',
        }"
      >
        <template #left>
          <div class="flex items-center gap-3">
            <UIcon name="i-lucide-message-circle" size="5" />
            <h1 class="font-semibold text-base">Chat</h1>
            <UBadge
              :color="
                status === 'open'
                  ? 'success'
                  : status === 'connecting'
                    ? 'warning'
                    : 'error'
              "
              variant="subtle"
              size="sm"
              :ui="{ base: 'font-mono uppercase' }"
            >
              {{ status }}
            </UBadge>
          </div>
        </template>

        <template #right>
          <span class="hidden sm:flex items-center gap-1">
            <UIcon name="i-lucide-users" size="3.5" />
            {{ connectedClients }} connected
          </span>

          <span v-if="latency !== null" class="hidden md:inline tabular-nums">
            {{ latency }} ms
          </span>

          <span v-if="serverTime" class="hidden lg:inline tabular-nums">
            {{ serverTime }}
          </span>

          <UButton
            v-if="!isOpen"
            size="xs"
            color="neutral"
            variant="subtle"
            icon="i-lucide-plug-zap"
            @click="() => socket.open()"
          >
            Connect
          </UButton>
          <UButton
            v-else
            size="xs"
            color="neutral"
            variant="subtle"
            icon="i-lucide-plug-x"
            @click="() => socket.close()"
          >
            Disconnect
          </UButton>
        </template>
      </UHeader>

      <!-- feed: each message renders itself, as UChatMessages does internally -->
      <UScrollArea
        ref="scrollAreaRef"
        class="flex-1 min-h-0"
        :ui="{ root: 'px-4 py-4' }"
      >
        <div v-if="chatMessages.length">
          <UChatMessage
            v-for="msg in chatMessages"
            :key="msg.id"
            v-bind="msg"
            :ui="{ container: 'max-w-[75%]' }"
          />
        </div>

        <div v-else class="flex flex-col justify-center h-full">
          <UEmpty
            icon="i-lucide-message-square-dashed"
            size="lg"
            variant="naked"
            :title="
              status === 'open'
                ? 'Waiting for messages…'
                : 'Not connected to the hub.'
            "
            :description="
              status === 'open'
                ? 'Say hi — everything you send is broadcast to the room.'
                : 'Press Connect to join the chat.'
            "
          />
        </div>
      </UScrollArea>

      <!-- composer -->
      <UFooter
        :ui="{
          root: 'border-t border-default bg-default/10 pb-[max(0.75rem,env(safe-area-inset-bottom))]',
          container: 'p-3 sm:p-4',
        }"
      >
        <UChatPrompt
          v-model="draft"
          :placeholder="isOpen ? 'Type a message' : 'Waiting for the hub'"
          :disabled="!isOpen"
          :rows="1"
          :maxrows="5"
          :autofocus="false"
          autoresize
          @submit.prevent="submit"
        >
          <template #footer>
            <span />
            <UChatPromptSubmit :disabled="!isOpen || !draft.trim()" />
          </template>
        </UChatPrompt>
      </UFooter>
    </UMain>
  </UApp>
</template>
