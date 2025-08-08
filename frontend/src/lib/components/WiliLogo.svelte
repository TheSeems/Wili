<script lang="ts">
  interface Props {
    variant?: 'auto' | 'light' | 'dark' | 'colored';
    className?: string;
    opticalAdjust?: boolean; // legacy CSS-based adjust; prefer intrinsic SVG shift
  }

  let { variant = 'auto', className = '', opticalAdjust = false }: Props = $props();

  const adjustClass = opticalAdjust
    ? '-translate-x-[2px] sm:-translate-x-[3px] md:-translate-x-[4px]'
    : '';
  const shiftX = -2; // Intrinsic art shift to remove left-side padding inside viewBox
</script>


<!-- Auto variant: toggle between light/dark via media class -->
<div style="display:inline-block; line-height:0;">
<svg viewBox="20 0 300 108.838" class={className} aria-label="Wili logo">
    <path d="M100 59.994c-18.925 0-32.533-7.46-25.86-14.138C67.467 52.533 60.001 38.92 60.001 20c0 18.919-7.418 32.533-14.14 25.856C52.584 52.533 38.926 59.994 20 59.994c18.926 0 32.584 7.468 25.862 14.139 6.722-6.671 14.14 6.938 14.14 25.868 0-18.93 7.466-32.54 14.137-25.868-6.67-6.671 6.936-14.139 25.862-14.139" xmlns="http://www.w3.org/2000/svg" fill="#111"/>
    <path d="M120 38.3h19.641l11.365 30.143 16.43-33.848 16.677 33.972L195.85 38.3h19.518l-29.154 70.537-18.777-36.071-18.53 36.071zm123.039-10.314q-2.965 2.903-7.165 2.903-4.324 0-7.35-2.903t-3.027-7.103q0-4.077 3.026-6.98t7.35-2.903q4.2 0 7.166 2.903t2.964 6.98q0 4.2-2.964 7.103m-16.43 10.191h18.53v67.82h-18.53zm36.936-25.57h18.53v93.39h-18.53zm53.49 15.379q-2.965 2.903-7.165 2.903-4.324 0-7.35-2.903t-3.027-7.103q0-4.077 3.027-6.98T309.87 11q4.2 0 7.165 2.903t2.965 6.98q0 4.2-2.965 7.103m-16.43 10.191h18.53v67.82h-18.53z" fill="#111"/>
    </svg>
<span class="sr-only">Wili</span>
</div>

<style>
  :global(.dark .dark\:block) { display: block; }
  :global(.dark .dark\:hidden) { display: none; }
</style>


