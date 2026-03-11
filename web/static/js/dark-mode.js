// ── Dark Mode Toggle ──────────────────────────────────────────
// Reads/writes 'atstex_theme' in localStorage.
// Applies data-theme="dark" on <html> element.
(function () {
  const THEME_KEY = 'atstex_theme';

  // Apply saved theme immediately (before paint)
  const saved = localStorage.getItem(THEME_KEY);
  if (saved === 'dark') {
    document.documentElement.setAttribute('data-theme', 'dark');
  }

  document.addEventListener('DOMContentLoaded', () => {
    const btn = document.getElementById('dark-mode-toggle');
    if (!btn) return;

    const updateIcon = () => {
      const isDark = document.documentElement.getAttribute('data-theme') === 'dark';
      const icon = btn.querySelector('i');
      const label = btn.querySelector('.sidebar-nav-text');
      if (icon) {
        icon.className = isDark ? 'ph ph-sun text-2xl' : 'ph ph-moon text-2xl';
      }
      if (label) {
        label.textContent = isDark ? 'Mode Terang' : 'Mode Gelap';
      }
    };

    updateIcon();

    btn.addEventListener('click', (e) => {
      e.preventDefault();
      const isDark = document.documentElement.getAttribute('data-theme') === 'dark';
      if (isDark) {
        document.documentElement.removeAttribute('data-theme');
        localStorage.setItem(THEME_KEY, 'light');
      } else {
        document.documentElement.setAttribute('data-theme', 'dark');
        localStorage.setItem(THEME_KEY, 'dark');
      }
      updateIcon();
    });
  });
})();
