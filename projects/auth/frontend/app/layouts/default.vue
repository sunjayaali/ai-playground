<script setup lang="ts">
import type { NavigationMenuItem } from "@nuxt/ui";

const router = useRouter();
const { signOut, user } = useAuth();

async function doSignOut() {
  await signOut();
  await router.push("/login");
}

const menuItems = computed<NavigationMenuItem[]>(() => [
  { label: "Home", to: "/" },
]);
</script>

<template>
  <UHeader>
    <UNavigationMenu orientation="horizontal" :items="menuItems" />

    <template #body>
      <UNavigationMenu :items="menuItems" orientation="vertical" />
    </template>

    <template #right>
      <UColorModeButton />
      <UDropdownMenu
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
