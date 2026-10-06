<script setup lang="ts">
import { useWebSocket } from "@vueuse/core";
import { v7 as uuidv7 } from "uuid";

enum MessageType {
  ConnectedClients = "connected_clients",
  System = "system",
  Echo = "echo",
}

type Message =
  | {
      type: MessageType.ConnectedClients;
      count: number;
    }
  | {
      type: MessageType.System;
      text: string;
      ts: number;
    }
  | {
      type: MessageType.Echo;
      text: string;
      ts: number;
      from: string;
    };

const connectedClients = ref(0);
const messages = ref<Message[]>([]);

const scrollAreaRef = useTemplateRef("scrollAreaRef");
watch(
  () => messages.value.length,
  async () => {
    await nextTick();

    const viewport = scrollAreaRef.value?.$el;
    if (!viewport) return;

    viewport.scrollTo({
      top: viewport.scrollHeight,
      behavior: "smooth",
    });
  },
);

const echoMessages = computed(() =>
  messages.value.filter((message) => message.type === MessageType.Echo),
);

const { status, open, close, send } = useWebSocket(
  useRuntimeConfig().public.wsUrl,
  {
    autoReconnect: {
      retries: Infinity,
      delay: 1000,
    },
    onMessage(ws, event) {
      console.log(event.data);
      const msg = JSON.parse(event.data) as Message;
      if (msg.type === MessageType.ConnectedClients) {
        connectedClients.value = msg.count;
        return;
      }

      if (msg.type === MessageType.Echo) {
        messages.value.push(msg);
        return;
      }
    },
  },
);

const selfId = uuidv7();
const isConnected = computed(() => status.value === "OPEN");

function isMine(msg: Message) {
  return msg.type === MessageType.Echo && msg.from === selfId;
}

const draft = ref("");
const submit = () => {
  send(
    JSON.stringify({
      type: "echo",
      text: draft.value.trim(),
      from: selfId,
    }),
  );
  draft.value = "";
};
</script>

<template>
  <div class="h-dvh flex flex-col">
    <UHeader title="Chat" :toggle="false">
      <template #title>asd</template>

      <template #left>
        <h1>Chat</h1>
        <UBadge
          variant="subtle"
          :color="
            status === 'OPEN'
              ? 'success'
              : status === 'CONNECTING'
                ? 'secondary'
                : 'error'
          "
        >
          {{ status }}
        </UBadge>
      </template>

      <template #right>
        <span class="items-center gap-1 hidden sm:flex text-muted">
          <UIcon name="i-lucide-users" />
          {{ connectedClients }} connected
        </span>

        <span class="hidden sm:flex text-muted"> 123 ms </span>

        <UButton
          variant="subtle"
          color="neutral"
          @click="isConnected ? close() : open()"
          >{{ isConnected ? "Disconnect" : "Connect" }}</UButton
        >
      </template>
    </UHeader>

    <UMain class="flex min-h-0 flex-1 flex-col">
      <UScrollArea ref="scrollAreaRef" :ui="{ root: 'p-4' }">
        <UChatMessages
          :should-auto-scroll="true"
          :should-scroll-to-bottom="true"
        >
          <UChatMessage
            v-for="(message, i) in echoMessages"
            :id="`${message.type}-${i}`"
            :key="`${message.type}-${i}`"
            :ui="{
              container: 'max-w-[75%]',
            }"
            compact
            role="user"
            :side="isMine(message) ? 'right' : 'left'"
            :variant="isMine(message) ? 'solid' : 'soft'"
            :parts="[
              {
                type: 'text',
                text: message.text,
              },
            ]"
          ></UChatMessage>
        </UChatMessages>
      </UScrollArea>
    </UMain>

    <div class="border-t border-default p-4">
      <UChatPrompt v-model="draft" :maxrows="8" @submit="submit">
        <UChatPromptSubmit class="self-end" />
      </UChatPrompt>
    </div>
  </div>
</template>
