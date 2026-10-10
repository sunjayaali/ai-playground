<script setup lang="ts">
import type { NavigationMenuItem } from "@nuxt/ui";

const authStore = useAuthStore();
const router = useRouter();
const colorMode = useColorMode();

const { user } = storeToRefs(authStore);

const navItems = computed<NavigationMenuItem[]>(() => [
  {
    label: "Home",
    icon: "i-lucide-house",
    to: "/",
  },
  {
    label: "Blank",
    icon: "i-lucide-file",
    to: "/blank",
  },
]);

async function signOut() {
  await authStore.signOut();
  await router.push("/login");
}

const route = useRoute();

// Page title comes from route meta, falling back to the route name.
const pageTitle = computed(
  () =>
    (route.meta.title as string | undefined) ?? route.name?.toString() ?? "",
);

const dropdownItems = computed(() => [
  [
    {
      type: "label" as const,
      label: user.value?.username,
    },
  ],
  [
    {
      label: "Profile",
      icon: "i-lucide-user",
    },
    {
      label: "Settings",
      icon: "i-lucide-settings",
    },
  ],
  [
    {
      label: "Appearance",
      icon: "i-lucide-sun-moon",
      children: [
        {
          label: "Light",
          icon: "i-lucide-sun",
          type: "checkbox" as const,
          checked: colorMode.preference === "light",
          onUpdateChecked: () => (colorMode.preference = "light"),
        },
        {
          label: "Dark",
          icon: "i-lucide-moon",
          type: "checkbox" as const,
          checked: colorMode.preference === "dark",
          onUpdateChecked: () => (colorMode.preference = "dark"),
        },
        {
          label: "System",
          icon: "i-lucide-monitor",
          type: "checkbox" as const,
          checked: colorMode.preference === "system",
          onUpdateChecked: () => (colorMode.preference = "system"),
        },
      ],
    },
  ],
  [
    {
      label: "Sign Out",
      icon: "i-lucide-log-out",
      onSelect: signOut,
    },
  ],
]);
</script>

<template>
  <UDashboardGroup>
    <UDashboardSidebar collapsible resizable>
      <template #header="{ collapsed }">
        <div class="flex items-center gap-2 px-2">
          <UIcon
            name="i-lucide-shield-check"
            class="size-5 shrink-0 text-primary"
          />
          <span v-if="!collapsed" class="truncate font-semibold"
            >Auth Playground</span
          >
        </div>
      </template>

      <template #default="{ collapsed }">
        <UNavigationMenu
          :collapsed="collapsed"
          :items="navItems"
          orientation="vertical"
        />
      </template>

      <template #footer="{ collapsed }">
        <UDropdownMenu
          :items="dropdownItems"
          :ui="{
            content: collapsed
              ? 'w-48'
              : 'w-(--reka-dropdown-menu-trigger-width)',
          }"
        >
          <UButton
            block
            variant="ghost"
            color="neutral"
            icon="i-lucide-circle-user"
            :label="user?.username"
            :square="collapsed"
            :trailing-icon="collapsed ? undefined : 'i-lucide-chevrons-up-down'"
          />
        </UDropdownMenu>
      </template>
    </UDashboardSidebar>

    <UDashboardPanel>
      <template #header>
        <UDashboardNavbar :title="pageTitle">
          <template #leading>
            <UDashboardSidebarCollapse />
          </template>
        </UDashboardNavbar>
      </template>

      <template #body>
        <slot />
      </template>
    </UDashboardPanel>
  </UDashboardGroup>
</template>
