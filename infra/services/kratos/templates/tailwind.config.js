function withOpacity(variableName) {
  return ({ opacityValue }) => {
    if (opacityValue !== undefined) {
      return `rgba(var(${variableName}), ${opacityValue})`;
    }
    return `rgb(var(${variableName}))`;
  };
}

module.exports = {
  mode: "jit",
  content: ["./**/*.html"],
  darkMode: "media",
  theme: {
    fontFamily: {
      display: ["Poppins"],
      sans: ["Inter"],
      mono: ['"Overpass Mono"', '"Source Code Pro"'],
    },
    extend: {
      colors: {
        primary: "#F35167",
        secondary: "#7DD043",
      },
      // Token colours
      textColor: {
        strong: withOpacity("--text-strong"),
        medium: withOpacity("--text-medium"),
        weak: withOpacity("--text-weak"),
        disabled: withOpacity("--text-disabled"),
        primary: withOpacity("--text-primary"),
        error: withOpacity("--text-error"),
      },
      backgroundColor: {
        container: withOpacity("--bg-container"),
        "container-hover": withOpacity("--bg-container-hover"),
        strong: withOpacity("--bg-strong"),
        disabled: withOpacity("--bg-disabled"),
        primary: withOpacity("--bg-primary"),
        "container-primary": withOpacity("--bg-container-primary"),
        "container-primary-hover": withOpacity("--bg-container-primary-hover"),
        "container-primary-active": withOpacity(
          "--bg-container-primary-active"
        ),
        snackbar: withOpacity("--bg-snackbar"),
      },
      borderColor: {
        base: withOpacity("--border"),
        focus: withOpacity("--border-focus"),
        hover: withOpacity("--border-hover"),
        active: withOpacity("--border-active"),
        error: withOpacity("--border-error"),
      },
      ringColor: {
        base: withOpacity("--border"),
        focus: withOpacity("--border-focus"),
        hover: withOpacity("--border-hover"),
        active: withOpacity("--border-active"),
        error: withOpacity("--border-error"),
      },
      outlineColor: {
        base: withOpacity("--border"),
        focus: withOpacity("--border-focus"),
        hover: withOpacity("--border-hover"),
        active: withOpacity("--border-active"),
        error: withOpacity("--border-error"),
      },
      divideColor: {
        base: withOpacity("--border"),
        focus: withOpacity("--border-focus"),
        hover: withOpacity("--border-hover"),
        active: withOpacity("--border-active"),
        error: withOpacity("--border-error"),
      },
    },
  },
  plugins: [],
};
