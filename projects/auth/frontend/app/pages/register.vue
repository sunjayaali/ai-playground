<script setup lang="ts">
import * as z from "zod";
import type { FormSubmitEvent } from "@nuxt/ui";

const auth = useAuth();
const toast = useToast();
const router = useRouter();

definePageMeta({
  layout: "auth",
  middleware: "guest",
});

const schema = z.object({
  username: z.string().min(3, "Must be at least 3 characters"),
  password: z.string().min(8, "Must be at least 8 characters"),
});

type Schema = z.output<typeof schema>;

const serverError = ref<string | null>(null);

async function onSubmit(event: FormSubmitEvent<Schema>) {
  serverError.value = null;
  try {
    await auth.register(event.data.username, event.data.password);
    toast.add({
      title: "Account created",
      color: "success",
      icon: "i-lucide-check-circle",
    });
    router.push("/");
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
        :fields="[
          {
            type: 'text',
            name: 'username',
            label: 'Username',
            placeholder: 'Enter your username',
            required: true,
          },
          {
            name: 'password',
            label: 'Password',
            type: 'password',
            placeholder: 'At least 8 characters',
            required: true,
          },
        ]"
        title="Create your account"
        description="Sign up to get a token pair and access the dashboard."
        icon="i-lucide-user-plus"
        :submit="{ label: 'Register', block: true }"
        @submit="onSubmit"
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
