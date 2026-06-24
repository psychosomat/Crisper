<script>
import logo from '../assets/logo.svg'
export let processing = false
</script>

<header class="flex flex-col items-center pt-12 pb-8 select-none">
  <img
    src={logo}
    alt="Crisper"
    class="w-14 h-auto mb-3"
    class:logo-active={processing}
  />
  <h1 class="text-xs font-mono tracking-[0.3em] uppercase text-text-secondary">Crisper</h1>
</header>

<style>
  @keyframes logo-spin-pulse {
    0% {
      transform: rotate(0deg) scale(1);
      opacity: 1;
    }
    25% {
      transform: rotate(90deg) scale(1.08);
      opacity: 0.85;
    }
    50% {
      transform: rotate(180deg) scale(1);
      opacity: 1;
    }
    75% {
      transform: rotate(270deg) scale(1.08);
      opacity: 0.85;
    }
    100% {
      transform: rotate(360deg) scale(1);
      opacity: 1;
    }
  }

  .logo-active {
    animation: logo-spin-pulse 3s linear infinite;
  }
</style>
