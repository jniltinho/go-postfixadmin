/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./views/**/*.html"],
  theme: {
    extend: {
      colors: {
        'brand': {
          'primary': '#3B82F6',
          'secondary': '#60A5FA',
          'cta': '#F97316',
          'background': '#F8FAFC',
          'text': '#1E293B',
        }
      },
      fontFamily: {
        'sans': ['Fira Sans', 'sans-serif'],
        'mono': ['Fira Code', 'monospace'],
      },
      backdropBlur: {
        'glass': '12px',
      }
    },
  },
  plugins: [],
}
