<script setup lang="ts">
definePageMeta({
  middleware: "auth",
});
  
const { user } = useAuth();
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
              <div class="flex items-center gap-2 font-medium text-muted">
                <UIcon name="i-lucide-fingerprint" />
                Identity
              </div>
            </template>

            <div class="font-mono text-sm">
              {{ user?.id }}
            </div>
          </UCard>
        </UPageGrid>
      </UPageBody>
    </UPage>
  </UContainer>
</template>
