<script setup lang="ts">
definePageMeta({
  middleware: "auth",
});

const toast = useToast();
const { user, refresh, fetchUser } = useAuth();

const refreshedAt = ref<Date | null>(null);
const identityLoadedAt = ref<Date | null>(null);

async function refreshSession() {
  if (await refresh()) {
    refreshedAt.value = new Date();
    toast.add({
      title: "Session refreshed",
      color: "success",
      icon: "i-lucide-check-circle",
    });
  } else {
    // refresh() cleared the identity state — the
    // session is unrecoverable, so leave the app.
    await navigateTo("/login", { replace: true });
  }
}

async function refreshIdentity() {
  if (await fetchUser()) {
    identityLoadedAt.value = new Date();
    toast.add({
      title: "Identity refreshed",
      color: "success",
      icon: "i-lucide-check-circle",
    });
  } else {
    // fetchUser() already retried after a token
    // refresh; still failing means the session is
    // gone — kick to login.
    await navigateTo("/login", { replace: true });
  }
}
</script>

<template>
  <UContainer>
    <UPage>
      <UPageBody>
        <UPageGrid
          :ui="{
            base: 'grid-cols-1 sm:grid-cols-2!',
          }"
        >
          <UCard title="Title">
            <template #header>
              <div class="flex items-center justify-between gap-2">
                <div class="flex items-center gap-2 font-medium text-muted">
                  <UIcon name="i-lucide-fingerprint" />
                  Identity
                </div>
                <UButton
                  icon="i-lucide-refresh-cw"
                  variant="ghost"
                  color="neutral"
                  size="sm"
                  @click="refreshIdentity"
                />
              </div>
            </template>

            <div class="flex flex-col gap-2 text-sm">
              <div class="flex items-center justify-between">
                <span class="text-muted">User ID</span>
                <span class="font-mono">{{ user?.id ?? "—" }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-muted">Username</span>
                <span class="font-mono">{{ user?.username ?? "—" }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-muted">Loaded</span>
                <span class="font-mono">
                  {{ identityLoadedAt ? identityLoadedAt.toLocaleTimeString() : "—" }}
                </span>
              </div>
            </div>
          </UCard>

          <UCard title="Title">
            <template #header>
              <div class="flex items-center justify-between gap-2">
                <div class="flex items-center gap-2 font-medium text-muted">
                  <UIcon name="i-lucide-refresh-cw" />
                  Session
                </div>
                <UButton
                  icon="i-lucide-refresh-cw"
                  variant="ghost"
                  color="neutral"
                  size="sm"
                  @click="refreshSession"
                />
              </div>
            </template>

            <div class="flex flex-col gap-2 text-sm">
              <div class="flex items-center justify-between">
                <span class="text-muted">Last refreshed</span>
                <span class="font-mono">
                  {{
                    refreshedAt
                      ? refreshedAt.toLocaleTimeString()
                      : "—"
                  }}
                </span>
              </div>
            </div>
          </UCard>
        </UPageGrid>
      </UPageBody>
    </UPage>
  </UContainer>
</template>
