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
const toast = useToast();

const serverError = ref<string | null>(null);

const schema = z.object({
  username: z.string().min(3, "Must be at least 3 characters"),
  password: z.string().min(8, "Must be at least 8 characters"),
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
    placeholder: "At least 8 characters",
    required: true,
  },
];

async function handleSubmit(event: FormSubmitEvent<z.output<typeof schema>>) {
  serverError.value = null;
  try {
    await authStore.register(event.data.username, event.data.password);
    toast.add({
      title: "Account created",
      color: "success",
      icon: "i-lucide-check-circle",
    });
    await router.push("/");
  } catch (err) {
    // 409 comes back when the name is taken.
    serverError.value =
      err instanceof ApiError ? err.message : "Could not register";
  }
}
</script>

<template>
  <div class="flex min-h-dvh items-center justify-center">
    <UPageCard class="w-full max-w-md">
      <UAuthForm
        :schema="schema"
        :fields="fields"
        title="Create your account"
        description="Sign up to get a token pair and access the dashboard."
        icon="i-lucide-user-plus"
        :submit="{ label: 'Register', block: true }"
        @submit="handleSubmit"
      >
        <template #validation>
          <UAlert
            v-if="serverError"
            color="error"
            icon="i-lucide-alert-circle"
            :title="serverError"
            variant="subtle"
          />
        </template>
        <template #footer>
          <p class="text-center text-sm text-muted">
            Already have an account?
            <ULink to="/login" class="text-primary font-medium">
              Sign in
            </ULink>
            .
          </p>
        </template>
      </UAuthForm>
    </UPageCard>
  </div>
</template>
