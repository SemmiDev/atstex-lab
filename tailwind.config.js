/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./web/templates/**/*.html", "./web/static/js/**/*.js"],
  theme: {
    extend: {
      colors: {
        bg: "var(--bg)",
        surface: "var(--surface)",
        surface2: "var(--surface2)",
        border: "var(--border)",
        text: "var(--text)",
        muted: "var(--muted)",
        accent: "var(--accent)",
        accent2: "var(--accent2)",
        accent3: "var(--accent3)",
        accent4: "var(--accent4)",
        error: "var(--error)",
        warn: "var(--warn)",
      },
      fontFamily: {
        ui: ["'DM Sans'", "sans-serif"],
        mono: ["'Fira Code'", "monospace"],
        disp: ["'DM Serif Display'", "serif"],
      }
    },
  },
  plugins: [],
}

