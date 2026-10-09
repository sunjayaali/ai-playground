<script setup lang="ts">
import type { NavigationMenuItem } from "@nuxt/ui";

const authStore = useAuthStore();
const router = useRouter();

const { user, isAuthenticated } = storeToRefs(authStore);

const menuItems: NavigationMenuItem[] = [
  { label: "Home", to: "/" },
  { label: "Blank", to: "/blank" },
];

async function signOut() {
  await authStore.signOut();
  await router.push("/login");
}
</script>

<template>
  <div>
    <UHeader>
      <UNavigationMenu
        orientation="horizontal"
        :items="menuItems"
        color="neutral"
      />

      <template #body>
        <UNavigationMenu
          :items="menuItems"
          orientation="vertical"
          color="neutral"
        />
      </template>

      <template #right>
        <UColorModeButton />
        <UDropdownMenu
          v-if="isAuthenticated"
          :content="{
            align: 'end',
          }"
          :items="[
            {
              label: 'Sign Out',
              icon: 'i-lucide-log-out',
              onSelect: signOut,
            },
          ]"
        >
          <UButton
            variant="ghost"
            color="neutral"
            icon="i-lucide-circle-user"
            :label="user?.username"
          />
        </UDropdownMenu>
      </template>
    </UHeader>

    <UMain>
      <slot />
    </UMain>
  </div>
</template>
