<script setup lang="ts">
import type { NavigationMenuItem } from "@nuxt/ui";

const router = useRouter();
const { signOut, user, isAuthenticated } = useAuth();

async function doSignOut() {
  await signOut();
  await router.push("/login");
}

const menuItems = computed<NavigationMenuItem[]>(() => [
  { label: "Home", to: "/" },
  { label: "Blank", to: "/blank" },
]);
</script>

<template>
  <UHeader>
    <UNavigationMenu orientation="horizontal" :items="menuItems" color="neutral" />

    <template #body>
      <UNavigationMenu :items="menuItems" orientation="vertical" color="neutral" />
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
            onSelect: doSignOut,
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
</template>
