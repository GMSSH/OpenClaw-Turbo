<template>
  <div class="cyber-worker-page">
    <div class="coming-soon fade-in-up">
      <div class="icon-wrap">
        <svg viewBox="0 0 80 80" width="80" height="80" fill="none">
          <rect x="12" y="12" width="56" height="40" rx="8" stroke="currentColor" stroke-width="2.5"/>
          <circle cx="30" cy="30" r="4" fill="currentColor" opacity="0.6"/>
          <circle cx="50" cy="30" r="4" fill="currentColor" opacity="0.6"/>
          <path d="M28 40h24" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"/>
          <path d="M28 62h24M40 52v10" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"/>
          <path d="M6 28l-2-4M74 28l2-4" stroke="currentColor" stroke-width="2" stroke-linecap="round" opacity="0.4"/>
        </svg>
      </div>
      <h2 class="title">赛博员工</h2>
      <p class="subtitle">功能开发中，敬请期待</p>
      <div class="dots">
        <span class="dot"></span>
        <span class="dot"></span>
        <span class="dot"></span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.cyber-worker-page {
  width: 100%; height: 100%;
  display: flex; align-items: center; justify-content: center;
  overflow: hidden;
}

.coming-soon {
  display: flex; flex-direction: column; align-items: center; gap: 12px;
  text-align: center; padding: 40px;
}

.fade-in-up { animation: fadeInUp 0.4s ease; }
@keyframes fadeInUp { from { opacity: 0; transform: translateY(12px); } to { opacity: 1; transform: none; } }

.icon-wrap {
  color: var(--jm-accent-3);
  margin-bottom: 8px;
  animation: float 3s ease-in-out infinite;
}
@keyframes float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-8px); }
}

.title {
  font-size: 18px; font-weight: 600;
  color: var(--jm-accent-7); margin: 0;
}

.subtitle {
  font-size: 13px; color: var(--jm-accent-4); margin: 0;
}

.dots {
  display: flex; gap: 6px; margin-top: 8px;
}
.dot {
  width: 6px; height: 6px; border-radius: 50%;
  background: var(--jm-primary-1); opacity: 0.4;
  animation: pulse 1.4s ease-in-out infinite;
}
.dot:nth-child(2) { animation-delay: 0.2s; }
.dot:nth-child(3) { animation-delay: 0.4s; }
@keyframes pulse {
  0%, 100% { opacity: 0.2; transform: scale(0.8); }
  50% { opacity: 1; transform: scale(1.2); }
}
</style>
