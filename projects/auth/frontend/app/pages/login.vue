<script setup lang="ts">
import type { FormSubmitEvent } from "@nuxt/ui";
import { z } from "zod";

definePageMeta({
  layout: "auth",
  middleware: "guest",
});

const { login } = useAuth();
const serverError = ref<string | null>(null);
const router = useRouter();

const schema = z.object({
  username: z.string().min(1, "Username is required"),
  password: z.string().min(1, "Password is required"),
});

async function handleSubmit(e: FormSubmitEvent<z.output<typeof schema>>) {
  serverError.value = null;
  try {
    await login(e.data.username, e.data.password);
    router.push("/");
  } catch (error: unknown) {
    serverError.value =
      error instanceof ApiError ? error.message : "Could not sign in";
  }
}
</script>

<template>
  <div class="min-h-dvh flex items-center justify-center">
    <UPageCard class="w-full max-w-md">
      <UAuthForm
        icon="i-lucide-log-in"
        title="Welcome Back!"
        description="Sign in to your account."
        :submit="{ label: 'Sign in', block: true }"
        @submit="handleSubmit"
        :schema="schema"
        :fields="[
          {
            type: 'text',
            name: 'username',
            label: 'Username',
            placeholder: 'Enter your username',
            required: true,
          },
          {
            type: 'password',
            name: 'password',
            label: 'Password',
            placeholder: 'Enter your password',
            required: true,
          },
        ]"
      >
        <template #validation>
          <UAlert
            v-if="serverError"
            color="error"
            variant="subtle"
            icon="i-lucide-alert-circle"
            :title="serverError"
          />
        </template>

        <template #footer>
          <p class="text-center text-sm text-muted">
            Don't have an account?
            <ULink to="/register" class="text-primary font-medium">
              Register
            </ULink>
            .
          </p>
        </template>
      </UAuthForm>
    </UPageCard>
  </div>
</template>
