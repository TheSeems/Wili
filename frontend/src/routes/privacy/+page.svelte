<script lang="ts">
  import T from "$lib/components/T.svelte";
  // Static privacy policy for MVP
  const lastUpdated = "2025-08-10";
  let { data } = $props();
  void data;
</script>

<section
  class="container mx-auto max-w-3xl px-4 py-12 text-[0.95rem] text-gray-700 dark:text-gray-300"
>
  <h1 class="mb-2 text-3xl font-bold"><T key="privacy.title" fallback="Privacy Policy" /></h1>
  <p class="mb-8 opacity-70">
    <T
      key="privacy.lastUpdated"
      values={{ date: lastUpdated }}
      fallback={`Last updated: ${lastUpdated}`}
    />
  </p>

  <div class="space-y-5 leading-7">
    <p>
      <T
        key="privacy.intro"
        fallback="This Privacy Policy explains how Wili collects, uses, and protects your information when you use our service."
      />
    </p>

    <h2 class="mt-8 text-xl font-bold">
      <T key="privacy.sections.collect.title" fallback="Information We Collect" />
    </h2>
    <p>
      <T
        key="privacy.sections.collect.body"
        fallback="We collect your basic profile information from Yandex ID during authentication."
      />
    </p>

    <h2 class="mt-8 text-xl font-bold">
      <T key="privacy.sections.use.title" fallback="How We Use Information" />
    </h2>
    <p>
      <T
        key="privacy.sections.use.body"
        fallback="We use your information to provide and improve the wishlist service."
      />
    </p>

    <h2 class="mt-8 text-xl font-bold">
      <T key="privacy.sections.share.title" fallback="Data Sharing" />
    </h2>
    <p><T key="privacy.sections.share.body" fallback="We do not sell your personal data." /></p>

    <h2 class="mt-8 text-xl font-bold">
      <T key="privacy.sections.security.title" fallback="Security" />
    </h2>
    <p>
      <T
        key="privacy.sections.security.body"
        fallback="We take reasonable measures to protect your data."
      />
    </p>

    <h2 class="mt-8 text-xl font-bold">
      <T key="privacy.sections.choices.title" fallback="Your Choices" />
    </h2>
    <p>
      <T
        key="privacy.sections.choices.body"
        fallback="You can update your profile information in the app."
      />
    </p>

    <h2 class="mt-8 text-xl font-bold">
      <T key="privacy.sections.contact.title" fallback="Contact" />
    </h2>
    <p>
      <T
        key="privacy.sections.contact.body"
        fallback="If you have questions about this Privacy Policy, please open an issue on our GitHub repository."
      />
      <a
        class="underline"
        href="https://github.com/theseems/wili"
        target="_blank"
        rel="noopener noreferrer">GitHub</a
      >.
    </p>
  </div>
</section>
