<script setup lang="ts">
import * as z from "zod";
import type { AuthFormField, FormSubmitEvent } from "@nuxt/ui";
import { ApiError } from "~/services/auth.api";

definePageMeta({
  layout: "auth",
  middleware: "guest",
});

const authStore = useAuthStore();
const router = useRouter();

const serverError = ref<string | null>(null);

const schema = z.object({
  username: z.string().min(1, "Username is required"),
  password: z.string().min(1, "Password is required"),
});

const fields: AuthFormField[] = [
  {
    type: "text",
    name: "username",
    label: "Username",
    placeholder: "Enter your username",
    required: true,
  },
  {
    name: "password",
    label: "Password",
    type: "password",
    placeholder: "Enter your password",
    required: true,
  },
];

async function submit(payload: { username: string; password: string }) {
  serverError.value = null;
  try {
    await authStore.login(payload.username, payload.password);
    await router.push("/");
  } catch (error: unknown) {
    console.log(error);
    serverError.value =
      error instanceof ApiError ? error.message : "Could not sign in";
  }
}

function handleSubmit(event: FormSubmitEvent<z.output<typeof schema>>) {
  void submit(event.data);
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
        :schema="schema"
        :fields="fields"
        @submit="handleSubmit"
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
